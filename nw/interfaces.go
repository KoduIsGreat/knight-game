package nw

import "github.com/KoduIsGreat/knight-game/common"

type StateManager[T any] interface {
	Update(dt float64)
	ApplyInputToState(ci common.ClientInput)
	InitClientEntity(clientID string)
	RemoveClientEntity(clientID string)
	Get() T
}

type ClientStateManager[T any] interface {
	Update(dt float64)
	ReconcileState(msg common.ServerStateMessage)
	UpdateLocal(input string)
	InputSeq() uint32
	GetCurrent() T
	GetTarget() *T
	SetClientID(string)
	ClientID() string
}

type GameClient[T any] interface {
	SendInputToServer(input string)
	State() ClientStateManager[T]
	ClientID() string
	RecvFromServer() <-chan common.ServerStateMessage
	QuitChan() <-chan struct{}
}
