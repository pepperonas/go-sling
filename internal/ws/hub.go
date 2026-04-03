package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // LAN-only, allow all origins
	},
}

type Hub struct {
	clients    map[string]*Client
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Peer connected: %s (%s)", client.Name, client.ID)
			h.broadcastPeerList()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("Peer disconnected: %s (%s)", client.Name, client.ID)
			h.broadcastPeerList()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := NewClient(h, conn)
	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) broadcastPeerList() {
	h.mu.RLock()
	peers := make([]PeerInfo, 0, len(h.clients))
	for _, c := range h.clients {
		peers = append(peers, PeerInfo{
			ID:       c.ID,
			Name:     c.Name,
			Browser:  c.Browser,
			OS:       c.OS,
			JoinedAt: c.JoinedAt.Unix(),
		})
	}
	h.mu.RUnlock()

	msg := SignalMessage{
		Type:  "peer-list",
		Peers: peers,
	}
	data, _ := json.Marshal(msg)

	h.mu.RLock()
	for _, client := range h.clients {
		select {
		case client.send <- data:
		default:
		}
	}
	h.mu.RUnlock()
}

func (h *Hub) sendTo(targetID string, data []byte) {
	h.mu.RLock()
	client, ok := h.clients[targetID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.send <- data:
		default:
		}
	}
}
