package snake

import (
	"sync"
	"time"

	. "github.com/KoduIsGreat/knight-game/common"
)

const (
	interpolationTimeMs = 100 // Time in milliseconds for interpolation
)

type ClientStateManager struct {
	clientID         string
	currentState     GameState
	targetState      GameState
	gameStateMutex   sync.Mutex
	inputHistory     map[uint32]string
	stateHistory     map[uint32]GameState
	interpolateUntil time.Time
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

func (s *ClientStateManager) ReconcileGameState(serverMessage ServerStateMessage) {
	s.gameStateMutex.Lock()
	defer s.gameStateMutex.Unlock()

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
		s.targetState = serverGameState
		s.interpolateUntil = time.Now().Add(interpolationTimeMs * time.Millisecond)
	}
}

// updateLocalGameState updates the client's local game state.
func (s *ClientStateManager) UpdateLocalGameState(input string) {
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

	if input != opposite[snake.Direction] {
		snake.Direction = input
	}
}

// updateGameState updates the local game state, including interpolation.
func (s *ClientStateManager) updateGameState() {
	s.gameStateMutex.Lock()
	defer s.gameStateMutex.Unlock()

	// Interpolation
	if !s.interpolateUntil.IsZero() && time.Now().Before(s.interpolateUntil) {
		factor := float32(interpolationTimeMs-(s.interpolateUntil.Sub(time.Now()).Milliseconds())) / float32(interpolationTimeMs)
		s.currentState = s.interpolateStates(factor)
	} else if !s.interpolateUntil.IsZero() {
		s.currentState = s.targetState
		s.interpolateUntil = time.Time{}
	}

	// Move the client's snake
	snake, exists := s.currentState.Snakes[s.clientID]
	if exists {
		moveSnake(snake, int(s.currentState.World.ToInt32().Width), int(s.currentState.World.ToInt32().Height), s.currentState.FoodItems)
	}
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
