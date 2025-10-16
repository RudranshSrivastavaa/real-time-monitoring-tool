// services/websocket_service.go
package services

import (
	//"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/websocket"
	"monitoring-tool/models"
)

// WebSocketHub manages all WebSocket connections
type WebSocketHub struct {
	// Connected clients
	Clients map[*WebSocketClient]bool

	// Inbound messages from clients
	Broadcast chan models.WebSocketMessage

	// Register requests from clients
	Register chan *WebSocketClient

	// Unregister requests from clients
	Unregister chan *WebSocketClient

	// Mutex for thread-safe operations
	mutex sync.RWMutex
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	// WebSocket connection
	Conn *websocket.Conn

	// Buffered channel of outbound messages
	Send chan models.WebSocketMessage

	// Reference to the hub
	Hub *WebSocketHub

	// Client ID for tracking
	ID string
}

// WebSocket upgrader configuration
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from React development server
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" || origin == ""
	},
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		Clients:    make(map[*WebSocketClient]bool),
		Broadcast:  make(chan models.WebSocketMessage, 256),
		Register:   make(chan *WebSocketClient),
		Unregister: make(chan *WebSocketClient),
	}
}

// Run starts the WebSocket hub
func (h *WebSocketHub) Run() {
	log.Println("🔌 WebSocket hub started")
	
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			h.mutex.Unlock()
			
			log.Printf("📱 Client connected: %s (Total: %d)", client.ID, len(h.Clients))
			
			// Send welcome message
			welcome := models.WebSocketMessage{
				Type: "connection_established",
				Data: map[string]string{
					"status":  "connected",
					"message": "Real-time monitoring connected",
				},
			}
			
			select {
			case client.Send <- welcome:
			default:
				h.unregisterClient(client)
			}

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				h.mutex.Unlock()
				log.Printf("📱 Client disconnected: %s (Total: %d)", client.ID, len(h.Clients))
			} else {
				h.mutex.Unlock()
			}

		case message := <-h.Broadcast:
			h.mutex.RLock()
			clientCount := len(h.Clients)
			h.mutex.RUnlock()
			
			if clientCount > 0 {
				log.Printf("📡 Broadcasting to %d clients: %s", clientCount, message.Type)
			}
			
			h.mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					// Client's send buffer is full, remove the client
					h.mutex.RUnlock()
					h.unregisterClient(client)
					h.mutex.RLock()
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// unregisterClient safely removes a client
func (h *WebSocketHub) unregisterClient(client *WebSocketClient) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, ok := h.Clients[client]; ok {
		delete(h.Clients, client)
		close(client.Send)
		client.Conn.Close()
	}
}

// GetClientCount returns the number of connected clients
func (h *WebSocketHub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.Clients)
}

// BroadcastToAll sends a message to all connected clients
func (h *WebSocketHub) BroadcastToAll(messageType string, data interface{}) {
	message := models.WebSocketMessage{
		Type: messageType,
		Data: data,
	}
	
	select {
	case h.Broadcast <- message:
	default:
		log.Println("⚠️  Broadcast channel is full, message dropped")
	}
}

// NewWebSocketClient creates a new WebSocket client
func NewWebSocketClient(conn *websocket.Conn, hub *WebSocketHub, clientID string) *WebSocketClient {
	return &WebSocketClient{
		Conn: conn,
		Send: make(chan models.WebSocketMessage, 256),
		Hub:  hub,
		ID:   clientID,
	}
}

// WritePump handles writing messages to the WebSocket connection
func (c *WebSocketClient) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
for {
    select {
    case message, ok := <-c.Send:
        if !ok {
            c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
            return
        }

        // Set write deadline
        c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
        
        // Send JSON message
        if err := c.Conn.WriteJSON(message); err != nil {
            log.Printf("WebSocket write error: %v", err)
            return
        }
        
    case <-time.After(60 * time.Second):
        // Ping to keep connection alive
        if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
            log.Printf("WebSocket ping error: %v", err)
            return
        }
    }
}

}

// ReadPump handles reading messages from the WebSocket connection
func (c *WebSocketClient) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler for keep-alive
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message models.WebSocketMessage
		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Handle client messages (ping, subscribe to specific monitors, etc.)
		c.handleClientMessage(message)
	}
}


// handleClientMessage processes messages received from clients
func (c *WebSocketClient) handleClientMessage(message models.WebSocketMessage) {
	switch message.Type {
	case "ping":
		// Respond with pong
		pong := models.WebSocketMessage{
			Type: "pong",
			Data: map[string]interface{}{
				"timestamp": time.Now(),
				"client_id": c.ID,
			},
		}
		select {
		case c.Send <- pong:
		default:
		}
		
	case "subscribe_monitor":
		// Handle monitor-specific subscriptions (future feature)
		log.Printf("Client %s subscribed to monitor updates", c.ID)
		
	default:
		log.Printf("Unknown message type from client %s: %s", c.ID, message.Type)
	}
}