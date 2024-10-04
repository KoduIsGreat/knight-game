package main

import (
	"fmt"
	"os"

	"github.com/KoduIsGreat/knight-game/nw"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running client: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	return nw.StartClient()
}
