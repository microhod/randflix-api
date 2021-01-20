package config

import (
	"encoding/json"

	"github.com/kelseyhightower/envconfig"
)

const (
	appName = "RANDFLIXAPI"
)

// Config describes the api configuration
type Config struct {
	StorageKind   string            `default:"MemStore"`
	StorageConfig map[string]string `default:""`
}

func (c *Config) String() string {
	s, _ := json.MarshalIndent(c, "", "\t")
	return string(s)
}

// GetConfig returns the current config from environment variables
func GetConfig() (*Config, error) {
	var c Config
	err := envconfig.Process(appName, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
