package main

import (
	"log"
	"net/http"

	"github.com/server/internal/handler"
	"github.com/server/internal/hub"
)

var h = hub.NewHub()

func main() {
	go h.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handler.HandleWebSocket(h, w, r)
	})

	addr := "localhost:8080"
	log.Printf("Server started at %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
