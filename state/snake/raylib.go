package snake

import (
	"github.com/KoduIsGreat/knight-game/nw"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type RaylibRenderer struct {
	WindowWidth  int32
	WindowHeight int32
	Title        string
	Camera       rl.Camera2D
	CameraTarget rl.Vector2
}

func NewRaylibRenderer() RaylibRenderer {
	camera := rl.Camera2D{
		Offset:   rl.NewVector2(float32(1200)/2, float32(1000)/2),
		Target:   rl.NewVector2(0, 0),
		Rotation: 0.0,
		Zoom:     1.0,
	}
	return RaylibRenderer{
		WindowWidth:  1200,
		WindowHeight: 1000,
		Title:        "Snake",
		Camera:       camera,
		CameraTarget: camera.Target,
	}
}

func (r RaylibRenderer) Init() {
	rl.InitWindow(r.WindowWidth, r.WindowHeight, r.Title)
}

func (r RaylibRenderer) Close() {
	rl.CloseWindow()
}
func (r RaylibRenderer) ShouldClose() bool {
	return rl.WindowShouldClose()
}

func (r RaylibRenderer) Render(m nw.ClientStateManager[GameState]) {
	rl.BeginDrawing()
	rl.BeginMode2D(r.Camera)
	rl.ClearBackground(rl.RayWhite)
	s := m.GetCurrent()
	clientId := m.ClientID()
	for _, snake := range s.Snakes {
		color := rl.Green
		if snake.ID == clientId {
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
	for _, food := range s.FoodItems {
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
