package config

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

const (
	// AppName is the name of the app to be used as prefix of environment variables
	AppName = "RANDFLIXAPI"
)

// Config describes the api configuration
type Config struct {
	StorageKind string `default:"MemStore"`
}

func (c *Config) String() string {
	s, _ := json.MarshalIndent(c, "", "\t")
	return string(s)
}

// GetConfig returns the current config from environment variables
func GetConfig() (*Config, error) {
	var c Config
	err := envconfig.Process(AppName, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
