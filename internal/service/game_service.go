package service

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/siddarth/go-app/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const (
	// GameWidth is the width of the game board
	GameWidth = 20
	// GameHeight is the height of the game board
	GameHeight = 15
	// GameTickInterval is the interval between game ticks
	GameTickInterval = 200 * time.Millisecond
	// ScorePerDot is the score awarded for collecting a dot
	ScorePerDot = 10
)

// gameService implements domain.GameService
type gameService struct {
	repo          domain.GameRepository
	logger        *slog.Logger
	tracer        trace.Tracer
	gameLoops     map[string]context.CancelFunc
	gameLoopMu    sync.RWMutex
	rng           *rand.Rand
}

// NewGameService creates a new game service
func NewGameService(repo domain.GameRepository, logger *slog.Logger) domain.GameService {
	return &gameService{
		repo:       repo,
		logger:     logger,
		tracer:     otel.Tracer("game-service"),
		gameLoops:  make(map[string]context.CancelFunc),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateGame creates a new game session
func (s *gameService) CreateGame(ctx context.Context, sessionID string) (*domain.Game, error) {
	ctx, span := s.tracer.Start(ctx, "CreateGame")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	if sessionID == "" {
		err := fmt.Errorf("session ID cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	game := s.initializeGame(sessionID)

	if err := s.repo.Save(ctx, game); err != nil {
		s.logger.ErrorContext(ctx, "failed to save game",
			"session_id", sessionID,
			"error", err,
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save game")
		return nil, fmt.Errorf("failed to save game: %w", err)
	}

	s.logger.InfoContext(ctx, "game created",
		"session_id", sessionID,
		"dots_count", game.DotsLeft,
	)

	return game, nil
}

// initializeGame creates a new game with initial state
func (s *gameService) initializeGame(sessionID string) *domain.Game {
	game := &domain.Game{
		ID:        sessionID,
		Board:     make([][]rune, GameHeight),
		Player:    domain.Position{X: 1, Y: 1},
		Ghosts: []domain.Ghost{
			{Position: domain.Position{X: GameWidth - 2, Y: GameHeight - 2}, Direction: domain.DirectionLeft},
			{Position: domain.Position{X: GameWidth - 2, Y: 1}, Direction: domain.DirectionLeft},
			{Position: domain.Position{X: 1, Y: GameHeight - 2}, Direction: domain.DirectionRight},
		},
		Score:     0,
		PlayerDir: domain.DirectionNone,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initialize board with maze
	maze := []string{
		"####################",
		"#..................#",
		"#.##.##.##.##.##.###",
		"#..................#",
		"#.##.##....##.##.###",
		"#......##.##......##",
		"#.##.##....##.##.###",
		"#..................#",
		"#.##.##.##.##.##.###",
		"#..................#",
		"#.##....##....##.###",
		"#......##.##......##",
		"#.##....##....##.###",
		"#..................#",
		"####################",
	}

	for i := 0; i < GameHeight; i++ {
		game.Board[i] = make([]rune, GameWidth)
		mazeRow := maze[i]
		for j := 0; j < GameWidth; j++ {
			if j < len(mazeRow) {
				game.Board[i][j] = rune(mazeRow[j])
				if mazeRow[j] == '.' {
					game.DotsLeft++
				}
			} else {
				game.Board[i][j] = '#'
			}
		}
	}

	return game
}

// GetGame retrieves a game by session ID
func (s *gameService) GetGame(ctx context.Context, sessionID string) (*domain.Game, error) {
	ctx, span := s.tracer.Start(ctx, "GetGame")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	if sessionID == "" {
		err := fmt.Errorf("session ID cannot be empty")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	game, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "game not found")
		return nil, fmt.Errorf("game not found: %w", err)
	}

	return game, nil
}

// SetPlayerDirection sets the player's movement direction
func (s *gameService) SetPlayerDirection(ctx context.Context, sessionID string, dir domain.Direction) error {
	ctx, span := s.tracer.Start(ctx, "SetPlayerDirection")
	defer span.End()

	span.SetAttributes(
		attribute.String("session.id", sessionID),
		attribute.String("direction", dir.String()),
	)

	game, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "game not found")
		return fmt.Errorf("game not found: %w", err)
	}

	game.PlayerDir = dir
	game.UpdatedAt = time.Now()

	if err := s.repo.Save(ctx, game); err != nil {
		s.logger.ErrorContext(ctx, "failed to update game",
			"session_id", sessionID,
			"error", err,
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to save game")
		return fmt.Errorf("failed to update game: %w", err)
	}

	return nil
}

// GetGameState retrieves the current game state
func (s *gameService) GetGameState(ctx context.Context, sessionID string) (*domain.GameState, error) {
	ctx, span := s.tracer.Start(ctx, "GetGameState")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	game, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "game not found")
		return nil, fmt.Errorf("game not found: %w", err)
	}

	state := game.ToGameState(GameWidth, GameHeight)
	
	span.SetAttributes(
		attribute.Int("score", state.Score),
		attribute.Int("dots_left", state.DotsLeft),
		attribute.Bool("game_over", state.GameOver),
		attribute.Bool("won", state.Won),
	)

	return &state, nil
}

// RestartGame restarts a game session
func (s *gameService) RestartGame(ctx context.Context, sessionID string) (*domain.Game, error) {
	ctx, span := s.tracer.Start(ctx, "RestartGame")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Stop existing game loop
	s.stopGameLoop(sessionID)

	// Delete old game
	if err := s.repo.Delete(ctx, sessionID); err != nil {
		s.logger.WarnContext(ctx, "failed to delete old game",
			"session_id", sessionID,
			"error", err,
		)
	}

	// Create new game
	return s.CreateGame(ctx, sessionID)
}

// DeleteGame removes a game session
func (s *gameService) DeleteGame(ctx context.Context, sessionID string) error {
	ctx, span := s.tracer.Start(ctx, "DeleteGame")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Stop game loop
	s.stopGameLoop(sessionID)

	// Delete from repository
	if err := s.repo.Delete(ctx, sessionID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete game",
			"session_id", sessionID,
			"error", err,
		)
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to delete game")
		return fmt.Errorf("failed to delete game: %w", err)
	}

	s.logger.InfoContext(ctx, "game deleted", "session_id", sessionID)
	return nil
}

// StartGameLoop starts the game loop for a session
func (s *gameService) StartGameLoop(ctx context.Context, sessionID string) error {
	ctx, span := s.tracer.Start(ctx, "StartGameLoop")
	defer span.End()

	span.SetAttributes(attribute.String("session.id", sessionID))

	// Check if game exists
	if !s.repo.Exists(ctx, sessionID) {
		err := fmt.Errorf("game not found: %s", sessionID)
		span.RecordError(err)
		span.SetStatus(codes.Error, "game not found")
		return err
	}

	// Create cancellable context for the game loop
	loopCtx, cancel := context.WithCancel(context.Background())

	s.gameLoopMu.Lock()
	// Stop existing loop if any
	if existingCancel, exists := s.gameLoops[sessionID]; exists {
		existingCancel()
	}
	s.gameLoops[sessionID] = cancel
	s.gameLoopMu.Unlock()

	// Start game loop in goroutine
	go s.runGameLoop(loopCtx, sessionID)

	s.logger.InfoContext(ctx, "game loop started", "session_id", sessionID)
	return nil
}

// runGameLoop runs the game loop until context is cancelled or game ends
func (s *gameService) runGameLoop(ctx context.Context, sessionID string) {
	ticker := time.NewTicker(GameTickInterval)
	defer ticker.Stop()
	defer s.cleanupGameLoop(sessionID)

	s.logger.Info("game loop running", "session_id", sessionID)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("game loop stopped", "session_id", sessionID)
			return
		case <-ticker.C:
			if err := s.gameTick(ctx, sessionID); err != nil {
				s.logger.Error("game tick failed",
					"session_id", sessionID,
					"error", err,
				)
				return
			}
		}
	}
}

