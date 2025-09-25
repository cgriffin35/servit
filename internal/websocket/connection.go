package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cgriffin35/servit/internal/tunnel"
	"github.com/gorilla/websocket"
)

type Handler struct {
	tunnelManager *tunnel.Manager
	upgrader      websocket.Upgrader
}

func NewHandler(tm *tunnel.Manager) *Handler {
	return &Handler{
		tunnelManager: tm,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// For development, allow all origins. Restrict in production!
				origin := r.Header.Get("Origin")
        return strings.HasSuffix(origin, "servit.app") || 
               strings.HasSuffix(origin, ".servit.app")
			},
		},
	}
}

func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket connection established from: %s", r.RemoteAddr)

	// First message must be the tunnel registration
	var registerMsg struct {
		TunnelID string `json:"tunnelId"`
	}

	if err := conn.ReadJSON(&registerMsg); err != nil {
		log.Printf("Failed to read registration message: %v", err)
		conn.WriteJSON(map[string]string{"error": "Send tunnelId first: " + err.Error()})
		return
	}

	log.Printf("Registration message received: %s", registerMsg.TunnelID)

	tunnelConn := h.tunnelManager.RegisterTunnel(registerMsg.TunnelID, conn)
	log.Printf("Tunnel registered: %s", tunnelConn.ID)

	conn.SetPingHandler(func(message string) error {
		log.Printf("Received ping from client %s", tunnelConn.ID)
		return conn.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(time.Second))
	})

	// Set up pong handler
	conn.SetPongHandler(func(message string) error {
		log.Printf("Received pong from client %s", tunnelConn.ID)
		return nil
	})

	// Listen for responses from the client
	for {
		// Log raw message first to see what's coming in
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Client %s disconnected: %v", tunnelConn.ID, err)
			h.tunnelManager.RemoveTunnel(tunnelConn.ID)
			break
		}

		log.Printf("Raw message received: %s", string(messageBytes))

		// Try to parse as JSON
		var response tunnel.ProxyResponse
		if err := json.Unmarshal(messageBytes, &response); err != nil {
			log.Printf("Failed to parse message as ProxyResponse: %v", err)
			continue
		}

		log.Printf("Parsed response for request: %s", response.RequestID)

		if channelInterface, exists := tunnelConn.ActiveRequests.Load(response.RequestID); exists {
			channel := channelInterface.(tunnel.ResponseChannel)
			log.Printf("Forwarding response to HTTP handler for request: %s", response.RequestID)
			channel <- &response
		} else {
			log.Printf("Received response for unknown request ID: %s", response.RequestID)
		}
	}

}
