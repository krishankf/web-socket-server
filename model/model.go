package model

import "time"

type Message struct {
	ID          string    `json:"id"`
	// Use string for Content to match typical JSON/text payloads from clients.
	Content     string    `json:"content"`
	Sender      string    `json:"sender"`
	Time        time.Time `json:"time"`
	Receiver    string    `json:"receiver"`
	IsBroadcast bool      `json:"is_broadcast"`
}
