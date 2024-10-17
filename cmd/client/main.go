package main

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/nw"
	"github.com/KoduIsGreat/knight-game/state/snake"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowWidth  = 1200
	windowHeight = 1000
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running server: %v\n", err)
	}
}

type Game struct {
	client       *nw.Client[snake.GameState]
	renderEngine snake.RaylibRenderer
}

func (g *Game) handleInput() {
	var input string
	if rl.IsKeyPressed(rl.KeyW) {
		input = "UP"
	} else if rl.IsKeyPressed(rl.KeyS) {
		input = "DOWN"
	} else if rl.IsKeyPressed(rl.KeyA) {
		input = "LEFT"
	} else if rl.IsKeyPressed(rl.KeyD) {
		input = "RIGHT"
	}
	if rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) {
		g.renderEngine.Camera.Zoom += 0.1
	}
	if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
		g.renderEngine.Camera.Zoom -= 0.1
		if g.renderEngine.Camera.Zoom < 0.1 {
			g.renderEngine.Camera.Zoom = 0.1
		}
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		g.renderEngine.Camera.Target.X -= 10
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		g.renderEngine.Camera.Target.X += 10
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		g.renderEngine.Camera.Target.Y -= 10
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		g.renderEngine.Camera.Target.Y += 10
	}

	if input != "" {
		g.client.State().UpdateLocal(input)
		g.client.SendInputToServer(input)
	}
}

func run() error {
	sm := snake.NewClientStateManger()
	renderer := snake.NewRaylibRenderer()

	renderer.Init()
	defer renderer.Close()
	g := Game{
		client:       nw.NewClient(sm, nw.ClientOpts{}),
		renderEngine: renderer,
	}
	for !g.renderEngine.ShouldClose() {
		select {
		case msg := <-g.client.RecvFromServer():
			g.client.State().ReconcileState(msg)
		case <-g.client.QuitChan():
			return nil
		default:
			break
		}
		g.handleInput()
		g.client.State().Update(float64(rl.GetFrameTime()))
		g.renderEngine.Render(g.client.State())
	}

	return nil

}
