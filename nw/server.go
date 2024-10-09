package nw

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	. "github.com/KoduIsGreat/knight-game/common"
	quic "github.com/quic-go/quic-go"
)

type Server[T any] struct {
	address string

	tlsConfig  *tls.Config
	quicConfig *quic.Config

	state StateManager[T]
	// defines the Tick of the server
	tickRate time.Duration
	log      *log.Logger
	// map of <remote_address:quic.StreamID> to Client
	clients map[string]*client
	// channel for new clients joining the server
	newClients        chan *client
	removeClients     chan *client
	clientInputs      chan ClientInput
	clientInputQueues map[string][]ClientInput
}

type serverStateMessage struct {
	GameState       any
	AcknowledgedSeq map[string]uint32
}

// represents a client connected to the server
type client struct {
	ID           string
	conn         quic.Connection
	stream       quic.Stream
	sendChan     chan string
	quitChan     chan struct{}
	lastSequence uint32
}

func (c *client) writer() {
	for {
		select {
		case msg, ok := <-c.sendChan:
			if !ok {
				return
			}
			if _, err := c.stream.Write([]byte(msg)); err != nil {
				log.Println("Error sending message to client:", err)
				return
			}
		case <-c.quitChan:
			return
		}
	}
}

func (c *client) reader(removedClients chan *client, clientInputs chan ClientInput) {
	defer func() {
		c.quitChan <- struct{}{}
		removedClients <- c
	}()

	buf := make([]byte, 1024)
	for {
		n, err := c.stream.Read(buf)
		if err != nil {
			log.Println("Client disconnected:", err)
			return
		}
		if buf[0] == 'j' {
			continue
		}

		var inputMsg struct {
			Sequence uint32 `json:"sequence"`
			Input    string `json:"input"`
		}

		err = json.Unmarshal(buf[:n], &inputMsg)
		if err != nil {
			log.Println("Error parsing client input:", err)
			continue
		}

		clientInputs <- ClientInput{
			ClientID: c.ID,
			Input:    inputMsg.Input,
			Sequence: inputMsg.Sequence,
		}
	}
}

func NewServer[T any](sm StateManager[T], opts ...ServerOption[T]) *Server[T] {
	log := log.New(os.Stdout, "server: ", log.Lshortfile)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{GenerateSelfSignedTLSCertificate()},
		NextProtos:   []string{"snake-game"},
	}
	s := &Server[T]{
		address:           address,
		tlsConfig:         tlsConfig,
		quicConfig:        &quic.Config{},
		state:             sm,
		tickRate:          gameInterval,
		log:               log,
		clients:           make(map[string]*client),
		newClients:        make(chan *client),
		removeClients:     make(chan *client),
		clientInputs:      make(chan ClientInput),
		clientInputQueues: map[string][]ClientInput{},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// StartServer starts a QUIC server and listens for client connections.
func (s *Server[T]) Listen() error {

	// Listen on a QUIC address
	listener, err := quic.ListenAddr(s.address, s.tlsConfig, nil)
	if err != nil {
		return err
	}

	s.log.Println("Server is listening on", address)
	// Start the broadcaster goroutine
	go s.gameLoop()
	// Accept client connections
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		go s.handleClient(conn)
	}
}

// handleClient handles individual client connections
func (s *Server[T]) handleClient(conn quic.Connection) {
	clientID := conn.RemoteAddr().String()
	fmt.Println("New client connected:", clientID)

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		log.Println("Error accepting stream:", err)
		return
	}

	client := &client{
		ID:           clientID,
		conn:         conn,
		stream:       stream,
		sendChan:     make(chan string, 10),
		quitChan:     make(chan struct{}),
		lastSequence: 0,
	}

	// Add the client to the server
	go client.writer()
	go client.reader(s.removeClients, s.clientInputs)
	s.newClients <- client
}

func (s *Server[T]) sendJoinMsg(client *client) {
	joinMsg := struct {
		ClientID string `json:"clientID"`
	}{
		ClientID: client.ID,
	}
	data, err := json.Marshal(joinMsg)
	if err != nil {
		log.Println("Error marshaling join message:", err)
		return
	}
	message := string(data) + "\n"

	fmt.Println("acking join message for client:", message)
	client.sendChan <- message
}

func (s *Server[T]) gameLoop() {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()
	for {
		select {
		case client := <-s.newClients:
			fmt.Printf("Adding client %s to server\n", client.ID)
			s.clients[client.ID] = client
			s.state.InitClientEntity(client.ID)
			s.sendJoinMsg(client)
			s.clientInputQueues[client.ID] = []ClientInput{}
			broadcastGameState(s.state.Get(), s.clients)
		case client := <-s.removeClients:
			s.state.RemoveClientEntity(client.ID)
			delete(s.clientInputQueues, client.ID)
			close(client.sendChan)
			s.log.Println("Client removed, clients count:", len(s.clients))
		case input := <-s.clientInputs:
			s.log.Println("Client input received:", input)
			queue := s.clientInputQueues[input.ClientID]
			queue = append(queue, input)
			s.clientInputQueues[input.ClientID] = queue
		case <-ticker.C:
			s.processInputs()
			s.state.Update(s.tickRate.Seconds())
			broadcastGameState(s.state.Get(), s.clients)
		}
	}
}

func (s *Server[T]) processInputs() {
	for clientID, queue := range s.clientInputQueues {
		client := s.clients[clientID]
		sort.Slice(queue, func(i, j int) bool {
			return queue[i].Sequence < queue[j].Sequence
		})

		newQueue := make([]ClientInput, 0)
		for _, input := range queue {
			if input.Sequence > client.lastSequence {
				s.state.ApplyInputToState(input)
				client.lastSequence = input.Sequence
			}
			if input.Sequence > client.lastSequence {
				newQueue = append(newQueue, input)
			}
		}
		s.clientInputQueues[clientID] = newQueue
	}
}

func broadcastGameState(gameState any, clients map[string]*client) {
	ackSeq := make(map[string]uint32)
	for clientID, client := range clients {
		ackSeq[clientID] = client.lastSequence
	}

	serverMessage := serverStateMessage{
		GameState:       gameState,
		AcknowledgedSeq: ackSeq,
	}

	data, err := json.Marshal(serverMessage)
	if err != nil {
		log.Println("Error marshaling game state:", err)
		return
	}
	message := string(data) + "\n"

	for _, client := range clients {
		select {
		case client.sendChan <- message:
		}
	}
}
