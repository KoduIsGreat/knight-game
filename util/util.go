package util

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

func IsOnGround(pos rl.Vector2) bool {
	if pos.Y <= 0 {
		return true
	}
	return false
}

func IsNotOnGround(pos rl.Vector2) bool {
	if pos.Y > 0 {
		return true
	}
	return false
}
