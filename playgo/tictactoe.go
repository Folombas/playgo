package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Board [3][3]rune

func NewBoard() Board {
	return Board{}
}

func (b *Board) Print() {
	fmt.Println("  1 2 3")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d ", i+1)
		for j := 0; j < 3; j++ {
			if b[i][j] == 0 {
				fmt.Print(".")
			} else {
				fmt.Printf("%c", b[i][j])
			}
			if j < 2 {
				fmt.Print("|")
			}
		}
		fmt.Println()
		if i < 2 {
			fmt.Println("  -----")
		}
	}
}

func (b *Board) MakeMove(row, col int, player rune) bool {
	if row < 0 || row > 2 || col < 0 || col > 2 || b[row][col] != 0 {
		return false
	}
	b[row][col] = player
	return true
}

func (b *Board) CheckWin(player rune) bool {
	// Check rows
	for i := 0; i < 3; i++ {
		if b[i][0] == player && b[i][1] == player && b[i][2] == player {
			return true
		}
	}
	// Check columns
	for j := 0; j < 3; j++ {
		if b[0][j] == player && b[1][j] == player && b[2][j] == player {
			return true
		}
	}
	// Check diagonals
	if b[0][0] == player && b[1][1] == player && b[2][2] == player {
		return true
	}
	if b[0][2] == player && b[1][1] == player && b[2][0] == player {
		return true
	}
	return false
}

func (b *Board) IsFull() bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b[i][j] == 0 {
				return false
			}
		}
	}
	return true
}

func main() {
	board := NewBoard()
	currentPlayer := 'X'
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Tic-Tac-Toe ===")
	fmt.Println("Player 1: X")
	fmt.Println("Player 2: O")
	fmt.Println()

	for {
		board.Print()
		fmt.Printf("\nPlayer %c's turn. Enter row and column (e.g., '1 2'): ", currentPlayer)

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		parts := strings.Split(input, " ")
		if len(parts) != 2 {
			fmt.Println("Invalid input. Please enter row and column separated by space.")
			continue
		}

		row, err1 := strconv.Atoi(parts[0])
		col, err2 := strconv.Atoi(parts[1])
		if err1 != nil || err2 != nil || row < 1 || row > 3 || col < 1 || col > 3 {
			fmt.Println("Invalid input. Row and column must be between 1 and 3.")
			continue
		}

		if !board.MakeMove(row-1, col-1, currentPlayer) {
			fmt.Println("That position is already taken or invalid. Try again.")
			continue
		}

		if board.CheckWin(currentPlayer) {
			board.Print()
			fmt.Printf("\nPlayer %c wins!\n", currentPlayer)
			break
		}

		if board.IsFull() {
			board.Print()
			fmt.Println("\nIt's a draw!")
			break
		}

		if currentPlayer == 'X' {
			currentPlayer = 'O'
		} else {
			currentPlayer = 'X'
		}
	}
}
