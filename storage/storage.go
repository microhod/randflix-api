package storage

import (
	"fmt"
	"reflect"

	"github.com/microhod/randflix-api/config"
	"github.com/microhod/randflix-api/model/title"
)

// Storage provides storage functions for the api
type Storage interface {
	// RandomTitle gets a random title from storage
	RandomTitle(filters ...title.Filter) (*title.Title, error)
	// AddTitle adds a title to storage
	AddTitle(t *title.Title) (*title.Title, error)
	// UpdateTitle replaces a title in storage
	UpdateTitle(t *title.Title) (*title.Title, error)
	// GetTitle retrieces a title from storage by id
	GetTitle(id string) (*title.Title, error)
}

// Config encapsulates config.StorageConfig, so that we can define methods on it in this package
type Config struct {
	config.StorageConfig
}

// CreateStorage is a factory method to create storage of any type, based on the config
func CreateStorage(config *config.Config) (Storage, error) {

	// Convert to local Config struct
	c := &Config{*config.Storage}

	// Constructors should have the form 'New<StorageKind>'
	name := fmt.Sprintf("New%s", c.Kind)
	f := reflect.ValueOf(c).MethodByName(name)

	if !f.IsValid() {
		return nil, fmt.Errorf("Storage kind not supported: %s", c.Kind)
	}

	v := f.Call(nil)

	if len(v) != 2 {
		return nil, fmt.Errorf("Expected 2 return values from method %s, got: %d", name, len(v))
	}

	s, ok := v[0].Interface().(Storage)
	if !ok {
		return nil, fmt.Errorf("Expected 1st return value to be *Storage from method %s, got: %s", name, reflect.TypeOf(v[0].Interface()))
	}

	err, ok := v[1].Interface().(error)
	if !(ok || v[1].IsNil()) {
		return nil, fmt.Errorf("Expected 2nd return value to be error from method %s, got: %s", name, reflect.TypeOf(v[0].Interface()))
	}

	return s, err
}
