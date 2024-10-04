package common

import (
	"encoding/json"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID        string     `json:"id"`
	Segments  []Position `json:"segments"`
	Direction string     `json:"direction"`
}

type FoodItem struct {
	Position `json:"position"`
}

type GameState struct {
	Snakes    map[string]*Snake `json:"snakes"`
	FoodItems []FoodItem        `json:"foodItems"`
	World     rl.Rectangle
}

func (s GameState) Serialize() ([]byte, error) {
	return json.Marshal(s)
}
