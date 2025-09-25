package proxy

import (
	"encoding/base64"
	"io"
	"os"

	"log"
	"log/syslog"
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

func setupLogger() *log.Logger {
	// For production, log to syslog
	logger, err := syslog.NewLogger(syslog.LOG_INFO, 0)
	if err != nil {
			// Fallback to stdout
			return log.New(os.Stdout, "TUNNEL: ", log.Ldate|log.Ltime|log.Lshortfile)
	}
	return logger
}

func NewHandler(tm *tunnel.Manager) *Handler {
	return &Handler{
		tunnelManager: tm,
		timeout:       30 * time.Second, // Request timeout
	}
}

func (h *Handler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	tunnelID := vars["tunnelId"]

	tunnelConn, exists := h.tunnelManager.GetTunnel(tunnelID)
	if !exists {
		http.Error(w, "Tunnel not found", http.StatusNotFound)
		return
	}


	// Create response channel
	responseChan := make(tunnel.ResponseChannel, 1)
	requestID, err := utils.GenerateRequestID()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store the channel for this request
	tunnelConn.ActiveRequests.Store(requestID, responseChan)
	defer tunnelConn.ActiveRequests.Delete(requestID)

	// Serialize and send request
	proxyReq, err := h.serializeHTTPRequest(r, requestID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Reset the deadline after successful write
	tunnelConn.WSConn.SetWriteDeadline(time.Time{})

	if err := tunnelConn.WSConn.WriteJSON(proxyReq); err != nil {

		// Send proper WebSocket close message before removing
		closeMsg := websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "Tunnel connection failed")
		tunnelConn.WSConn.WriteMessage(websocket.CloseMessage, closeMsg)

		// Give time for close message to be sent
		time.Sleep(100 * time.Millisecond)

		h.tunnelManager.RemoveTunnel(tunnelID)
		http.Error(w, "Tunnel connection failed", http.StatusBadGateway)
		return
	}

	// Wait for response with timeout
	select {
	case response := <-responseChan:

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
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			w.Write(decoded)
		} else {
			// Regular text content
			w.Write([]byte(response.Body))
		}

	case <-time.After(30 * time.Second):
		http.Error(w, "Request timeout", http.StatusGatewayTimeout)
	}

	log.Printf("HTTP %s %s %s %d %v", 
        r.Method, 
        r.URL.Path, 
        r.RemoteAddr, 
				r.Response.StatusCode, 
        time.Since(start))
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
