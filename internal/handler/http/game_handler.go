package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siddarth/go-app/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// GameHandler handles HTTP requests for game operations
type GameHandler struct {
	gameService domain.GameService
	logger      *slog.Logger
	tracer      trace.Tracer
}

// NewGameHandler creates a new game handler
func NewGameHandler(gameService domain.GameService, logger *slog.Logger) *GameHandler {
	return &GameHandler{
		gameService: gameService,
		logger:      logger,
		tracer:      otel.Tracer("game-handler"),
	}
}

// StartGameRequest represents the start game request
type StartGameRequest struct {
	SessionID string `json:"sessionId,omitempty"`
}

// StartGameResponse represents the start game response
type StartGameResponse struct {
	SessionID string            `json:"sessionId"`
	State     domain.GameState  `json:"state"`
}

// MoveRequest represents a player move request
type MoveRequest struct {
	Direction string `json:"direction" binding:"required,oneof=up down left right"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

// RegisterRoutes registers all game routes
func (h *GameHandler) RegisterRoutes(r *gin.Engine) {
	// Serve static files
	r.Static("/static", "./static")

	// Serve index.html at root
	r.GET("/", h.ServeIndex)

	// Health check
	r.GET("/health", h.Health)

	// API routes
	api := r.Group("/api/game")
	{
		api.POST("/start", h.StartGame)
		api.GET("/state", h.GetGameState)
		api.POST("/move", h.MovePlayer)
		api.POST("/restart", h.RestartGame)
	}
}

// ServeIndex serves the index.html file
func (h *GameHandler) ServeIndex(c *gin.Context) {
	c.File("./static/index.html")
}

// Health handles health check requests
func (h *GameHandler) Health(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "Health")
	defer span.End()

	response := HealthResponse{
		Status:  "ok",
		Service: "pacman-game",
		Time:    time.Now().Format(time.RFC3339),
	}

	h.logger.DebugContext(ctx, "health check")
	c.JSON(http.StatusOK, response)
}

// StartGame handles starting a new game
func (h *GameHandler) StartGame(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "StartGame")
	defer span.End()

	// Get session ID from header or generate new one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Create game
	game, err := h.gameService.CreateGame(ctx, sessionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to create game",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusInternalServerError, "Failed to create game", err)
		return
	}

	// Start game loop
	if err := h.gameService.StartGameLoop(ctx, sessionID); err != nil {
		h.logger.ErrorContext(ctx, "failed to start game loop",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusInternalServerError, "Failed to start game loop", err)
		return
	}

	// Get game state
	state := game.ToGameState(20, 15) // Using constants from service

	response := StartGameResponse{
		SessionID: sessionID,
		State:     state,
	}

	h.logger.InfoContext(ctx, "game started",
		"session_id", sessionID,
	)

	c.JSON(http.StatusOK, response)
}

// GetGameState handles retrieving game state
func (h *GameHandler) GetGameState(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "GetGameState")
	defer span.End()

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		h.respondError(c, http.StatusBadRequest, "Session ID required", nil)
		return
	}

	span.SetAttributes(attribute.String("session.id", sessionID))

	state, err := h.gameService.GetGameState(ctx, sessionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get game state",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusNotFound, "Game not found", err)
		return
	}

	c.JSON(http.StatusOK, state)
}

// MovePlayer handles player movement
func (h *GameHandler) MovePlayer(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "MovePlayer")
	defer span.End()

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		h.respondError(c, http.StatusBadRequest, "Session ID required", nil)
		return
	}

	span.SetAttributes(attribute.String("session.id", sessionID))

	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WarnContext(ctx, "invalid move request",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("direction", req.Direction))

	// Parse direction
	dir, ok := domain.ParseDirection(req.Direction)
	if !ok {
		h.respondError(c, http.StatusBadRequest, "Invalid direction", nil)
		return
	}

	// Set player direction
	if err := h.gameService.SetPlayerDirection(ctx, sessionID, dir); err != nil {
		h.logger.ErrorContext(ctx, "failed to set player direction",
			"session_id", sessionID,
			"direction", req.Direction,
			"error", err,
		)
		h.respondError(c, http.StatusNotFound, "Game not found", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RestartGame handles restarting a game
func (h *GameHandler) RestartGame(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "RestartGame")
	defer span.End()

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
	}

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Restart game
	game, err := h.gameService.RestartGame(ctx, sessionID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to restart game",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusInternalServerError, "Failed to restart game", err)
		return
	}

	// Start game loop
	if err := h.gameService.StartGameLoop(ctx, sessionID); err != nil {
		h.logger.ErrorContext(ctx, "failed to start game loop",
			"session_id", sessionID,
			"error", err,
		)
		h.respondError(c, http.StatusInternalServerError, "Failed to start game loop", err)
		return
	}

	// Get game state
	state := game.ToGameState(20, 15)

	response := StartGameResponse{
		SessionID: sessionID,
		State:     state,
	}

	h.logger.InfoContext(ctx, "game restarted",
		"session_id", sessionID,
	)

	c.JSON(http.StatusOK, response)
}

// respondError sends an error response
func (h *GameHandler) respondError(c *gin.Context, statusCode int, message string, err error) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	if err != nil {
		h.logger.Error("handler error",
			"status", statusCode,
			"message", message,
			"error", err,
		)
	}

	c.JSON(statusCode, response)
}

