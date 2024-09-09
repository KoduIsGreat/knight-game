package sprite

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	maxFrameSpeed = 15
	minFrameSpeed = 1
)

type Sprite2dAnimator struct {
	SpriteSheet   rl.Texture2D
	NumFrames     float32
	Pos           *rl.Vector2
	speed         int
	currentFrame  float32
	frameRec      rl.Rectangle
	framesCounter int
}

func (s *Sprite2dAnimator) SetSpeed(speed int) {
	s.speed = speed
}

type Option func(s *Sprite2dAnimator)

func NewSprite2dAnimator(spriteSheet rl.Texture2D, numFrames float32, pos *rl.Vector2) *Sprite2dAnimator {
	return &Sprite2dAnimator{
		SpriteSheet:   spriteSheet,
		NumFrames:     numFrames,
		Pos:           pos,
		frameRec:      rl.NewRectangle(pos.X, pos.Y, float32(spriteSheet.Width/int32(numFrames)), float32(spriteSheet.Height)),
		speed:         3,
		framesCounter: 0,
	}
}

func (s *Sprite2dAnimator) Destroy() {
	rl.UnloadTexture(s.SpriteSheet)
}

func (s *Sprite2dAnimator) Update(dt float64) {
	s.framesCounter++
	if s.framesCounter >= (60 / s.speed) {
		s.framesCounter = 0
		s.currentFrame++
		if s.currentFrame > s.NumFrames-1 {
			s.currentFrame = 0
		}
		s.frameRec.X = s.currentFrame * float32(s.SpriteSheet.Width) / (s.NumFrames)
	}
	if rl.IsKeyPressed(rl.KeyRight) {
		s.speed++
	} else if rl.IsKeyPressed(rl.KeyLeft) {
		s.speed--
	}
	if s.speed > maxFrameSpeed {
		s.speed = maxFrameSpeed
	} else if s.speed < minFrameSpeed {
		s.speed = minFrameSpeed
	}
}

func (s *Sprite2dAnimator) Render() {
	rl.DrawTextureRec(s.SpriteSheet, s.frameRec, *s.Pos, rl.White) // Draw part of the texture
}
