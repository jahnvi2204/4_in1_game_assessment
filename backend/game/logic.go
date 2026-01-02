package game

const (
	ROWS      = 6
	COLS      = 7
	WIN_LENGTH = 4
)

type MoveResult struct {
	Success bool
	Message string
	Row     int
}

type WinResult struct {
	Won       bool
	Direction string
}

func CreateBoard() [][]interface{} {
	board := make([][]interface{}, ROWS)
	for i := range board {
		board[i] = make([]interface{}, COLS)
	}
	return board
}

func MakeMove(board [][]interface{}, column int, playerID interface{}) *MoveResult {
	if column < 0 || column >= COLS {
		return &MoveResult{Success: false, Message: "Invalid column"}
	}

	// Find the lowest available row in the column
	for row := ROWS - 1; row >= 0; row-- {
		if board[row][column] == nil {
			board[row][column] = playerID
			return &MoveResult{Success: true, Row: row}
		}
	}

	return &MoveResult{Success: false, Message: "Column is full"}
}

func CheckWin(board [][]interface{}, row, col int) *WinResult {
	playerID := board[row][col]
	if playerID == nil {
		return &WinResult{Won: false}
	}

	// Check horizontal
	if checkDirection(board, row, col, 0, 1, playerID) {
		return &WinResult{Won: true, Direction: "horizontal"}
	}

	// Check vertical
	if checkDirection(board, row, col, 1, 0, playerID) {
		return &WinResult{Won: true, Direction: "vertical"}
	}

	// Check diagonal (top-left to bottom-right)
	if checkDirection(board, row, col, 1, 1, playerID) {
		return &WinResult{Won: true, Direction: "diagonal"}
	}

	// Check diagonal (top-right to bottom-left)
	if checkDirection(board, row, col, 1, -1, playerID) {
		return &WinResult{Won: true, Direction: "diagonal"}
	}

	return &WinResult{Won: false}
}

func checkDirection(board [][]interface{}, startRow, startCol, deltaRow, deltaCol int, playerID interface{}) bool {
	count := 1 // Count the current piece

	// Check in positive direction
	row := startRow + deltaRow
	col := startCol + deltaCol
	for row >= 0 && row < ROWS && col >= 0 && col < COLS && board[row][col] == playerID {
		count++
		row += deltaRow
		col += deltaCol
	}

	// Check in negative direction
	row = startRow - deltaRow
	col = startCol - deltaCol
	for row >= 0 && row < ROWS && col >= 0 && col < COLS && board[row][col] == playerID {
		count++
		row -= deltaRow
		col -= deltaCol
	}

	return count >= WIN_LENGTH
}

func IsBoardFull(board [][]interface{}) bool {
	for col := 0; col < COLS; col++ {
		if board[0][col] == nil {
			return false
		}
	}
	return true
}

func GetValidMoves(board [][]interface{}) []int {
	validMoves := []int{}
	for col := 0; col < COLS; col++ {
		if board[0][col] == nil {
			validMoves = append(validMoves, col)
		}
	}
	return validMoves
}

func EvaluatePosition(board [][]interface{}, playerID, opponentID interface{}) int {
	score := 0

	// Check all possible 4-in-a-row positions
	for row := 0; row < ROWS; row++ {
		for col := 0; col < COLS; col++ {
			// Horizontal
			score += evaluateLine(board, row, col, 0, 1, playerID, opponentID)
			// Vertical
			score += evaluateLine(board, row, col, 1, 0, playerID, opponentID)
			// Diagonal \
			score += evaluateLine(board, row, col, 1, 1, playerID, opponentID)
			// Diagonal /
			score += evaluateLine(board, row, col, 1, -1, playerID, opponentID)
		}
	}

	return score
}

func evaluateLine(board [][]interface{}, startRow, startCol, deltaRow, deltaCol int, playerID, opponentID interface{}) int {
	playerCount := 0
	opponentCount := 0
	emptyCount := 0

	for i := 0; i < WIN_LENGTH; i++ {
		row := startRow + i*deltaRow
		col := startCol + i*deltaCol

		if row < 0 || row >= ROWS || col < 0 || col >= COLS {
			return 0 // Out of bounds
		}

		cell := board[row][col]
		if cell == playerID {
			playerCount++
		} else if cell == opponentID {
			opponentCount++
		} else {
			emptyCount++
		}
	}

	// Scoring
	if opponentCount > 0 && playerCount > 0 {
		return 0 // Blocked line
	}

	if playerCount == WIN_LENGTH {
		return 10000 // Win
	}
	if opponentCount == WIN_LENGTH {
		return -10000 // Opponent wins (should be blocked)
	}
	if opponentCount == WIN_LENGTH-1 && emptyCount == 1 {
		return -1000 // Opponent about to win (must block)
	}
	if playerCount == WIN_LENGTH-1 && emptyCount == 1 {
		return 1000 // Bot about to win
	}
	if playerCount == WIN_LENGTH-2 && emptyCount == 2 {
		return 100 // Potential win
	}
	if opponentCount == WIN_LENGTH-2 && emptyCount == 2 {
		return -100 // Opponent potential win
	}

	return playerCount*10 - opponentCount*10
}

