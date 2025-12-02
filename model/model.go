package model

import "time"

type Message struct {
	ID          string    `json:"id"`
	Content     []byte    `json:"content"`
	Sender      string    `json:"sender"`
	Time        time.Time `json:"time"`
	Receiver    string    `json:"receiver"`
	IsBroadcast bool      `json:"is_broadcast"`
}
