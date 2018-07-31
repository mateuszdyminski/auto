package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/mateuszdyminski/auto/indexer/pkg/server"
	"github.com/mateuszdyminski/auto/server/pkg/config"
	"github.com/mateuszdyminski/auto/server/pkg/search"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	healthy int32 = 1
	ready   int32 = 1
)

type Server struct {
	mux     *http.ServeMux
	service *search.FlightService
}

func NewServer(service *search.FlightService, options ...func(*Server)) *Server {
	s := &Server{service: service, mux: http.NewServeMux()}

	for _, f := range options {
		f(s)
	}

	// register flights handlers
	s.mux.HandleFunc("/api/flights", s.search)
	s.mux.HandleFunc("/wsapi/ws", s.serveWs)

	// register generic handlers
	s.mux.HandleFunc("/healthz", s.healthz)
	s.mux.HandleFunc("/readyz", s.readyz)
	s.mux.HandleFunc("/version", s.version)
	s.mux.Handle("/metrics", promhttp.Handler())

	// Register pprof handlers
	s.mux.HandleFunc("/debug/pprof/", pprof.Index)
	s.mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	s.mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	s.mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	s.mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", runtime.Version())

	s.mux.ServeHTTP(w, r)
}

func ListenAndServe(service *search.FlightService, cfg *config.Config, cancelCtx context.Context) {
	inst := server.NewInstrument()
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      inst.Wrap(NewServer(service)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  15 * time.Second,
	}

	// run server in background
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server crashed")
		}
	}()

	// wait for SIGTERM or SIGINT
	<-cancelCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.GracefulShutdownTimeout)*time.Second)
	defer cancel()

	// all calls to /healthz and /readyz will fail from now on
	atomic.StoreInt32(&healthy, 0)
	atomic.StoreInt32(&ready, 0)

	time.Sleep(3 * time.Second)

	log.Info().Msgf("Shutting down HTTP server with timeout: %v", time.Duration(cfg.GracefulShutdownTimeout)*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("HTTP server graceful shutdown failed")
	} else {
		log.Info().Msg("HTTP server stopped")
	}
}
