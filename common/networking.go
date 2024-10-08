package common

type ClientInput struct {
	ClientID string
	Input    string
	Sequence uint32
}

type ServerStateMessage struct {
	GameState       GameState         `json:"gameState"`
	AcknowledgedSeq map[string]uint32 `json:"acknowledgedSeq"`
}

type ServerMessage struct {
	Type string `json:"type"`
}
