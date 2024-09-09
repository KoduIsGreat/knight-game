package game

import (
	"time"

	"github.com/gen2brain/raylib-go/physics"
	"github.com/gen2brain/raylib-go/raylib"
	"github.com/gofrs/uuid"
)

const (
	ScreenHeight = 720
	ScreenWidth  = 1280
	TargetFPS    = 60
)

var (
	BackgroundColor = rl.RayWhite
)

type Game interface {
	Init()
	AddComponent(c Component)
	Process()
}

type game struct {
	Title      string
	components []Component
	Camera     rl.Camera2D
}

type Component interface {
	Update(dt float64)
	Render()
	AddChild(c Component)
	GetChildren() []Component
	Destroy()
}

type BaseComponent struct {
	GUID     uuid.UUID
	Children []Component
}

type Option func(g *game)

func NewGame(title string) Game {
	g := &game{Title: title}
	g.Init()
	return g
}

func (g *game) Init() {
	rl.InitWindow(ScreenWidth, ScreenHeight, g.Title)
	// physics.Init()
	rl.SetTargetFPS(TargetFPS)
}

func (g *game) AddComponent(c Component) {
	g.components = append(g.components, c)
}

func (g *game) Process() {
	for !rl.WindowShouldClose() {
		since := time.Duration(0)
		physics.Update()
		for _, c := range g.components {
			ct := time.Now()
			c.Update(since.Seconds())
			rl.BeginDrawing()
			rl.BeginMode2D(g.Camera)
			rl.ClearBackground(BackgroundColor)
			c.Render()
			rl.EndMode2D()
			rl.EndDrawing()
			since = time.Since(ct)
		}
	}
}
