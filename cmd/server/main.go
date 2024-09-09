package main

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/networking/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
	}
}

func run() error {
	fmt.Println("Server started on :1111")
	s := server.NewServer("udp", ":1111")
	return s.Listen()
}
