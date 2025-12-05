package hub

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/server/model"
)

// Client represents a connected WebSocket client.
type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
	Id   string
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	Clients    map[string]*Client
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the Hub's main loop for managing clients.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.Id] = client
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.Id]; ok {
				delete(h.Clients, client.Id)
				close(client.Send)
			}
			h.mu.Unlock()

		case message := <-h.Broadcast:
			fmt.Println("received message at hub:", string(message))
			if message == nil {
				continue
			}

			var msg model.Message
			err := json.Unmarshal(message, &msg)
			if err != nil {
				fmt.Printf("error unmarshaling message: %v\n", err)
				continue
			}

			h.mu.Lock()
			if msg.IsBroadcast {
				fmt.Println("broadcasting message to all clients")
				for clientId, client := range h.Clients {
					if clientId == msg.Sender {
						continue
					}

					fmt.Println("sending message to client", client.Id)
					select {
					case client.Send <- message:
						fmt.Println("sent message to client:", client.Id)
					default:
						// remove slow/unresponsive client
						close(client.Send)
						delete(h.Clients, client.Id)
					}
				}
			} else {
				fmt.Println("sending message to client:", msg.Receiver)
				if target, ok := h.Clients[msg.Receiver]; ok {
					select {
					case target.Send <- []byte(msg.Content):
					default:
						// remove slow/unresponsive target
						close(target.Send)
						delete(h.Clients, target.Id)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}
