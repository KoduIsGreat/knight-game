package nw

import (
	"crypto/tls"
	"log"
	"time"

	quic "github.com/quic-go/quic-go"
)

type ServerOption[T any] func(*Server[T])

func WithLogger[T any](log *log.Logger) ServerOption[T] {
	return func(s *Server[T]) {
		s.log = log
	}
}
func WithTickRate[T any](rate time.Duration) ServerOption[T] {
	return func(s *Server[T]) {
		s.tickRate = rate
	}
}
func WithStateManager[T any](sm StateManager[T]) ServerOption[T] {
	return func(s *Server[T]) {
		s.state = sm
	}
}
func WithTLSConfig[T any](tlsConfig *tls.Config) ServerOption[T] {
	return func(s *Server[T]) {
		s.tlsConfig = tlsConfig
	}
}

func WithQuicConfig[T any](quicConfig *quic.Config) ServerOption[T] {
	return func(s *Server[T]) {
		s.quicConfig = quicConfig
	}
}
