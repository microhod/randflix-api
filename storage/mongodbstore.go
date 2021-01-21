package storage

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gobeam/mongo-go-pagination"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/microhod/randflix-api/model/title"
)

const (
	envPrefix = "MONGO"
)

// MongoStore is storage using mongodb
type MongoStore struct {
	client *mongo.Client
	titles *mongo.Collection
	config mongoConfig
}

type mongoConfig struct {
	URI              string        `required:"true"`
	Database         string        `default:"randflix"`
	Collection       string        `default:"titles"`
	OperationTimeout time.Duration `default:"10s"`
}

// NewMongoStore creates a new mongodb client, based on the config passed in
func (c *Config) NewMongoStore() (Storage, error) {

	var mc mongoConfig
	err := c.ProcessStorageConfig(&mc)
	if err != nil {
		return nil, fmt.Errorf("failed to load mongo config: %s", err)
	}

	client, err := mc.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %s", err)
	}
	log.Printf("(storage): initiated connection to mongodb: %s", mc.URI)

	db := client.Database(mc.Database)
	collection := db.Collection(mc.Collection)

	return &MongoStore{
		client: client,
		titles: collection,
		config: mc,
	}, nil
}

func (mc mongoConfig) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.OperationTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mc.URI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %s", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongo db '%s': %s", mc.URI, err)
	}

	return client, nil
}

// Disconnect disconnects from the mongodb
func (m *MongoStore) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	if err := m.client.Disconnect(ctx); err != nil {
		panic(err)
	}
}

func (m *MongoStore) RandomTitle(filters ...title.Filter) (*title.Title, error) {

	// TODO: add filter support

	// $sample handles random sampling for us
	sampleOne := []bson.D{bson.D{{"$sample", bson.D{{"size", 1}}}}}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	cursor, err := m.titles.Aggregate(ctx, sampleOne)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments){
		return nil, fmt.Errorf("failed to sample: %s", nil)
	}

	var rawDocuments []bson.Raw
	if err = cursor.All(ctx, &rawDocuments); err != nil {
		return nil, fmt.Errorf("failed to read cursor from sample aggregation")
	}

	if len(rawDocuments) == 0 {
		return nil, nil
	}

	var title *title.Title
	if bson.Unmarshal(rawDocuments[0], &title); err != nil {
		return nil, fmt.Errorf("failed to unmarshall bson to title: %s", err)
	}

	return title, nil
}

func (m *MongoStore) AddTitle(t *title.Title) (*title.Title, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	_, err := m.titles.InsertOne(ctx, t)
	
	return t, err
}

func (m *MongoStore) UpdateTitle(t *title.Title) (*title.Title, error) {
	filter := bson.M{"_id": t.ID}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	_, err := m.titles.ReplaceOne(ctx, filter, t)

	return t, err
}

func (m *MongoStore) GetTitle(id string) (*title.Title, error) {
	filter := bson.M{"_id": id}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	var title *title.Title
	err := m.titles.FindOne(ctx, filter).Decode(&title)
	
	// we don't want to return an error if the title was not found
	if errors.Is(err, mongo.ErrNoDocuments) {
		err = nil
	}

	return title, err
}

func (m *MongoStore) ListTitles(pageSize int, page int) ([]*title.Title, error) {

	// empty filter
	filter := bson.M{}

	data, err := mongopagination.New(m.titles).Filter(filter).Limit(int64(pageSize)).Page(int64(page)).Find()
	if err != nil {
		return nil, fmt.Errorf("failed to get page %d of titles with a page size of %d: %s", page, pageSize, err)
	}

	titles := []*title.Title{}
	for _, raw := range data.Data {
		var title *title.Title
		
		if bson.Unmarshal(raw, &title); err != nil {
			return nil, fmt.Errorf("failed to unmarshall bson to title: %s", err)
		}

		titles = append(titles, title)
	}

	return titles, nil
}
