package nw

import (
	"crypto/tls"
	"log"
	"time"
)

type ServerOption func(*Server)

func WithLogger(log *log.Logger) ServerOption {
	return func(s *Server) {
		s.log = log
	}
}
func WithTickRate(rate time.Duration) ServerOption {
	return func(s *Server) {
		s.tickRate = rate
	}
}
func WithStateManager(sm StateManager) ServerOption {
	return func(s *Server) {
		s.state = sm
	}
}
func WithTLSConfig(tlsConfig *tls.Config) ServerOption {
	return func(s *Server) {
		s.tlsConfig = tlsConfig
	}
}
