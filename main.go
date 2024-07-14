package main

import (
	// "fmt"

	// "log"

	"github.com/gen2brain/raylib-go/raylib"
)

const (
	maxFrameSpeed = 15
	minFrameSpeed = 1
)

type Sprite2dAnimator struct {
	SpriteSheet   rl.Texture2D
	pos           rl.Vector2
	NumFrames     float32
	Speed         int
	currentFrame  float32
	frameRec      rl.Rectangle
	framesCounter int
}

func NewSprite2dAnimator(spriteSheet rl.Texture2D, numFrames float32, speed int) *Sprite2dAnimator {
	return &Sprite2dAnimator{
		SpriteSheet:   spriteSheet,
		NumFrames:     numFrames,
		frameRec:      rl.NewRectangle(0, 0, float32(spriteSheet.Width/int32(numFrames)), float32(spriteSheet.Height)),
		Speed:         speed,
		framesCounter: 0,
	}
}

func (s *Sprite2dAnimator) Destroy() {
	// rl.UnloadTexture(s.SpriteSheet)
}

func (s *Sprite2dAnimator) Update(dt float32) {
	s.framesCounter++
	if s.framesCounter >= (60 / s.Speed) {
		s.framesCounter = 0
		s.currentFrame++
		if s.currentFrame > s.NumFrames-1 {
			s.currentFrame = 0
		}
		s.frameRec.X = s.currentFrame * float32(s.SpriteSheet.Width) / (s.NumFrames)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		s.Speed++
	} else if rl.IsKeyPressed(rl.KeyLeft) {
		s.Speed--
	}
	if s.Speed > maxFrameSpeed {
		s.Speed = maxFrameSpeed
	} else if s.Speed < minFrameSpeed {
		s.Speed = minFrameSpeed
	}
}

func (s *Sprite2dAnimator) Render() {
	rl.ClearBackground(rl.RayWhite)
	pos := rl.NewVector2(350.0, 280.0)
	rl.DrawTextureRec(s.SpriteSheet, s.frameRec, pos, rl.White) // Draw part of the texture
}

func main() {
	screenWidth := int32(1280)
	screenHeight := int32(720)

	rl.InitWindow(screenWidth, screenHeight, "raylib [textures] example - texture loading and drawing")

	// NOTE: Textures MUST be loaded after Window initialization (OpenGL context is required)
	ss := rl.LoadTexture("./assets/KnightIdle_ss.png") // Texture loading
	// scarfy := rl.LoadTexture("./assets/scarfy.png")
	animator := NewSprite2dAnimator(ss, 2, 2)
	defer animator.Destroy()

	// position := rl.NewVector2(350.0, 280.0)
	// framesSpeed := 8 // Number of spritesheet frames shown by second

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		// rl.DrawTexture(scarfy, 15, 40, rl.White)
		// rl.DrawTexture(ss, 350, 40, rl.White)
		// rl.DrawRectangleLines(15, 40, scarfy.Width, scarfy.Height, rl.Lime)
		// rl.DrawRectangleLines(15+int32(frameRec.X), 40+int32(frameRec.Y), int32(frameRec.Width), int32(frameRec.Height), rl.Red)

		// rl.DrawText("FRAME SPEED: ", 165, 210, 10, rl.DarkGray)
		// rl.DrawText(fmt.Sprintf("%02d FPS", framesSpeed), 575, 210, 10, rl.DarkGray)
		// rl.DrawText("PRESS RIGHT/LEFT KEYS to CHANGE SPEED!", 290, 240, 10, rl.DarkGray)

		animator.Update(rl.GetFrameTime())
		animator.Render()
		// for i := 0; i < maxFrameSpeed; i++ {
		// 	if i < framesSpeed {
		// 		rl.DrawRectangle(int32(250+21*i), 205, 20, 20, rl.Red)
		// 	}
		// 	rl.DrawRectangleLines(int32(250+21*i), 205, 20, 20, rl.Maroon)
		// }

		// rl.DrawTextureRec(scarfy, frameRec, position, rl.White) // Draw part of the texture

		// rl.DrawText("(c) Scarfy sprite by Eiden Marsal", screenWidth-200, screenHeight-20, 10, rl.Gray)

		rl.EndDrawing()
	}

	// rl.UnloadTexture(scarfy)

	rl.CloseWindow()
}
