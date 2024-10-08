package snake

import (
	"fmt"
	"time"

	. "github.com/KoduIsGreat/knight-game/common"
	"github.com/KoduIsGreat/knight-game/nw"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	interpolationTimeMs = 100 // Time in milliseconds for interpolation
)

type ClientStateManager struct {
	clientID         string
	currentState     GameState
	targetState      *GameState
	inputHistory     map[uint32]string
	stateHistory     map[uint32]GameState
	inputSequence    uint32
	interpolateUntil time.Time
}

func newGameState() GameState {
	return GameState{
		Snakes:    make(map[string]*Snake),
		FoodItems: []FoodItem{},
		World:     rl.NewRectangle(0, 0, 600, 600),
	}
}

func NewClientStateManger[T GameState]() *ClientStateManager {
	return &ClientStateManager{
		currentState: newGameState(),
		targetState:  nil,
		inputHistory: make(map[uint32]string),
		stateHistory: make(map[uint32]GameState),
	}
}

var _ nw.ClientStateManager[GameState] = &ClientStateManager{}

func (s *ClientStateManager) GetCurrent() GameState {
	return s.currentState
}

func (s *ClientStateManager) GetTarget() *GameState {
	return s.targetState
}

func (s *ClientStateManager) SetClientID(clientId string) {
	s.clientID = clientId
}
func (s *ClientStateManager) ClientID() string {
	return s.clientID
}

// ReconcileGameState reconciles the client's game state with the server's game state.
// It is called when the client receives a new server state message at the beginning of the frame
func (s *ClientStateManager) ReconcileState(serverMessage ServerStateMessage) {
	serverGameState := serverMessage.GameState
	acknowledgedSeq := serverMessage.AcknowledgedSeq[s.clientID]

	// Remove acknowledged inputs and states
	for seq := range s.inputHistory {
		if seq <= acknowledgedSeq {
			delete(s.inputHistory, seq)
			delete(s.stateHistory, seq)
		}
	}

	// Compare server state with local predicted state
	predictedState, exists := s.stateHistory[acknowledgedSeq]
	if !exists || !compareGameStates(s.clientID, serverGameState, predictedState) {
		// Discrepancy found, start interpolation
		s.targetState = &serverGameState
		s.interpolateUntil = time.Now().Add(interpolationTimeMs * time.Millisecond)
	}
}

// updateLocalGameState updates the client's local game state.
func (s *ClientStateManager) UpdateLocal(input string) {
	s.inputSequence++
	fmt.Printf("client id %s\nsequence %d\n", s.clientID, s.inputSequence)
	snake, exists := s.currentState.Snakes[s.clientID]
	if !exists {
		// Initialize snake if not exists
		snake = &Snake{
			ID:        s.clientID,
			Segments:  []Position{{X: 10, Y: 10}},
			Direction: "RIGHT",
		}
		s.currentState.Snakes[s.clientID] = snake
	}

	opposite := map[string]string{
		"UP":    "DOWN",
		"DOWN":  "UP",
		"LEFT":  "RIGHT",
		"RIGHT": "LEFT",
	}

	if input == "UP" || input == "DOWN" || input == "LEFT" || input == "RIGHT" {
		if snake.Direction != opposite[input] {
			snake.Direction = input
		}
	}
	s.inputHistory[s.inputSequence] = input
	s.stateHistory[s.inputSequence] = s.currentState
}

// updateGameState updates the local game state, including interpolation.
func (s *ClientStateManager) Update(dt float64) {
	// Interpolation
	if s.targetState != nil && time.Now().Before(s.interpolateUntil) {
		elapsed := float32(time.Since(s.interpolateUntil.Add(-interpolationTimeMs * time.Millisecond)).Milliseconds())
		factor := elapsed / float32(interpolationTimeMs)
		if factor > 1.0 {
			factor = 1.0
		} else if factor < 0.0 {
			factor = 0.0
		}
		s.currentState = s.interpolateStates(factor)
	} else if s.targetState != nil {
		s.currentState = *s.targetState
		s.targetState = nil
	}

	// Move the client's snake
	snake, exists := s.currentState.Snakes[s.clientID]
	fmt.Println("snakeExists: ", exists, s.clientID)
	if exists {
		moveSnake(snake, int(s.currentState.World.ToInt32().Width), int(s.currentState.World.ToInt32().Height), s.currentState.FoodItems, s.currentState.Snakes)
	}
}

func (s *ClientStateManager) interpolateStates(factor float32) GameState {

	interpolated := GameState{
		Snakes:    make(map[string]*Snake),
		FoodItems: s.targetState.FoodItems, // Assuming food positions don't need interpolation
	}

	for id, snake := range s.targetState.Snakes {
		currentSnake, exists := s.currentState.Snakes[id]
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

// lerp performs linear interpolation between two values.
func lerp(a, b, t float32) float32 {
	return a + t*(b-a)
}

// compareGameStates compares two game states for the client's snake.
func compareGameStates(clientID string, a, b GameState) bool {
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
