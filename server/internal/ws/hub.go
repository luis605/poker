package ws

import (
	"context"
)

type Hub struct {
	// Arrays
	clients map[*Client]bool
	broadcast chan []byte

	// requests
	register chan *Client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Clean up clients on shutdown
			for client := range h.clients {
				delete(h.clients, client)
				client.Close()
			}
			return

		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.Close()
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
				delete(h.clients, client)
				client.Close()
				}
			}
		}
	}
}