# Migration Guide: Legacy to Clean Architecture

## Overview

This guide helps you migrate from the legacy single-file `main.go` implementation to the new Clean Architecture structure.

## What Changed?

### Before (Legacy)
```
main.go (462 lines)
├── All code in one file
├── No separation of concerns
├── Tightly coupled to Gin
├── No observability
├── Manual CORS middleware
├── Deprecated rand usage
├── Goroutine leaks
└── Hard to test
```

### After (Refactored)
```
cmd/server/main.go          # Entry point only
internal/
  ├── domain/               # Business entities
  ├── service/              # Business logic
  ├── repository/           # Data access
  ├── handler/http/         # HTTP layer
  ├── middleware/           # Cross-cutting concerns
  └── config/               # Configuration
pkg/
  └── observability/        # Shared utilities
```

## Key Improvements

### 1. ✅ Architecture
- **Before**: Everything in one file
- **After**: Clean separation into layers (domain, service, repository, handler)

### 2. ✅ Context Propagation
- **Before**: No context usage
- **After**: Context passed through all layers for cancellation and tracing

### 3. ✅ Observability
- **Before**: No structured logging, no tracing
- **After**: Structured logging with `slog` and distributed tracing with OpenTelemetry

### 4. ✅ Error Handling
- **Before**: Minimal error handling
- **After**: Comprehensive error handling with wrapped errors and context

### 5. ✅ Goroutine Management
- **Before**: Game loop goroutines could leak
- **After**: Proper cancellation with context, no leaks

### 6. ✅ Configuration
- **Before**: Hardcoded values
- **After**: Environment-based configuration with validation

### 7. ✅ Random Number Generation
- **Before**: Deprecated `rand.Seed()` and global `rand`
- **After**: Local `rand.New()` instance per service

### 8. ✅ Code Duplication
- **Before**: Move logic duplicated 3 times
- **After**: Extracted to `Position.Move()` method

### 9. ✅ Testability
- **Before**: Hard to test due to tight coupling
- **After**: Interface-driven design, easy to mock and test

### 10. ✅ Graceful Shutdown
- **Before**: Abrupt termination
- **After**: Graceful shutdown with resource cleanup

## Migration Steps

### Step 1: Update Dependencies

```bash
# The go.mod has been updated with new dependencies
go mod tidy
```

### Step 2: Build the New Application

```bash
# Build the refactored version
go build -o pacman-game cmd/server/main.go
```

### Step 3: Test the New Application

```bash
# Run the new version
./pacman-game

# Or run directly
go run cmd/server/main.go
```

### Step 4: Verify API Compatibility

The API endpoints remain the same:
- ✅ `POST /api/game/start` - Start new game
- ✅ `GET /api/game/state` - Get game state
- ✅ `POST /api/game/move` - Move player
- ✅ `POST /api/game/restart` - Restart game
- ✅ `GET /health` - Health check
- ✅ `GET /` - Serve UI

Test each endpoint to ensure compatibility:

```bash
# Health check
curl http://localhost:8080/health

# Start game
curl -X POST http://localhost:8080/api/game/start

# Move player (use session ID from start response)
curl -X POST \
  -H "X-Session-ID: <session-id>" \
  -H "Content-Type: application/json" \
  -d '{"direction": "up"}' \
  http://localhost:8080/api/game/move
```

### Step 5: Configure for Your Environment

```bash
# Production configuration example
PORT=8080 \
GIN_MODE=release \
LOG_LEVEL=info \
LOG_FORMAT=json \
TRACING_ENABLED=true \
./pacman-game
```

### Step 6: Remove Legacy Code

Once you've verified the new version works:

```bash
# Backup the old main.go
mv main.go main.go.legacy

# Update any scripts that referenced the old main.go
# to use cmd/server/main.go instead
```

## Code Comparison

### Random Number Generation

**Before:**
```go
func main() {
    rand.Seed(time.Now().UnixNano()) // Deprecated
}

func (g *game) moveGhosts() {
    if rand.Intn(100) < 30 { // Global rand
        // ...
    }
}
```

**After:**
```go
type gameService struct {
    rng *rand.Rand
}

func NewGameService(...) domain.GameService {
    return &gameService{
        rng: rand.New(rand.NewSource(time.Now().UnixNano())),
    }
}

func (s *gameService) moveGhosts(game *domain.Game) {
    if s.rng.Intn(100) < 30 { // Instance-specific rand
        // ...
    }
}
```

### Move Logic

**Before:**
```go
// Duplicated 3 times in the code
switch dir {
case up:
    newPos.Y--
case down:
    newPos.Y++
case left:
    newPos.X--
case right:
    newPos.X++
}
```

**After:**
```go
// Single method, reused everywhere
func (p Position) Move(dir Direction) Position {
    newPos := p
    switch dir {
    case DirectionUp:
        newPos.Y--
    case DirectionDown:
        newPos.Y++
    case DirectionLeft:
        newPos.X--
    case DirectionRight:
        newPos.X++
    }
    return newPos
}

// Usage
newPos := game.Player.Move(game.PlayerDir)
```

