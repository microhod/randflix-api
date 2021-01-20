package storage

import (
	"context"
	"fmt"
	"log"
	"time"

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
}

type mongoConfig struct {
	URI            string        `required:"true"`
	ConnectTimeout time.Duration `default:"10s"`
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

	return &MongoStore{client: client}, nil
}

func (mc mongoConfig) connect() (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mc.ConnectTimeout)
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
	if err := m.client.Disconnect(context.Background()); err != nil {
		panic(err)
	}
}

func (m *MongoStore) RandomTitle(filters ...title.Filter) (*title.Title, error) {
	return nil, nil
}

func (m *MongoStore) AddTitle(t *title.Title) (*title.Title, error) {
	return nil, nil
}

func (m *MongoStore) UpdateTitle(t *title.Title) (*title.Title, error) {
	return nil, nil
}

func (m *MongoStore) GetTitle(id string) (*title.Title, error) {
	return nil, nil
}

func (m *MongoStore) ListTitles(pageSize int, page int) ([]*title.Title, error) {
	return nil, nil
}
