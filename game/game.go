package game

import (
	"log"
	"time"

	"github.com/gen2brain/raylib-go/physics"
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gofrs/uuid"
)

type Game interface {
	Init()
	AddComponent(c Component)
	Process()
}

type game struct {
	Title string
	Fps   float64
	Resolution
	components []Component
}

type Component interface {
	Update(dt float64)
	Render()
	Physics(dt float64)
	AddChild(c Component)
	GetChildren() []Component
	Destroy()
}

type BaseComponent struct {
	GUID     uuid.UUID
	Children []Component
}

type Resolution struct {
	Width  int32
	Height int32
}

type Option func(g *game)

func NewGame(title string, r Resolution) Game {
	g := &game{Title: title, Resolution: r}
	g.Init()
	return g
}

func (g *game) Init() {
	// do some things here
	rl.InitWindow(g.Resolution.Width, g.Resolution.Height, g.Title)
	physics.Init()
}

func (g *game) AddComponent(c Component) {
	g.components = append(g.components, c)
}

func (g *game) Process() {
	log.Println("we gaming")
	for !rl.WindowShouldClose() {
		since := time.Duration(0)
		for _, c := range g.components {
			ct := time.Now()
			c.Update(since.Seconds())
			physics.Update()
			c.Physics(since.Seconds())
			rl.BeginDrawing()
			rl.ClearBackground(rl.RayWhite)
			c.Render()
			rl.EndDrawing()
			since = time.Since(ct)
		}
	}
}
