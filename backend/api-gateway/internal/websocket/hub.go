package websocket

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub управляет WebSocket-подключениями и рассылкой сообщений
type Hub struct {
	clients    map[*Client]bool // активные клиенты
	broadcast  chan []byte      // канал для отправки сообщений всем клиентам
	register   chan *Client     // регистрация нового клиента
	unregister chan *Client     // отключение клиента
	mu         sync.RWMutex
}

// Client представляет одно WebSocket-подключение
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte // буферизованный канал для отправки сообщений
}

// NewHub создаёт новый Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run запускает основной цикл hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client registered. Total clients: %d", len(h.clients))
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client unregistered. Total clients: %d", len(h.clients))
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Если канал заполнен, закрываем соединение
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast отправляет сообщение всем клиентам (можно вызывать извне)
func (h *Hub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser отправляет сообщение конкретному пользователю (по userID)
// Это можно реализовать позже, когда будем хранить userID в клиенте
