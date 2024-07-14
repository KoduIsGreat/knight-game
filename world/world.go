package world

import (
	"github.com/KoduIsGreat/knight-game/game"
	// "github.com/gen2brain/raylib-go/physics"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type World struct {
	game.BaseComponent
}

func NewWorld() *World {
	world := &World{}
	world.Init()
	return world
}

func (w *World) Init() {

}

// AddChild implements game.Component.
func (w *World) AddChild(c game.Component) {
	w.BaseComponent.Children = append(w.BaseComponent.Children, c)
}

// Children implements game.Component.
func (w *World) GetChildren() []game.Component {
	return w.BaseComponent.Children
}

// Destroy implements game.Component.
func (w *World) Destroy() {
}

// Physics implements game.Component.
func (w *World) Physics(dt float64) {
	panic("unimplemented")
}

// Pos implements game.Component.
func (w *World) Pos() rl.Vector2 {
	panic("unimplemented")
}

// Render implements game.Component.
func (w *World) Render() {
	panic("unimplemented")
}

// Update implements game.Component.
func (w *World) Update(dt float64) {
	panic("unimplemented")
}

var _ game.Component = &World{}
