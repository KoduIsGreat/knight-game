package main

import "fmt"

type MessageHeader uint8

const (
	MsgConnect MessageHeader = iota
	MsgCreateLobby
	MsgJoinLobby
	MsgDeleteLobby
	MsgAuth
	MsgDisconnect
	MsgInput
)

func main() {
	fmt.Printf("%d, %b\n", MsgConnect, MsgConnect)
	fmt.Printf("%d, %b\n", MsgAuth, MsgAuth)
	fmt.Printf("%d, %b\n", MsgDisconnect, MsgDisconnect)
	fmt.Printf("%d, %b\n", MsgInput, MsgInput)
}

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
	return buf
}

func (m *Message) Unpack(buf []byte) {
	m.header = MessageHeader(buf[0])
	m.data.Fmt = buf[1]
	m.data.Size = uint16(buf[2])<<8 | uint16(buf[3])
	m.data.Data = make([]byte, m.data.Size)
	copy(m.data.Data, buf[4:])
}
