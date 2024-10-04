package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	quic "github.com/quic-go/quic-go"
)

const (
	serverAddress       = "localhost:4242"
	windowWidth         = 800
	windowHeight        = 600
	interpolationTimeMs = 100
	worldWidth          = 1000
	worldHeight         = 1000
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID        string     `json:"id"`
	Segments  []Position `json:"segments"`
	Direction string     `json:"direction"`
}

type FoodItem struct {
	Position `json:"position"`
}

type GameState struct {
	Snakes    map[string]*Snake `json:"snakes"`
	FoodItems []FoodItem        `json:"foodItems"`
}

type InputMessage struct {
	Sequence uint32 `json:"sequence"`
	Input    string `json:"input"`
}

type ServerStateMessage struct {
	GameState       GameState         `json:"gameState"`
	AcknowledgedSeq map[string]uint32 `json:"acknowledgedSeq"`
}

type Client struct {
	stream           quic.Stream
	sendChan         chan []byte
	recvChan         chan ServerStateMessage
	quitChan         chan struct{}
	inputSequence    uint32
	inputHistory     map[uint32]string
	stateHistory     map[uint32]GameState
	gameState        GameState
	targetGameState  *GameState
	interpolateUntil time.Time
	clientID         string
	camera           rl.Camera2D
	cameraTarget     rl.Vector2
}

func main() {
	rl.InitWindow(windowWidth, windowHeight, "Snake Game")
	defer rl.CloseWindow()

	client := &Client{
		sendChan:     make(chan []byte, 10),
		recvChan:     make(chan ServerStateMessage, 10),
		quitChan:     make(chan struct{}),
		inputHistory: make(map[uint32]string),
		stateHistory: make(map[uint32]GameState),
		gameState:    GameState{Snakes: make(map[string]*Snake)},
	}

	client.initializeCamera()
	client.connectToServer()
	client.startNetworkHandlers()
	client.gameLoop()
}

func (c *Client) initializeCamera() {
	c.camera = rl.Camera2D{
		Offset:   rl.NewVector2(float32(windowWidth)/2, float32(windowHeight)/2),
		Target:   rl.NewVector2(0, 0),
		Rotation: 0.0,
		Zoom:     1.0,
	}
	c.cameraTarget = c.camera.Target
}

func (c *Client) connectToServer() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"snake-game"},
	}

	session, err := quic.DialAddr(context.Background(), serverAddress, tlsConfig, nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}

	c.stream, err = session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal("Failed to open stream:", err)
	}

	c.clientID = session.LocalAddr().String()
}

func (c *Client) startNetworkHandlers() {
	go func() {
		for {
			select {
			case data := <-c.sendChan:
				_, err := c.stream.Write(data)
				if err != nil {
					log.Println("Error sending data:", err)
					c.quitChan <- struct{}{}
					return
				}
			case <-c.quitChan:
				return
			}
		}
	}()

	go func() {
		reader := bufio.NewReader(c.stream)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Error receiving game state:", err)
				c.quitChan <- struct{}{}
				return
			}

			var serverMessage ServerStateMessage
			err = json.Unmarshal([]byte(message), &serverMessage)
			if err != nil {
				log.Println("Error parsing game state:", err)
				continue
			}

			c.recvChan <- serverMessage
		}
	}()
}

func (c *Client) gameLoop() {
	for !rl.WindowShouldClose() {
		select {
		case serverMessage := <-c.recvChan:
			c.reconcileGameState(serverMessage)
		case <-c.quitChan:
			return
		default:
			break
		}

		c.handleInput()
		c.updateGameState()

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		c.renderGameState()
		rl.EndDrawing()
	}
}

func (c *Client) handleInput() {
	var input string
	if rl.IsKeyPressed(rl.KeyUp) {
		input = "UP"
	} else if rl.IsKeyPressed(rl.KeyDown) {
		input = "DOWN"
	} else if rl.IsKeyPressed(rl.KeyLeft) {
		input = "LEFT"
	} else if rl.IsKeyPressed(rl.KeyRight) {
		input = "RIGHT"
	}

	if rl.IsKeyPressed(rl.KeyEqual) || rl.IsKeyPressed(rl.KeyKpAdd) {
		c.camera.Zoom += 0.1
	}
	if rl.IsKeyPressed(rl.KeyMinus) || rl.IsKeyPressed(rl.KeyKpSubtract) {
		c.camera.Zoom -= 0.1
		if c.camera.Zoom < 0.1 {
			c.camera.Zoom = 0.1
		}
	}

	if input != "" {
		c.inputSequence++
		seq := c.inputSequence

		c.updateLocalGameState(input)
		c.inputHistory[seq] = input
		c.stateHistory[seq] = c.gameState

		c.sendInputToServer(seq, input)
	}
}

func (c *Client) sendInputToServer(seq uint32, input string) {
	msg := InputMessage{
		Sequence: seq,
		Input:    input,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling input:", err)
		return
	}
	c.sendChan <- append(data, '\n')
}

