package main

import (
	"fmt"
	"time"

	"github.com/KoduIsGreat/knight-game/nw"
	"github.com/KoduIsGreat/knight-game/state/snake"
	"github.com/quic-go/quic-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
	}
}

func run() error {
	sm := snake.NewServerStateManager()
	s := nw.NewServer(sm, nw.WithQuicConfig[snake.GameState](&quic.Config{
		KeepAlivePeriod: time.Second,
		MaxIdleTimeout:  time.Minute * 15,
	}))
	return s.Listen()
}
