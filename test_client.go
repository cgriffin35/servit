package main

import (
    "log"
    "time"
    "github.com/gorilla/websocket"
)

func main() {
    conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/tunnel", nil)
    if err != nil {
        log.Fatal("Dial error:", err)
    }
    defer conn.Close()

    log.Println("Connected successfully!")

    // Send registration message (as required by your protocol)
    registerMsg := map[string]string{"tunnelId": "test123"}
    if err := conn.WriteJSON(registerMsg); err != nil {
        log.Fatal("Write error:", err)
    }

    log.Println("Registration sent. Keeping connection open for 5 seconds...")
    time.Sleep(5 * time.Second)
    log.Println("Test completed successfully!")
}