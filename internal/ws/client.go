package ws

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 512 * 1024 // 512KB for signaling messages
)

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	ID       string
	Name     string
	Browser  string
	OS       string
	Headless bool
	JoinedAt time.Time
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		ID:       generateClientID(),
		Name:     generateName(),
		JoinedAt: time.Now(),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// Send initial welcome with client ID
	welcome := SignalMessage{
		Type: "welcome",
		From: c.ID,
		Payload: map[string]any{
			"id":   c.ID,
			"name": c.Name,
		},
	}
	if data, err := marshalJSON(welcome); err == nil {
		c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		c.conn.WriteMessage(websocket.TextMessage, data)
	}

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func generateClientID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

var adjectives = []string{"Swift", "Bright", "Cool", "Fast", "Bold", "Calm", "Keen", "Sharp", "Smart", "Warm"}
var nouns = []string{"Fox", "Hawk", "Wolf", "Bear", "Lynx", "Owl", "Puma", "Raven", "Tiger", "Eagle"}

func generateName() string {
	ai, _ := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	ni, _ := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	suffix := make([]byte, 2)
	rand.Read(suffix)
	return fmt.Sprintf("%s-%s-%s", adjectives[ai.Int64()], nouns[ni.Int64()], hex.EncodeToString(suffix)[:3])
}