func (c *Client) updateLocalGameState(input string) {
	snake, exists := c.gameState.Snakes[c.clientID]
	if !exists {
		snake = &Snake{
			ID:        c.clientID,
			Segments:  []Position{{X: 10, Y: 10}},
			Direction: "RIGHT",
		}
		c.gameState.Snakes[c.clientID] = snake
	}

	opposite := map[string]string{
		"UP":    "DOWN",
		"DOWN":  "UP",
		"LEFT":  "RIGHT",
		"RIGHT": "LEFT",
	}

	if input != opposite[snake.Direction] {
		snake.Direction = input
	}
}

func (c *Client) updateGameState() {
	if c.targetGameState != nil && time.Now().Before(c.interpolateUntil) {
		elapsed := float32(time.Since(c.interpolateUntil.Add(-interpolationTimeMs * time.Millisecond)).Milliseconds())
		factor := elapsed / float32(interpolationTimeMs)
		if factor > 1.0 {
			factor = 1.0
		} else if factor < 0.0 {
			factor = 0.0
		}
		c.gameState = interpolateStates(c.gameState, *c.targetGameState, factor)
	} else if c.targetGameState != nil {
		c.gameState = *c.targetGameState
		c.targetGameState = nil
	}

	snake, exists := c.gameState.Snakes[c.clientID]
	if exists {
		moveSnake(snake)

		// Smoothly interpolate camera position
		if len(snake.Segments) > 0 {
			head := snake.Segments[0]
			targetX := float32(head.X*10 + 5)
			targetY := float32(head.Y*10 + 5)

			// Interpolate camera target position
			c.cameraTarget.X += (targetX - c.cameraTarget.X) * 0.1
			c.cameraTarget.Y += (targetY - c.cameraTarget.Y) * 0.1

			c.camera.Target = c.cameraTarget
		}
	}
}

func (c *Client) renderGameState() {
	rl.BeginMode2D(c.camera)

	for _, snake := range c.gameState.Snakes {
		color := rl.Green
		if snake.ID == c.clientID {
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

	for _, food := range c.gameState.FoodItems {
		rl.DrawCircle(
			int32(food.X*10+5),
			int32(food.Y*10+5),
			5,
			rl.Red,
		)
	}

	rl.EndMode2D()
}

func (c *Client) reconcileGameState(serverMessage ServerStateMessage) {
	serverGameState := serverMessage.GameState
	acknowledgedSeq := serverMessage.AcknowledgedSeq[c.clientID]

	for seq := range c.inputHistory {
		if seq <= acknowledgedSeq {
			delete(c.inputHistory, seq)
			delete(c.stateHistory, seq)
		}
	}

	predictedState, exists := c.stateHistory[acknowledgedSeq]
	if !exists || !compareGameStates(serverGameState, predictedState, c.clientID) {
		c.targetGameState = &serverGameState
		c.interpolateUntil = time.Now().Add(interpolationTimeMs * time.Millisecond)
	}
}

func compareGameStates(a, b GameState, clientID string) bool {
	snakeA, existsA := a.Snakes[clientID]
	snakeB, existsB := b.Snakes[clientID]
	if !existsA || !existsB {
		return false
	}
	if len(snakeA.Segments) != len(snakeB.Segments) {
		return false
	}
	for i := range snakeA.Segments {
		if snakeA.Segments[i] != snakeB.Segments[i] {
			return false
		}
	}
	return true
}

func moveSnake(snake *Snake) {
	head := snake.Segments[0]
	newHead := head

	switch snake.Direction {
	case "UP":
		newHead.Y -= 1
	case "DOWN":
		newHead.Y += 1
	case "LEFT":
		newHead.X -= 1
	case "RIGHT":
		newHead.X += 1
	}

	if newHead.X < 0 {
		newHead.X = worldWidth/10 - 1
	} else if newHead.X >= worldWidth/10 {
		newHead.X = 0
	}

	if newHead.Y < 0 {
		newHead.Y = worldHeight/10 - 1
	} else if newHead.Y >= worldHeight/10 {
		newHead.Y = 0
	}

	snake.Segments = append([]Position{newHead}, snake.Segments...)
	snake.Segments = snake.Segments[:len(snake.Segments)-1]
}

func interpolateStates(current, target GameState, factor float32) GameState {
	interpolated := GameState{
		Snakes:    make(map[string]*Snake),
		FoodItems: target.FoodItems,
	}

	for id, snake := range target.Snakes {
		currentSnake, exists := current.Snakes[id]
		if !exists {
			interpolated.Snakes[id] = snake
			continue
		}

		interpolatedSnake := &Snake{
			ID:        id,
			Direction: snake.Direction,
			Segments:  make([]Position, len(snake.Segments)),
		}

		for i := range snake.Segments {
			var currentPos Position
			if i < len(currentSnake.Segments) {
				currentPos = currentSnake.Segments[i]
			} else {
				currentPos = snake.Segments[i]
			}

			interpolatedSnake.Segments[i] = Position{
				X: int(lerp(float32(currentPos.X), float32(snake.Segments[i].X), factor)),
				Y: int(lerp(float32(currentPos.Y), float32(snake.Segments[i].Y), factor)),
			}
		}

		interpolated.Snakes[id] = interpolatedSnake
	}

	return interpolated
}

func lerp(a, b, t float32) float32 {
	return a + t*(b-a)
}
