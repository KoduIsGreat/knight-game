package main

import (
	embed "embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed assets/*
var assets embed.FS

var (
	screenWidth       int32 = 1000
	screenHeight      int32 = 480
	camera            rl.Camera2D
	running           bool     = true
	bkgColor          rl.Color = rl.NewColor(147, 211, 196, 255)
	grassSprite       rl.Texture2D
	fenceSprite       rl.Texture2D
	HillsSprite       rl.Texture2D
	WaterSprite       rl.Texture2D
	TilledDirtSprite  rl.Texture2D
	HouseSprite       rl.Texture2D
	tex               rl.Texture2D
	playerSprite      rl.Texture2D
	playerSrc         rl.Rectangle
	playerDst         rl.Rectangle
	playerSpeed       float32 = 3
	playerMoving      bool
	playerIsPressedUp bool
	playerIsPressedDn bool
	playerIsPressedLt bool
	playerIsPressedRt bool

	tileDst rl.Rectangle
	tileSrc rl.Rectangle
	tileMap []int
	srcMap  []string

	mapWidth, mapHeight int
	frameCount          int
	playerFrame         int
	playerDir           playerDirection
	music               rl.Music
	musicPaused         bool = false
)

type playerDirection int

const (
	playerDown playerDirection = iota
	playerUp
	playerLeft
	playerRight
)

func drawScene() {
	for i := 0; i < len(tileMap); i++ {
		if tileMap[i] != 0 {
			tileDst.X = tileDst.Width * float32(i%mapWidth)
			tileDst.Y = tileDst.Height * float32(i/mapWidth)
			switch srcMap[i] {
			case "g":
				tex = grassSprite
			case "w":
				tex = WaterSprite
			case "f":
				tex = fenceSprite
			case "h":
				tex = HillsSprite
			case "t":
				tex = TilledDirtSprite
			case "H":
				tex = HouseSprite
			}

			tileSrc.X = tileSrc.Width * float32(tileMap[i]-1%int(tex.Width/int32(tileSrc.Width)))
			tileSrc.Y = tileSrc.Height * float32(tileMap[i]-1%int(tex.Width/int32(tileSrc.Width)))
			fmt.Println(i, tileMap[i], srcMap[i], tileSrc.X, tileSrc.Y, tileSrc.Width, tileSrc.Height, tileDst.X, tileDst.Y, tileDst.Width, tileDst.Height)
			rl.DrawTexturePro(tex, tileSrc, tileDst, rl.NewVector2(tileDst.Width, tileDst.Height), 0, rl.White)
		}

	}
	rl.DrawTexturePro(playerSprite, playerSrc, playerDst, rl.NewVector2(playerDst.Width, playerDst.Height), 0, rl.White)
}
func input() {
	if rl.IsKeyDown(rl.KeyW) {
		playerMoving = true
		playerDir = playerUp
		playerIsPressedUp = true
	}
	if rl.IsKeyDown(rl.KeyS) {
		playerMoving = true
		playerDir = playerDown
		playerIsPressedDn = true
	}
	if rl.IsKeyDown(rl.KeyA) {
		playerMoving = true
		playerDir = playerLeft
		playerIsPressedLt = true
	}
	if rl.IsKeyDown(rl.KeyD) {
		playerMoving = true
		playerDir = playerRight
		playerIsPressedRt = true
	}

	if rl.IsKeyPressed(rl.KeyM) {
		musicPaused = !musicPaused
	}
}
func update() {
	running = !rl.WindowShouldClose()

	playerSrc.X = playerSrc.Width * float32(playerFrame)
	if playerMoving {
		if playerIsPressedUp {
			playerDst.Y -= playerSpeed
		}
		if playerIsPressedDn {
			playerDst.Y += playerSpeed
		}

		if playerIsPressedLt {
			playerDst.X -= playerSpeed
		}
		if playerIsPressedRt {
			playerDst.X += playerSpeed
		}
		if frameCount%8 == 1 {
			playerFrame++
		}
	} else if frameCount%45 == 1 {
		playerFrame++
	}
	frameCount++
	if playerFrame > 3 {
		playerFrame = 0
	}
	if !playerMoving && playerFrame > 1 {
		playerFrame = 0
	}
	playerSrc.X = playerSrc.Width * float32(playerFrame)
	playerSrc.Y = playerSrc.Height * float32(playerDir)

	rl.UpdateMusicStream(music)
	if musicPaused {
		rl.PauseMusicStream(music)
	} else {
		rl.ResumeMusicStream(music)
	}
	camera.Target = rl.NewVector2(playerDst.X-playerDst.Width/2, playerDst.Y-playerDst.Height/2)
	playerMoving = false
	playerIsPressedUp, playerIsPressedDn, playerIsPressedLt, playerIsPressedRt = false, false, false, false
}

func render() {
	rl.BeginDrawing()
	rl.ClearBackground(bkgColor)
	rl.BeginMode2D(camera)
	drawScene()
	rl.EndMode2D()
	rl.EndDrawing()
}

func quit() {
	rl.UnloadTexture(grassSprite)
	rl.UnloadTexture(playerSprite)
	rl.CloseWindow()
}

func loadMap() {
	f, err := os.ReadFile("assets/map.txt")
	if err != nil {
		panic(err)
	}
	remNewLines := strings.ReplaceAll(string(f), "\n", " ")
	if len(remNewLines) == 0 {
		panic("no map data")
	}
	mapWidth = -1
	mapHeight = -1
	tileStr := strings.Split(remNewLines, " ")
	// fmt.Println(tileStr)
	for i := 0; i < len(tileStr); i++ {
		val, _ := strconv.Atoi(tileStr[i])
		fmt.Println(tileStr[i])
		if mapWidth == -1 {
			mapWidth = val
		} else if mapHeight == -1 {
			mapHeight = val
		} else if i < mapWidth*mapHeight+2 {
			tileMap = append(tileMap, val)
		} else {
			if tileStr[i] != "" {
				srcMap = append(srcMap, tileStr[i])
			}
		}
	}
	if len(tileMap) > mapWidth*mapHeight {
		tileMap = tileMap[:len(tileMap)-1]
	}
}

func init() {
	rl.InitWindow(screenWidth, screenHeight, "sproutlings")
	rl.SetExitKey(0)
	rl.SetTargetFPS(60)
	tileDst = rl.NewRectangle(0, 0, 16, 16)
	tileSrc = rl.NewRectangle(0, 0, 16, 16)

	grassSprite = rl.LoadTexture("assets/Tilesets/Grass.png")
	WaterSprite = rl.LoadTexture("assets/Tilesets/Water.png")
	fenceSprite = rl.LoadTexture("assets/Tilesets/Fences.png")
	HillsSprite = rl.LoadTexture("assets/Tilesets/Hills.png")
	TilledDirtSprite = rl.LoadTexture("assets/Tilesets/TilledDirt.png")
	HouseSprite = rl.LoadTexture("assets/Tilesets/WoodenHouse.png")
	playerSprite = rl.LoadTexture("assets/Characters/BasicCharakterSpriteSheet.png")
	playerSrc = rl.NewRectangle(0, 0, 48, 48)
	playerDst = rl.NewRectangle(200, 200, 100, 100)
	rl.InitAudioDevice()

	music = rl.LoadMusicStream("assets/audio/music/8bitarp.mp3")
	musicPaused = true
	// rl.PlayMusicStream(music)
	camera = rl.NewCamera2D(rl.NewVector2(float32(screenWidth)/2, float32(screenHeight/2)), rl.NewVector2(playerDst.X-playerDst.Width/2, playerDst.Y-playerDst.Height/2), 0, 1.5)
	loadMap()
}

func main() {
	for running {
		update()
		input()
		render()
	}
	quit()
}
