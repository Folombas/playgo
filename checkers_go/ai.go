package main

// Minimax AI для шашек с alpha-beta отсечением
// Go365 Day 99 - Улучшение AI

import (
	"math"
	"math/rand"
)

const (
	AIDepthEasy   = 2
	AIDepthMedium = 4
	AIDepthHard   = 6
)

// AIDifficulty уровни
type AIDifficulty int

const (
	Easy AIDifficulty = iota
	Medium
	Hard
)

var currentDifficulty = Medium

// EvaluateBoard оценивает позицию
// Положительное = преимущество чёрных (AI), отрицательное = белых (игрок)
func EvaluateBoard(board [N][N]int, aiColor int) int {
	score := 0
	playerColor := 3 - aiColor // WHITE=1, BLACK=2

	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			piece := board[r][c]
			if piece == NONE {
				continue
			}

			var value int
			isKing := piece == WK || piece == BK
			isAI := piece == aiColor || piece == aiColor+2

			if isKing {
				value = 5 // Дамка стоит 5
			} else {
				value = 1 // Обычная шашка = 1
			}

			// Бонус за продвижение вперёд
			if !isKing {
				if isAI && aiColor == BLACK {
					value += float64(r) * 0.1 // Чем дальше, тем лучше
				} else if !isAI && aiColor == WHITE {
					value += float64(N-1-r) * 0.1
				}
			}

			// Бонус за центр
			centerDist := math.Abs(float64(c)-3.5) + math.Abs(float64(r)-3.5)
			value += (7 - centerDist) * 0.05

			if isAI {
				score += int(value * 100)
			} else {
				score -= int(value * 100)
			}
		}
	}

	return score
}

// Minimax с alpha-beta отсечением
func Minimax(board [N][N]int, depth int, alpha, beta int, maximizing bool, aiColor int) int {
	if depth == 0 {
		return EvaluateBoard(board, aiColor)
	}

	currentColor := BLACK
	if !maximizing {
		currentColor = WHITE
	}

	moves := GetAllMoves(board, currentColor)
	if len(moves) == 0 {
		// Нет ходов = проигрыш
		if maximizing {
			return -10000
		}
		return 10000
	}

	if maximizing {
		maxEval := -math.MaxInt32
		for _, move := range moves {
			newBoard := ApplyMove(board, move)
			eval := Minimax(newBoard, depth-1, alpha, beta, false, aiColor)
			maxEval = int(math.Max(float64(maxEval), float64(eval)))
			alpha = int(math.Max(float64(alpha), float64(eval)))
			if beta <= alpha {
				break // Beta cutoff
			}
		}
		return maxEval
	} else {
		minEval := math.MaxInt32
		for _, move := range moves {
			newBoard := ApplyMove(board, move)
			eval := Minimax(newBoard, depth-1, alpha, beta, true, aiColor)
			minEval = int(math.Min(float64(minEval), float64(eval)))
			beta = int(math.Min(float64(beta), float64(eval)))
			if beta <= alpha {
				break // Alpha cutoff
			}
		}
		return minEval
	}
}

// GetBestMove находит лучший ход для AI
func GetBestMove(board [N][N]int, aiColor int, difficulty AIDifficulty) Move {
	depth := AIDepthEasy
	switch difficulty {
	case Medium:
		depth = AIDepthMedium
	case Hard:
		depth = AIDepthHard
	}

	moves := GetAllMoves(board, aiColor)
	if len(moves) == 0 {
		return Move{}
	}

	// Easy: случайный ход иногда
	if difficulty == Easy && rand.Float64() < 0.3 {
		return moves[rand.Intn(len(moves))]
	}

	bestMove := moves[0]
	bestEval := -math.MaxInt32

	for _, move := range moves {
		newBoard := ApplyMove(board, move)
		eval := Minimax(newBoard, depth-1, -math.MaxInt32, math.MaxInt32, false, aiColor)
		if eval > bestEval {
			bestEval = eval
			bestMove = move
		}
	}

	return bestMove
}

// Move представляет ход
type Move struct {
	FromR, FromC int
	ToR, ToC     int
	Captures     [][2]int // Взятые шашки [r, c]
	Promoted     bool
}

// GetAllMoves получает все возможные ходы для цвета
func GetAllMoves(board [N][N]int, color int) []Move {
	moves := []Move{}

	// Сначала проверяем обязательные взятия
	captureMoves := GetCaptureMoves(board, color)
	if len(captureMoves) > 0 {
		return captureMoves
	}

	// Иначе обычные ходы
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if board[r][c] == color || board[r][c] == color+2 { // Regular или king
				moves = append(moves, GetRegularMoves(board, r, c)...)
			}
		}
	}

	return moves
}

