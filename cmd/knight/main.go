package main

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/game"
	"github.com/KoduIsGreat/knight-game/player"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		return
	}

}
func run() error {
	g := game.NewGame("Knight Game")
	p := player.NewPlayer()
	g.AddComponent(p)

	g.Process()
	return nil
}
