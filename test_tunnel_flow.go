package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    "github.com/gorilla/websocket"
)

func main() {
    // Connect and register
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/tunnel", nil)
    if err != nil {
        log.Fatal("Dial error:", err)
    }
    defer conn.Close()

    // Register our tunnel
    if err := conn.WriteJSON(map[string]string{"tunnelId": "test123"}); err != nil {
        log.Fatal("Write error:", err)
    }

    log.Println("Tunnel registered. Waiting for requests...")

    // Set up graceful shutdown
    interrupt := make(chan os.Signal, 1)
    signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

    // Listen for incoming requests
    go func() {
        for {
            var request map[string]interface{}
            if err := conn.ReadJSON(&request); err != nil {
                log.Println("Read error (probably normal shutdown):", err)
                return
            }
            
            log.Printf("Received request: %+v", request)
            
            // Simulate processing and send a response
            response := map[string]interface{}{
                "requestId":  request["requestId"],
                "statusCode": 200,
                "headers":    map[string]interface{}{"Content-Type": []string{"application/json"}},
                "body":       []byte(`{"message": "Hello from test client!", "requestId": "` + request["requestId"].(string) + `"}`),
            }
            
            if err := conn.WriteJSON(response); err != nil {
                log.Println("Write response error:", err)
                return
            }
            
            log.Printf("Sent response for request: %s", request["requestId"])
        }
    }()

    // Make test HTTP request after a delay
    go func() {
        time.Sleep(2 * time.Second)
        log.Println("Making test HTTP request...")
        resp, err := http.Get("http://localhost:8080/test123/api/test")
        if err != nil {
            log.Println("HTTP request error:", err)
            return
        }
        defer resp.Body.Close()
        
        log.Printf("HTTP Response Status: %s", resp.Status)
    }()

    // Wait for interrupt signal
    <-interrupt
    log.Println("Shutting down gracefully...")
    
    // Send close message
    conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
    time.Sleep(1 * time.Second)
}