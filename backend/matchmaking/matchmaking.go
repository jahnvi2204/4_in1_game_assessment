package matchmaking

import (
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	ID        string
	Username  string
	Conn      *websocket.Conn
	Connected bool
	IsBot     bool
}

type MatchResult struct {
	Matched bool
	Player1 *Player
	Player2 *Player
}

type Service struct {
	gameManager    GameManager
	timeout        time.Duration
	waitingPlayers []*Player
	botTimers      map[string]*time.Timer
}

type GameManager interface {
	CreateGame(player1, player2 interface{}) interface{} // Returns *game.Game, accepts *game.Player
}

func NewService(gameManager GameManager, timeout time.Duration) *Service {
	return &Service{
		gameManager:    gameManager,
		timeout:        timeout,
		waitingPlayers: []*Player{},
		botTimers:      make(map[string]*time.Timer),
	}
}

func (s *Service) AddPlayer(player *Player) *MatchResult {
	// Remove any existing bot timer for this player
	if timer, exists := s.botTimers[player.ID]; exists {
		timer.Stop()
		delete(s.botTimers, player.ID)
	}

	// Check if there's a waiting player
	if len(s.waitingPlayers) > 0 {
		opponent := s.waitingPlayers[0]
		s.waitingPlayers = s.waitingPlayers[1:]
		return &MatchResult{
			Matched: true,
			Player1: opponent,
			Player2: player,
		}
	}

	// Add to waiting queue
	s.waitingPlayers = append(s.waitingPlayers, player)
	return &MatchResult{Matched: false}
}

func (s *Service) RemovePlayer(conn *websocket.Conn) {
	// Remove from waiting queue
	newWaiting := []*Player{}
	for _, p := range s.waitingPlayers {
		if p.Conn != conn {
			newWaiting = append(newWaiting, p)
		}
	}
	s.waitingPlayers = newWaiting

	// Clear bot timer if exists
	for playerID, timer := range s.botTimers {
		player := s.findPlayerByID(playerID)
		if player == nil || player.Conn == conn {
			timer.Stop()
			delete(s.botTimers, playerID)
		}
	}
}

func (s *Service) ScheduleBotMatch(player *Player, callback func(*Player)) {
	timer := time.AfterFunc(s.timeout, func() {
		// Check if player is still waiting
		if s.isPlayerWaiting(player.ID) {
			s.removeWaitingPlayer(player.ID)
			delete(s.botTimers, player.ID)
			callback(player)
		}
	})

	s.botTimers[player.ID] = timer
}

func (s *Service) findPlayerByID(id string) *Player {
	for _, p := range s.waitingPlayers {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (s *Service) isPlayerWaiting(id string) bool {
	return s.findPlayerByID(id) != nil
}

func (s *Service) removeWaitingPlayer(id string) {
	newWaiting := []*Player{}
	for _, p := range s.waitingPlayers {
		if p.ID != id {
			newWaiting = append(newWaiting, p)
		}
	}
	s.waitingPlayers = newWaiting
}

