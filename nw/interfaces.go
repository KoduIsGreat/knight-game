package nw

type StateManager[T any] interface {
	Update(dt float64)
	ApplyInputToState(ci ClientInput)
	InitClientEntity(clientID string)
	RemoveClientEntity(clientID string)
	Get() T
}

type ClientStateManager[T any] interface {
	Update(dt float64)
	ReconcileState(msg ServerStateMessage[T])
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
	RecvFromServer() <-chan ServerStateMessage[T]
	QuitChan() <-chan struct{}
	Start()
	Promote(clientId string)
	KickFromLobby(clientId string)
	JoinLobby(lobbyId string)
	SyncLobbies()
	Lobby() *Lobby
	Lobbies() LobbiesSync
	IsStarted() bool
	LeaveLobby()
}
