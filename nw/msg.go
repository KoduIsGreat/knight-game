package nw

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
)

const (
	MaxMessageSize = 1024
)

const (
	MsgEnd  uint8 = '\n'
	MsgSept uint8 = '|'
)

type messageData struct {
	Size uint16
	Fmt  MessageFmt
	Data []byte
}

type Message struct {
	header MessageHeader
	data   messageData
}

func (m Message) String() string {
	return fmt.Sprintf("Message{header: %s, fmt: %s, size: %d, data: %s}", m.header.String(), m.data.Fmt.String(), m.data.Size, m.data.Data)
}

func (m Message) HexString() string {
	return hex.EncodeToString(m.Pack())
}

func (m *Message) Pack() []byte {
	buf := make([]byte, 3+len(m.data.Data))
	buf[0] = byte(m.header)
	buf[1] = byte(m.data.Fmt)
	buf[2] = byte(m.data.Size >> 8)
	buf[3] = byte(m.data.Size)
	copy(buf[4:], m.data.Data)
	return append(buf, MsgEnd)
}

func (m *Message) Unpack(buf []byte) error {
	if len(buf) < 4 {
		return fmt.Errorf("invalid message buffer")
	}
	m.header = MessageHeader(buf[0])
	m.data.Fmt = MessageFmt(buf[1])
	m.data.Size = uint16(buf[2])<<8 | uint16(buf[3])
	m.data.Data = make([]byte, m.data.Size)
	fmt.Println("Unpack data", m.data.Size, m.data.Fmt, m.data.Data, buf[4:])
	copy(m.data.Data, buf[4:])

	fmt.Println("Unpack data", m.data.Size, m.data.Fmt, m.data.Data, buf[4:])
	return nil
}

type MessageHandler interface {
	Handle(Message) error
}

type MessageHandlerFunc func(Message) error

func (f MessageHandlerFunc) Handle(m Message) error {
	return f(m)
}

func (m Message) EncodeTo(w io.Writer) error {
	if _, err := w.Write(m.Pack()); err != nil {
		return err
	}
	return nil
}
func (m *Message) DecodeFrom(r io.Reader) error {
	buf := make([]byte, MaxMessageSize)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	m.Unpack(buf[:n])
	return nil
}

func NewMessage(header MessageHeader, fmt MessageFmt, data []byte) Message {
	return Message{
		header: header,
		data: messageData{
			Size: uint16(len(data)),
			Fmt:  fmt,
			Data: data,
		},
	}
}

type NetworkMessageEncoder interface {
	EncodeTo(io.Writer) error
}

type NetworkMessageDecoder interface {
	DecodeFrom(io.Reader, any) error
}

func NewAuthMessage(fmt MessageFmt, clientID string, token []byte) Message {
	data := append([]byte(clientID), token...)
	return NewMessage(MsgAuth, fmt, data)
}

func NewAuthAckMessage(fmt MessageFmt, clientID string) Message {
	return NewMessage(MsgAuthAck, fmt, []byte(clientID))
}

func NewConnectMessage(fmt MessageFmt, clientID string) Message {
	return NewMessage(MsgConnect, fmt, []byte(clientID))
}

func NewDisconnectMessage(fmt MessageFmt, clientID string) Message {
	return NewMessage(MsgDisconnect, fmt, []byte(clientID))
}

func NewLobbyCreateMessage(fmt MessageFmt, clientId string) Message {
	return NewMessage(MsgLobbyCreate, fmt, []byte(clientId))
}

func NewLobbyCreatedMessage(fmt MessageFmt, lobbyID string) Message {
	return NewMessage(MsgLobbyCreated, fmt, []byte(lobbyID))
}

func NewLobbyJoinMessage(f MessageFmt, lobbyID, clientId string) Message {
	msg := fmt.Sprintf("%s|%s", lobbyID, clientId)
	switch f {
	case FmtJSON:
		data, _ := json.Marshal(lobbyMsg{LobbyID: lobbyID, ClientID: clientId})
		return NewMessage(MsgLobbyClientJoin, f, data)
	case FmtText:
		return NewMessage(MsgLobbyClientJoin, f, []byte(msg))
	case FmtBinary:
		return NewMessage(MsgLobbyClientJoin, f, []byte(msg))
	}
	return NewMessage(MsgLobbyClientJoin, f, []byte(msg))
}

type lobbyMsg struct {
	LobbyID  string `json:"lobbyId"`
	ClientID string `json:"clientId"`
}

func NewLobbyLeaveMessage(f MessageFmt, lobbyID, clientId string) Message {
	msg := fmt.Sprintf("%s|%s", lobbyID, clientId)
	switch f {
	case FmtJSON:
		data, _ := json.Marshal(lobbyMsg{LobbyID: lobbyID, ClientID: clientId})
		return NewMessage(MsgLobbyClientLeave, f, data)
	case FmtText:
		return NewMessage(MsgLobbyClientLeave, f, []byte(msg))
	case FmtBinary:
		return NewMessage(MsgLobbyClientLeave, f, []byte(msg))
	}
	return NewMessage(MsgLobbyClientLeave, f, []byte(msg))
}

func NewGameStateMessage(f MessageFmt, state any) (Message, error) {
	var data []byte
	switch f {
	case FmtJSON:
		var err error
		data, err = json.Marshal(state)
		if err != nil {
			return Message{}, err
		}
	default:
		return Message{}, fmt.Errorf("unsupported message format")
	}

	return NewMessage(MsgServerState, f, data), nil
}

func NewClientInputMessage(f MessageFmt, ci ClientInput) (Message, error) {
	var data []byte
	switch f {
	case FmtJSON:
		var err error
		data, err = json.Marshal(ci)
		if err != nil {
			return Message{}, err
		}
	default:
		return Message{}, fmt.Errorf("unsupported message format")
	}

	return NewMessage(MsgClientInput, f, data), nil

}
