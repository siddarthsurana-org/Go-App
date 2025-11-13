# Pac-Man Game Server - Refactored

A production-ready Pac-Man game server built with Go, following Clean Architecture principles and modern best practices.

## 🚀 Features

- **Clean Architecture**: Modular, testable, and maintainable codebase
- **Observability**: Structured logging and distributed tracing with OpenTelemetry
- **Graceful Shutdown**: Proper cleanup of resources on shutdown
- **Configuration Management**: Environment-based configuration
- **Type-Safe**: Strong typing throughout the application
- **Thread-Safe**: Concurrent game sessions with proper synchronization
- **RESTful API**: Well-designed HTTP endpoints
- **Middleware Stack**: CORS, logging, tracing, and recovery middleware

## 📋 Prerequisites

- Go 1.21 or higher
- A modern web browser

## 🛠️ Installation

```bash
# Clone the repository
git clone <repository-url>
cd Go-App/MlQDc

# Install dependencies
go mod download

# Build the application
go build -o pacman-game cmd/server/main.go
```

## 🎮 Running the Application

### Development Mode

```bash
# Run with default configuration
go run cmd/server/main.go

# Run with debug logging
LOG_LEVEL=debug go run cmd/server/main.go

# Run on custom port
PORT=9000 go run cmd/server/main.go
```

### Production Mode

```bash
# Build the binary
go build -o pacman-game cmd/server/main.go

# Run the binary
./pacman-game

# Run with custom configuration
PORT=8080 GIN_MODE=release LOG_FORMAT=json ./pacman-game
```

## 🌐 Access the Game

Once the server is running, open your browser and navigate to:
```
http://localhost:8080
```

## 📚 API Documentation

### Health Check
```bash
curl http://localhost:8080/health
```

### Start New Game
```bash
curl -X POST http://localhost:8080/api/game/start
```

Response:
```json
{
  "sessionId": "session-1234567890",
  "state": {
    "board": [...],
    "player": {"x": 1, "y": 1},
    "ghosts": [...],
    "score": 0,
    "dotsLeft": 100,
    "gameOver": false,
    "won": false
  }
}
```

### Get Game State
```bash
curl -H "X-Session-ID: session-1234567890" \
  http://localhost:8080/api/game/state
```

### Move Player
```bash
curl -X POST \
  -H "X-Session-ID: session-1234567890" \
  -H "Content-Type: application/json" \
  -d '{"direction": "up"}' \
  http://localhost:8080/api/game/move
```

Valid directions: `up`, `down`, `left`, `right`

### Restart Game
```bash
curl -X POST \
  -H "X-Session-ID: session-1234567890" \
  http://localhost:8080/api/game/restart
```

## ⚙️ Configuration

Configure the application using environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (`debug` or `release`) | `release` |
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `LOG_FORMAT` | Log format (`json` or `text`) | `json` |
| `SERVICE_NAME` | Service name for tracing | `pacman-game` |
| `SERVICE_VERSION` | Service version | `1.0.0` |
| `ENVIRONMENT` | Environment name | `development` |
| `TRACING_ENABLED` | Enable OpenTelemetry tracing | `true` |
| `READ_TIMEOUT` | HTTP read timeout | `30s` |
| `WRITE_TIMEOUT` | HTTP write timeout | `30s` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `10s` |

Example:
```bash
PORT=9000 \
LOG_LEVEL=debug \
LOG_FORMAT=text \
TRACING_ENABLED=false \
go run cmd/server/main.go
```

## 🏗️ Architecture

The application follows Clean Architecture with clear separation of concerns:

```
cmd/server/          # Application entry point
internal/
  config/           # Configuration management
  domain/           # Business entities and interfaces
  service/          # Business logic
  repository/       # Data access layer
  handler/http/     # HTTP handlers
  middleware/       # HTTP middleware
pkg/
  observability/    # Logging and tracing utilities
```

For detailed architecture documentation, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/service/...
```

## 📊 Observability

### Structured Logging

The application uses Go's `log/slog` for structured logging:

```bash
# JSON format (default)
{"time":"2024-11-13T10:30:00Z","level":"INFO","msg":"server listening","port":"8080"}

# Text format
LOG_FORMAT=text go run cmd/server/main.go
```

### Distributed Tracing

OpenTelemetry tracing is enabled by default. Traces are exported to stdout for development.

To disable tracing:
```bash
TRACING_ENABLED=false go run cmd/server/main.go
```

For production, configure an OTLP endpoint:
```bash
TRACING_ENDPOINT=http://jaeger:4318 go run cmd/server/main.go
```

## 🔒 Security

- **Input Validation**: All inputs are validated
- **Error Handling**: Errors don't leak sensitive information
- **CORS**: Configurable CORS policies
- **Graceful Degradation**: Handles errors without crashing
- **Resource Cleanup**: Proper cleanup prevents resource leaks

## 🐳 Docker Support

```bash
# Build Docker image
docker build -t pacman-game .

# Run container
docker run -p 8080:8080 pacman-game

# Run with custom configuration
docker run -p 9000:9000 -e PORT=9000 -e LOG_LEVEL=debug pacman-game
```

## 📝 Development

### Code Organization

- **Domain Layer**: Pure business logic, no external dependencies
- **Service Layer**: Use cases and business logic implementation
- **Repository Layer**: Data access and storage
- **Handler Layer**: HTTP request/response handling
- **Middleware**: Cross-cutting concerns (logging, tracing, CORS)

### Adding New Features

1. Define domain entities and interfaces in `internal/domain/`
2. Implement business logic in `internal/service/`
3. Add data access in `internal/repository/`
4. Create HTTP handlers in `internal/handler/http/`
5. Wire dependencies in `cmd/server/main.go`

### Code Quality

```bash
# Format code
go fmt ./...

# Run linters
golangci-lint run

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🚀 Deployment

### Binary Deployment

```bash
# Build for production
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o pacman-game \
  cmd/server/main.go

# Run
PORT=8080 GIN_MODE=release ./pacman-game
```

### Kubernetes Deployment

Helm charts and Kubernetes manifests are available in:
- `helm/todo-app/` - Helm chart
- `k8s/` - Kubernetes manifests
- `argocd/` - ArgoCD configurations

```bash
# Deploy with Helm
helm install pacman-game ./helm/todo-app

# Deploy with kubectl
kubectl apply -f k8s/
```

## 🤝 Contributing

1. Follow Go best practices and conventions
2. Write tests for new features
3. Update documentation
4. Run linters and formatters
5. Ensure all tests pass

## 📄 License

[Your License Here]

## 📧 Contact

[Your Contact Information]

## 🙏 Acknowledgments

- Clean Architecture by Robert C. Martin
- OpenTelemetry community
- Go community

## 🔄 Migration from Legacy

The old `main.go` contained all code in a single file. The refactored version:

✅ **Improvements:**
- Separated concerns into layers
- Added observability (logging, tracing)
- Implemented graceful shutdown
- Added context propagation
- Fixed goroutine leaks
- Added proper error handling
- Made code testable
- Added configuration management
- Used modern Go patterns (slog, context)

To migrate:
1. Use the new entry point: `go run cmd/server/main.go`
2. The API endpoints remain the same
3. Configure via environment variables
4. The old `main.go` can be removed

## 📖 Additional Resources

- [ARCHITECTURE.md](./ARCHITECTURE.md) - Detailed architecture documentation
- [Go Documentation](https://golang.org/doc/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Gin Framework](https://gin-gonic.com/)

