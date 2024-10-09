package common

import "io"

type ClientInput struct {
	ClientID string
	Input    string
	Sequence uint32
}

type ServerStateMessage struct {
	GameState       GameState         `json:"gameState"`
	AcknowledgedSeq map[string]uint32 `json:"acknowledgedSeq"`
}

const (
	NetworkMessageJoin  byte = 'J'
	NetworkMessageState byte = 'S'
)

type NetworkMessageEncoder interface {
	Encode() []byte
	EncodeTo(io.Writer) error
}

type NetworkMessageDecoder interface {
	Decode([]byte, any) error
	DecodeFrom(io.Reader, any) error
}
