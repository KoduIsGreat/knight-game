package main

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/nw"
	"github.com/KoduIsGreat/knight-game/state/snake"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
	}
}

func run() error {
	sm := snake.NewServerStateManager()
	s := nw.NewServer(sm)
	return s.Listen()
}
