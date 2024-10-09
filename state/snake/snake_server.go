package snake

import (
	"math/rand"

	. "github.com/KoduIsGreat/knight-game/common"
	"github.com/KoduIsGreat/knight-game/nw"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type ServerStateManager struct {
	state             GameState
	clientInputQueues map[string][]ClientInput
}

func NewServerStateManager() *ServerStateManager {
	world := rl.NewRectangle(0, 0, 600, 600)
	foodItems := generateRandomFoodItems(40, world)
	return &ServerStateManager{
		state: GameState{
			World:     world,
			Snakes:    make(map[string]*Snake),
			FoodItems: foodItems,
		},
		clientInputQueues: make(map[string][]ClientInput),
	}
}

var _ nw.StateManager[GameState] = &ServerStateManager{}

// generate food items within the world bounds
func generateRandomFoodItems(num int, world rl.Rectangle) []FoodItem {
	foodItems := make([]FoodItem, num)
	for i := 0; i < num; i++ {
		foodItems[i] = FoodItem{
			Position: Position{
				X: rand.Intn(int(world.Width) / 10),
				Y: rand.Intn(int(world.Height) / 10),
			},
		}
	}
	return foodItems
}

func (s *ServerStateManager) Update(dt float64) {
	updateGameState(s.state)
}

func (s *ServerStateManager) Get() GameState {
	return s.state
}

func (s *ServerStateManager) ApplyInputToState(ci ClientInput) {
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
	s.clientInputQueues[clientID] = make([]ClientInput, 0)

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

func updateGameState(gameState GameState) {
	for _, snake := range gameState.Snakes {
		moveSnake(snake, int(gameState.World.ToInt32().Width), int(gameState.World.ToInt32().Height), gameState.FoodItems, gameState.Snakes)
	}
}

// move snake but respect world bounds
// expand snake by adding a new tail if it eats food
func moveSnake(snake *Snake, worldWidth, worldHeight int, foodItems []FoodItem, allSnakes map[string]*Snake) {
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
		newHead.X = 600/10 - 1
	} else if newHead.X >= 600/10 {
		newHead.X = 0
	}

	if newHead.Y < 0 {
		newHead.Y = 600/10 - 1
	} else if newHead.Y >= 600/10 {
		newHead.Y = 0
	}
	// Check for self-collision
	if snakeCollidesWithSelf(snake, newHead) {
		snake.Segments = []Position{{
			X: 1 + rand.Intn(600/10-2),
			Y: 1 + rand.Intn(600/10-2),
		}}
		snake.Direction = "RIGHT"
		return
	}

	// Check for collision with other snakes
	for _, otherSnake := range allSnakes {
		if otherSnake.ID != snake.ID {
			if snakeCollidesWithOther(newHead, otherSnake) {
				if len(snake.Segments) > len(otherSnake.Segments) {
					// Eat the smaller snake
					snake.Segments = append(snake.Segments, otherSnake.Segments...)
					snake.Segments = []Position{{
						X: 1 + rand.Intn(600/10-2),
						Y: 1 + rand.Intn(600/10-2),
					}}
					snake.Direction = "RIGHT"
				} else {
					// Die and respawn
					snake.Segments = []Position{{
						X: 1 + rand.Intn(600/10-2),
						Y: 1 + rand.Intn(600/10-2),
					}}
					snake.Direction = "RIGHT"
					return
				}
			}
		}
	}

	// check if snake eats food
	for i, food := range foodItems {
		if newHead == food.Position {
			// add new tail
			snake.Segments = append(snake.Segments, snake.Segments[len(snake.Segments)-1])
			// remove food
			foodItems = append(foodItems[:i], foodItems[i+1:]...)
			break
		}
	}

	snake.Segments = append([]Position{newHead}, snake.Segments...)
	snake.Segments = snake.Segments[:len(snake.Segments)-1]
}

func snakeCollidesWithSelf(snake *Snake, newHead Position) bool {
	for _, segment := range snake.Segments[1:] {
		if newHead == segment {
			return true
		}
	}
	return false
}

func snakeCollidesWithOther(newHead Position, otherSnake *Snake) bool {
	for _, segment := range otherSnake.Segments {
		if newHead == segment {
			return true
		}
	}
	return false
}
