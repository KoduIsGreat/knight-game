package nw

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"sync"

	. "github.com/KoduIsGreat/knight-game/common"
	quic "github.com/quic-go/quic-go"
)

type Client struct {
	stream        quic.Stream
	sendMutex     sync.Mutex
	inputSequence uint32
	state         ClientStateManager
	clientID      string
}

// connectToServer establishes a QUIC connection to the server.
func (c *Client) ConnectToServer() {
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

	// Use session's local address as client ID
	c.clientID = session.LocalAddr().String()
}

// startReceivingGameState starts a goroutine to receive game state updates.
func (c *Client) StartReceivingGameState() {
	go func() {
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

			c.state.ReconcileState(serverMessage)
		}
	}()
}
