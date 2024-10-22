package nw

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"log"

	quic "github.com/quic-go/quic-go"
)

type Client[T any] struct {
	stream quic.Stream
	// sendChan is used to send messages to the server
	sendChan chan Message
	// recvChan is used to receive messages from the server
	recvChan chan Message
	// gameStateChan is used to game state from the server
	gameStateChan chan ServerStateMessage[T]
	// quitChan is used to signal the network handlers to stop
	quitChan chan struct{}
	// state is the client's state manager
	state          ClientStateManager[T]
	joinLobbyChan  chan string
	leaveLobbyChan chan string
	lobby          *Lobby
	// clientID is the client's ID determined by the server
	clientID string
}

type Lobby struct {
	ID               string
	OwnerClientID    string
	ReadyClients     map[string]bool
	ConnectedClients map[string]otherClient
	Started          bool
	Countdown        int
}
type otherClient struct {
}

type ClientOpts struct {
	ServerAddress string
	TLSConfig     *tls.Config
}

// NewClient creates a new client with the given state manager.
func NewClient[T any](state ClientStateManager[T], co ClientOpts) *Client[T] {
	c := &Client[T]{
		sendChan:      make(chan Message),
		recvChan:      make(chan Message),
		gameStateChan: make(chan ServerStateMessage[T]),
		quitChan:      make(chan struct{}),
		state:         state,
	}
	c.connectToServer(co)
	c.waitUntilConnected()
	c.startNetworkHandlers()
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

func (c *Client[T]) JoinLobby(lobbyID string) {
	c.joinLobbyChan <- lobbyID
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
		case MsgServerState:
			msg, err := ServerStateMessageFromMessage[T](msg)
			if err != nil {
				fmt.Println("Error decoding server state message:", err)
			}
			c.gameStateChan <- msg
		case MsgLobbyGameStart:

		}

		return nil
	}))
}
