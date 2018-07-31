package main

import (
	"flag"

	"github.com/mateuszdyminski/auto/indexer/pkg/signals"
	"github.com/mateuszdyminski/auto/server/pkg/config"
	"github.com/mateuszdyminski/auto/server/pkg/search"
	"github.com/mateuszdyminski/auto/server/pkg/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var configPath string
var debug bool

func init() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}

	flag.StringVar(&configPath, "config", "../../config/conf.toml", "config path")
	flag.BoolVar(&debug, "debug", false, "sets log level to debug")
}

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatal().Msgf("can't load config file. err: %s", err)
	}

	ctx := signals.SetupSignalContext()

	srv, err := search.NewFlightService(cfg, ctx)
	if err != nil {
		log.Fatal().Msgf("can't create indexer. err: %s", err)
	}

	server.ListenAndServe(srv, cfg, ctx)
}
