// Package websocket provides WebSocket connection management for real-time communication.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	upgrader websocket.Upgrader
	clients  map[string]*websocket.Conn
	register chan *Client
	unregister chan *Client
	broadcast chan []byte
}

// Client represents a WebSocket client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	Send     chan []byte
	AgentID  string
	LastPing time.Time
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients:    make(map[string]*websocket.Conn),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run starts the WebSocket manager
func (ws *WebSocketManager) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.clients[client.ID] = client.Conn
			fmt.Printf("Client %s connected\n", client.ID)

		case client := <-ws.unregister:
			if conn, ok := ws.clients[client.ID]; ok {
				delete(ws.clients, client.ID)
				conn.Close()
				fmt.Printf("Client %s disconnected\n", client.ID)
			}

		case message := <-ws.broadcast:
			for id, conn := range ws.clients {
				err := conn.WriteMessage(websocket.TextMessage, message)
				if err != nil {
					fmt.Printf("Error sending message to client %s: %v\n", id, err)
					conn.Close()
					delete(ws.clients, id)
				}
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (ws *WebSocketManager) HandleWebSocket(c *gin.Context) {
	conn, err := ws.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	clientID := c.Query("client_id")
	agentID := c.Query("agent_id")

	client := &Client{
		ID:       clientID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		AgentID:  agentID,
		LastPing: time.Now(),
	}

	ws.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(ws)
}

// writePump pumps messages from the websocket connection to the hub
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

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
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

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump(ws *WebSocketManager) {
	defer func() {
		ws.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.LastPing = time.Now()
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// Handle incoming message
		ws.handleMessage(c, message)
	}
}

// handleMessage processes incoming WebSocket messages
func (ws *WebSocketManager) handleMessage(client *Client, message []byte) {
	// TODO: Parse and handle different message types
	fmt.Printf("Received message from client %s: %s\n", client.ID, string(message))
	
	// Echo back for now
	client.Send <- message
}

// BroadcastMessage sends a message to all connected clients
func (ws *WebSocketManager) BroadcastMessage(message []byte) {
	ws.broadcast <- message
}

// SendToAgent sends a message to a specific agent
func (ws *WebSocketManager) SendToAgent(agentID string, message []byte) {
	for id, conn := range ws.clients {
		// TODO: Match by agent ID instead of client ID
		if id == agentID {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Printf("Error sending message to agent %s: %v\n", agentID, err)
			}
		}
	}
}

// GetConnectedAgents returns list of connected agent IDs
func (ws *WebSocketManager) GetConnectedAgents() []string {
	var agents []string
	for id := range ws.clients {
		agents = append(agents, id)
	}
	return agents
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	AgentID   string                 `json:"agent_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewWebSocketMessage creates a new WebSocket message
func NewWebSocketMessage(msgType, agentID string, data map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		Type:      msgType,
		AgentID:   agentID,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// ToJSON converts message to JSON
func (m *WebSocketMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

