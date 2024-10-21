package nw

import (
	"fmt"
	"io"
)

type MessageHeader uint8

const (
	MsgAuth MessageHeader = iota
	MsgAuthAck
	MsgConnect
	MsgDisconnect
	MsgLobbyCreate
	MsgLobbyCreated
	MsgLobbyClientReady
	MsgLobbyClientJoin
	MsgLobbyClientLeave
	MsgClientInput
	MsgServerState
)

const (
	MsgEnd  uint8 = '\n'
	MsgSept uint8 = '|'
)

type MessageFmt uint8

const (
	FmtText MessageFmt = iota
	FmtBinary
	FmtJSON
)

type messageData struct {
	Size uint16
	Fmt  uint8
	Data []byte
}

type Message struct {
	header MessageHeader
	data   messageData
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
	m.data.Fmt = buf[1]
	m.data.Size = uint16(buf[2])<<8 | uint16(buf[3])
	m.data.Data = make([]byte, m.data.Size)
	copy(m.data.Data, buf[4:])
	return nil
}

func (m *Message) EncodeTo(w io.Writer) error {
	if _, err := w.Write(m.Pack()); err != nil {
		return err
	}
	return nil
}
func (m *Message) DecodeFrom(r io.Reader) error {
	buf := make([]byte, 1024)
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
			Fmt:  uint8(fmt),
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

func NewLobbyJoinMessage(fmt MessageFmt, lobbyID, clientId string) Message {
	return NewMessage(MsgLobbyClientJoin, fmt, []byte(lobbyID+clientId))
}

func NewLobbyLeaveMessage(fmt MessageFmt, lobbyID, clientId string) Message {
	return NewMessage(MsgLobbyClientLeave, fmt, []byte(lobbyID+clientId))
}
