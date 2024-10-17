package snake

import (
	"github.com/KoduIsGreat/knight-game/nw"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ServerStateManager struct {
	state             GameState
	clientInputQueues map[string][]nw.ClientInput
}

func NewServerStateManager() *ServerStateManager {
	world := rl.NewRectangle(0, 0, 1000, 1000)
	foodItems := spawnFoodItems(80, world)
	return &ServerStateManager{
		state: GameState{
			World:     world,
			Snakes:    make(map[string]*Snake),
			FoodItems: foodItems,
		},
		clientInputQueues: make(map[string][]nw.ClientInput),
	}
}

var _ nw.StateManager[GameState] = &ServerStateManager{}

func (s *ServerStateManager) Update(dt float64) {
	updateGameState(s.state)
}

func (s *ServerStateManager) Get() GameState {
	return s.state
}

func (s *ServerStateManager) ApplyInputToState(ci nw.ClientInput) {
	snake, exists := s.state.Snakes[ci.ClientID]
	if !exists {
		return
	}

	oppositeDirections := map[string]string{
		"UP":    "DOWN",
		"DOWN":  "UP",
		"LEFT":  "RIGHT",
		"RIGHT": "LEFT",
	}
	input := ci.Input

	if snake.Direction != oppositeDirections[input] {
		snake.Direction = input
	}
}

func (s *ServerStateManager) InitClientEntity(clientID string) {
	s.clientInputQueues[clientID] = make([]nw.ClientInput, 0)
	s.state.Snakes[clientID] = &Snake{
		ID:        clientID,
		Segments:  []Position{{X: 0, Y: 0}},
		Direction: "RIGHT",
	}
}
func (s *ServerStateManager) RemoveClientEntity(clientId string) {
	delete(s.state.Snakes, clientId)
	delete(s.clientInputQueues, clientId)
}
