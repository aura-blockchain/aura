package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	MessageTypeSubscribe        MessageType = "subscribe"
	MessageTypeUnsubscribe      MessageType = "unsubscribe"
	MessageTypeAssistantUpdate  MessageType = "assistant_update"
	MessageTypeScoreUpdate      MessageType = "score_update"
	MessageTypeIRCompletion     MessageType = "ir_completion"
	MessageTypeHeartbeatAlert   MessageType = "heartbeat_alert"
	MessageTypeMisbehavior      MessageType = "misbehavior_report"
	MessageTypeError            MessageType = "error"
	MessageTypePing             MessageType = "ping"
	MessageTypePong             MessageType = "pong"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      MessageType            `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp int64                  `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	ID            string
	Conn          *websocket.Conn
	Send          chan WSMessage
	Subscriptions map[string]bool
	mu            sync.RWMutex
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan WSMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WSMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client registered: %s. Total clients: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.mu.Lock()
				delete(h.clients, client)
				h.mu.Unlock()
				close(client.Send)
				log.Printf("Client unregistered: %s. Remaining clients: %d", client.ID, len(h.clients))
			}

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				// Check if client is subscribed to this message type
				if h.shouldSendToClient(client, message) {
					select {
					case client.Send <- message:
					default:
						// Client buffer full, unregister
						h.mu.RUnlock()
						h.unregister <- client
						h.mu.RLock()
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// shouldSendToClient checks if a message should be sent to a client
func (h *Hub) shouldSendToClient(client *Client, message WSMessage) bool {
	client.mu.RLock()
	defer client.mu.RUnlock()

	// Always send error and pong messages
	if message.Type == MessageTypeError || message.Type == MessageTypePong {
		return true
	}

	// Check subscription
	channel := getChannelForMessageType(message.Type)
	if channel == "" {
		return false
	}

	return client.Subscriptions[channel]
}

// getChannelForMessageType maps message types to subscription channels
func getChannelForMessageType(msgType MessageType) string {
	switch msgType {
	case MessageTypeAssistantUpdate:
		return "assistants"
	case MessageTypeScoreUpdate:
		return "scores"
	case MessageTypeIRCompletion:
		return "completions"
	case MessageTypeHeartbeatAlert:
		return "alerts"
	case MessageTypeMisbehavior:
		return "misbehavior"
	default:
		return ""
	}
}

// BroadcastAssistantUpdate broadcasts an assistant update
func (h *Hub) BroadcastAssistantUpdate(data map[string]interface{}) {
	h.broadcast <- WSMessage{
		Type:      MessageTypeAssistantUpdate,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// BroadcastScoreUpdate broadcasts a score update
func (h *Hub) BroadcastScoreUpdate(data map[string]interface{}) {
	h.broadcast <- WSMessage{
		Type:      MessageTypeScoreUpdate,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// BroadcastIRCompletion broadcasts an IR completion
func (h *Hub) BroadcastIRCompletion(data map[string]interface{}) {
	h.broadcast <- WSMessage{
		Type:      MessageTypeIRCompletion,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// BroadcastHeartbeatAlert broadcasts a heartbeat alert
func (h *Hub) BroadcastHeartbeatAlert(data map[string]interface{}) {
	h.broadcast <- WSMessage{
		Type:      MessageTypeHeartbeatAlert,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// BroadcastMisbehaviorReport broadcasts a misbehavior report
func (h *Hub) BroadcastMisbehaviorReport(data map[string]interface{}) {
	h.broadcast <- WSMessage{
		Type:      MessageTypeMisbehavior,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		c.handleMessage(msg, hub)
	}
}

// handleMessage handles incoming client messages
func (c *Client) handleMessage(msg WSMessage, hub *Hub) {
	switch msg.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(msg.Data)
	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(msg.Data)
	case MessageTypePing:
		c.Send <- WSMessage{
			Type:      MessageTypePong,
			Data:      make(map[string]interface{}),
			Timestamp: time.Now().Unix(),
		}
	default:
		log.Printf("Unknown message type: %s", msg.Type)
	}
}

// handleSubscribe handles subscription requests
func (c *Client) handleSubscribe(data map[string]interface{}) {
	channel, ok := data["channel"].(string)
	if !ok {
		log.Printf("Invalid channel in subscribe request")
		return
	}

	c.mu.Lock()
	c.Subscriptions[channel] = true
	c.mu.Unlock()

	log.Printf("Client %s subscribed to %s", c.ID, channel)

	// Send confirmation
	c.Send <- WSMessage{
		Type: "subscribed",
		Data: map[string]interface{}{
			"channel": channel,
		},
		Timestamp: time.Now().Unix(),
	}
}

// handleUnsubscribe handles unsubscribe requests
func (c *Client) handleUnsubscribe(data map[string]interface{}) {
	channel, ok := data["channel"].(string)
	if !ok {
		log.Printf("Invalid channel in unsubscribe request")
		return
	}

	c.mu.Lock()
	delete(c.Subscriptions, channel)
	c.mu.Unlock()

	log.Printf("Client %s unsubscribed from %s", c.ID, channel)

	// Send confirmation
	c.Send <- WSMessage{
		Type: "unsubscribed",
		Data: map[string]interface{}{
			"channel": channel,
		},
		Timestamp: time.Now().Unix(),
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from clients
func ServeWs(hub *Hub, conn *websocket.Conn, clientID string) {
	client := &Client{
		ID:            clientID,
		Conn:          conn,
		Send:          make(chan WSMessage, 256),
		Subscriptions: make(map[string]bool),
	}

	hub.register <- client

	go client.writePump()
	go client.readPump(hub)
}
