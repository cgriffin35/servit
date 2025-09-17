package proxy

import (
	"encoding/base64"
	"io"

	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/cgriffin35/servit/internal/tunnel"
	"github.com/cgriffin35/servit/pkg/utils"
	"github.com/gorilla/mux"
)

type Handler struct {
	tunnelManager *tunnel.Manager
	timeout       time.Duration
}

func NewHandler(tm *tunnel.Manager) *Handler {
	return &Handler{
		tunnelManager: tm,
		timeout:       30 * time.Second, // Request timeout
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tunnelID := vars["tunnelId"]
	path := vars["path"]

	log.Printf("HTTP Request received: Method=%s, URL=%s, TunnelID=%s, Path=%s",
		r.Method, r.URL.String(), tunnelID, path)

	tunnelConn, exists := h.tunnelManager.GetTunnel(tunnelID)
	if !exists {
		log.Printf("Tunnel NOT found: %s", tunnelID)
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}

	log.Printf("Tunnel found: %s", tunnelID)

	// Create response channel
	responseChan := make(tunnel.ResponseChannel, 1)
	requestID, err := utils.GenerateRequestID()
	if err != nil {
		log.Printf("Failed to generate request ID: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Created response channel for request: %s", requestID)

	// Store the channel for this request
	tunnelConn.ActiveRequests.Store(requestID, responseChan)
	defer tunnelConn.ActiveRequests.Delete(requestID)

	// Serialize and send request
	proxyReq, err := h.serializeHTTPRequest(r, requestID)
	if err != nil {
		log.Printf("Failed to serialize request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Printf("Sending request to tunnel: %s", proxyReq.RequestID)

	// Reset the deadline after successful write
	tunnelConn.WSConn.SetWriteDeadline(time.Time{})

	if err := tunnelConn.WSConn.WriteJSON(proxyReq); err != nil {
		log.Printf("WebSocket write failed: %v", err)

		// Send proper WebSocket close message before removing
		closeMsg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Tunnel connection failed")
		tunnelConn.WSConn.WriteMessage(websocket.CloseMessage, closeMsg)

		// Give time for close message to be sent
		time.Sleep(100 * time.Millisecond)

		h.tunnelManager.RemoveTunnel(tunnelID)
		http.Error(w, "Tunnel connection failed", http.StatusBadGateway)
		return
	}

	log.Printf("Waiting for response for request: %s (timeout: 30s)", requestID)

	// Wait for response with timeout
	select {
	case response := <-responseChan:
		log.Printf("✅ SUCCESS: Received response for request: %s, Status: %d",
			response.RequestID, response.StatusCode)

		// Write response back to HTTP client
		for key, values := range response.Headers {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(response.StatusCode)

		// Handle base64 encoded content
		if response.IsBase64 {
			// Decode base64 content
			decoded, err := base64.StdEncoding.DecodeString(response.Body)
			if err != nil {
				log.Printf("Failed to decode base64: %v", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Write(decoded)
		} else {
			// Regular text content
			w.Write([]byte(response.Body))
		}

	case <-time.After(30 * time.Second):
		log.Printf("❌ TIMEOUT: No response received for request: %s", requestID)
		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
	}
}

func (h *Handler) serializeHTTPRequest(r *http.Request, requestID string) (*tunnel.ProxyRequest, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	return &tunnel.ProxyRequest{
		RequestID: requestID,
		Method:    r.Method,
		URL:       r.URL.String(),
		Headers:   r.Header,
		Body:      body,
	}, nil
}
