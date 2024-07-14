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
		i.p.ChangeState(PlayerStateRUNNING)
	}

}

type Running struct {
	p *Player
}

// Physics implements state.State.
func (r *Running) Physics(dt float64) {
	if rl.IsKeyDown(controls.MOVE_RIGHT) {
		r.p.body.Velocity.X = PlayerVelocity
	} else if rl.IsKeyDown(controls.MOVE_LEFT) {
		r.p.body.Velocity.X = -PlayerVelocity
	}
	if rl.IsKeyDown(controls.JUMP) && r.p.body.IsGrounded {
		r.p.ChangeState(PlayerStateJUMPING)
	}
}

// Update implements state.State.
func (r *Running) Update(dt float64) {
	if !rl.IsKeyPressed(controls.MOVE_LEFT) && !rl.IsKeyPressed(controls.MOVE_RIGHT) && r.p.body.IsGrounded {
		r.p.ChangeState(PlayerStateIDLE)
	}
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
	fmt.Println(j.p.body.Velocity.Y)
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
