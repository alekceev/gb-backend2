package server

import (
	"context"
	"net/http"
	"time"

	"gb-backend2/internal/app/starter"
	"gb-backend2/internal/app/store"
)

var _ starter.APIServer = &Server{}

type Server struct {
	srv   http.Server
	store *store.Store
}

func NewServer(addr string, h http.Handler, store *store.Store) *Server {
	s := &Server{
		store: store,
	}

	s.srv = http.Server{
		Addr:              addr,
		Handler:           h,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	return s
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	s.srv.Shutdown(ctx)
	cancel()
}

func (s *Server) Start() {
	go s.srv.ListenAndServe()
}
