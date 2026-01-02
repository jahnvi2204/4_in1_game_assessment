package game

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

type Game struct {
	ID           string
	Player1      *Player
	Player2      *Player
	Board        [][]interface{}
	CurrentPlayer string
	Status       string
	Winner       string
	Moves        []Move
	StartedAt    time.Time
	EndedAt      *time.Time
	LastMoveAt   time.Time
}

type Player struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	IsBot    bool
}

type Move struct {
	Player    string
	Column    int
	Row       int
	Timestamp time.Time
}

// Analytics interface to avoid circular dependency
type Analytics interface {
	TrackGameStart(game *Game)
	TrackMove(game *Game, column, row int)
	TrackGameEnd(game *Game)
}

type Manager struct {
	games          map[string]*Game
	db             *sql.DB
	analyticsService Analytics
	reconnectWindows map[string]*ReconnectWindow
}

type ReconnectWindow struct {
	PlayerID  string
	ExpiresAt time.Time
}

type GameMoveResult struct {
	Success bool
	Message string
	Game    *Game
}

type RejoinResult struct {
	Success bool
	Message string
	Game    *Game
}

type LeaderboardEntry struct {
	Username   string `json:"username"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
	Draws      int    `json:"draws"`
	TotalGames int    `json:"total_games"`
}

func NewManager(db *sql.DB, analyticsService Analytics) *Manager {
	return &Manager{
		games:            make(map[string]*Game),
		db:                db,
		analyticsService:  analyticsService,
		reconnectWindows:  make(map[string]*ReconnectWindow),
	}
}

func InitDB() (*sql.DB, error) {
	dbUser := getEnv("DB_USER", "postgres")
	dbHost := getEnv("DB_HOST", "localhost")
	dbName := getEnv("DB_NAME", "connectfour")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbPort := getEnv("DB_PORT", "5432")

	connStr := fmt.Sprintf("user=%s host=%s dbname=%s password=%s port=%s sslmode=disable",
		dbUser, dbHost, dbName, dbPassword, dbPort)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Initialize tables
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS games (
			id UUID PRIMARY KEY,
			player1_username VARCHAR(255),
			player2_username VARCHAR(255),
			winner VARCHAR(255),
			status VARCHAR(50),
			started_at TIMESTAMP,
			ended_at TIMESTAMP,
			duration_seconds INTEGER,
			moves JSONB
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS leaderboard (
			username VARCHAR(255) PRIMARY KEY,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			draws INTEGER DEFAULT 0,
			total_games INTEGER DEFAULT 0
		)
	`)
	return err
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func (m *Manager) CreateGame(player1, player2 *Player) *Game {
	gameID := uuid.New().String()
	game := &Game{
		ID:            gameID,
		Player1:       player1,
		Player2:       player2,
		Board:         CreateBoard(),
		CurrentPlayer: player1.ID,
		Status:        "active",
		Winner:        "",
		Moves:         []Move{},
		StartedAt:     time.Now(),
		LastMoveAt:    time.Now(),
	}

	m.games[gameID] = game

	// Track game start
	if m.analyticsService != nil {
		m.analyticsService.TrackGameStart(game)
	}

	return game
}

func (m *Manager) MakeMove(gameID string, column int, conn *websocket.Conn) *GameMoveResult {
	game, exists := m.games[gameID]
	if !exists {
		return &GameMoveResult{Success: false, Message: "Game not found"}
	}

	if game.Status != "active" {
		return &GameMoveResult{Success: false, Message: "Game is not active"}
	}

	// Verify it's the player's turn
	var player *Player
	if game.CurrentPlayer == game.Player1.ID {
		player = game.Player1
	} else {
		player = game.Player2
	}

	if player.IsBot {
		return &GameMoveResult{Success: false, Message: "Not your turn"}
	}
	if player.Conn != conn {
		return &GameMoveResult{Success: false, Message: "Not your turn"}
	}

	// Validate column
	if column < 0 || column >= 7 {
		return &GameMoveResult{Success: false, Message: "Invalid column"}
	}

	// Make move
	moveResult := MakeMove(game.Board, column, game.CurrentPlayer)
	if !moveResult.Success {
		return &GameMoveResult{Success: false, Message: moveResult.Message}
	}

	// Record move
	game.Moves = append(game.Moves, Move{
		Player:    game.CurrentPlayer,
		Column:    column,
		Row:       moveResult.Row,
		Timestamp: time.Now(),
	})

	game.LastMoveAt = time.Now()

	// Check for win
	winResult := CheckWin(game.Board, moveResult.Row, column)
	if winResult.Won {
		game.Status = "finished"
		game.Winner = game.CurrentPlayer
		now := time.Now()
		game.EndedAt = &now
		m.UpdateLeaderboard(game)
	} else if IsBoardFull(game.Board) {
		game.Status = "finished"
		game.Winner = "draw"
		now := time.Now()
		game.EndedAt = &now
		m.UpdateLeaderboard(game)
	} else {
		// Switch turns
		if game.CurrentPlayer == game.Player1.ID {
			if game.Player2.IsBot {
				game.CurrentPlayer = "bot"
			} else {
				game.CurrentPlayer = game.Player2.ID
			}
		} else {
			game.CurrentPlayer = game.Player1.ID
		}
	}

	// Track move
	if m.analyticsService != nil {
		m.analyticsService.TrackMove(game, column, moveResult.Row)
	}

	return &GameMoveResult{Success: true, Game: game}
}

