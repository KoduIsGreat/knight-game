package main

import (
	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {

	const (
		screenWidth  = 690
		screenHeight = 560
	)

	var (
		exitWindow          bool
		showClosingInputBox bool
	)
	closingInputBoxRect := rl.Rectangle{float32(rl.GetScreenWidth())/2 - 120, float32(rl.GetScreenHeight())/2 - 60, 240, 140}
	rl.InitWindow(screenWidth, screenHeight, "raygui - controls test suite")
	rl.SetTargetFPS(60)
	for !exitWindow { // Detect window close button or ESC key
		exitWindow = rl.WindowShouldClose()

		if rl.IsKeyPressed(rl.KeyEscape) {
			showClosingInputBox = !showClosingInputBox
		}

		if rl.IsKeyDown(rl.KeyLeftControl) && rl.IsKeyPressed(rl.KeyS) {
			showClosingInputBox = true
		}
		rl.BeginDrawing()

		rl.ClearBackground(rl.GetColor(uint(gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR))))
		if showClosingInputBox {
			rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.RayWhite, 0.8))

			var result int32 = gui.MessageBox(closingInputBoxRect, gui.IconText(gui.ICON_EXIT, "Close Window"), "Do you really want to exit?", "Yes;No")

			if (result == 0) || (result == 2) {
				showClosingInputBox = false
			} else if result == 1 {
				exitWindow = true
			}
		}
		rl.EndDrawing()
	}
}
