package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"reflect"
	"regexp"
	"strings"
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
	config *mongoConfig
}

type mongoConfig struct {
	URI              string        `required:"true"`
	Database         string        `default:"randflix"`
	Collection       string        `default:"titles"`
	OperationTimeout time.Duration `default:"10s"`
	Server           string
}

func (c *mongoConfig) String() string {
	// get copy of instance so that we don't override the URI in the real config
	conf := *c

	// mask basic auth password (if exists)
	basicAuthPattern := regexp.MustCompile(`(:\/\/[^\/]+:)[^\/]+(@)`)
	conf.URI = basicAuthPattern.ReplaceAllString(conf.URI, `$1*****$2`)

	s, _ := json.MarshalIndent(conf, "", "\t")
	return string(s)
}

// NewMongoStore creates a new mongodb client, based on the config passed in
func (c *Config) NewMongoStore() (Storage, error) {

	var mConfig mongoConfig
	err := c.ProcessStorageConfig(&mConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to load mongo config: %s", err)
	}
	mc := &mConfig
	mc.parseServer()

	log.Printf("mongo config: %s\n", mc.String())

	client, err := mc.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %s", err)
	}
	log.Printf("(storage): initiated connection to mongodb: %s", mc.Server)

	db := client.Database(mc.Database)
	collection := db.Collection(mc.Collection)

	return &MongoStore{
		client: client,
		titles: collection,
		config: mc,
	}, nil
}

// parse server from URI (the 'server' will be what we use for logging)
func (mc *mongoConfig) parseServer() {
	// remove authanctication part (if exists)
	parts := strings.Split(mc.URI, "@")

	var noAuth string
	if len(parts) == 1 {
		noAuth = parts[0]
	} else {
		noAuth = parts[1]
	}

	// remove any query params
	mc.Server = strings.Split(noAuth, "?")[0]
}

func (mc *mongoConfig) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.OperationTimeout)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mc.URI))
	// remove URI after use (as it may include passwords)
	mc.URI = "REDACTED"
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %s", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping mongo db '%s': %s", mc.Server, err)
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

// RandomTitle picks a random title based on the filters passed in
func (m *MongoStore) RandomTitle(titleFilters ...title.Filter) (*title.Title, error) {

	filters, err := m.parseFilters(titleFilters...)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filters: %s", err)
	}

	sampleOne := mongo.Pipeline{
		{
			{Key: "$match", Value: filters},
		},
		{
			// $sample handles random sampling for us
			{Key: "$sample", Value: bson.D{
				{Key: "size", Value: 1},
			}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	cursor, err := m.titles.Aggregate(ctx, sampleOne)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return nil, fmt.Errorf("failed to sample: %s", err)
	}

	var rawDocuments []bson.Raw
	if err = cursor.All(ctx, &rawDocuments); err != nil {
		return nil, fmt.Errorf("failed to read cursor from sample aggregation: %s", err)
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

// AddTitle adds the title passed in
func (m *MongoStore) AddTitle(t *title.Title) (*title.Title, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	_, err := m.titles.InsertOne(ctx, t)

	return t, err
}

// UpdateTitle updates the title passed in
func (m *MongoStore) UpdateTitle(t *title.Title) (*title.Title, error) {
	filter := bson.M{"_id": t.ID}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	_, err := m.titles.ReplaceOne(ctx, filter, t)

	return t, err
}

// GetTitle gets a single title by id, if it doesn't exist, it returns nil
func (m *MongoStore) GetTitle(id string) (*title.Title, error) {
	filter := bson.M{"_id": id}

	ctx, cancel := context.WithTimeout(context.Background(), m.config.OperationTimeout)
	defer cancel()

	var title *title.Title
	err := m.titles.FindOne(ctx, filter).Decode(&title)

	// we don't want to return an error if the title was not found
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}

	return title, err
}

// ListTitles lists all elements in the mongo store, by page and pageSize
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

func (m *MongoStore) parseFilters(titleFilters ...title.Filter) ([]bson.E, error) {

	filters := []bson.E{}

	for _, tf := range titleFilters {
		f, err := m.parseFilter(tf)
		if err != nil {
			return nil, fmt.Errorf("failed to parse filter: %s", err)
		}

		filters = append(filters, f)
	}

	return filters, nil
}

func (m *MongoStore) parseFilter(tf title.Filter) (bson.E, error) {

	var filter bson.E

	switch tf.(type) {
	case title.OnServiceFilter:
		filter = m.onService(tf.(title.OnServiceFilter).Service)
	case title.IsGenreFilter:
		filter = m.isGenre(tf.(title.IsGenreFilter).Genres...)
	case title.ScoreBetweenFilter:
		f := tf.(title.ScoreBetweenFilter)
		filter = m.scoreBetween(f.Kind, f.Min, f.Max)
	default:
		return bson.E{}, fmt.Errorf("unsupported title filter type: %s", reflect.TypeOf(tf))
	}

	return filter, nil
}

func (m *MongoStore) onService(name string) bson.E {
	if name == "" {
		return m.emptyFilter()
	}

	serviceID := fmt.Sprintf("services.%s.id", name)

	return bson.E{Key: serviceID, Value: bson.D{
		{Key: "$exists", Value: true},
	}}
}

func (m *MongoStore) isGenre(names ...string) bson.E {

	if names == nil {
		return m.emptyFilter()
	}

	return bson.E{Key: "genres", Value: bson.D{
		{Key: "$all", Value: names},
	}}
}

func (m *MongoStore) scoreBetween(kind string, min int, max int) bson.E {

	if kind == "" {
		return m.emptyFilter()
	}
	if max == 0 {
		max = math.MaxInt64
	}

	score := fmt.Sprintf("scores.%s", kind)

	return bson.E{Key: score, Value: bson.D{
		{Key: "$gte", Value: min},
		{Key: "$lte", Value: max},
	}}
}

func (m *MongoStore) emptyFilter() bson.E {
	return bson.E{}
}
