package config

var (
	defaultConfig = &Config{
		Storage: &StorageConfig{
			Kind: "MemStore",
		},
	}
)

// Config describes the api configuration
type Config struct {
	Storage *StorageConfig `json:"storage"`
}

// StorageConfig describes the generic storage configuration
type StorageConfig struct {
	Kind   string      `json:"kind"`
	Config interface{} `json:"config"`
}

func GetConfig() *Config {
	return defaultConfig
}
