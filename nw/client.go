package nw

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	quic "github.com/quic-go/quic-go"
)

type Client[T any] struct {
	stream quic.Stream
	// sendChan is used to send messages to the server
	sendChan chan Message
	// recvChan is used to receive messages from the server
	// gameStateChan is used to game state from the server
	gameStateChan chan ServerStateMessage[T]
	// quitChan is used to signal the network handlers to stop
	quitChan chan struct{}
	// state is the client's state manager
	state   ClientStateManager[T]
	Lobbies LobbiesSync
	lobby   *Lobby
	// clientID is the client's ID determined by the server
	clientID string
}

type Lobby struct {
	ID               string
	OwnerClientID    string `json:"ownerClientID"`
	ReadyClients     map[string]bool
	ConnectedClients map[string]otherClient
	MaxPlayers       int
	Started          bool `json:"started"`
	Countdown        int
}
type otherClient struct {
}

type ClientOpts struct {
	ServerAddress string
	TLSConfig     *tls.Config
	QuicConfig    *quic.Config
}

// NewClient creates a new client with the given state manager.
func NewClient[T any](state ClientStateManager[T], co ClientOpts) *Client[T] {
	c := &Client[T]{
		sendChan:      make(chan Message),
		gameStateChan: make(chan ServerStateMessage[T]),
		quitChan:      make(chan struct{}),
		state:         state,
	}
	c.connectToServer(co)
	c.waitUntilConnected()
	c.startNetworkHandlers()
	go func() {
		for {

			c.SyncLobbies()
			time.Sleep(time.Second * 5)
		}
	}()
	return c
}

func (c *Client[T]) State() ClientStateManager[T] {
	return c.state
}

func (c *Client[T]) Lobby() *Lobby {
	return c.lobby
}

func (c *Client[T]) makeClientInputMessage(input string) (Message, error) {
	ci := ClientInput{
		ClientID: c.clientID,
		Sequence: c.state.InputSeq(),
		Input:    input,
	}
	return NewClientInputMessage(FmtJSON, ci)
}

func (c *Client[T]) SendInputToServer(input string) {
	fmt.Println("Sending input:", input)
	msg, err := c.makeClientInputMessage(input)
	if err != nil {
		log.Println("Error creating client input message:", err)
	}
	c.sendChan <- msg
}

func (c *Client[T]) CreateLobby() {
	msg := NewMessage(MsgLobbyCreate, FmtText, []byte{})
	c.sendChan <- msg
}

func (c *Client[T]) SyncLobbies() {
	msg := NewMessage(MsgLobbiesSync, FmtText, []byte{})
	fmt.Println("Syncing lobbies...")
	c.sendChan <- msg
}

func (c *Client[T]) JoinLobby(lobbyID string) {
	data := fmt.Sprintf("%s|%s", lobbyID, c.clientID)
	msg := NewMessage(MsgLobbyClientJoin, FmtText, []byte(data))
	fmt.Println("Joining lobby:", lobbyID)
	c.sendChan <- msg
}

func (c *Client[T]) Start() {
	msg := NewMessage(MsgLobbyGameStart, FmtText, []byte{})
	fmt.Println("Starting game...")
	c.sendChan <- msg
}

func (c *Client[T]) LeaveLobby() {
	msg := NewMessage(MsgLobbyClientLeave, FmtText, []byte(c.clientID))
	fmt.Println("Leaving lobby:", c.lobby.ID)
	c.sendChan <- msg
}

func (c *Client[T]) KickFromLobby(clientID string) {
	msg := NewMessage(MsgLobbyKick, FmtText, []byte(clientID))
	fmt.Println("Kicking client from lobby:", clientID)
	c.sendChan <- msg
}

func (c Client[T]) IsStarted() bool {
	if c.lobby == nil {
		return false
	}
	return c.lobby.Started
}

func (c *Client[T]) Promote(clientID string) {
	msg := NewMessage(MsgLobbyPromote, FmtText, []byte(clientID))
	fmt.Println("Promoting client to host:", clientID)
	c.sendChan <- msg
}

func (c *Client[T]) RecvFromServer() <-chan ServerStateMessage[T] {
	return c.gameStateChan
}

func (c *Client[T]) QuitChan() <-chan struct{} {
	return c.quitChan
}

func (c *Client[T]) ClientID() string {
	return c.clientID
}

