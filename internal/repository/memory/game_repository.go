package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/siddarth/go-app/internal/domain"
)

// GameRepository implements domain.GameRepository using in-memory storage
type GameRepository struct {
	games map[string]*domain.Game
	mu    sync.RWMutex
}

// NewGameRepository creates a new in-memory game repository
func NewGameRepository() *GameRepository {
	return &GameRepository{
		games: make(map[string]*domain.Game),
	}
}

// Save persists a game to memory
func (r *GameRepository) Save(ctx context.Context, game *domain.Game) error {
	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}
	if game.ID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.games[game.ID] = game
	return nil
}

// FindByID retrieves a game by ID
func (r *GameRepository) FindByID(ctx context.Context, id string) (*domain.Game, error) {
	if id == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	
	game, exists := r.games[id]
	if !exists {
		return nil, fmt.Errorf("game not found: %s", id)
	}
	
	return game, nil
}

// Delete removes a game from storage
func (r *GameRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.games, id)
	return nil
}

// Exists checks if a game exists
func (r *GameRepository) Exists(ctx context.Context, id string) bool {
	if id == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	
	_, exists := r.games[id]
	return exists
}

