package nw

import "github.com/KoduIsGreat/knight-game/common"

type StateManager interface {
	Update(dt float64)
	ApplyInputToState(ci common.ClientInput)
	InitClientEntity(clientID string)
	RemoveClientEntity(clientID string)
	Get() common.GameState
}

type ClientStateManager interface {
	Update(dt float64)
	ReconcileState(msg common.ServerStateMessage)
}
