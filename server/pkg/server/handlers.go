package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/mateuszdyminski/auto/ingress/model"
	"github.com/mateuszdyminski/auto/server/pkg/version"
	"github.com/mateuszdyminski/auto/server/pkg/ws"
	"github.com/rs/zerolog/log"
)

func (s *Server) search(w http.ResponseWriter, req *http.Request) {
	from := req.URL.Query().Get("from") + "+01:00"
	to := req.URL.Query().Get("to") + "+01:00"
	query := req.URL.Query().Get("query")

	size, err := strconv.Atoi(req.URL.Query().Get("l"))
	if err != nil {
		size = 100
	}

	skip, err := strconv.Atoi(req.URL.Query().Get("s"))
	if err != nil {
		skip = 0
	}

	fromTime := time.Time{}
	if from != "" {
		if fromTime, err = time.Parse(time.RFC3339, from); err != nil {
			log.Warn().Msgf("Can't parse 'from' time: %s", from)
		}
	}

	toTime := time.Time{}
	if to != "" {
		if toTime, err = time.Parse(time.RFC3339, to); err != nil {
			log.Warn().Msgf("Can't parse 'to' time: %s", to)
		}
	}

	logs, err := s.service.Search(query, fromTime, toTime, size, skip)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	json, err := json.Marshal(logs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(json)
}

// serverWs handles websocket requests from the peer.
func (s *Server) serveWs(w http.ResponseWriter, req *http.Request) {
	log.Info().Msgf("Registering client for WS")

	ws.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	upg, err := ws.Upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Error().Msgf("Error %+v", err)
		return
	}

	c := &ws.Connection{Send: make(chan *model.FlightCrash, 256), Ws: upg}
	s.service.Ws.Register <- c
	go c.WritePump()
}

func (s *Server) version(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/version" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp := map[string]string{
		"version":       version.APP_VERSION,
		"buildTime":     version.BUILD_TIME,
		"gitVersion":    version.GIT_VERSION,
		"gitCommitHash": version.LAST_COMMIT_HASH,
		"gitCommitUser": version.LAST_COMMIT_USER,
		"gitCommitTime": version.LAST_COMMIT_TIME,
	}

	d, err := json.Marshal(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(d)
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&ready) == 1 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
		return
	}
	w.WriteHeader(http.StatusServiceUnavailable)
}
