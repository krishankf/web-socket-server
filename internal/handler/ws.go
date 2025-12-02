package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/server/internal/hub"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket handles the WebSocket connection upgrade
func HandleWebSocket(h *hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not upgrade connection", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	clientId := r.URL.Query().Get("id")
	fmt.Println("New client connected with ID:", clientId)

	client := &hub.Client{
		Hub:  h,
		Conn: conn,
		Send: make(chan []byte),
		Id:   clientId,
	}

	h.Register <- client

	for {
		_, msg, err := conn.ReadMessage()
		fmt.Println("message received at handler:", string(msg))
		if err != nil {
			h.Unregister <- client
			break
		}

		decodedMsg, err := base64.StdEncoding.DecodeString(string(msg))
		if err != nil {
			fmt.Printf("error decoding base64 message: %v\n", err)
			continue
		}

		h.Broadcast <- decodedMsg
		continue
	}
}