func (m *Manager) BotMakeMove(gameID string, column int) *GameMoveResult {
	game, exists := m.games[gameID]
	if !exists || game.Status != "active" {
		return &GameMoveResult{Success: false}
	}

	if game.CurrentPlayer != "bot" {
		return &GameMoveResult{Success: false}
	}

	moveResult := MakeMove(game.Board, column, "bot")
	if !moveResult.Success {
		return &GameMoveResult{Success: false}
	}

	game.Moves = append(game.Moves, Move{
		Player:    "bot",
		Column:    column,
		Row:       moveResult.Row,
		Timestamp: time.Now(),
	})

	game.LastMoveAt = time.Now()

	winResult := CheckWin(game.Board, moveResult.Row, column)
	if winResult.Won {
		game.Status = "finished"
		game.Winner = "bot"
		now := time.Now()
		game.EndedAt = &now
		m.UpdateLeaderboard(game)
	} else if IsBoardFull(game.Board) {
		game.Status = "finished"
		game.Winner = "draw"
		now := time.Now()
		game.EndedAt = &now
		m.UpdateLeaderboard(game)
	} else {
		game.CurrentPlayer = game.Player1.ID
	}

	if m.analyticsService != nil {
		m.analyticsService.TrackMove(game, column, moveResult.Row)
	}

	return &GameMoveResult{Success: true, Game: game}
}

func (m *Manager) RejoinGame(conn *websocket.Conn, username, gameID string) *RejoinResult {
	game, exists := m.games[gameID]
	if !exists {
		return &RejoinResult{Success: false, Message: "Game not found"}
	}

	// Check reconnect window
	reconnectInfo, hasWindow := m.reconnectWindows[gameID]
	if !hasWindow {
		return &RejoinResult{Success: false, Message: "Reconnection window expired"}
	}

	now := time.Now()
	if now.After(reconnectInfo.ExpiresAt) {
		delete(m.reconnectWindows, gameID)
		m.ForfeitGame(gameID, reconnectInfo.PlayerID, nil)
		return &RejoinResult{Success: false, Message: "Reconnection window expired"}
	}

	// Reconnect player
	if game.Player1.Username == username {
		game.Player1.Conn = conn
		delete(m.reconnectWindows, gameID)
		return &RejoinResult{Success: true, Game: game}
	} else if game.Player2.Username == username {
		game.Player2.Conn = conn
		delete(m.reconnectWindows, gameID)
		return &RejoinResult{Success: true, Game: game}
	}

	return &RejoinResult{Success: false, Message: "Username does not match this game"}
}

func (m *Manager) HandleDisconnect(conn *websocket.Conn, notifyCallback func(*Game)) {
	for gameID, game := range m.games {
		if game.Status != "active" {
			continue
		}

		var disconnectedPlayer *Player
		if game.Player1.Conn == conn {
			disconnectedPlayer = game.Player1
		} else if game.Player2.Conn == conn && !game.Player2.IsBot {
			disconnectedPlayer = game.Player2
		}

		if disconnectedPlayer != nil {
			// Set 30 second reconnect window
			expiresAt := time.Now().Add(30 * time.Second)
			m.reconnectWindows[gameID] = &ReconnectWindow{
				PlayerID:  disconnectedPlayer.ID,
				ExpiresAt: expiresAt,
			}

			// Notify opponent
			var opponent *Player
			if disconnectedPlayer == game.Player1 {
				opponent = game.Player2
			} else {
				opponent = game.Player1
			}

			if opponent.Conn != nil {
				opponent.Conn.WriteJSON(map[string]interface{}{
					"type":    "playerDisconnected",
					"message": fmt.Sprintf("%s disconnected. Reconnecting...", disconnectedPlayer.Username),
				})
			}

			// Schedule forfeit if not reconnected
			forfeitGameID := gameID
			forfeitPlayerID := disconnectedPlayer.ID
			time.AfterFunc(30*time.Second, func() {
				if _, exists := m.reconnectWindows[forfeitGameID]; exists {
					forfeitedGame := m.ForfeitGame(forfeitGameID, forfeitPlayerID, notifyCallback)
					if forfeitedGame != nil && notifyCallback != nil {
						notifyCallback(forfeitedGame)
					}
				}
			})
		}
	}
}

