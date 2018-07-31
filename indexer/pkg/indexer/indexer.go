package indexer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"sync"

	"github.com/mateuszdyminski/auto/indexer/pkg/config"
	"github.com/mateuszdyminski/auto/ingress/model"
	nats "github.com/nats-io/go-nats"
	"github.com/olivere/elastic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
	"googlemaps.github.io/maps"
)

type Indexer struct {
	nc               *nats.Conn
	conf             *config.Config
	ctx              *context.Context
	googleAPIcounter *prometheus.CounterVec
	mapClient        *maps.Client
}

func NewIndexer(conf *config.Config) (*Indexer, error) {
	nc, err := nats.Connect(conf.NATSAddress)
	if err != nil {
		return nil, err
	}

	proxyUrl, _ := url.Parse("http://10.144.1.10:8080")
	myClient := &http.Client{Transport: &http.Transport{
		Proxy:           http.ProxyURL(proxyUrl),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
	mapClient, err := maps.NewClient(maps.WithAPIKey(conf.APIKey), maps.WithHTTPClient(myClient))
	if err != nil {
		return nil, err
	}

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "requests_google_api_total",
			Help:      "The total number of Google API requests.",
		},
		[]string{"status"},
	)

	prometheus.MustRegister(counter)

	return &Indexer{nc: nc, conf: conf, mapClient: mapClient, googleAPIcounter: counter}, nil
}

func (i *Indexer) Start(cancelCtx context.Context) {
	log.Info().Msg("Start indexing flights...")
	i.indexFlights(i.streamFlights(cancelCtx))
}

func (i *Indexer) indexFlights(flights chan model.FlightCrash) {
	// connect to the cluster
	client, err := elastic.NewClient(elastic.SetURL(i.conf.Elastics...), elastic.SetSniff(false))
	if err != nil {
		log.Fatal().Msgf("Can't create elastic client. Err: %v", err)
	}

	exists, err := client.IndexExists("flights").Do(context.Background())
	if err != nil {
		log.Fatal().Msgf("Can't check if index exists. Err: %v", err)
	}

	if !exists {
		log.Info().Msg("Creating index 'flights'")
		// Create an index if not exists
		_, err = client.
			CreateIndex("flights").
			Do(context.Background())
		if err != nil {
			log.Fatal().Msgf("Can't create index. Err: %v", err)
		}
	}

	var enqued int
	bulkRequest := client.Bulk()
	for flight := range flights {
		log.Info().Msgf("Got flight crash: %+v", flight)

		// get coordinates from google maps API
		coordinates, err := i.coordinates(flight)
		if err == nil && coordinates != nil {
			flight.LocationGPS = coordinates
			data, err := json.Marshal(flight)
			if err != nil {
				log.Error().Msgf("can't marshal flight with coordinates. err: %v", err)
			} else {
				i.nc.Publish(i.conf.OutTopic, data)
			}
		} else {
			log.Error().Msgf("can't find gps coordinates for location: %s. err %+v", flight.Location, err)
		}

		if enqued > 0 && enqued%i.conf.BulkSize == 0 {
			if _, err := bulkRequest.Do(context.Background()); err != nil {
				log.Fatal().Msgf("Can't execute bulk. Err: %v", err)
			}

			log.Info().Msgf("Bulk with %v flights indexed! Total indexed flights: %v", i.conf.BulkSize, enqued)

			bulkRequest = client.Bulk()
		}

		bulkRequest.Add(
			elastic.NewBulkIndexRequest().
				Index("flights").
				Type("flight").
				Id(uuid.Must(uuid.NewV4()).String()).
				Doc(flight))

		enqued++
	}

	if bulkRequest.NumberOfActions() > 0 {
		if _, err := bulkRequest.Do(context.Background()); err != nil {
			log.Fatal().Msgf("Can't execute bulk. Err: %v", err)
		}
	}
}

func (i *Indexer) coordinates(flight model.FlightCrash) (*model.Location, error) {

	r := &maps.GeocodingRequest{
		Address: flight.Location,
	}

	geo, err := i.mapClient.Geocode(context.Background(), r)
	if err != nil {
		i.googleAPIcounter.WithLabelValues("500").Inc()
		return nil, err
	}

	i.googleAPIcounter.WithLabelValues("200").Inc()

	if len(geo) > 0 {
		return &model.Location{
			Latitude:  geo[0].Geometry.Location.Lat,
			Longitude: geo[0].Geometry.Location.Lng,
		}, nil
	}

	return nil, nil
}

func (i *Indexer) streamFlights(ctx context.Context) chan model.FlightCrash {
	out := make(chan model.FlightCrash)

	var mu sync.Mutex
	go func() {
		sub, err := i.nc.QueueSubscribe(i.conf.Topic, i.conf.QueueGroup, func(m *nats.Msg) {
			var flight model.FlightCrash
			if err := json.Unmarshal(m.Data, &flight); err != nil {
				log.Error().Msgf("Can't unmarshal data from queue! Err: %v", err)
				return
			}

			mu.Lock()
			out <- flight
			mu.Unlock()
		})

		if err != nil {
			log.Fatal().Msgf("Can't subscribe to topic: %s, queue group: %s, err: %s", i.conf.Topic, i.conf.QueueGroup, err)
		}

		go func() {
			<-ctx.Done()
			log.Info().Msgf("Work cancelled!")
			sub.Unsubscribe()
			log.Info().Msgf("Unsubscribed from %s topic!", i.conf.Topic)
			mu.Lock()
			close(out)
			log.Info().Msgf("Channel with flights closed!")
			mu.Unlock()
		}()
	}()

	return out
}
