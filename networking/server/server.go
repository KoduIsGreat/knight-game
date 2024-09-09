package server

import (
	// "context"
	"fmt"
	"sync"
	// "hash/fnv"
	"log/slog"
	"net"
	"os"
)

type Server struct {
	Type     string
	Addr     string
	shutdown chan struct{}
	handleFn func(conn *net.UDPConn, addr *net.UDPAddr, buf []byte)
	log      *slog.Logger
	clients  map[string]*net.UDPAddr
	mu       sync.RWMutex
}

const (
	UDP = "udp"
)

type HandleFunc func(conn *net.UDPConn, addr *net.UDPAddr, buf []byte)

func EchoHandler(s *Server) {
	s.handleFn = func(conn *net.UDPConn, addr *net.UDPAddr, buf []byte) {
		fmt.Println("buf", buf)
		if len(buf) == 0 {
			return
		}
		s.mu.Lock()
		s.clients[addr.String()] = addr
		for _, clientAddr := range s.clients {
			_, err := conn.WriteToUDP([]byte("hello from server"), clientAddr)
			if err != nil {
				s.log.Error("Error writing to UDP client %s: %v", clientAddr.String(), err)
			}
		}
		s.mu.Unlock()

		// s.log.Info("Received from client:", string(buf[0:]), slog.Any("id", id), slog.Any("addr", addr.String()))
		_, err := conn.WriteToUDP([]byte("Hello UDP Client"), addr)
		if err != nil {
			s.log.Error("Error writing to UDP connection: %v", err)
		}
	}
}

func (s *Server) Listen() error {
	addr, err := net.ResolveUDPAddr(UDP, s.Addr)
	if err != nil {
		s.log.Error("Error resolving UDP address: %v", err)
		return err
	}

	conn, err := net.ListenUDP(s.Type, addr)
	if err != nil {
		s.log.Error("Error listening on UDP connection: %v", err)
		return err
	}
	defer conn.Close()

	for {
		// s.log.Info("Waiting for message")
		buf := make([]byte, 16)
		_, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			s.log.Error("Error reading from UDP connection: %v", err)
			return err
		}

		// h := fnv.New64a()
		// if _, err := h.Write([]byte(addr.String())); err != nil {
		// 	s.log.Error("Error hashing address: %v", err)
		// 	return err
		// }
		// id := h.Sum64()
		// if _, ok := s.clients[id]; !ok {
		// 	s.log.Info("new client connected:", slog.Any("id", id), slog.Any("addr", addr.String()))
		// 	s.clients[id] = addr
		// }
		fmt.Println("buf", buf)
		s.handleFn(conn, addr, buf) // fmt.Println("len(clients)", len(s.clients))
		// for id, caddr := range s.clients {
		// 	var cbuf [16]byte
		// 	copy(cbuf[:], buf[0:])
		// 	fmt.Println("id, addr", id, caddr.String())
		// 	s.handleFn(conn, caddr, cbuf[:])
		// }
	}
}

type Option func(*Server)

func WithHandlerFunc(h HandleFunc) Option {
	return func(s *Server) {
		s.handleFn = h
	}
}

func WithLogger(l *slog.Logger) Option {
	return func(s *Server) {
		s.log = l
	}
}

func NewServer(t, a string, opts ...Option) *Server {
	th := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	s := &Server{
		Type:     t,
		Addr:     a,
		log:      slog.New(th),
		mu:       sync.RWMutex{},
		clients:  make(map[string]*net.UDPAddr),
		shutdown: make(chan struct{}),
	}

	for _, opt := range opts {
		opt(s)
	}
	if s.handleFn == nil {
		EchoHandler(s)
	}
	return s
}
