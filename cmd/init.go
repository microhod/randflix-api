package cmd

import (
	"fmt"

	"github.com/microhod/randflix-api/api"
	"github.com/microhod/randflix-api/config"
	"github.com/microhod/randflix-api/storage"
)

// InitAPI sets up all the basic objects for the API
func InitAPI() (*config.Config, storage.Storage, api.API, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, nil, api.API{}, fmt.Errorf("failed to get config: %s", err)
	}

	store, err := storage.CreateStorage(cfg)
	if err != nil {
		return nil, nil, api.API{}, fmt.Errorf("failed to create storage: %s", err)
	}

	api := api.API{Storage: store}

	return cfg, store, api, nil
}
