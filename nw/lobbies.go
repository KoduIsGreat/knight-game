package nw

import (
	"fmt"
	"log"
	"sort"
	"time"
)

type GameServer[T any] struct {
	ID                string
	clients           map[string]*client
	state             StateManager[T]
	clientInputs      chan ClientInput
	log               *log.Logger
	clientInputQueues map[string][]ClientInput
	tickRate          time.Duration
	newClients        chan *client
	removeClients     chan *client
	startChan         chan struct{}
	started           bool
}

func (s *GameServer[T]) Broadcast(msg Message) {
	for _, client := range s.clients {
		client.sendChan <- msg
	}
}

func (s *GameServer[T]) MakeServerStateMessage(gameState T) (Message, error) {
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

func (s *GameServer[T]) Start() {
	countDownTicker := time.NewTicker(time.Second)
	countDown := 10
	for {
		select {
		case <-countDownTicker.C:
			countDown--
			s.log.Println("Starting in", countDown)
			countDownMsg := fmt.Sprintf("Game Starting in %d", countDown)
			msg := NewMessage(MsgLobbyGameStart, FmtText, []byte(countDownMsg))
			s.Broadcast(msg)
			if countDown == 0 {
				s.startChan <- struct{}{}
				break
			}
		case <-s.startChan:
			countDownTicker.Stop()
			s.started = true
			goto Started
		}
	Started:
		break
	}
	go s.gameLoop()
}

func (s *GameServer[T]) ProcessInputs() {
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

func (s *GameServer[T]) gameLoop() {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()
	for {
		select {
		case client := <-s.newClients:
			fmt.Printf("Adding client %s to server\n", client.ID)
			s.clients[client.ID] = client
			s.state.InitClientEntity(client.ID)
			message := NewMessage(MsgConnect, FmtText, []byte(client.ID))
			client.sendChan <- message
			s.clientInputQueues[client.ID] = []ClientInput{}
			s.Broadcast(message)
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
			s.ProcessInputs()
			s.state.Update(s.tickRate.Seconds())
			msg, err := s.MakeServerStateMessage(s.state.Get())
			if err != nil {
				s.log.Println("Error making server state message:", err)
				continue
			}
			s.Broadcast(msg)
		}
	}
}
