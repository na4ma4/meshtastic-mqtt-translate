package health

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/na4ma4/meshtastic-mqtt-translate/internal/relay"
)

type WebServer struct {
	Logger *slog.Logger
	Port   int
	Relay  *relay.Relay
	srv    *http.Server
}

func NewServer(port int, logger *slog.Logger, relay *relay.Relay) *WebServer {
	return &WebServer{
		Logger: logger,
		Port:   port,
		Relay:  relay,
	}
}

func (s *WebServer) Start() <-chan error {
	if s.srv != nil {
		errChan := make(chan error, 1)
		errChan <- errors.New("server already started")
		return errChan
	}
	s.srv = &http.Server{
		Addr:         ":" + strconv.Itoa(s.Port),
		Handler:      s,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		IdleTimeout:  time.Second,
	}
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)
		s.Logger.Info("Starting health server", slog.String("bind", s.srv.Addr))
		errChan <- s.srv.ListenAndServe()
	}()
	return errChan
}

func (s *WebServer) Stop(ctx context.Context) error {
	if s.srv == nil {
		return nil
	}

	s.Logger.InfoContext(ctx, "Stopping health server", slog.String("bind", s.srv.Addr))
	err := s.srv.Shutdown(ctx)
	s.srv = nil
	return err
}

func (s *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Logger.Debug("Health check", slog.String("remote_addr", r.RemoteAddr))

	w.Header().Set("Content-Type", "application/json")
	status := s.Relay.GetStatus()
	if !status.Status {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
