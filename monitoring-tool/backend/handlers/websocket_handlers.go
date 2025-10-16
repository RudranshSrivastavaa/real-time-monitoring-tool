// handlers/websocket_handlers.go
package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	//"github.com/gorilla/websocket"

	"monitoring-tool/services"
)

type WebSocketHandler struct {
	hub *services.WebSocketHub
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *services.WebSocketHub) *WebSocketHandler {
	return &WebSocketHandler{
		hub: hub,
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
    // Use the upgrader defined in services/websocket_service.go
    conn, err := services.Upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "Failed to upgrade connection to WebSocket",
            "details": err.Error(),
        })
        return
    }

	// Generate client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Create new WebSocket client
	client := services.NewWebSocketClient(conn, h.hub, clientID)

	// Register client with hub
	h.hub.Register <- client

	// Start client goroutines
	go client.WritePump()
	go client.ReadPump()
}