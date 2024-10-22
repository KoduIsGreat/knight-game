package nw

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	quic "github.com/quic-go/quic-go"
	"golang.org/x/exp/rand"
)

type Server[T any] struct {
	address string

	tlsConfig  *tls.Config
	quicConfig *quic.Config

	lobbies map[string]*GameServer[T]
	state   StateManager[T]
	// defines the Tick of the server
	tickRate time.Duration
	log      *log.Logger
	// map of <remote_address:quic.StreamID> to Client
	clients map[string]*client

	// channel for handling joining clients
	newClients chan *client
	// channel for handling removing clients
	removeClients chan *client
	// channel for new lobbies being created on the server
	newLobbies chan string
	// channel for removing empty lobbies
	closeLobbies chan string
}

type ClientInput struct {
	ClientID string
	Input    string
	Sequence uint32
}

type ServerStateMessage[T any] struct {
	GameState       T
	AcknowledgedSeq map[string]uint32
}

// represents a client connected to the server
type client struct {
	ID           string
	nick         string
	conn         quic.Connection
	stream       quic.Stream
	sendChan     chan Message
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
			if err := msg.EncodeTo(c.stream); err != nil {
				log.Println("Error sending message to client:", err)
				return
			}
		case <-c.quitChan:
			return
		}
	}
}

func (c *client) reader(removedClients chan *client, mh MessageHandler) {
	defer func() {
		c.quitChan <- struct{}{}
		removedClients <- c
	}()

	for {
		var message Message
		if err := message.DecodeFrom(c.stream); err != nil {
			log.Println("error decoding message:", err)
			return
		}
		mh.Handle(message)
	}
}

func NewServer[T any](sm StateManager[T], opts ...ServerOption[T]) *Server[T] {
	log := log.New(os.Stdout, "server: ", log.Lshortfile)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{GenerateSelfSignedTLSCertificate()},
		NextProtos:   []string{"snake-game"},
	}
	s := &Server[T]{
		address:       address,
		tlsConfig:     tlsConfig,
		quicConfig:    &quic.Config{},
		state:         sm,
		tickRate:      gameInterval,
		log:           log,
		lobbies:       make(map[string]*GameServer[T]),
		newClients:    make(chan *client),
		removeClients: make(chan *client),
		newLobbies:    make(chan string),
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
	go s.loop()
	// Accept client connections
	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			return err
		}
		go s.handleClient(conn)
	}
}

func randomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
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
		sendChan:     make(chan Message, 10),
		quitChan:     make(chan struct{}),
		lastSequence: 0,
	}

	// Add the client to the server
	go client.writer()
	mh := MessageHandlerFunc(func(msg Message) error {
		switch msg.header {
		case MsgAuth:
			// TODO handle auth message
			client.sendChan <- NewMessage(MsgAuthAck, FmtText, []byte(client.ID))
		case MsgConnect:
			s.log.Println("Client connected:", client.ID)
			s.newClients <- client
			client.sendChan <- NewMessage(MsgConnect, FmtText, []byte(client.ID))
		case MsgDisconnect:
			s.log.Fatal("Client disconnected:", client.ID)
			s.removeClients <- client
			// TODO handle disconnect message
		case MsgLobbyClientReady:
			parts := strings.Split(string(msg.data.Data), "|")
			if len(parts) != 2 {
				return fmt.Errorf("invalid lobby join message")
			}
			lobbyID := parts[0]
			lobby, ok := s.lobbies[lobbyID]
			if !ok {
				return fmt.Errorf("lobby not found")
			}
			lobby.readyChan <- client
		case MsgLobbyCreate:
			s.newLobbies <- fmt.Sprintf("%s:%s", randomString(6), string(msg.data.Data))
		case MsgLobbyClientJoin:
			parts := strings.Split(string(msg.data.Data), "|")
			if len(parts) != 2 {
				return fmt.Errorf("invalid lobby join message")
			}
			lobbyID := parts[0]
			lobby := s.lobbies[lobbyID]
			if lobby == nil {
				return fmt.Errorf("lobby not found")
			}
			lobby.addClient(client)
		case MsgLobbyClientLeave:
			parts := strings.Split(string(msg.data.Data), "|")
			if len(parts) != 2 {
				return fmt.Errorf("invalid lobby leave message")
			}
			lobbyID := parts[0]
			lobby := s.lobbies[lobbyID]
			if lobby == nil {
				return fmt.Errorf("lobby not found")
			}
			lobby.removeClient(client)
			if len(lobby.clients) == 0 {
				s.closeLobbies <- lobbyID
			}
		}

		return nil
	})

	go client.reader(s.removeClients, mh)
	s.newClients <- client
}

func (s *Server[T]) loop() {
	for {
		select {
		case msg := <-s.newLobbies:
			parts := strings.Split(msg, ":")
			if len(parts) != 2 {
				s.log.Println("Invalid lobby create message")
				continue
			}
			newLobbyCode := parts[0]
			clientID := parts[1]

			for _, ok := s.lobbies[newLobbyCode]; ok == true; {
				newLobbyCode = randomString(6)
			}
			newLobby := NewGameServer(newLobbyCode, s.state)
			client, ok := s.clients[clientID]
			if !ok {
				s.log.Println("Client not found:", clientID)
				continue
			}
			newLobby.addClient(client)
			s.lobbies[newLobbyCode] = newLobby
			s.log.Println("New lobby created:", newLobbyCode)
			// Send the client the lobby code
			message := NewMessage(MsgLobbyCreated, FmtText, []byte(newLobbyCode))
			client.sendChan <- message

		case lobbyCode := <-s.closeLobbies:
			delete(s.lobbies, lobbyCode)
		case client := <-s.newClients:
			fmt.Printf("Adding client %s to server\n", client.ID)
			s.clients[client.ID] = client
			message := NewMessage(MsgConnect, FmtText, []byte(client.ID))
			client.sendChan <- message
		case client := <-s.removeClients:
			close(client.sendChan)
			s.log.Println("Client removed, clients count:", len(s.clients))
		}
	}
}
