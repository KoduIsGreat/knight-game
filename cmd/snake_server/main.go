package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	quic "github.com/quic-go/quic-go"
)

const (
	address      = "localhost:4242"
	gameInterval = time.Second / 30 // 30 ticks per second
)

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Snake struct {
	ID        string     `json:"id"`
	Segments  []Position `json:"segments"`
	Direction string     `json:"direction"`
}

type FoodItem struct {
	Position `json:"position"`
}

type GameState struct {
	Snakes    map[string]*Snake `json:"snakes"`
	FoodItems []FoodItem        `json:"foodItems"`
}

type Client struct {
	ID           string
	conn         quic.Connection
	stream       quic.Stream
	sendChan     chan string
	quitChan     chan struct{}
	lastSequence uint32
}

type ClientInput struct {
	ClientID string
	Input    string
	Sequence uint32
}

type ServerStateMessage struct {
	GameState       GameState         `json:"gameState"`
	AcknowledgedSeq map[string]uint32 `json:"acknowledgedSeq"`
}

func main() {
	if err := StartServer(); err != nil {
		log.Fatal(err)
	}
}

func StartServer() error {
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{generateTLSCertificate()},
		NextProtos:   []string{"snake-game"},
	}

	listener, err := quic.ListenAddr(address, tlsConfig, nil)
	if err != nil {
		return err
	}

	fmt.Println("Server is listening on", address)

	newClients := make(chan *Client)
	removedClients := make(chan *Client)
	clientInputs := make(chan ClientInput, 100)

	go gameLoop(newClients, removedClients, clientInputs)

	for {
		conn, err := listener.Accept(context.Background())
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleClient(conn, newClients, removedClients, clientInputs)
	}
}

func handleClient(conn quic.Connection, newClients chan *Client, removedClients chan *Client, clientInputs chan ClientInput) {
	clientID := conn.RemoteAddr().String()
	fmt.Println("New client connected:", clientID)

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		log.Println("Error accepting stream:", err)
		return
	}

	client := &Client{
		ID:           clientID,
		conn:         conn,
		stream:       stream,
		sendChan:     make(chan string, 10),
		quitChan:     make(chan struct{}),
		lastSequence: 0,
	}

	newClients <- client

	go clientWriter(client, removedClients)
	go clientReader(client, clientInputs, removedClients)
}

func clientWriter(client *Client, removedClients chan *Client) {
	for {
		select {
		case msg, ok := <-client.sendChan:
			if !ok {
				return
			}
			_, err := client.stream.Write([]byte(msg))
			if err != nil {
				log.Println("Error sending to client:", err)
				removedClients <- client
				return
			}
		case <-client.quitChan:
			return
		}
	}
}

func clientReader(client *Client, clientInputs chan ClientInput, removedClients chan *Client) {
	defer func() {
		client.quitChan <- struct{}{}
		removedClients <- client
	}()

	buf := make([]byte, 1024)
	for {
		n, err := client.stream.Read(buf)
		if err != nil {
			log.Println("Client disconnected:", err)
			return
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
			ClientID: client.ID,
			Input:    inputMsg.Input,
			Sequence: inputMsg.Sequence,
		}
	}
}

func gameLoop(newClients chan *Client, removedClients chan *Client, clientInputs chan ClientInput) {
	clients := make(map[string]*Client)
	clientInputQueues := make(map[string][]ClientInput)
	ticker := time.NewTicker(gameInterval)
	defer ticker.Stop()

	gameState := GameState{
		Snakes:    make(map[string]*Snake),
		FoodItems: []FoodItem{},
	}

	for {
		select {
		case client := <-newClients:
			clients[client.ID] = client
			gameState.Snakes[client.ID] = initializeSnakeForClient(client.ID)
			clientInputQueues[client.ID] = []ClientInput{}
		case client := <-removedClients:
			delete(clients, client.ID)
			delete(gameState.Snakes, client.ID)
			delete(clientInputQueues, client.ID)
			close(client.sendChan)
			fmt.Println("Client removed:", client.ID)
		case input := <-clientInputs:
			queue := clientInputQueues[input.ClientID]
			queue = append(queue, input)
			clientInputQueues[input.ClientID] = queue
		case <-ticker.C:
			processInputs(clientInputQueues, gameState, clients)
			updateGameState(gameState)
			broadcastGameState(gameState, clients)
		}
	}
}