// gameTick performs one game tick
func (s *gameService) gameTick(ctx context.Context, sessionID string) error {
	game, err := s.repo.FindByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("game not found: %w", err)
	}

	// Stop if game is over or won
	if game.GameOver || game.DotsLeft == 0 {
		s.logger.Info("game ended",
			"session_id", sessionID,
			"game_over", game.GameOver,
			"won", game.DotsLeft == 0,
		)
		return fmt.Errorf("game ended")
	}

	// Move player
	s.movePlayer(game)

	// Move ghosts
	s.moveGhosts(game)

	// Check collisions
	s.checkCollisions(game)

	// Update timestamp
	game.UpdatedAt = time.Now()

	// Save game state
	if err := s.repo.Save(ctx, game); err != nil {
		return fmt.Errorf("failed to save game: %w", err)
	}

	return nil
}

// movePlayer moves the player based on current direction
func (s *gameService) movePlayer(game *domain.Game) {
	if game.PlayerDir == domain.DirectionNone {
		return
	}

	newPos := game.Player.Move(game.PlayerDir)

	if game.IsValidPosition(newPos, GameWidth, GameHeight) {
		game.Player = newPos

		// Collect dot
		if game.Board[game.Player.Y][game.Player.X] == '.' {
			game.Board[game.Player.Y][game.Player.X] = ' '
			game.Score += ScorePerDot
			game.DotsLeft--
		}
	}
}

