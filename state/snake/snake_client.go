package snake

import (
	"encoding/json"
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
func (s *ClientStateManager) InputSeq() uint32 {
	return s.inputSequence
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
	// Always set the target state and start interpolation
	s.targetState = &serverGameState
	s.interpolateUntil = time.Now().Add(interpolationTimeMs * time.Millisecond)
}

// updateLocalGameState updates the client's local game state.
func (s *ClientStateManager) UpdateLocal(input string) {
	s.inputSequence++
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

func jsonPrettyState(gs any) []byte {
	b, _ := json.MarshalIndent(gs, "", "  ")
	return b
}

func (s *ClientStateManager) Update(dt float64) {
	// Continuous interpolation
	if s.targetState != nil {
		now := time.Now()
		if now.Before(s.interpolateUntil) {
			elapsed := float32(now.Sub(s.interpolateUntil.Add(-interpolationTimeMs * time.Millisecond)).Seconds())
			factor := elapsed / (float32(interpolationTimeMs) / 1000.0)
			factor = clamp(factor, 0.0, 1.0)
			s.currentState = s.interpolateStates(factor)
		} else {
			s.currentState = *s.targetState
			s.targetState = nil
		}
	}

	// Move the client's snake
	snake, exists := s.currentState.Snakes[s.clientID]
	if exists {
		moveSnake(snake, int(s.currentState.World.ToInt32().Width), int(s.currentState.World.ToInt32().Height), s.currentState.FoodItems, s.currentState.Snakes)
	}
}

func clamp(value, min, max float32) float32 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (s *ClientStateManager) interpolateStates(factor float32) GameState {
	interpolated := GameState{
		Snakes:    make(map[string]*Snake),
		FoodItems: s.targetState.FoodItems,
		World:     s.targetState.World,
	}

	worldWidth := 600
	worldHeight := 600

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

			targetPos := snake.Segments[i]

			interpolatedX := interpolateCoordinate(currentPos.X, targetPos.X, factor, worldWidth)
			interpolatedY := interpolateCoordinate(currentPos.Y, targetPos.Y, factor, worldHeight)

			interpolatedSnake.Segments[i] = Position{
				X: interpolatedX,
				Y: interpolatedY,
			}
		}

		interpolated.Snakes[id] = interpolatedSnake
	}

	return interpolated
}

func interpolateCoordinate(current, target int, factor float32, worldSize int) int {
	diff := target - current
	if abs(diff) > worldSize/2 {
		if diff > 0 {
			diff -= worldSize
		} else {
			diff += worldSize
		}
	}

	interpolated := int(lerp(float32(current), float32(current+diff), factor))
	return (interpolated + worldSize) % worldSize
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// lerp performs linear interpolation between two values.
func lerp(a, b, t float32) float32 {
	return a + t*(b-a)
}
