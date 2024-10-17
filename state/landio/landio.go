package landio

// Tile represents a single tile in the game world.

type Tile struct {
	TopLeft     Position
	BottomRight Position
}

func (t Tile) Contains(p Position) bool {
	return p.X >= t.TopLeft.X && p.X <= t.BottomRight.X && p.Y >= t.TopLeft.Y && p.Y <= t.BottomRight.Y
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Direction byte

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Player struct {
	Pos       Position
	Direction Direction
	CurrTile  *Tile
}

type World struct {
	Tiles   []Tile
	Players map[string]Player
}

func NewWorld() *World {
	return &World{
		Tiles:   []Tile{},
		Players: make(map[string]Player),
	}
}
