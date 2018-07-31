package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// Config holds configuration of feeder.
type Config struct {
	// NATS config
	NATSAddress string
	Topic       string

	// Elastisearch config
	Elastics []string

	// HTTP config
	HTTPPort                int
	GracefulShutdownTimeout int
}

// LoadConfig loads and unmarshal config from file passed as argument to func.
func LoadConfig(pathToConfig string) (*Config, error) {

	bytes, err := ioutil.ReadFile(pathToConfig)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := toml.Unmarshal(bytes, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
