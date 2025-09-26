package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cgriffin35/servit/internal/middleware"
	"github.com/cgriffin35/servit/internal/proxy"
	"github.com/cgriffin35/servit/internal/tunnel"
	"github.com/cgriffin35/servit/internal/websocket"
	"github.com/cgriffin35/servit/pkg/config"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	cfg := config.Load()

	// Initialize core components
	tunnelManager := tunnel.NewManager()
	wsHandler := websocket.NewHandler(tunnelManager)
	proxyHandler := proxy.NewHandler(tunnelManager)

	// Set up router
	router := mux.NewRouter()
	router.Use(middleware.Recovery)
	router.Use(middleware.RateLimit(100))

	// WebSocket endpoint for tunnel connections
	router.HandleFunc("/tunnel", wsHandler.HandleConnection)

	// Public HTTP traffic (catch-all for any path under a tunnel ID)
	router.HandleFunc("/{tunnelId}", proxyHandler.HandleRequest)
	router.HandleFunc("/{tunnelId}/{path:.*}", proxyHandler.HandleRequest)

	// Add CORS for development
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://" + cfg.Domain, "https://*." + cfg.Domain},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
	handler := c.Handler(router)

	// Configure server timeouts
	// server := &http.Server{
	// 	Addr:         ":" + strconv.Itoa(cfg.Port),
	// 	Handler:      handler,
	// 	ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
	// 	WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	// 	IdleTimeout:  120 * time.Second,
	// }

	// Start health check routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			tunnelManager.HealthCheck()
		}
	}()

	port := 80
	addr := ":" + strconv.Itoa(port)

	log.Printf("Server starting on port %d for domain %s", port, cfg.Domain)

	// Start server
	log.Printf("Server starting on port %d for domain %s", cfg.Port, cfg.Domain)
	log.Fatal(http.ListenAndServe(addr, handler))
}
