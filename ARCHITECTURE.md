# Architecture Documentation

## Overview

This application is a Pac-Man game server built using **Clean Architecture** principles in Go. The architecture emphasizes separation of concerns, testability, and maintainability.

## Architecture Layers

### 1. Domain Layer (`internal/domain/`)

The innermost layer containing business entities and interfaces. This layer has no dependencies on external frameworks or libraries.

**Files:**
- `game.go`: Core domain entities (Game, Position, Direction, Ghost) and service interfaces

**Key Principles:**
- Pure business logic
- Framework-agnostic
- Defines interfaces for repositories and services
- No external dependencies

### 2. Repository Layer (`internal/repository/`)

Implements data access and storage logic.

**Files:**
- `memory/game_repository.go`: In-memory implementation of GameRepository interface

**Key Features:**
- Implements domain.GameRepository interface
- Thread-safe with sync.RWMutex
- Can be easily replaced with database implementation

### 3. Service Layer (`internal/service/`)

Contains business logic and use cases.

**Files:**
- `game_service.go`: Implements game business logic, AI, and game loop

**Key Features:**
- Implements domain.GameService interface
- Game initialization and state management
- Player and ghost movement logic
- Collision detection
- Game loop management with context cancellation
- OpenTelemetry tracing integration

### 4. Handler Layer (`internal/handler/http/`)

Handles HTTP requests and responses.

**Files:**
- `game_handler.go`: HTTP handlers for game operations

**Key Features:**
- Framework-specific code isolated here
- Request validation
- Response formatting
- Delegates to service layer
- OpenTelemetry span creation

### 5. Middleware Layer (`internal/middleware/`)

Contains HTTP middleware components.

**Files:**
- `cors.go`: CORS configuration using gin-contrib/cors
- `logging.go`: Structured request logging
- `tracing.go`: OpenTelemetry distributed tracing
- `recovery.go`: Panic recovery middleware

### 6. Configuration Layer (`internal/config/`)

Manages application configuration.

**Files:**
- `config.go`: Configuration structure and environment variable loading

**Key Features:**
- Environment-based configuration
- Validation logic
- Type-safe configuration access

### 7. Observability Package (`pkg/observability/`)

Shared observability utilities.

**Files:**
- `logger.go`: Structured logger setup using slog
- `tracing.go`: OpenTelemetry initialization

### 8. Entry Point (`cmd/server/`)

Application entry point with dependency wiring.

**Files:**
- `main.go`: Application bootstrap, dependency injection, graceful shutdown

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── domain/
│   │   └── game.go              # Domain entities and interfaces
│   ├── handler/
│   │   └── http/
│   │       └── game_handler.go  # HTTP handlers
│   ├── middleware/
│   │   ├── cors.go              # CORS middleware
│   │   ├── logging.go           # Logging middleware
│   │   ├── recovery.go          # Recovery middleware
│   │   └── tracing.go           # Tracing middleware
│   ├── repository/
│   │   └── memory/
│   │       └── game_repository.go # In-memory storage
│   └── service/
│       └── game_service.go      # Business logic
├── pkg/
│   └── observability/
│       ├── logger.go            # Logger setup
│       └── tracing.go           # Tracing setup
├── static/
│   └── index.html               # Frontend
├── go.mod
├── go.sum
└── main.go                      # Legacy file (to be removed)
```

## Key Design Patterns

### 1. Dependency Injection
All dependencies are injected through constructors, making the code testable and flexible.

```go
gameRepo := memory.NewGameRepository()
gameService := service.NewGameService(gameRepo, logger)
gameHandler := httphandler.NewGameHandler(gameService, logger)
```

### 2. Interface-Driven Development
All layers interact through interfaces, not concrete implementations.

```go
type GameService interface {
    CreateGame(ctx context.Context, sessionID string) (*Game, error)
    // ... other methods
}
```

### 3. Context Propagation
`context.Context` is passed through all layers for:
- Request cancellation
- Deadline propagation
- Distributed tracing
- Request-scoped values

### 4. Graceful Shutdown
The server supports graceful shutdown with configurable timeout:
- Stops accepting new connections
- Waits for in-flight requests to complete
- Cleans up resources (game loops, tracers)

### 5. Game Loop Management
Each game session has its own game loop goroutine:
- Context-based cancellation
- Automatic cleanup on game end
- Prevention of goroutine leaks

## Observability

### Structured Logging
Uses Go's `log/slog` for structured, JSON-formatted logs:
- Consistent log format
- Contextual information
- Multiple log levels (debug, info, warn, error)

### Distributed Tracing
OpenTelemetry integration for tracing:
- Automatic span creation in middleware
- Service and handler-level spans
- Span attributes for debugging
- Trace propagation across services

### Configuration
Controlled via environment variables:
- `TRACING_ENABLED`: Enable/disable tracing
- `LOG_LEVEL`: Set log level
- `LOG_FORMAT`: json or text

## Security Features

### Input Validation
- Request binding with validation tags
- Direction validation
- Session ID validation
- Error messages don't leak sensitive information

### Error Handling
- All errors are wrapped with context
- Errors logged with appropriate levels
- Generic error messages sent to clients
- Detailed errors in logs

### Resource Management
- Game loops properly cancelled to prevent resource leaks
- Context timeouts on all operations
- Graceful shutdown prevents data loss

## Testing Strategy

### Unit Tests
Test individual components in isolation:
- Domain entities and methods
- Service business logic
- Repository operations
- Handler request/response handling

### Integration Tests
Test component interactions:
- Service + Repository integration
- Handler + Service integration
- End-to-end API tests

### Mocking
Use interfaces for easy mocking:
```go
type mockGameRepository struct {
    mock.Mock
}

