package nw

import (
	"fmt"
	"log"
	"sort"
	"time"
)

type GameServer[T any] struct {
	ID         string
	state      StateManager[T]
	maxClients int
	log        *log.Logger
	countdown  int
	tickRate   time.Duration

	OwnerID           string
	clients           map[string]*client
	clientInputs      chan ClientInput
	clientInputQueues map[string][]ClientInput
	newClients        chan *client
	promoteChan       chan string
	removeClients     chan *client
	readyChan         chan *client
	readyClients      map[string]bool
	startChan         chan struct{}
	stopChan          chan struct{}
	started           bool
}

func NewGameServerID() string {
	return fmt.Sprintf("game_%d", time.Now().Unix())
}

func NewGameServer[T any](id, ownerId string, state StateManager[T], opts ...GameServerOption[T]) *GameServer[T] {
	s := &GameServer[T]{
		ID: id,

		OwnerID:           ownerId,
		clients:           make(map[string]*client),
		state:             state,
		clientInputs:      make(chan ClientInput),
		log:               log.Default(),
		clientInputQueues: make(map[string][]ClientInput),
		tickRate:          time.Second / 30,
		newClients:        make(chan *client),
		removeClients:     make(chan *client),
		startChan:         make(chan struct{}),
		readyChan:         make(chan *client),
		readyClients:      make(map[string]bool),
		promoteChan:       make(chan string),

		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(s)
	}
	go s.handleLobbyActions()
	return s
}

func (s *GameServer[T]) promote(clientId string) {
	s.promoteChan <- clientId
}
func (s *GameServer[T]) kick(clientId string) {
	client, ok := s.clients[clientId]
	if !ok {
		s.log.Println("Client not found to kick")
		return
	}
	s.removeClients <- client
}

func (s *GameServer[T]) broadcast(msg Message) {
	for _, client := range s.clients {
		client.sendChan <- msg
	}
}

func (s *GameServer[T]) addClient(client *client) {
	s.newClients <- client
}

func (s *GameServer[T]) removeClient(client *client) {
	s.removeClients <- client
}

func (s *GameServer[T]) makeServerStateMessage(gameState T) (Message, error) {
	ackSeq := make(map[string]uint32)
	for clientID, client := range s.clients {
		ackSeq[clientID] = client.lastSequence
	}
	serverMessage := ServerStateMessage[T]{
		GameState:       gameState,
		AcknowledgedSeq: ackSeq,
	}
	return NewGameStateMessage(FmtJSON, serverMessage)
}

func (s *GameServer[T]) start() {
	countDownTicker := time.NewTicker(time.Second)
	countDown := 10
	for {
		select {
		case <-countDownTicker.C:
			countDown--
			s.log.Println("Starting in", countDown)
			countDownMsg := fmt.Sprintf(`{"countdown": %d}`, countDown)
			msg := NewMessage(MsgLobbyGameStarted, FmtJSON, []byte(countDownMsg))
			s.broadcast(msg)
			if countDown == 0 {
				s.startChan <- struct{}{}
				break
			}
		case <-s.startChan:
			countDownTicker.Stop()
			s.started = true
			msg := NewMessage(MsgLobbyGameStarted, FmtText, []byte{})
			s.broadcast(msg)
			goto Started
		}
	Started:
		break
	}
	go s.gameLoop()
}
func (s *GameServer[T]) stop() {
	s.log.Println("Stopping server")
	s.stopChan <- struct{}{}
	for _, client := range s.clients {
		s.removeClients <- client
	}
}

func (s *GameServer[T]) processInputs() {
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

func (s *GameServer[T]) handleLobbyActions() {
	for {
		select {
		case readyClient := <-s.readyChan:
			s.log.Println("Client ready msg:", readyClient.ID)
			s.readyClients[readyClient.ID] = !s.readyClients[readyClient.ID]
			s.broadcast(NewMessage(MsgLobbyClientReady, FmtText, []byte(readyClient.ID)))
		case client := <-s.newClients:
			fmt.Printf("Adding client %s to lobby %s\n", client.ID, s.ID)
			s.clients[client.ID] = client
			message := NewMessage(MsgLobbyClientJoin, FmtText, []byte(client.ID))
			client.sendChan <- message
			s.clientInputQueues[client.ID] = []ClientInput{}
			s.broadcast(message)
		case toPromote := <-s.promoteChan:
			_, ok := s.clients[toPromote]
			if !ok {
				s.log.Println("Client not found to promote")
				continue
			}
			s.OwnerID = toPromote
		case client := <-s.removeClients:
			s.state.RemoveClientEntity(client.ID)
			message := NewMessage(MsgLobbyClientLeave, FmtText, []byte(client.ID))
			delete(s.clientInputQueues, client.ID)
			close(client.sendChan)
			s.broadcast(message)
			s.log.Println("Client removed, clients count:", len(s.clients))
		case <-s.startChan:
			s.log.Println("attempting to start game")
			var allReady bool = true
			for _, client := range s.clients {
				allReady = s.readyClients[client.ID] && allReady
			}
			if !allReady {
				s.log.Println("Not all clients are ready")
				s.broadcast(NewMessage(MsgLobbyClientsNotReady, FmtText, []byte("Not all clients are ready")))
				continue
			}
			s.log.Println("Starting game")
			s.start()

		}
	}
}

func (s *GameServer[T]) gameLoop() {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopChan:
			return
		case input := <-s.clientInputs:
			s.log.Println("Client input received:", input)
			queue := s.clientInputQueues[input.ClientID]
			queue = append(queue, input)
			s.clientInputQueues[input.ClientID] = queue
		case <-ticker.C:
			s.processInputs()
			s.state.Update(s.tickRate.Seconds())
			msg, err := s.makeServerStateMessage(s.state.Get())
			if err != nil {
				s.log.Println("Error making server state message:", err)
				continue
			}
			s.broadcast(msg)
		}
	}
}
