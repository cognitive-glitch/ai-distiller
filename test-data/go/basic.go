package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Config holds application configuration
type Config struct {
	Port    int    `json:"port"`
	Host    string `json:"host"`
	Timeout time.Duration
}

// Logger interface for structured logging
type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, err error, fields ...Field)
	Debug(msg string, fields ...Field)
}

// Field represents a log field
type Field struct {
	Key   string
	Value interface{}
}

// Server represents our HTTP server
type Server struct {
	config     *Config
	logger     Logger
	httpServer *http.Server
	mu         sync.RWMutex
	handlers   map[string]http.HandlerFunc
}

// NewServer creates a new server instance
func NewServer(config *Config, logger Logger) *Server {
	return &Server{
		config:   config,
		logger:   logger,
		handlers: make(map[string]http.HandlerFunc),
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      s,
		ReadTimeout:  s.config.Timeout,
		WriteTimeout: s.config.Timeout,
	}

	// Start server in goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Server failed to start", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	return s.Shutdown()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}

// ServeHTTP implements http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	handler, ok := s.handlers[r.URL.Path]
	s.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	handler(w, r)
}

// RegisterHandler registers a new route handler
func (s *Server) RegisterHandler(path string, handler http.HandlerFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[path] = handler
}

// Response represents an API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// handleHealth is a health check endpoint
func handleHealth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// processRequest processes incoming requests with timeout
func processRequest(ctx context.Context, data []byte) ([]byte, error) {
	resultCh := make(chan []byte, 1)
	errCh := make(chan error, 1)

	go func() {
		// Simulate processing
		time.Sleep(100 * time.Millisecond)
		resultCh <- append([]byte("processed: "), data...)
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Generic constraint example
type Number interface {
	int | int64 | float64
}

// Sum calculates sum of numbers (generic function)
func Sum[T Number](nums []T) T {
	var sum T
	for _, n := range nums {
		sum += n
	}
	return sum
}

// Option represents an optional value
type Option[T any] struct {
	value *T
}

// Some creates an Option with a value
func Some[T any](value T) Option[T] {
	return Option[T]{value: &value}
}

// None creates an empty Option
func None[T any]() Option[T] {
	return Option[T]{value: nil}
}

// IsNone checks if Option is empty
func (o Option[T]) IsNone() bool {
	return o.value == nil
}

// Unwrap returns the value or panics
func (o Option[T]) Unwrap() T {
	if o.value == nil {
		panic("called Unwrap on None")
	}
	return *o.value
}

// Constants
const (
	DefaultPort    = 8080
	DefaultTimeout = 30 * time.Second
	MaxRetries     = 3
)

// Variables
var (
	ErrNotFound     = fmt.Errorf("not found")
	ErrUnauthorized = fmt.Errorf("unauthorized")
	defaultLogger   Logger
)