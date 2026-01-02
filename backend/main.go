package main

import (
	"connect-four/analytics"
	"connect-four/bot"
	"connect-four/game"
	"connect-four/matchmaking"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type Server struct {
	gameManager      *game.Manager
	matchmaking      *matchmaking.Service
	botPlayer        *bot.Player
	analyticsService *analytics.Service
}

// Adapter to make game.Manager implement matchmaking.GameManager interface
type gameManagerAdapter struct {
	manager *game.Manager
}

func (a *gameManagerAdapter) CreateGame(player1, player2 interface{}) interface{} {
	p1 := player1.(*game.Player)
	p2 := player2.(*game.Player)
	return a.manager.CreateGame(p1, p2)
}

func main() {
	// Initialize database
	db, err := game.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize analytics service
	analyticsService, err := analytics.NewService()
	if err != nil {
		log.Printf("Warning: Analytics service initialization failed: %v", err)
		log.Println("Continuing without analytics...")
		analyticsService = nil
	}

	// Initialize services
	gameManager := game.NewManager(db, analyticsService)
	// Create adapter for matchmaking interface
	gameManagerAdapter := &gameManagerAdapter{manager: gameManager}
	matchmakingService := matchmaking.NewService(gameManagerAdapter, 10*time.Second)
	botPlayer := bot.NewPlayer()

	server := &Server{
		gameManager:      gameManager,
		matchmaking:      matchmakingService,
		botPlayer:        botPlayer,
		analyticsService: analyticsService,
	}

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/api/leaderboard", server.getLeaderboard).Methods("GET")
	r.HandleFunc("/api/health", server.healthCheck).Methods("GET")
	r.HandleFunc("/ws", server.handleWebSocket)

	// Handle favicon and root
	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}).Methods("GET")
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Connect Four API Server"))
	}).Methods("GET")

	// CORS middleware
	r.Use(corsMiddleware)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) getLeaderboard(w http.ResponseWriter, r *http.Request) {
	leaderboard, err := s.gameManager.GetLeaderboard()
	if err != nil {
		http.Error(w, "Failed to fetch leaderboard", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Println("New WebSocket connection")

	// Handle messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			s.matchmaking.RemovePlayer(conn)
			s.gameManager.HandleDisconnect(conn, s.notifyPlayers)
			break
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			s.sendError(conn, "Invalid message format")
			continue
		}

		switch msgType {
		case "join":
			username, _ := msg["username"].(string)
			s.handleJoin(conn, username)
		case "rejoin":
			username, _ := msg["username"].(string)
			gameID, _ := msg["gameId"].(string)
			s.handleRejoin(conn, username, gameID)
		case "makeMove":
			gameID, _ := msg["gameId"].(string)
			column, _ := msg["column"].(float64)
			s.handleMakeMove(conn, gameID, int(column))
		default:
			s.sendError(conn, "Unknown message type")
		}
	}
}

func (s *Server) handleJoin(conn *websocket.Conn, username string) {
	if username == "" {
		s.sendError(conn, "Username is required")
		return
	}

	matchPlayer := &matchmaking.Player{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Username:  username,
		Conn:      conn,
		Connected: true,
	}

	matchResult := s.matchmaking.AddPlayer(matchPlayer)

	if matchResult.Matched {
		// Convert matchmaking.Player to game.Player
		player1 := convertToGamePlayer(matchResult.Player1)
		player2 := convertToGamePlayer(matchResult.Player2)
		// Start game with matched player
		game := s.gameManager.CreateGame(player1, player2)
		s.notifyPlayers(game)
	} else {
		// Waiting for opponent
		s.sendMessage(conn, map[string]interface{}{
			"type":    "waiting",
			"message": "Waiting for opponent...",
		})

		// Schedule bot match if no opponent joins
		s.matchmaking.ScheduleBotMatch(matchPlayer, func(p *matchmaking.Player) {
			botPlayer := convertToGamePlayer(&matchmaking.Player{
				ID:        "bot",
				Username:  "Bot",
				Conn:      nil,
				Connected: true,
				IsBot:     true,
			})
			player1 := convertToGamePlayer(p)
			game := s.gameManager.CreateGame(player1, botPlayer)
			s.notifyPlayers(game)

			// Bot makes first move if it's bot's turn
			if game.CurrentPlayer == "bot" {
				time.AfterFunc(500*time.Millisecond, func() {
					s.botPlayer.MakeMove(game, s.gameManager, s.notifyPlayers)
				})
			}
		})
	}
}

