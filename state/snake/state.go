package snake

import (
	"math/rand"

	rl "github.com/gen2brain/raylib-go/raylib"
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
	World     rl.Rectangle
}

// generate food items within the world bounds
// ensure food items do not overlap with other food items
func spawnFoodItems(num int, world rl.Rectangle) []FoodItem {
	foodItems := make([]FoodItem, 0, num)
	occupied := make(map[Position]bool)

	for len(foodItems) < num {
		position := Position{
			X: rand.Intn(int(world.Width) / 10),
			Y: rand.Intn(int(world.Height) / 10),
		}

		if !occupied[position] {
			foodItems = append(foodItems, FoodItem{Position: position})
			occupied[position] = true
		}
	}

	return foodItems
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
		newHead.X = worldWidth/10 - 1
	} else if newHead.X >= worldWidth/10 {
		newHead.X = 0
	}

	if newHead.Y < 0 {
		newHead.Y = worldHeight/10 - 1
	} else if newHead.Y >= worldHeight/10 {
		newHead.Y = 0
	}
	// Check for self-collision
	if snakeCollidesWithSelf(snake, newHead) {
		respawnSnake(snake, worldWidth, worldHeight)
		return
	}

	// Check for collision with other snakes
	for _, otherSnake := range allSnakes {
		if otherSnake.ID != snake.ID {
			if snakeCollidesWithOther(newHead, otherSnake) {
				if len(snake.Segments) > len(otherSnake.Segments) {
					// Eat the smaller snake
					snake.Segments = append(snake.Segments, otherSnake.Segments...)
					respawnSnake(otherSnake, worldWidth, worldHeight)
				} else {
					// Die and respawn
					respawnSnake(snake, worldWidth, worldHeight)
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

func respawnSnake(snake *Snake, worldWidth, worldHeight int) {
	snake.Segments = []Position{{
		X: 1 + rand.Intn(worldWidth/10-2),
		Y: 1 + rand.Intn(worldHeight/10-2),
	}}
	snake.Direction = "RIGHT"
}
