package nw

import (
	"fmt"
	"testing"
)

func TestMsgPack(t *testing.T) {
	tests := []struct {
		name string
		f    func() Message
		want string
	}{

		{
			name: "auth ack message",
			f: func() Message {
				return NewAuthAckMessage(FmtText, "client1")
			},
			want: "Message{header: MsgAuthAck, fmt: FmtText, size: 7, data: client1}",
		},
		{
			name: "connect message",
			f: func() Message {
				return NewConnectMessage(FmtText, "client1")
			},
			want: "Message{header: MsgConnect, fmt: FmtText, size: 7, data: client1}",
		},
		{
			name: "disconnect message",
			f: func() Message {
				return NewDisconnectMessage(FmtText, "client1")
			},
			want: "Message{header: MsgDisconnect, fmt: FmtText, size: 7, data: client1}",
		},
		{
			name: "lobby create message",
			f: func() Message {
				return NewLobbyCreateMessage(FmtText, "client1")
			},
			want: "Message{header: MsgLobbyCreate, fmt: FmtText, size: 7, data: client1}",
		},
		{
			name: "lobby created message",
			f: func() Message {
				return NewLobbyCreatedMessage(FmtText, "lobby1")
			},
			want: "Message{header: MsgLobbyCreated, fmt: FmtText, size: 6, data: lobby1}",
		},
		{
			name: "lobby join message",
			f: func() Message {
				return NewLobbyJoinMessage(FmtText, "lobby1", "client1")
			},
			want: "Message{header: MsgLobbyClientJoin, fmt: FmtText, size: 14, data: lobby1|client1}",
		},
		{
			name: "lobby leave message",
			f: func() Message {
				return NewLobbyLeaveMessage(FmtText, "lobby1", "client1")
			},
			want: "Message{header: MsgLobbyClientLeave, fmt: FmtText, size: 14, data: lobby1|client1}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.f()
			m.Pack()
			gMsg := m.String()

			wMsg := tt.want
			if gMsg != wMsg {
				t.Errorf("got %s, want %s", gMsg, wMsg)
			}
		})
	}
}

func TestMsgPackUnpack(t *testing.T) {
	msg := NewAuthAckMessage(FmtText, "client1")
	bytes := msg.Pack()
	var unpacked Message
	unpacked.Unpack(bytes)

	fmt.Println("Unpack data", unpacked.data.Size, unpacked.data.Fmt, unpacked.data.Data, msg.data.Data)
	if unpacked.header != msg.header {
		t.Errorf("\ngot %d\nwant %d", unpacked.header, msg.header)
	}
	if unpacked.data.Fmt != msg.data.Fmt {
		t.Errorf("\ngot %d\nwant %d", unpacked.data.Fmt, msg.data.Fmt)
	}
	if unpacked.data.Size != msg.data.Size {
		t.Errorf("\ngot %d\nwant %d", unpacked.data.Size, msg.data.Size)
	}
	if string(unpacked.data.Data) != string(msg.data.Data) {
		t.Errorf("\ngot %s\nwant %s", unpacked.data.Data, msg.data.Data)
	}
}

func TestMsgUnpack(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want Message
	}{
		{
			name: "empty message",
			msg:  ``,
			want: Message{
				header: MsgAuth,
				data: messageData{
					Size: 0,
					Fmt:  FmtText,
					Data: []byte{},
				},
			},
		},
		{
			name: "simple message",
			msg:  `01000007636c69656e740a`,
			want: Message{
				header: MsgAuthAck,
				data: messageData{
					Size: 7,
					Fmt:  FmtText,
					Data: []byte("client1"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Message
			m.Unpack([]byte(tt.msg))
			gMsg := m.String()

			wMsg := tt.want.String()
			if gMsg != wMsg {
				t.Errorf("got %s, want %s", gMsg, wMsg)
			}
		})
	}
}