// GetCaptureMoves получает все ходы со взятием
func GetCaptureMoves(board [N][N]int, color int) []Move {
	moves := []Move{}
	isKing := func(p int) bool { return p == WK || p == BK }

	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			piece := board[r][c]
			if piece != color && piece != color+2 {
				continue
			}

			king := isKing(piece)
			dirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}

			for _, d := range dirs {
				if !king {
					// Обычные шашки бьют только вперёд
					if color == WHITE && d[0] < 0 {
						continue
					}
					if color == BLACK && d[0] > 0 {
						continue
					}
				}

				mr, mc := r+d[0]*2, c+d[1]*2
				if mr >= 0 && mr < N && mc >= 0 && mc < N {
					midR, midC := r+d[0], c+d[1]
					midPiece := board[midR][midC]
					if board[mr][mc] == NONE && midPiece != NONE && midPiece != color && midPiece != color+2 {
						captures := [][2]int{{midR, midC}}
						// Проверяем мульти-взятия
						newBoard := board
						newBoard[r][c] = NONE
						newBoard[midR][midC] = NONE
						newBoard[mr][mc] = piece
						
						moreCaptures := GetContinuedCaptures(newBoard, mr, mc, piece)
						if len(moreCaptures) > 0 {
							captures = append(captures, moreCaptures...)
							mr2, mc2 := moreCaptures[len(moreCaptures)-1]
							moves = append(moves, Move{
								FromR: r, FromC: c,
								ToR: mr2, ToC: mc2,
								Captures: captures,
							})
						} else {
							moves = append(moves, Move{
								FromR: r, FromC: c,
								ToR: mr, ToC: mc,
								Captures: captures,
							})
						}
					}
				}
			}
		}
	}

	return moves
}

// GetContinuedCaptures проверяет дополнительные взятия после первого
func GetContinuedCaptures(board [N][N]int, r, c int, piece int) [][2]int {
	isKing := piece == WK || piece == BK
	color := piece
	if piece == WK || piece == WP {
		color = WHITE
	} else {
		color = BLACK
	}

	dirs := [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	bestCapture := [][2]int{}

	for _, d := range dirs {
		if !isKing {
			if color == WHITE && d[0] < 0 {
				continue
			}
			if color == BLACK && d[0] > 0 {
				continue
			}
		}

		mr, mc := r+d[0]*2, c+d[1]*2
		if mr >= 0 && mr < N && mc >= 0 && mc < N {
			midR, midC := r+d[0], c+d[1]
			midPiece := board[midR][midC]
			if board[mr][mc] == NONE && midPiece != NONE && midPiece != color && midPiece != color+2 {
				newBoard := board
				newBoard[r][c] = NONE
				newBoard[midR][midC] = NONE
				newBoard[mr][mc] = piece

				more := GetContinuedCaptures(newBoard, mr, mc, piece)
				capture := append([][2]int{{midR, midC}}, more...)
				if len(capture) > len(bestCapture) {
					bestCapture = capture
				}
			}
		}
	}

	return bestCapture
}

// GetRegularMoves получает обычные ходы без взятия
func GetRegularMoves(board [N][N]int, r, c int) []Move {
	moves := []Move{}
	piece := board[r][c]
	king := piece == WK || piece == BK
	color := piece
	if piece == WP || piece == WK {
		color = WHITE
	} else {
		color = BLACK
	}

	var dirs [][2]int
	if king {
		dirs = [][2]int{{-1, -1}, {-1, 1}, {1, -1}, {1, 1}}
	} else if color == WHITE {
		dirs = [][2]int{{-1, -1}, {-1, 1}}
	} else {
		dirs = [][2]int{{1, -1}, {1, 1}}
	}

	for _, d := range dirs {
		nr, nc := r+d[0], c+d[1]
		if nr >= 0 && nr < N && nc >= 0 && nc < N && board[nr][nc] == NONE {
			moves = append(moves, Move{
				FromR: r, FromC: c,
				ToR: nr, ToC: nc,
				Promoted: (color == WHITE && nr == 0) || (color == BLACK && nr == N-1),
			})
		}
	}

	return moves
}

// ApplyMove применяет ход к доске
func ApplyMove(board [N][N]int, move Move) [N][N]int {
	newBoard := board
	piece := newBoard[move.FromR][move.FromC]
	newBoard[move.FromR][move.FromC] = NONE
	newBoard[move.ToR][move.ToC] = piece

	// Удаляем взятые шашки
	for _, cap := range move.Captures {
		newBoard[cap[0]][cap[1]] = NONE
	}

	// Превращение в дамку
	if move.Promoted {
		if piece == WP {
			newBoard[move.ToR][move.ToC] = WK
		} else if piece == BP {
			newBoard[move.ToR][move.ToC] = BK
		}
	}

	return newBoard
}
