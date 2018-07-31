package search

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/mateuszdyminski/auto/ingress/model"
	"github.com/mateuszdyminski/auto/server/pkg/config"
	"github.com/mateuszdyminski/auto/server/pkg/ws"
	nats "github.com/nats-io/go-nats"
	"github.com/olivere/elastic"
	"github.com/rs/zerolog/log"
)

type FlightService struct {
	cfg *config.Config
	esc *elastic.Client
	nc  *nats.Conn
	Ws  *ws.Hub
}

func NewFlightService(cfg *config.Config, ctx context.Context) (*FlightService, error) {
	// connet to NATS
	nc, err := nats.Connect(cfg.NATSAddress)
	if err != nil {
		return nil, err
	}

	// connect to Elastic
	esc, err := elastic.NewClient(elastic.SetURL(cfg.Elastics...), elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}

	// turn on the WebSockets server
	ws := ws.NewHub()
	go ws.Run()

	fs := &FlightService{cfg: cfg, esc: esc, Ws: ws, nc: nc}
	go func() {
		if err := fs.Run(ctx); err != nil {
			log.Error().Msgf("error during collecting flight crashes")
		}
	}()

	return fs, nil
}

func (s *FlightService) Run(ctx context.Context) error {
	if s.nc == nil {
		return errors.New("nats client connected! exiting")
	}

	log.Info().Msgf("Starting subscribing to topic %s!", s.cfg.Topic)

	var wg sync.WaitGroup
	wg.Add(1)

	sub, err := s.nc.Subscribe(s.cfg.Topic, func(m *nats.Msg) {
		l := new(model.FlightCrash)

		if err := json.Unmarshal(m.Data, l); err != nil {
			log.Error().Msgf("can't unmarshall flight crash. err: %v", err)
			return
		}

		log.Info().Msgf("got flight crash: %v", l)

		// send flight to all WS clients
		s.Ws.Broadcast <- l
	})

	if err != nil {
		log.Error().Msgf("Error during subscription to NATS topic: %s! err: %v", s.cfg.Topic, err)
	} else {
		log.Info().Msgf("Subscribed!")
	}

	go func() {
		<-ctx.Done()
		log.Info().Msgf("Got cancel signal. Exiting Flight Service!")
		sub.Unsubscribe()
		wg.Done()
	}()

	wg.Wait()
	return nil
}

func (s *FlightService) Search(query string, from, to time.Time, size, skip int) (*Response, error) {
	// Create and execute finder
	res, err := NewFinder().Query(query).From(from).To(to).Size(size).Skip(skip).Sort("-time").Find(s.esc)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// Response holds information about queried data and total number of hits.
type Response struct {
	Data  interface{} `json:"data,omitempty"`
	Total int64       `json:"total,omitempty"`
}
