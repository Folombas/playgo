package main

import "testing"

func TestNewBoard(t *testing.T) {
	board := NewBoard()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] != 0 {
				t.Errorf("Expected empty board, but position (%d, %d) has value %c", i, j, board[i][j])
			}
		}
	}
}

func TestMakeMove(t *testing.T) {
	board := NewBoard()
	if !board.MakeMove(0, 0, 'X') {
		t.Error("Failed to make valid move")
	}
	if board[0][0] != 'X' {
		t.Error("Move was not recorded correctly")
	}
}

func TestCheckWinRows(t *testing.T) {
	board := NewBoard()
	board.MakeMove(0, 0, 'X')
	board.MakeMove(0, 1, 'X')
	board.MakeMove(0, 2, 'X')
	if !board.CheckWin('X') {
		t.Error("Should detect win in row")
	}
}

func TestCheckWinColumns(t *testing.T) {
	board := NewBoard()
	board.MakeMove(0, 1, 'O')
	board.MakeMove(1, 1, 'O')
	board.MakeMove(2, 1, 'O')
	if !board.CheckWin('O') {
		t.Error("Should detect win in column")
	}
}

func TestCheckWinDiagonals(t *testing.T) {
	board := NewBoard()
	board.MakeMove(0, 0, 'X')
	board.MakeMove(1, 1, 'X')
	board.MakeMove(2, 2, 'X')
	if !board.CheckWin('X') {
		t.Error("Should detect win in diagonal")
	}
}

func TestIsFull(t *testing.T) {
	board := NewBoard()
	if board.IsFull() {
		t.Error("Empty board should not be full")
	}
	// Fill the board
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			board.MakeMove(i, j, 'X')
		}
	}
	if !board.IsFull() {
		t.Error("Full board should be detected as full")
	}
}