func initializeSnakeForClient(clientID string) *Snake {
	return &Snake{
		ID:        clientID,
		Segments:  []Position{{X: 10, Y: 10}},
		Direction: "RIGHT",
	}
}

func processInputs(clientInputQueues map[string][]ClientInput, gameState GameState, clients map[string]*Client) {
	for clientID, queue := range clientInputQueues {
		client := clients[clientID]
		sort.Slice(queue, func(i, j int) bool {
			return queue[i].Sequence < queue[j].Sequence
		})

		newQueue := make([]ClientInput, 0)
		for _, input := range queue {
			if input.Sequence > client.lastSequence {
				applyInputToGameState(clientID, input.Input, gameState)
				client.lastSequence = input.Sequence
			}
			if input.Sequence > client.lastSequence {
				newQueue = append(newQueue, input)
			}
		}
		clientInputQueues[clientID] = newQueue
	}
}

func applyInputToGameState(clientID string, input string, gameState GameState) {
	snake, exists := gameState.Snakes[clientID]
	if !exists {
		return
	}

	oppositeDirections := map[string]string{
		"UP":    "DOWN",
		"DOWN":  "UP",
		"LEFT":  "RIGHT",
		"RIGHT": "LEFT",
	}

	if input == "UP" || input == "DOWN" || input == "LEFT" || input == "RIGHT" {
		if snake.Direction != oppositeDirections[input] {
			snake.Direction = input
		}
	}
}

func updateGameState(gameState GameState) {
	for _, snake := range gameState.Snakes {
		moveSnake(snake)
	}
}

func moveSnake(snake *Snake) {
	head := snake.Segments[0]
	newHead := head

	switch snake.Direction {
	case "UP":
		newHead.Y -= 1
	case "DOWN":
		newHead.Y += 1
	case "LEFT":
		newHead.X -= 1
	case "RIGHT":
		newHead.X += 1
	}

	snake.Segments = append([]Position{newHead}, snake.Segments...)
	snake.Segments = snake.Segments[:len(snake.Segments)-1]
}

func broadcastGameState(gameState GameState, clients map[string]*Client) {
	ackSeq := make(map[string]uint32)
	for clientID, client := range clients {
		ackSeq[clientID] = client.lastSequence
	}

	serverMessage := ServerStateMessage{
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

func generateTLSCertificate() tls.Certificate {
	keyPEM := []byte(`-----BEGIN EC PRIVATE KEY-----
MIGkAgEBBDDCjE0tOOwGP2UdSkYfHFyWvvtQnsbOFhIT1cciEwmBqlMX47gU2FMa
lKWSk6EhwnegBwYFK4EEACKhZANiAATNP7I/+RzLGqQYcQg76aV6MXhZFIJFs9mk
7tIsqNKMecWdKHqJ2P+5zSRXokpXNbXiMVtAaRNYNXUcjz3VZYXQCsUAc2yKHOus
PwTiEyGL5EX/9phn0wKqYdNDqBUKSdc=
-----END EC PRIVATE KEY-----`)
	certPEM := []byte(`-----BEGIN CERTIFICATE-----
MIICHDCCAaKgAwIBAgIUKLFAoZHTnqjOONHoJlHgcFnIu+wwCgYIKoZIzj0EAwIw
RTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoMGElu
dGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yNDEwMDIxOTE1MDhaFw0zNDA5MzAx
OTE1MDhaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYD
VQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwdjAQBgcqhkjOPQIBBgUrgQQA
IgNiAATNP7I/+RzLGqQYcQg76aV6MXhZFIJFs9mk7tIsqNKMecWdKHqJ2P+5zSRX
okpXNbXiMVtAaRNYNXUcjz3VZYXQCsUAc2yKHOusPwTiEyGL5EX/9phn0wKqYdND
qBUKSdejUzBRMB0GA1UdDgQWBBRL1Z7wj2175u2XqXBqcnTDnVqIETAfBgNVHSME
GDAWgBRL1Z7wj2175u2XqXBqcnTDnVqIETAPBgNVHRMBAf8EBTADAQH/MAoGCCqG
SM49BAMCA2gAMGUCMFUaPjRAeZAtBQ2FEbNcrGjogQKmiIpkmYJkQEl56li94HEp
Uq7/BS9p0x3GRSjRfQIxAOGOg29v0FQ4Dn7nUK7En8513Njhot/n5N0ZGw+GdztG
d/fYrIe4UfvCz2/SNewdPQ==
-----END CERTIFICATE-----
`)

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal("Error generating TLS certificate:", err)
	}
	return cert
}