func (m *mockGameRepository) Save(ctx context.Context, game *domain.Game) error {
    args := m.Called(ctx, game)
    return args.Error(0)
}
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `release` |
| `LOG_LEVEL` | Log level | `info` |
| `LOG_FORMAT` | Log format (json/text) | `json` |
| `SERVICE_NAME` | Service name for tracing | `pacman-game` |
| `SERVICE_VERSION` | Service version | `1.0.0` |
| `ENVIRONMENT` | Environment name | `development` |
| `TRACING_ENABLED` | Enable tracing | `true` |
| `READ_TIMEOUT` | HTTP read timeout | `30s` |
| `WRITE_TIMEOUT` | HTTP write timeout | `30s` |
| `SHUTDOWN_TIMEOUT` | Graceful shutdown timeout | `10s` |

## Running the Application

### Development
```bash
go run cmd/server/main.go
```

### Production
```bash
# Build
go build -o pacman-game cmd/server/main.go

# Run
./pacman-game
```

### With Custom Configuration
```bash
PORT=9000 LOG_LEVEL=debug TRACING_ENABLED=false go run cmd/server/main.go
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Serve game UI |
| GET | `/health` | Health check |
| POST | `/api/game/start` | Start new game |
| GET | `/api/game/state` | Get game state |
| POST | `/api/game/move` | Move player |
| POST | `/api/game/restart` | Restart game |

## Future Improvements

1. **Database Integration**: Replace in-memory repository with Redis/PostgreSQL
2. **WebSocket Support**: Real-time game updates
3. **Metrics**: Add Prometheus metrics
4. **Authentication**: Add JWT-based authentication
5. **Rate Limiting**: Implement distributed rate limiting
6. **Leaderboard**: Add persistent leaderboard
7. **Multi-player**: Support for multiplayer games
8. **Circuit Breakers**: Add resilience patterns
9. **API Documentation**: Add OpenAPI/Swagger docs
10. **E2E Tests**: Add comprehensive integration tests

## Maintenance

### Adding New Features
1. Define domain entities/interfaces in `internal/domain/`
2. Implement business logic in `internal/service/`
3. Add repository methods if needed in `internal/repository/`
4. Create HTTP handlers in `internal/handler/http/`
5. Wire dependencies in `cmd/server/main.go`

### Replacing Dependencies
Thanks to interface-driven design, replacing implementations is straightforward:
- Replace in-memory repo with database repo
- Replace stdout trace exporter with Jaeger/OTLP
- Replace Gin with another HTTP framework (isolated in handler layer)

## References

- [Clean Architecture by Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [OpenTelemetry Go Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

