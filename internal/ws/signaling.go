package ws

import (
	"encoding/json"
	"log"
)

type PeerInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Browser  string `json:"browser,omitempty"`
	OS       string `json:"os,omitempty"`
	JoinedAt int64  `json:"joinedAt"`
}

type SignalMessage struct {
	Type    string         `json:"type"`
	From    string         `json:"from,omitempty"`
	To      string         `json:"to,omitempty"`
	Payload map[string]any `json:"payload,omitempty"`
	Peers   []PeerInfo     `json:"peers,omitempty"`
}

func marshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (c *Client) handleMessage(raw []byte) {
	var msg SignalMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		log.Printf("Invalid message from %s: %v", c.ID, err)
		return
	}

	msg.From = c.ID

	switch msg.Type {
	case "join":
		// Update client info from browser
		if name, ok := msg.Payload["name"].(string); ok && name != "" {
			c.Name = name
		}
		if browser, ok := msg.Payload["browser"].(string); ok {
			c.Browser = browser
		}
		if osName, ok := msg.Payload["os"].(string); ok {
			c.OS = osName
		}
		c.hub.broadcastPeerList()

	case "offer", "answer", "ice-candidate":
		// Relay signaling messages to target peer
		if msg.To == "" {
			log.Printf("Signaling message from %s has no target", c.ID)
			return
		}
		data, _ := json.Marshal(msg)
		c.hub.sendTo(msg.To, data)

	case "transfer-request", "transfer-accept", "transfer-reject":
		// Relay transfer negotiation
		if msg.To == "" {
			return
		}
		data, _ := json.Marshal(msg)
		c.hub.sendTo(msg.To, data)

	default:
		log.Printf("Unknown message type from %s: %s", c.ID, msg.Type)
	}
}
