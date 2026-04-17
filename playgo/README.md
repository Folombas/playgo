# Tic-Tac-Toe Game in Go

A simple command-line Tic-Tac-Toe (Three in a Row) game implemented in Go.

## Files

- `tictactoe.go` - Main game source code
- `README.md` - This file with instructions

## How to Play

1. Two players take turns marking spaces in a 3×3 grid
2. Player 1 uses 'X', Player 2 uses 'O'
3. The first player to get 3 of their marks in a row (horizontally, vertically, or diagonally) wins
4. If all 9 spaces are filled without a winner, the game is a draw

## Building and Running

```bash
# Navigate to the playgo directory
cd playgo

# Run the game
go run tictactoe.go
```

## Controls

When prompted, enter the row and column numbers (1-3) separated by a space:
- Example: `1 2` for top-middle position
- Rows: 1 (top), 2 (middle), 3 (bottom)
- Columns: 1 (left), 2 (center), 3 (right)