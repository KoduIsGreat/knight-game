package main

import (
	"fmt"
	"time"

	"github.com/KoduIsGreat/knight-game/nw"
	"github.com/KoduIsGreat/knight-game/state/snake"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/quic-go/quic-go"

	gui "github.com/gen2brain/raylib-go/raygui"
)

const (
	windowWidth  = 1200
	windowHeight = 1000
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error running client: %v\n", err)
	}
}

type Game struct {
	client       *nw.Client[snake.GameState]
	started      bool
	renderEngine snake.RaylibRenderer
	filterText   string
	activeTab    int32
}

func (g *Game) gameLoop() {
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
	g.renderEngine.Render(g.client.State())
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

const (
	// ServerLobbyBrowser is the index of the server lobby browser tab.
	ServerLobbyBrowser = 0
	// Settings is the index of the settings tab.
	Settings = 1
	// Lobby is the index of the lobby tab.
	Lobby = 2
)

func (g *Game) renderLobbies() {
	gui.SetStyle(gui.BUTTON, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	if gui.Button(rl.NewRectangle(10, 75, 100, 20), "Refresh") {
		g.client.SyncLobbies()
	}

	gui.SetStyle(gui.BUTTON, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	if gui.Button(rl.NewRectangle(115, 75, 100, 20), "Create") {
		g.client.CreateLobby()
		g.client.SyncLobbies()
		g.activeTab = Lobby

	}

	codeRect := rl.NewRectangle(10, 100, 100, 20)
	rl.DrawRectangleRounded(codeRect, 0, 0, rl.Black)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(codeRect, "Code")

	ownerRect := rl.NewRectangle(110, 100, 100, 20)
	rl.DrawRectangleRounded(ownerRect, 0, 0, rl.Black)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(ownerRect, "Owner")
	playerRect := rl.NewRectangle(210, 100, 100, 20)
	rl.DrawRectangleRounded(playerRect, 0, 0, rl.Black)
	gui.SetStyle(gui.LABEL, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	gui.Label(playerRect, "Players")
	for i, lobby := range g.client.Lobbies.Lobbies {

		codeCellRect := rl.NewRectangle(10, 105+float32((i+1)*20), 100, 20)
		rl.DrawRectangleRec(codeCellRect, rl.Gray)
		gui.Label(codeCellRect, lobby.Code)
		ownerCellRect := rl.NewRectangle(110, 105+float32((i+1)*20), 100, 20)
		rl.DrawRectangleRec(ownerCellRect, rl.Gray)
		gui.Label(ownerCellRect, lobby.OwnerID)
		playerCellRect := rl.NewRectangle(210, 105+float32((i+1)*20), 100, 20)
		rl.DrawRectangleRec(playerCellRect, rl.Gray)
		gui.Label(playerCellRect, fmt.Sprintf("%d/%d", lobby.NumClients, lobby.MaxClients))

		gui.SetStyle(gui.BUTTON, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
		if gui.Button(rl.NewRectangle(310, 105+float32((i+1)*20), 100, 20), "Join") {
			g.client.JoinLobby(lobby.Code)
		}
		i++
	}
}

func (g *Game) renderSettings() {

}

func (g *Game) renderServerLobbyBrowser() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.GetColor(uint(gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR))))
	rl.DrawText(fmt.Sprintf("Server: %s"), 10, 10, 20, rl.Black)
	gui.Label(rl.NewRectangle(10, 40, 100, 20), "Lobbies")
	tabs := []string{"Lobbies", "Settings"}
	if g.client.Lobby() != nil {
		tabs = append(tabs, g.client.Lobby().ID)
	}
	gui.TabBar(rl.NewRectangle(10, 40, 100, 20), tabs, &g.activeTab)

	switch g.activeTab {
	case ServerLobbyBrowser:
		g.renderLobbies()
	case Settings:
		g.renderSettings()
	case Lobby:
		g.renderLobby()
	}

	rl.EndDrawing()
}

func (g *Game) renderLobby() {
	rl.BeginDrawing()
	rl.ClearBackground(rl.GetColor(uint(gui.GetStyle(gui.DEFAULT, gui.BACKGROUND_COLOR))))
	rl.DrawText("Lobby", 10, 10, 20, rl.Black)
	gui.Label(rl.NewRectangle(10, 40, 100, 20), "Lobby")
	lobby := g.client.Lobby()
	gui.Label(rl.NewRectangle(10, 40, 100, 20), lobby.ID)
	gui.Label(rl.NewRectangle(110, 40, 100, 20), lobby.OwnerClientID)
	gui.Label(rl.NewRectangle(210, 40, 100, 20), fmt.Sprintf("%d/%d", len(lobby.ConnectedClients), lobby.MaxPlayers))

	gui.SetStyle(gui.BUTTON, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	if gui.Button(rl.NewRectangle(310, 40, 100, 20), "Start") {
		g.client.Start()
	}

	gui.SetStyle(gui.BUTTON, gui.TEXT_ALIGNMENT, gui.TEXT_ALIGN_CENTER)
	if gui.Button(rl.NewRectangle(310, 60, 100, 20), "Leave") {
		g.client.LeaveLobby()
	}
	rl.EndDrawing()
}

func run() error {
	sm := snake.NewClientStateManger()
	renderer := snake.NewRaylibRenderer()

	renderer.Init()
	defer renderer.Close()
	g := Game{
		client:       nw.NewClient(sm, nw.ClientOpts{QuicConfig: &quic.Config{KeepAlivePeriod: time.Second, MaxIdleTimeout: time.Minute * 15}}),
		renderEngine: renderer,
	}
	for !g.renderEngine.ShouldClose() {
		if g.client.IsStarted() {
			g.gameLoop()
		} else {
			g.renderServerLobbyBrowser()
		}
	}

	return nil

}
