//go:build ignore

// Test Pattern 2: Aliased Imports
// Tests imports with explicit aliases and their usage

package main

import (
	f "fmt"
	"os"
	stdlog "log"
	req "net/http"
	"context"
	j "encoding/json"
	_ "net/http/pprof" // Blank import for side effects
	"database/sql"
	"sync"
)

// Not using: os, context, database/sql, sync

type Server struct {
	logger *stdlog.Logger
}

func NewServer() *Server {
	// Using aliased log package
	logger := stdlog.New(os.Stdout, "[SERVER] ", stdlog.LstdFlags)
	return &Server{logger: logger}
}

func (s *Server) Start() {
	// Using aliased fmt
	f.Println("Starting server with aliases")

	// Using aliased http as req
	req.HandleFunc("/health", func(w req.ResponseWriter, r *req.Request) {
		// Using aliased json as j
		response := map[string]string{
			"status": "healthy",
		}

		w.Header().Set("Content-Type", "application/json")
		j.NewEncoder(w).Encode(response)
	})

	// Using the logger
	s.logger.Println("Server configured")

	// The pprof import is a blank import and should always be kept
	// It registers the pprof handlers as a side effect

	// Start server
	req.ListenAndServe(":8081", nil)
}