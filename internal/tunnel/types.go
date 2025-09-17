package tunnel

import (
	"sync"

	"github.com/gorilla/websocket"
)

// ResponseChannel is where the HTTP handler waits for the tunnel response
type ResponseChannel chan *ProxyResponse

// Tunnel represents an active connection from a client
type Tunnel struct {
	ID             string
	WSConn         *websocket.Conn
	ActiveRequests sync.Map // map[string]chan<- *ProxyResponse
}

// ProxyRequest represents an HTTP request to be forwarded to the client
type ProxyRequest struct {
	RequestID string              `json:"requestId"`
	Method    string              `json:"method"`
	URL       string              `json:"url"`
	Headers   map[string][]string `json:"headers"` // This uses net/http types indirectly
	Body      []byte              `json:"body,omitempty"`
}

// ProxyResponse represents an HTTP response from the client
type ProxyResponse struct {
	RequestID  string              `json:"requestId"`
	StatusCode int                 `json:"statusCode"`
	Headers    map[string][]string `json:"headers"` // This uses net/http types indirectly
	Body       string              `json:"body,omitempty"`
    IsBase64   bool                `json:"isBase64,omitempty"` 
}
