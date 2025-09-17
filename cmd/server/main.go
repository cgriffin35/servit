package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cgriffin35/servit/internal/proxy"
	"github.com/cgriffin35/servit/internal/tunnel"
	"github.com/cgriffin35/servit/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Initialize core components
	tunnelManager := tunnel.NewManager()
	wsHandler := websocket.NewHandler(tunnelManager)
	proxyHandler := proxy.NewHandler(tunnelManager)

	// Set up router
	router := mux.NewRouter()

	// WebSocket endpoint for tunnel connections
	router.HandleFunc("/tunnel", wsHandler.HandleConnection)

	// Public HTTP traffic (catch-all for any path under a tunnel ID)
	router.HandleFunc("/{tunnelId}", proxyHandler.HandleRequest)
	router.HandleFunc("/{tunnelId}/{path:.*}", proxyHandler.HandleRequest)

	// Add CORS for development
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	})
	handler := c.Handler(router)

	// Start health check routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			tunnelManager.HealthCheck()
		}
	}()

	// Start server
	log.Println("Tunnel server starting on :8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", handler))
}
