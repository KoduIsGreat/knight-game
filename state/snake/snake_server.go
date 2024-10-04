package snake

import (
	"math/rand"

	. "github.com/KoduIsGreat/knight-game/common"
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
		state: GameState{World: world,
			Snakes:    make(map[string]*Snake),
			FoodItems: foodItems,
		},
		clientInputQueues: make(map[string][]ClientInput),
	}
}

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

	if input == "UP" || input == "DOWN" || input == "LEFT" || input == "RIGHT" {
		if snake.Direction != oppositeDirections[input] {
			snake.Direction = input
		}
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
		moveSnake(snake, int(gameState.World.ToInt32().Width), int(gameState.World.ToInt32().Height), gameState.FoodItems)
	}
}

// move snake but respect world bounds
// expand snake by adding a new tail if it eats food
func moveSnake(snake *Snake, worldWidth, worldHeight int, foodItems []FoodItem) {
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
