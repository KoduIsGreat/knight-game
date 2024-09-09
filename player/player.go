package player

import (
	"log"

	"github.com/KoduIsGreat/knight-game/game"
	"github.com/KoduIsGreat/knight-game/sprite"
	"github.com/gen2brain/raylib-go/physics"
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/gofrs/uuid"
)

const PlayerVelocity = 0.5
const PlayerAnimationSpeed = 2

type Player struct {
	game.BaseComponent
	activeState PlayerState
	states      map[PlayerState]State
	body        *physics.Body
	camera      rl.Camera2D
	*sprite.Sprite2dAnimator
}

var _ game.Component = &Player{}

func (p *Player) ChangeState(ns PlayerState) {
	log.Println("Changing state to ", ns.String())
	p.activeState = ns
}

func NewPlayer() *Player {
	uid, _ := uuid.NewV4()
	player := &Player{
		BaseComponent: game.BaseComponent{
			GUID:     uid,
			Children: make([]game.Component, 0),
		},
	}
	player.Init()
	return player
}

func (p *Player) Init() {
	p.states = map[PlayerState]State{
		PlayerStateIDLE:    &Idle{p: p},
		PlayerStateRUNNING: &Running{p: p},
		PlayerStateJUMPING: &Jumping{p: p},
		// PlayerStateFALLING: &Falling{p: p},
	}
	p.activeState = PlayerStateIDLE
	sheet := rl.LoadTexture("./assets/KnightIdle_ss.png")
	playerPosition := rl.NewVector2(20, 10)
	p.camera = rl.NewCamera2D(rl.NewVector2(game.ScreenWidth/2, game.ScreenHeight/2), rl.NewVector2(playerPosition.X-(float32(sheet.Width)/2), playerPosition.Y-(float32(sheet.Height)/2)), 0, 1)
	p.body = physics.NewBodyRectangle(playerPosition, float32(sheet.Width), float32(sheet.Height), .3)
	p.body.UseGravity = false
	p.Sprite2dAnimator = sprite.NewSprite2dAnimator(sheet, 2, &p.body.Position)
}

func (p *Player) Update(dt float64) {
	if rl.IsKeyPressed(rl.KeyDelete) {
		p.ChangeState(PlayerStateJUMPING)
	}
	if rl.IsKeyDown(rl.KeyA) || rl.IsKeyDown(rl.KeyLeft) {
		p.ChangeState(PlayerStateRUNNING)
	}
	p.states[p.activeState].Update(dt)
	p.camera.Target = rl.NewVector2(p.body.Position.X-(float32(p.Sprite2dAnimator.SpriteSheet.Width)/p.Sprite2dAnimator.NumFrames), p.body.Position.Y-(float32(p.SpriteSheet.Height)/p.Sprite2dAnimator.NumFrames))
	p.Sprite2dAnimator.Update(dt)
	for _, child := range p.Children {
		child.Update(dt)
	}
}

func (p *Player) Render() {
	p.Sprite2dAnimator.Render()
	for _, child := range p.Children {
		child.Render()
	}
}

func (p *Player) Physics(dt float64) {
	p.states[p.activeState].Physics(dt)
	for _, child := range p.Children {
		child.Physics(dt)
	}
}

func (p *Player) Destroy() {
	p.Sprite2dAnimator.Destroy()
	p.body.Destroy()
}

func (p *Player) AddChild(gc game.Component) {
	p.Children = append(p.Children, gc)
}

func (p *Player) GetChildren() []game.Component {
	return p.Children
}