### Game Loop

**Before:**
```go
func (gm *gameManager) runGameLoop(sessionID string) {
    ticker := time.NewTicker(200 * time.Millisecond)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            // No way to cancel from outside
            // Potential goroutine leak
        }
    }
}
```

**After:**
```go
func (s *gameService) runGameLoop(ctx context.Context, sessionID string) {
    ticker := time.NewTicker(GameTickInterval)
    defer ticker.Stop()
    defer s.cleanupGameLoop(sessionID)

    for {
        select {
        case <-ctx.Done():
            return // Proper cancellation
        case <-ticker.C:
            // Game tick with error handling
        }
    }
}
```

### CORS Middleware

**Before:**
```go
// Manual CORS implementation (13 lines)
r.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    // ... many lines of manual header setting
    c.Next()
})
```

**After:**
```go
// Using proper CORS library (clean and tested)
func CORS() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Session-ID"},
        AllowCredentials: true,
    }
    return cors.New(config)
}
```

### Error Handling

**Before:**
```go
// Minimal error handling
game := gameMgr.getGame(sessionID)
if game == nil {
    c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
    return
}
```

**After:**
```go
// Comprehensive error handling with logging and tracing
state, err := h.gameService.GetGameState(ctx, sessionID)
if err != nil {
    h.logger.ErrorContext(ctx, "failed to get game state",
        "session_id", sessionID,
        "error", err,
    )
    h.respondError(c, http.StatusNotFound, "Game not found", err)
    return
}
```

## Testing Strategy

### Unit Tests Example

**Domain:**
```go
func TestPositionMove(t *testing.T) {
    pos := domain.Position{X: 5, Y: 5}
    newPos := pos.Move(domain.DirectionUp)
    assert.Equal(t, 4, newPos.Y)
    assert.Equal(t, 5, newPos.X)
}
```

**Service:**
```go
func TestCreateGame(t *testing.T) {
    repo := &mockGameRepository{}
    service := NewGameService(repo, logger)
    
    game, err := service.CreateGame(ctx, "test-session")
    assert.NoError(t, err)
    assert.NotNil(t, game)
    assert.Equal(t, "test-session", game.ID)
}
```

**Handler:**
```go
func TestStartGameHandler(t *testing.T) {
    mockService := &mockGameService{}
    handler := NewGameHandler(mockService, logger)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    
    handler.StartGame(c)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

## Deployment Changes

### Docker

Update your Dockerfile to use the new entry point:

**Before:**
```dockerfile
CMD ["./main"]
```

**After:**
```dockerfile
CMD ["./pacman-game"]
# Or during build:
RUN go build -o pacman-game cmd/server/main.go
```

### Kubernetes

Update your deployments to use environment variables:

```yaml
env:
- name: PORT
  value: "8080"
- name: LOG_LEVEL
  value: "info"
- name: TRACING_ENABLED
  value: "true"
- name: GIN_MODE
  value: "release"
```

## Rollback Plan

If you need to rollback to the legacy version:

1. Keep the old `main.go` as `main.go.legacy`
2. Test the new version thoroughly in staging
3. Have monitoring in place to detect issues
4. If issues occur:
   ```bash
   mv main.go main.go.new
   mv main.go.legacy main.go
   go build -o pacman-game main.go
   ./pacman-game
   ```

## Common Issues

### Issue 1: Module Path
**Problem**: Import errors due to module path
**Solution**: Ensure `go.mod` has `module github.com/siddarth/go-app`

### Issue 2: Missing Dependencies
**Problem**: `go build` fails with missing packages
**Solution**: Run `go mod tidy`

### Issue 3: Port Already in Use
**Problem**: Server fails to start
**Solution**: Change port via environment: `PORT=9000 ./pacman-game`

### Issue 4: Static Files Not Found
**Problem**: UI doesn't load
**Solution**: Ensure `static/` directory exists in the same directory as the binary

## Performance Comparison

| Metric | Before | After |
|--------|--------|-------|
| Lines of Code (main) | 462 | 120 (entry point only) |
| Testability | Hard | Easy (interface-driven) |
| Goroutine Leaks | Yes | No (context cancellation) |
| Error Handling | Minimal | Comprehensive |
| Observability | None | Full (logging + tracing) |
| Configuration | Hardcoded | Environment-based |
| Code Duplication | High | Low |
| Separation of Concerns | None | Complete |

## Support

For questions or issues during migration:
1. Check [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed architecture docs
2. Check [README_NEW.md](./README_NEW.md) for usage documentation
3. Review code comments in the refactored files
4. Run tests to verify behavior: `go test ./...`

## Conclusion

The refactored version provides:
- ✅ Better maintainability
- ✅ Improved testability
- ✅ Production-ready observability
- ✅ Proper resource management
- ✅ Type safety
- ✅ Modern Go patterns
- ✅ Scalable architecture

The migration is **backward compatible** - all API endpoints work the same way!