// moveGhosts moves all ghosts with AI behavior
func (s *gameService) moveGhosts(game *domain.Game) {
	for i := range game.Ghosts {
		ghost := &game.Ghosts[i]
		dir := ghost.Direction

		// 30% chance to change direction randomly
		if s.rng.Intn(100) < 30 {
			dir = domain.Direction(s.rng.Intn(4))
		} else {
			// Try to move towards player
			dx := game.Player.X - ghost.Position.X
			dy := game.Player.Y - ghost.Position.Y

			if abs(dx) > abs(dy) {
				if dx > 0 {
					dir = domain.DirectionRight
				} else {
					dir = domain.DirectionLeft
				}
			} else {
				if dy > 0 {
					dir = domain.DirectionDown
				} else {
					dir = domain.DirectionUp
				}
			}
		}

		newPos := ghost.Position.Move(dir)

		if game.IsValidPosition(newPos, GameWidth, GameHeight) {
			ghost.Position = newPos
			ghost.Direction = dir
		} else {
			// Try random direction if current doesn't work
			dirs := []domain.Direction{
				domain.DirectionUp,
				domain.DirectionDown,
				domain.DirectionLeft,
				domain.DirectionRight,
			}
			s.rng.Shuffle(len(dirs), func(i, j int) {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			})
			for _, d := range dirs {
				newPos := ghost.Position.Move(d)
				if game.IsValidPosition(newPos, GameWidth, GameHeight) {
					ghost.Position = newPos
					ghost.Direction = d
					break
				}
			}
		}
	}
}

// checkCollisions checks if player collided with any ghost
func (s *gameService) checkCollisions(game *domain.Game) {
	for _, ghost := range game.Ghosts {
		if game.Player.Equals(ghost.Position) {
			game.GameOver = true
			s.logger.Info("game over - collision",
				"session_id", game.ID,
				"player_position", game.Player,
				"ghost_position", ghost.Position,
			)
			return
		}
	}
}

// stopGameLoop stops the game loop for a session
func (s *gameService) stopGameLoop(sessionID string) {
	s.gameLoopMu.Lock()
	defer s.gameLoopMu.Unlock()

	if cancel, exists := s.gameLoops[sessionID]; exists {
		cancel()
		delete(s.gameLoops, sessionID)
	}
}

// cleanupGameLoop cleans up game loop resources
func (s *gameService) cleanupGameLoop(sessionID string) {
	s.gameLoopMu.Lock()
	defer s.gameLoopMu.Unlock()

	delete(s.gameLoops, sessionID)
}

// abs returns absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

