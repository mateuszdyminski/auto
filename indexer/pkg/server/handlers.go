package server

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/mateuszdyminski/auto/indexer/pkg/version"
)

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
