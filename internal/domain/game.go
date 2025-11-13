package domain

import (
	"context"
	"time"
)

// Direction represents movement direction
type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
	DirectionNone
)

// String returns the string representation of direction
func (d Direction) String() string {
	switch d {
	case DirectionUp:
		return "up"
	case DirectionDown:
		return "down"
	case DirectionLeft:
		return "left"
	case DirectionRight:
		return "right"
	default:
		return "none"
	}
}

// ParseDirection converts string to Direction
func ParseDirection(s string) (Direction, bool) {
	switch s {
	case "up":
		return DirectionUp, true
	case "down":
		return DirectionDown, true
	case "left":
		return DirectionLeft, true
	case "right":
		return DirectionRight, true
	default:
		return DirectionNone, false
	}
}

// Position represents a coordinate on the game board
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Move calculates new position based on direction
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

// Equals checks if two positions are the same
func (p Position) Equals(other Position) bool {
	return p.X == other.X && p.Y == other.Y
}

// Ghost represents a ghost entity in the game
type Ghost struct {
	Position  Position
	Direction Direction
}

// Game represents the core game entity
type Game struct {
	ID         string
	Board      [][]rune
	Player     Position
	Ghosts     []Ghost
	Score      int
	DotsLeft   int
	GameOver   bool
	PlayerDir  Direction
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// GameState represents the serializable game state for API responses
type GameState struct {
	Board    [][]string `json:"board"`
	Player   Position   `json:"player"`
	Ghosts   []Position `json:"ghosts"`
	Score    int        `json:"score"`
	DotsLeft int        `json:"dotsLeft"`
	GameOver bool       `json:"gameOver"`
	Won      bool       `json:"won"`
}

// ToGameState converts Game to GameState
func (g *Game) ToGameState(width, height int) GameState {
	board := make([][]string, height)
	for i := 0; i < height; i++ {
		board[i] = make([]string, width)
		for j := 0; j < width; j++ {
			board[i][j] = string(g.Board[i][j])
		}
	}

	ghostPositions := make([]Position, len(g.Ghosts))
	for i, ghost := range g.Ghosts {
		ghostPositions[i] = ghost.Position
	}

	return GameState{
		Board:    board,
		Player:   g.Player,
		Ghosts:   ghostPositions,
		Score:    g.Score,
		DotsLeft: g.DotsLeft,
		GameOver: g.GameOver,
		Won:      g.DotsLeft == 0,
	}
}

// IsValidPosition checks if a position is valid and not a wall
func (g *Game) IsValidPosition(pos Position, width, height int) bool {
	if pos.X < 0 || pos.X >= width || pos.Y < 0 || pos.Y >= height {
		return false
	}
	return g.Board[pos.Y][pos.X] != '#'
}

// GameService defines the interface for game business logic
type GameService interface {
	// CreateGame creates a new game session
	CreateGame(ctx context.Context, sessionID string) (*Game, error)
	
	// GetGame retrieves a game by session ID
	GetGame(ctx context.Context, sessionID string) (*Game, error)
	
	// SetPlayerDirection sets the player's movement direction
	SetPlayerDirection(ctx context.Context, sessionID string, dir Direction) error
	
	// GetGameState retrieves the current game state
	GetGameState(ctx context.Context, sessionID string) (*GameState, error)
	
	// RestartGame restarts a game session
	RestartGame(ctx context.Context, sessionID string) (*Game, error)
	
	// DeleteGame removes a game session
	DeleteGame(ctx context.Context, sessionID string) error
	
	// StartGameLoop starts the game loop for a session
	StartGameLoop(ctx context.Context, sessionID string) error
}

// GameRepository defines the interface for game storage
type GameRepository interface {
	// Save persists a game to storage
	Save(ctx context.Context, game *Game) error
	
	// FindByID retrieves a game by ID
	FindByID(ctx context.Context, id string) (*Game, error)
	
	// Delete removes a game from storage
	Delete(ctx context.Context, id string) error
	
	// Exists checks if a game exists
	Exists(ctx context.Context, id string) bool
}

