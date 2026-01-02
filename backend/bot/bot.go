package bot

import (
	"connect-four/game"
)

type Player struct{}

func NewPlayer() *Player {
	return &Player{}
}

func (b *Player) MakeMove(g *game.Game, gameManager *game.Manager, notifyCallback func(*game.Game)) {
	if g.Status != "active" || g.CurrentPlayer != "bot" {
		return
	}

	opponentID := g.Player1.ID
	botID := "bot"

	// Get valid moves
	validMoves := game.GetValidMoves(g.Board)
	if len(validMoves) == 0 {
		return
	}

	bestColumn := validMoves[0]
	bestScore := -999999
	blockingColumn := -1

	// Strategy priority:
	// 1. Check if bot can win
	// 2. Check if opponent can win (block)
	// 3. Make best strategic move

	// First, check if opponent can win immediately (must block)
	for _, col := range validMoves {
		testBoard := copyBoard(g.Board)
		moveResult := game.MakeMove(testBoard, col, opponentID)
		if moveResult.Success {
			winCheck := game.CheckWin(testBoard, moveResult.Row, col)
			if winCheck.Won {
				blockingColumn = col
				break // Must block this
			}
		}
	}

	// If we found a blocking move, use it
	if blockingColumn != -1 {
		b.executeMove(gameManager, g, blockingColumn, notifyCallback)
		return
	}

	// Check if bot can win
	for _, col := range validMoves {
		testBoard := copyBoard(g.Board)
		moveResult := game.MakeMove(testBoard, col, botID)
		if !moveResult.Success {
			continue
		}

		winCheck := game.CheckWin(testBoard, moveResult.Row, col)
		if winCheck.Won {
			// Bot wins - make this move immediately
			b.executeMove(gameManager, g, col, notifyCallback)
			return
		}
	}

	// Evaluate all moves and pick the best
	for _, col := range validMoves {
		testBoard := copyBoard(g.Board)
		moveResult := game.MakeMove(testBoard, col, botID)
		if !moveResult.Success {
			continue
		}

		// Score this move
		score := game.EvaluatePosition(testBoard, botID, opponentID)

		// Prefer center columns (better strategic position)
		centerDistance := abs(col - 3)
		score += (3 - centerDistance) * 5

		if score > bestScore {
			bestScore = score
			bestColumn = col
		}
	}

	// Make the best move
	b.executeMove(gameManager, g, bestColumn, notifyCallback)
}

func (b *Player) executeMove(gameManager *game.Manager, g *game.Game, column int, notifyCallback func(*game.Game)) {
	result := gameManager.BotMakeMove(g.ID, column)
	if result.Success {
		updatedGame := result.Game

		// Notify players
		if notifyCallback != nil {
			notifyCallback(updatedGame)
		}
	}
}

func copyBoard(board [][]interface{}) [][]interface{} {
	newBoard := make([][]interface{}, len(board))
	for i, row := range board {
		newBoard[i] = make([]interface{}, len(row))
		copy(newBoard[i], row)
	}
	return newBoard
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

