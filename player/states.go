package player

import (
	"fmt"

	"github.com/KoduIsGreat/knight-game/controls"
	. "github.com/KoduIsGreat/knight-game/state"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Idle struct {
	p *Player
}

// Physics implements state.Machine.
func (i *Idle) Physics(dt float64) {
	return
}

// Update implements state.Machine.
func (i *Idle) Update(dt float64) {
	if rl.IsKeyPressed(rl.KeyD) || rl.IsKeyPressed(rl.KeyA) {
		fmt.Println("Key is pressed ")
		i.p.ChangeState(PlayerStateRUNNING)
	}
	if rl.IsKeyPressed((rl.KeyDelete)) {
		i.p.ChangeState(PlayerStateJUMPING)
	}
}

type Running struct {
	p *Player
}

// Physics implements state.State.
func (r *Running) Physics(dt float64) {
	fmt.Println("velocity ")
	if rl.IsKeyDown(rl.KeyD) {
		r.p.body.Velocity.X = PlayerVelocity
		r.p.body.Position.X += r.p.body.Velocity.X
		fmt.Println("velocity ", r.p.body.Velocity.X)
		fmt.Println("position ", r.p.body.Position.X)
	} else if rl.IsKeyDown(rl.KeyA) {
		r.p.body.Velocity.X = -PlayerVelocity
		r.p.body.Position.X -= r.p.body.Velocity.X
	}
	if rl.IsKeyDown(rl.KeyDelete) && r.p.body.IsGrounded {
		r.p.ChangeState(PlayerStateJUMPING)
	}
}

// Update implements state.State.
func (r *Running) Update(dt float64) {
	stateToChangeTo := PlayerStateIDLE
	if rl.IsKeyPressed(rl.KeyD) || rl.IsKeyPressed(rl.KeyA) {
		stateToChangeTo = PlayerStateRUNNING
	}
	r.p.ChangeState(stateToChangeTo)
}

type Jumping struct {
	p *Player
}

// Physics implements state.State.
func (j *Jumping) Physics(dt float64) {
	if rl.IsKeyDown(controls.MOVE_RIGHT) {
		j.p.body.Velocity.X = PlayerVelocity * .3
	} else if rl.IsKeyDown(controls.MOVE_LEFT) {
		j.p.body.Velocity.X = -PlayerVelocity * .3
	}
	j.p.body.Velocity.Y = -PlayerVelocity * 4
}

// Update implements state.State.
func (j *Jumping) Update(dt float64) {
}

type Falling struct {
	p *Player
}

// Physics implements state.State.
func (f *Falling) Physics(dt float64) {
	if f.p.body.IsGrounded {
		f.p.ChangeState(PlayerStateIDLE)
	}
}

// Update implements state.State.
func (f *Falling) Update(dt float64) {
}

var _ State = &Idle{}
var _ State = &Running{}
var _ State = &Jumping{}
var _ State = &Falling{}
