package nw

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	. "github.com/KoduIsGreat/knight-game/common"
	quic "github.com/quic-go/quic-go"
)

type Client[T any] struct {
	stream        quic.Stream
	inputSequence uint32
	// sendChan is used to send messages to the server
	sendChan chan []byte
	// recvChan is used to receive messages from the server
	recvChan chan ServerStateMessage
	// quitChan is used to signal the network handlers to stop
	quitChan chan struct{}
	// state is the client's state manager
	state ClientStateManager[T]

	// clientID is the client's ID determined by the server
	clientID string
}

// NewClient creates a new client with the given state manager.
func NewClient[T any](state ClientStateManager[T]) *Client[T] {
	c := &Client[T]{
		sendChan: make(chan []byte),
		recvChan: make(chan ServerStateMessage),
		quitChan: make(chan struct{}),
		state:    state,
	}
	c.connectToServer()
	c.waitUntilConnected()
	c.startNetworkHandlers()
	return c
}

func (c *Client[T]) State() ClientStateManager[T] {
	return c.state
}

func (c *Client[T]) SendInputToServer(input string) {
	fmt.Println("Sending input:", input)
	msg := ClientInput{
		ClientID: c.clientID,
		Sequence: c.inputSequence,
		Input:    input,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println("Error marshaling input:", err)
		return
	}
	c.sendChan <- append(data, '\n')
}

func (c *Client[T]) RecvFromServer() <-chan ServerStateMessage {
	return c.recvChan
}

func (c *Client[T]) QuitChan() <-chan struct{} {
	return c.quitChan
}

func (c *Client[T]) ClientID() string {
	return c.clientID
}

// connectToServer establishes a QUIC connection to the server.
func (c *Client[T]) connectToServer() {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"snake-game"},
	}

	session, err := quic.DialAddr(context.Background(), address, tlsConfig, nil)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}

	c.stream, err = session.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal("Failed to open stream:", err)
	}
	c.stream.Write([]byte("join\n"))
}

func (c *Client[T]) writer() {
	for {
		select {
		case msg := <-c.sendChan:
			if _, err := c.stream.Write(msg); err != nil {
				log.Println("Error sending message to server:", err)
				return
			}
		case <-c.quitChan:
			return
		}
	}
}

func (c *Client[T]) reader() {
	defer func() {
		c.quitChan <- struct{}{}
	}()

	reader := bufio.NewReader(c.stream)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error receiving game state:", err)
			return
		}

		var serverMessage ServerStateMessage
		err = json.Unmarshal([]byte(message), &serverMessage)
		if err != nil {
			log.Println("Error parsing game state:", err)
			continue
		}
		c.recvChan <- serverMessage
	}
}

func (c *Client[T]) waitUntilConnected() {
	fmt.Println("Waiting for client ID...")
	reader := bufio.NewReader(c.stream)
	for {
		if c.clientID != "" {
			return
		}
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error receiving game state:", err)
			return
		}
		type joinMessage struct {
			ClientID string `json:"clientID"`
		}
		var joinMsg joinMessage
		err = json.Unmarshal([]byte(message), &joinMsg)
		if err != nil {
			log.Println("Error parsing join message:", err)
			continue
		}
		fmt.Println("Received client ID:", joinMsg.ClientID)
		c.clientID = joinMsg.ClientID
		c.state.SetClientID(c.clientID)
	}
}

func (c *Client[T]) startNetworkHandlers() {
	go c.writer()
	go c.reader()
}