// connectToServer establishes a QUIC connection to the server.
func (c *Client[T]) connectToServer(co ClientOpts) {
	if co.ServerAddress == "" {
		co.ServerAddress = "localhost:4242"
	}
	if co.TLSConfig == nil {
		co.TLSConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"snake-game"},
		}
	}
	session, err := quic.DialAddr(context.Background(), co.ServerAddress, co.TLSConfig, nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}

	c.stream, err = session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal("Failed to open stream:", err)
	}
	msg := NewMessage(MsgConnect, FmtText, []byte{})
	if err := msg.EncodeTo(c.stream); err != nil {
		log.Fatal("Failed to send connect message:", err)
	}
}

func (c *Client[T]) writer() {
	for {
		select {
		case msg := <-c.sendChan:
			if err := msg.EncodeTo(c.stream); err != nil {
				log.Println("Error sending message to server:", err)
				return
			}
		case <-c.quitChan:
			return
		}
	}
}

func (c *Client[T]) reader(mh MessageHandler) {
	defer func() {
		c.quitChan <- struct{}{}
	}()
	reader := bufio.NewReader(c.stream)
	for {
		var msg Message
		if err := msg.DecodeFrom(reader); err != nil {
			log.Println("error decoding message:", err)
			return
		}
		mh.Handle(msg)
	}
}

func (c *Client[T]) waitUntilConnected() {
	fmt.Println("Waiting for client ID...")
	reader := bufio.NewReader(c.stream)
	for {
		if c.clientID != "" {
			return
		}
		var connectMsg Message
		if err := connectMsg.DecodeFrom(reader); err != nil {
			log.Println("Error decoding connect message:", err)
			return
		}
		clientId := string(connectMsg.data.Data)
		fmt.Println("Received client ID:", clientId)
		c.clientID = clientId
		c.state.SetClientID(c.clientID)
	}
}

func (c *Client[T]) startNetworkHandlers() {
	go c.writer()
	go c.reader(MessageHandlerFunc(func(msg Message) error {
		switch msg.header {
		case MsgLobbyCreated:
			lobbyID := string(msg.data.Data)
			fmt.Println("Lobby created:", lobbyID)
			c.lobby = &Lobby{
				ID:               lobbyID,
				ReadyClients:     make(map[string]bool),
				ConnectedClients: make(map[string]otherClient),
				Countdown:        10,
			}
		case MsgLobbiesSynced:
			fmt.Println("Lobbies synced")
			if err := json.Unmarshal(msg.data.Data, &c.Lobbies); err != nil {
				fmt.Println("Error decoding lobbies sync message:", err)
			}
		case MsgServerState:
			msg, err := ServerStateMessageFromMessage[T](msg)
			if err != nil {
				fmt.Println("Error decoding server state message:", err)
			}
			c.gameStateChan <- msg
		case MsgLobbyClientJoin:
			clientID := string(msg.data.Data)
			fmt.Println("Client joined lobby:", clientID)
			if c.lobby == nil {
				c.lobby = &Lobby{
					OwnerClientID:    c.clientID,
					ReadyClients:     make(map[string]bool),
					ConnectedClients: make(map[string]otherClient),
					Countdown:        10,
				}
			}
			c.lobby.ConnectedClients[clientID] = otherClient{}
			c.lobby.ReadyClients[clientID] = false
		case MsgLobbyPromote:
			clientID := string(msg.data.Data)
			fmt.Println("Client promoted to host:", clientID)
			c.lobby.OwnerClientID = clientID
		case MsgLobbyClientLeave, MsgLobbyKick:
			clientID := string(msg.data.Data)
			fmt.Println("Client left lobby:", clientID)
			delete(c.lobby.ConnectedClients, clientID)
			if c.clientID == clientID {
				c.lobby = nil
			}
		case MsgLobbyClientReady:
			clientID := string(msg.data.Data)
			fmt.Println("Client ready:", clientID)
			c.lobby.ReadyClients[clientID] = !c.lobby.ReadyClients[clientID]
		case MsgLobbyGameStarted:
			type countdownMsg struct {
				Countdown int `json:"countdown"`
			}
			switch msg.data.Fmt {
			case FmtText:
				c.lobby.Started = true
			case FmtJSON:
				var cm countdownMsg
				if err := json.Unmarshal(msg.data.Data, &cm); err != nil {
					fmt.Println("Error decoding countdown message:", err)
				}
				c.lobby.Countdown = cm.Countdown
			}
		}

		return nil
	}))
}
