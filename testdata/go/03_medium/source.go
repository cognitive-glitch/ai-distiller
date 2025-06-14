// server.go
package server

import "fmt"

// Config holds server configuration.
type Config struct {
	Port int
	Host string
}

// Validate checks the configuration.
func (c *Config) Validate() error {
	if c.Port <= 0 {
		return &ConfigError{Field: "Port", Reason: "must be positive"}
	}
	return nil
}

// ConfigError is a custom error for configuration issues.
type ConfigError struct {
	Field  string
	Reason string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error on field '%s': %s", e.Field, e.Reason)
}

// Server represents a network server.
// It embeds Config, promoting its fields.
type Server struct {
	Config
	listener string
}

// Start validates config and starts the server.
func (s *Server) Start() error {
	// The AI should understand that s.Validate() is a promoted method from Config.
	if err := s.Validate(); err != nil {
		return err // This could be a *ConfigError
	}
	// ... start listening on s.Host and s.Port
	fmt.Printf("Starting server on %s:%d\n", s.Host, s.Port)
	return nil
}