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
	Clients     map[*Client]bool
	ClientsById map[string]*Client
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	mu          sync.Mutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		ClientsById: make(map[string]*Client),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
	}
}

// Run starts the Hub's main loop for managing clients.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.ClientsById[client.Id] = client
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				if client.Id != "" {
					if idClient, ok := h.ClientsById[client.Id]; ok && idClient == client {
						delete(h.ClientsById, client.Id)
					}
				}
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
				for client := range h.Clients {
					select {
					case client.Send <- msg.Content:
						fmt.Println("sent message to client:", client.Id)
					default:
						// remove slow/unresponsive client
						close(client.Send)
						delete(h.Clients, client)
						if client.Id != "" {
							if idc, ok := h.ClientsById[client.Id]; ok && idc == client {
								delete(h.ClientsById, client.Id)
							}
						}
					}
				}
			} else {
				fmt.Println("sending message to client:", msg.Receiver)
				if target, ok := h.ClientsById[msg.Receiver]; ok {
					select {
					case target.Send <- msg.Content:
					default:
						// remove slow/unresponsive target
						close(target.Send)
						delete(h.Clients, target)
						if target.Id != "" {
							if idc, ok := h.ClientsById[target.Id]; ok && idc == target {
								delete(h.ClientsById, target.Id)
							}
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}
