package tunnel

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Manager struct {
	tunnels sync.Map // map[string]*Tunnel (tunnelID -> Tunnel)
}

func NewManager() *Manager {
	return &Manager{}
}

// RegisterTunnel adds a new tunnel connection
func (m *Manager) RegisterTunnel(tunnelID string, wsConn *websocket.Conn) *Tunnel {
	tunnel := &Tunnel{
		ID:     tunnelID,
		WSConn: wsConn,
	}

	m.tunnels.Store(tunnelID, tunnel)
	return tunnel
}

// GetTunnel retrieves a tunnel by ID
func (m *Manager) GetTunnel(tunnelID string) (*Tunnel, bool) {
	value, ok := m.tunnels.Load(tunnelID)
	if !ok {
		return nil, false
	}
	return value.(*Tunnel), true
}

func (m *Manager) GetActiveTunnels() []string {
	var tunnelIDs []string
	m.tunnels.Range(func(key, value any) bool {
		tunnelIDs = append(tunnelIDs, key.(string))
		return true
	})
	return tunnelIDs
}

func (m *Manager) GetActiveTunnelCount() int {
	count := 0
	m.tunnels.Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}

// RemoveTunnel cleans up a disconnected tunnel
func (m *Manager) RemoveTunnel(tunnelID string) {
	if tunnel, ok := m.GetTunnel(tunnelID); ok {
		tunnel.WSConn.Close()
		m.tunnels.Delete(tunnelID)
	}
}

// HealthCheck removes stale connections
func (m *Manager) HealthCheck() {
	m.tunnels.Range(func(key, value any) bool {
		tunnel := value.(*Tunnel)
		// Set a write deadline to check if connection is still alive
		tunnel.WSConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := tunnel.WSConn.WriteMessage(websocket.PingMessage, nil); err != nil {
			m.RemoveTunnel(tunnel.ID)
		}
		return true
	})
}
