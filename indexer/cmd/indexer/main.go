package main

import (
	"flag"
	"sync"

	"github.com/mateuszdyminski/auto/indexer/pkg/config"
	"github.com/mateuszdyminski/auto/indexer/pkg/indexer"
	"github.com/mateuszdyminski/auto/indexer/pkg/server"
	"github.com/mateuszdyminski/auto/indexer/pkg/signals"
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

	idx, err := indexer.NewIndexer(cfg)
	if err != nil {
		log.Fatal().Msgf("can't create indexer. err: %s", err)
	}

	ctx := signals.SetupSignalContext()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		server.ListenAndServe(cfg, ctx)
		wg.Done()
	}()

	go func() {
		idx.Start(ctx)
		wg.Done()
	}()

	wg.Wait()
}
