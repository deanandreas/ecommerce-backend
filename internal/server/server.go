package server

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Closer interface {
	Close()
}

type Server struct {
	port int
	http *http.Server
	db   Closer
	*Handlers
}

func NewAPI() (*Server, error) {
	port, dbURL, err := GetENV()
	if err != nil {
		return nil, err
	}
	pool, handlers, err := Handler(dbURL)
	if err != nil {
		return nil, err
	}

	app := &Server{
		port:     port,
		db:       pool,
		Handlers: handlers,
	}

	app.http = &http.Server{
		Addr:         fmt.Sprintf(":%d", app.port),
		Handler:      http.TimeoutHandler(app.RegisterRoutes(), 60*time.Second, "server time out"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 40 * time.Second,
		IdleTimeout:  2 * time.Minute,
	}

	return app, nil
}

func (s *Server) Start() error {
	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	err := s.http.Shutdown(ctx)
	s.db.Close()
	return err
}
