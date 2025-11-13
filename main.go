package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	width  = 20
	height = 15
)

type direction int

const (
	up direction = iota
	down
	left
	right
	none
)

type position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type gameState struct {
	Board    [][]string `json:"board"`
	Player   position   `json:"player"`
	Ghosts   []position `json:"ghosts"`
	Score    int        `json:"score"`
	DotsLeft int        `json:"dotsLeft"`
	GameOver bool       `json:"gameOver"`
	Won      bool       `json:"won"`
}

type game struct {
	board     [][]rune
	player    position
	ghosts    []position
	ghostDirs []direction
	score     int
	dotsLeft  int
	gameOver  bool
	playerDir direction
	mu        sync.RWMutex
	rng       *rand.Rand
}

type gameManager struct {
	games map[string]*game
	mu    sync.RWMutex
}

func newGameManager() *gameManager {
	return &gameManager{
		games: make(map[string]*game),
	}
}

func (gm *gameManager) getGame(sessionID string) *game {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.games[sessionID]
}

func (gm *gameManager) createGame(sessionID string) *game {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	g := newGame()
	gm.games[sessionID] = g
	return g
}

func (gm *gameManager) deleteGame(sessionID string) {
	gm.mu.Lock()
	defer gm.mu.Unlock()
	delete(gm.games, sessionID)
}

func newGame() *game {
	g := &game{
		board:     make([][]rune, height),
		player:    position{X: 1, Y: 1},
		ghosts:    []position{{X: width - 2, Y: height - 2}, {X: width - 2, Y: 1}, {X: 1, Y: height - 2}},
		ghostDirs: []direction{left, left, right},
		score:     0,
		playerDir: none,
		rng:       rand.New(rand.NewSource(time.Now().UnixNano())),
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

	for i := 0; i < height; i++ {
		g.board[i] = make([]rune, width)
		mazeRow := maze[i]
		for j := 0; j < width; j++ {
			if j < len(mazeRow) {
				g.board[i][j] = rune(mazeRow[j])
				if mazeRow[j] == '.' {
					g.dotsLeft++
				}
			} else {
				g.board[i][j] = '#'
			}
		}
	}

	return g
}

func (g *game) isValidMove(pos position) bool {
	if pos.X < 0 || pos.X >= width || pos.Y < 0 || pos.Y >= height {
		return false
	}
	return g.board[pos.Y][pos.X] != '#'
}

func (g *game) setDirection(dir direction) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.playerDir = dir
}

func (g *game) movePlayer() {
	g.mu.Lock()
	defer g.mu.Unlock()

	newPos := g.player

	switch g.playerDir {
	case up:
		newPos.Y--
	case down:
		newPos.Y++
	case left:
		newPos.X--
	case right:
		newPos.X++
	}

	if g.isValidMove(newPos) {
		g.player = newPos

		// Collect dot
		if g.board[g.player.Y][g.player.X] == '.' {
			g.board[g.player.Y][g.player.X] = ' '
			g.score += 10
			g.dotsLeft--
		}
	}
}

func (g *game) moveGhosts() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for i := range g.ghosts {
		ghost := g.ghosts[i]
		var dir direction

		// 30% chance to change direction randomly
		if g.rng.Intn(100) < 30 {
			dir = direction(g.rng.Intn(4))
		} else {
			// Try to move towards player using current direction as base
			dir = g.ghostDirs[i]
			dx := g.player.X - ghost.X
			dy := g.player.Y - ghost.Y

			if abs(dx) > abs(dy) {
				if dx > 0 {
					dir = right
				} else {
					dir = left
				}
			} else {
				if dy > 0 {
					dir = down
				} else {
					dir = up
				}
			}
		}

		newPos := ghost
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

		if g.isValidMove(newPos) {
			g.ghosts[i] = newPos
			g.ghostDirs[i] = dir
		} else {
			// Try random direction if current doesn't work
			dirs := []direction{up, down, left, right}
			g.rng.Shuffle(len(dirs), func(i, j int) {
				dirs[i], dirs[j] = dirs[j], dirs[i]
			})
			for _, d := range dirs {
				newPos := ghost
				switch d {
				case up:
					newPos.Y--
				case down:
					newPos.Y++
				case left:
					newPos.X--
				case right:
					newPos.X++
				}
				if g.isValidMove(newPos) {
					g.ghosts[i] = newPos
					g.ghostDirs[i] = d
					break
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (g *game) checkCollisions() {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, ghost := range g.ghosts {
		if g.player.X == ghost.X && g.player.Y == ghost.Y {
			g.gameOver = true
			return
		}
	}
}

func (g *game) checkWin() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.dotsLeft == 0
}

func (g *game) getState() gameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	board := make([][]string, height)
	for i := 0; i < height; i++ {
		board[i] = make([]string, width)
		for j := 0; j < width; j++ {
			board[i][j] = string(g.board[i][j])
		}
	}

	return gameState{
		Board:    board,
		Player:   g.player,
		Ghosts:   g.ghosts,
		Score:    g.score,
		DotsLeft: g.dotsLeft,
		GameOver: g.gameOver,
		Won:      g.dotsLeft == 0,
	}
}

func (gm *gameManager) runGameLoop(sessionID string) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		g := gm.getGame(sessionID)
		if g == nil {
			return
		}

		if g.gameOver || g.checkWin() {
			return
		}

		g.movePlayer()
		g.moveGhosts()
		g.checkCollisions()
	}
}

func main() {
	// Set Gin to release mode for production-like behavior
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	r := gin.Default()

	// Enable CORS for web client
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	gameMgr := newGameManager()

	// Serve static files
	r.Static("/static", "./static")

	// Serve index.html at root
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "pacman-game",
		})
	})

	// Start new game
	r.POST("/api/game/start", func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
		}

		game := gameMgr.createGame(sessionID)
		go gameMgr.runGameLoop(sessionID)

		c.JSON(http.StatusOK, gin.H{
			"sessionID": sessionID,
			"state":     game.getState(),
		})
	})

	// Get game state
	r.GET("/api/game/state", func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session ID required"})
			return
		}

		game := gameMgr.getGame(sessionID)
		if game == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		c.JSON(http.StatusOK, game.getState())
	})

	// Move player
	r.POST("/api/game/move", func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "session ID required"})
			return
		}

		game := gameMgr.getGame(sessionID)
		if game == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}

		var req struct {
			Direction string `json:"direction"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var dir direction
		switch req.Direction {
		case "up":
			dir = up
		case "down":
			dir = down
		case "left":
			dir = left
		case "right":
			dir = right
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid direction"})
			return
		}

		game.setDirection(dir)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Restart game
	r.POST("/api/game/restart", func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			sessionID = fmt.Sprintf("session-%d", time.Now().UnixNano())
		}

		gameMgr.deleteGame(sessionID)
		game := gameMgr.createGame(sessionID)
		go gameMgr.runGameLoop(sessionID)

		c.JSON(http.StatusOK, gin.H{
			"sessionID": sessionID,
			"state":     game.getState(),
		})
	})

	// Start server
	port := ":8080"
	fmt.Printf("🎮 Pacman Game Server starting on http://localhost%s\n", port)
	fmt.Println("Open your browser and navigate to http://localhost:8080")
	log.Fatal(r.Run(port))
}