func (m *Manager) ForfeitGame(gameID, forfeitingPlayerID string, notifyCallback func(*Game)) *Game {
	game, exists := m.games[gameID]
	if !exists || game.Status != "active" {
		return nil
	}

	game.Status = "finished"
	now := time.Now()
	game.EndedAt = &now

	// Determine winner
	if game.Player1.ID == forfeitingPlayerID {
		if game.Player2.IsBot {
			game.Winner = "bot"
		} else {
			game.Winner = game.Player2.ID
		}
	} else {
		game.Winner = game.Player1.ID
	}

	m.SaveGame(game)
	m.UpdateLeaderboard(game)
	if m.analyticsService != nil {
		m.analyticsService.TrackGameEnd(game)
	}

	// Notify players if callback provided
	if notifyCallback != nil {
		notifyCallback(game)
	}

	delete(m.games, gameID)
	delete(m.reconnectWindows, gameID)

	return game
}

func (m *Manager) SaveGame(game *Game) {
	if game.Status != "finished" {
		return
	}

	var duration *int
	if game.EndedAt != nil {
		d := int(game.EndedAt.Sub(game.StartedAt).Seconds())
		duration = &d
	}

	movesJSON, _ := json.Marshal(game.Moves)

	_, err := m.db.Exec(
		`INSERT INTO games (id, player1_username, player2_username, winner, status, started_at, ended_at, duration_seconds, moves)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		game.ID, game.Player1.Username, game.Player2.Username, game.Winner, game.Status,
		game.StartedAt, game.EndedAt, duration, movesJSON,
	)
	if err != nil {
		log.Printf("Error saving game: %v", err)
	}
}

func (m *Manager) UpdateLeaderboard(game *Game) {
	if game.Status != "finished" {
		return
	}

	// Update player1
	var player1Wins, player1Losses, player1Draws int
	if game.Winner == game.Player1.ID {
		player1Wins = 1
	} else if game.Winner != "draw" {
		player1Losses = 1
	} else {
		player1Draws = 1
	}

	_, err := m.db.Exec(
		`INSERT INTO leaderboard (username, wins, losses, draws, total_games)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (username) 
		 DO UPDATE SET 
		   wins = leaderboard.wins + $2,
		   losses = leaderboard.losses + $3,
		   draws = leaderboard.draws + $4,
		   total_games = leaderboard.total_games + $5`,
		game.Player1.Username, player1Wins, player1Losses, player1Draws, 1,
	)
	if err != nil {
		log.Printf("Error updating leaderboard: %v", err)
	}

	// Update player2 (skip bot)
	if !game.Player2.IsBot {
		var player2Wins, player2Losses, player2Draws int
		if game.Winner == "bot" {
			player2Wins = 1
		} else if game.Winner == game.Player2.ID {
			player2Wins = 1
		} else if game.Winner != "draw" && game.Winner != game.Player2.ID {
			player2Losses = 1
		} else if game.Winner == "draw" {
			player2Draws = 1
		}

		_, err := m.db.Exec(
			`INSERT INTO leaderboard (username, wins, losses, draws, total_games)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (username) 
			 DO UPDATE SET 
			   wins = leaderboard.wins + $2,
			   losses = leaderboard.losses + $3,
			   draws = leaderboard.draws + $4,
			   total_games = leaderboard.total_games + $5`,
			game.Player2.Username, player2Wins, player2Losses, player2Draws, 1,
		)
		if err != nil {
			log.Printf("Error updating leaderboard: %v", err)
		}
	}
}

func (m *Manager) GetLeaderboard() ([]LeaderboardEntry, error) {
	rows, err := m.db.Query(`
		SELECT username, wins, losses, draws, total_games
		FROM leaderboard
		ORDER BY wins DESC, total_games DESC
		LIMIT 100
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(&entry.Username, &entry.Wins, &entry.Losses, &entry.Draws, &entry.TotalGames)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (m *Manager) GetGame(gameID string) *Game {
	return m.games[gameID]
}

