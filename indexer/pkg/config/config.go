package config

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

// Config holds configuration of feeder.
type Config struct {
	// NATS config
	NATSAddress string
	Topic       string
	OutTopic    string
	QueueGroup  string

	// Elastisearch config
	Elastics []string
	BulkSize int

	// HTTP config
	HTTPPort                int
	GracefulShutdownTimeout int

	// Google maps api
	APIKey string
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

	conf.APIKey = os.Getenv("GOOGLE_MAPS_API_KEY")

	return &conf, nil

}
