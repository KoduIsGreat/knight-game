package main

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/common"
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
	client       *nw.Client[common.GameState]
	camera       rl.Camera2D
	cameraTarget rl.Vector2
}

func (g *Game) renderGameState() {
	rl.BeginDrawing()
	rl.BeginMode2D(g.camera)
	rl.ClearBackground(rl.RayWhite)
	state := g.client.State().GetCurrent()

	rl.DrawRectangleLines(
		state.World.ToInt32().X,
		state.World.ToInt32().Y,
		state.World.ToInt32().Width,
		state.World.ToInt32().Height,
		rl.Black,
	)

	for _, snake := range state.Snakes {
		color := rl.Green
		if snake.ID == g.client.State().ClientID() {
			color = rl.Blue
		}
		for _, segment := range snake.Segments {
			rl.DrawRectangle(
				int32(segment.X*10),
				int32(segment.Y*10),
				10,
				10,
				color,
			)
		}
	}

	for _, food := range state.FoodItems {
		rl.DrawCircle(
			int32(food.X*10+5),
			int32(food.Y*10+5),
			5,
			rl.Red,
		)
	}

	rl.EndMode2D()
	rl.EndDrawing()
}

func (g *Game) Loop() {
	for !rl.WindowShouldClose() {
		select {
		case msg := <-g.client.RecvFromServer():
			g.client.State().ReconcileState(msg)
		case <-g.client.QuitChan():
			return
		default:
			break
		}
		g.handleInput()
		g.client.State().Update(float64(rl.GetFrameTime()))
		g.renderGameState()
	}

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
		g.camera.Zoom += 0.1
	}
	if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
		g.camera.Zoom -= 0.1
		if g.camera.Zoom < 0.1 {
			g.camera.Zoom = 0.1
		}
	}
	if rl.IsKeyDown(rl.KeyLeft) {
		g.camera.Target.X -= 10
	}
	if rl.IsKeyDown(rl.KeyRight) {
		g.camera.Target.X += 10
	}
	if rl.IsKeyDown(rl.KeyDown) {
		g.camera.Target.Y -= 10
	}
	if rl.IsKeyDown(rl.KeyUp) {
		g.camera.Target.Y += 10
	}

	if input != "" {
		g.client.State().UpdateLocal(input)
		g.client.SendInputToServer(input)
	}
}

func (g *Game) initializeCamera() {
	g.camera = rl.Camera2D{
		Offset:   rl.NewVector2(float32(windowWidth)/2, float32(windowHeight)/2),
		Target:   rl.NewVector2(0, 0),
		Rotation: 0.0,
		Zoom:     1.0,
	}
	g.cameraTarget = g.camera.Target
}

func run() error {
	rl.InitWindow(windowWidth, windowHeight, "Snake Game")
	defer rl.CloseWindow()
	sm := snake.NewClientStateManger()
	g := Game{
		client: nw.NewClient(sm),
	}
	g.Loop()

	return nil

}