func convertToGamePlayer(mp *matchmaking.Player) *game.Player {
	return &game.Player{
		ID:       mp.ID,
		Username: mp.Username,
		Conn:     mp.Conn,
		IsBot:    mp.IsBot,
	}
}

func (s *Server) handleRejoin(conn *websocket.Conn, username, gameID string) {
	result := s.gameManager.RejoinGame(conn, username, gameID)
	if result.Success {
		s.notifyPlayers(result.Game)
		// Notify opponent
		if result.Game.Player1.Conn != nil {
			s.sendMessage(result.Game.Player1.Conn, map[string]interface{}{
				"type":     "playerReconnected",
				"username": username,
			})
		}
		if result.Game.Player2.Conn != nil {
			s.sendMessage(result.Game.Player2.Conn, map[string]interface{}{
				"type":     "playerReconnected",
				"username": username,
			})
		}
	} else {
		s.sendError(conn, result.Message)
	}
}

func (s *Server) handleMakeMove(conn *websocket.Conn, gameID string, column int) {
	result := s.gameManager.MakeMove(gameID, column, conn)

	if !result.Success {
		s.sendError(conn, result.Message)
		return
	}

	game := result.Game
	s.notifyPlayers(game)

	// Check if game ended
	if game.Status == "finished" {
		s.gameManager.SaveGame(game)
		if s.analyticsService != nil {
			s.analyticsService.TrackGameEnd(game)
		}
	} else if game.CurrentPlayer == "bot" && game.Player2.IsBot {
		// Bot makes move
		time.AfterFunc(500*time.Millisecond, func() {
			s.botPlayer.MakeMove(game, s.gameManager, s.notifyPlayers)
		})
	}
}

func (s *Server) notifyPlayers(game *game.Game) {
	// Convert board to use usernames instead of IDs for frontend
	boardForFrontend := make([][]interface{}, len(game.Board))
	for i, row := range game.Board {
		boardForFrontend[i] = make([]interface{}, len(row))
		for j, cell := range row {
			if cell == nil {
				boardForFrontend[i][j] = nil
			} else if cell == game.Player1.ID {
				boardForFrontend[i][j] = game.Player1.Username
			} else if cell == game.Player2.ID || cell == "bot" {
				boardForFrontend[i][j] = game.Player2.Username
			} else {
				boardForFrontend[i][j] = cell
			}
		}
	}

	// Convert currentPlayer to username for frontend
	currentPlayerForFrontend := game.CurrentPlayer
	if game.CurrentPlayer == game.Player1.ID {
		currentPlayerForFrontend = game.Player1.Username
	} else if game.CurrentPlayer == game.Player2.ID || game.CurrentPlayer == "bot" {
		currentPlayerForFrontend = game.Player2.Username
	}

	// Convert winner to username for frontend
	winnerForFrontend := game.Winner
	if game.Winner == game.Player1.ID {
		winnerForFrontend = game.Player1.Username
	} else if game.Winner == game.Player2.ID || game.Winner == "bot" {
		winnerForFrontend = game.Player2.Username
	}

	gameState := map[string]interface{}{
		"type": "gameState",
		"game": map[string]interface{}{
			"id":            game.ID,
			"board":         boardForFrontend,
			"currentPlayer": currentPlayerForFrontend,
			"player1": map[string]interface{}{
				"username": game.Player1.Username,
				"isBot":    game.Player1.IsBot,
			},
			"player2": map[string]interface{}{
				"username": game.Player2.Username,
				"isBot":    game.Player2.IsBot,
			},
			"status": game.Status,
			"winner": winnerForFrontend,
		},
	}

	if game.Player1.Conn != nil {
		s.sendMessage(game.Player1.Conn, gameState)
	}
	if game.Player2.Conn != nil {
		s.sendMessage(game.Player2.Conn, gameState)
	}
}

func (s *Server) sendMessage(conn *websocket.Conn, msg map[string]interface{}) {
	if conn != nil {
		conn.WriteJSON(msg)
	}
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	s.sendMessage(conn, map[string]interface{}{
		"type":    "error",
		"message": message,
	})
}
