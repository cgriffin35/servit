# Servit Tunnel Server

A high-performance, scalable reverse proxy tunnel server written in Go. This server enables secure public access to locally hosted web services by establishing persistent WebSocket tunnels from clients (like Android devices) to this public-facing server.

## Features
- WebSocket Tunneling: Secure, persistent connections for reliable tunneling

- HTTP Reverse Proxy: Seamless HTTP request/response forwarding

- Multi-Client Support: Handle thousands of simultaneous tunnel connections

- Automatic Health Checks: Built-in connection monitoring and cleanup

- CORS Enabled: Cross-origin resource sharing for web compatibility

- Simple Protocol: Easy-to-implement JSON-based communication protocol

- High Performance: Built with Netty-inspired Gorilla WebSocket for optimal performance

## Architecture
```
Public Internet → Go Server (VPS) → WebSocket Tunnel → Android Client → Local Server
Local Server → Android Client → WebSocket Tunnel → Go Server → Public Internet
```
## Installation
### Prerequisites
- Go 1.18+
- Public VPS (DigitalOcean, AWS, Linode, etc.)
- Domain name (optional, but recommended)

### 1. Clone and Build
```
# Clone the repository
git clone <your-repo-url>
cd servit-tunnel-server

# Install dependencies
go mod download

# Build the server
go build -o tunnel-server cmd/server/main.go

# Alternatively, run directly
go run cmd/server/main.go

```

### 2. Configuration
The server uses environment variables for configuration:
```bash
# Copy example environment file
cp .env.example .env

# Edit with your settings
PORT=8080
ALLOWED_ORIGINS=*
LOG_LEVEL=info
```

### 3. Deployment
```bash
# Run on your VPS (listens on all interfaces)
./tunnel-server

# Or run with specific port
PORT=9090 ./tunnel-server

# Run in background with nohup
nohup ./tunnel-server > server.log 2>&1 &
```

## Project Structure
```
tunnel-server/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── tunnel/
│   │   ├── manager.go           # Tunnel connection management
│   │   └── types.go             # Data structures and types
│   ├── proxy/
│   │   └── http_handler.go      # HTTP request handling
│   └── websocket/
│       └── connection.go        # WebSocket connection handling
├── pkg/
│   └── utils/
│       ├── requestid.go         # Unique request ID generation
│       └── serialization.go     # HTTP serialization utilities
├── go.mod
├── go.sum
└── README.md
```

## API Endpoints
### WebSocket Connection
- Endpoint: /tunnel
- Method: WebSocket upgrade
- Purpose: Establish persistent tunnel connection

### HTTP Proxy
- Endpoint: /{tunnelId} and /{tunnelId}/{path:.*}
- Method: Any HTTP method
- Purpose: Proxy HTTP requests through tunnels

##  Protocol Specification
### Client Registration
First message from client must be a JSON registration:
```json
{
  "tunnelId": "unique-client-id"
}
```

### HTTP Request Forwarding
Server → Client format:
```json
{
  "requestId": "abc123def456",
  "method": "GET",
  "url": "http://example.com/path",
  "headers": {
    "Content-Type": ["application/json"],
    "User-Agent": ["curl/7.68.0"]
  },
  "body": "base64-encoded-data"  // For binary content
}
```

### HTTP Response Format
Client → Server format:
```json
{
  "requestId": "abc123def456",
  "statusCode": 200,
  "headers": {
    "Content-Type": ["text/html"],
    "Cache-Control": ["no-cache"]
  },
  "body": "<html>...</html>",    // String for text content
  "isBase64": false              // true if body is base64-encoded
}
```
## Health Checking
The server automatically performs health checks:

```go
// Health check runs every 30 seconds
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        tunnelManager.HealthCheck()
    }
}()
```
Unresponsive connections are automatically cleaned up.

## Security Considerations
### Current Security
- CORS configured for development (AllowedOrigins: ["*"])
- WebSocket origin checking disabled for testing

### Production Security Checklist
- Enable proper CORS restrictions
- Implement authentication/authorization
- Add TLS/SSL termination
- Rate limiting per client
- Request validation and sanitization
- API key authentication for tunnel registration
- IP whitelisting/blacklisting

## Monitoring and Logging
### Log Levels
- DEBUG: Detailed connection information
- INFO: Operational messages
- WARN: Non-critical issues
- ERROR: Critical failures

### Key Metrics to Monitor
- Active tunnel connections
- Request/response throughput
- Error rates
- Memory usage
- Connection latency

## Testing
### Unit Tests
```bash
go test ./internal/... -v
```
### Integration Testing
```bash
# Test WebSocket connection
wscat -c ws://localhost:8080/tunnel

# Test HTTP proxy
curl -v http://localhost:8080/test-tunnel/health
```
### Load Testing
```bash
# Install hey load testing tool
go install github.com/rakyll/hey@latest

# Test with 1000 requests
hey -n 1000 http://localhost:8080/test-tunnel/
```
## Performance Optimization
### Current Optimizations
- Goroutine-based concurrency
- Sync.Map for thread-safe tunnel management
- Connection pooling and reuse
- Memory-efficient serialization
### Further Optimizations
- Add connection pooling
- Implement response caching
- Add compression for large responses
- Implement load balancing for multiple instances

## Troubleshooting
### Common Issues
1. WebSocket Connection Failures
```bash
# Check if port is open
netstat -tlnp | grep 8080

# Test WebSocket manually
wscat -c ws://your-server:8080/tunnel
```
2. Tunnel Not Found Errors
- Verify client registration message format
- Check tunnel ID matching
3. Timeout Issues
- Increase timeout durations in http_handler.go
- Check network latency between client and server

## Debug Mode
Enable debug logging by setting `LOG_LEVEL=debug` environment variable.

## Scaling
### Vertical Scaling
- Increase server resources (CPU, RAM)
- Optimize Go garbage collection settings
### Horizontal Scaling
- Deploy multiple instances behind load balancer
- Use shared Redis store for tunnel management
- Implement sticky sessions for WebSocket connections

## Client Implementation
### Android Client Requirements
- WebSocket client library
- JSON serialization/deserialization
- Local HTTP server integration
- Base64 encoding for binary content

### Example Client Flow
1. Establish WebSocket connection to /tunnel
2. Send registration message with unique tunnel ID
3. Listen for incoming HTTP requests
4. Forward requests to local server
5. Send responses back through WebSocket

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Roadmap
- TLS/SSL termination
- Authentication system
- Rate limiting
- Metrics and monitoring dashboard
- Docker containerization
- Kubernetes deployment manifests
- CLI management tool
- Web administration interface

## Acknowledgments
- Gorilla WebSocket for excellent WebSocket implementation
- Gorilla Mux for HTTP routing
- Go standard library for robust networking primitives

