// Package config содержит все константы и конфигурацию игры.
package config

// Board dimensions
const (
	Cols  = 8
	Rows  = 8
	Tile  = 64
	GemTypes = 6
	MatchMin = 3
)

// Layout
const (
	BoardOffX = 32
	BoardOffY = 80
	HUD       = 60
	WinW      = Cols*Tile + BoardOffX*2
	WinH      = Rows*Tile + BoardOffY + HUD
)

// Gameplay
const (
	TargetScore  = 5000
	CascadeDelay = 25 // frames before cascade check
)

// Game states
type State int

const (
	StateMenu State = iota
	StatePlay
	StatePause
	StateOptions
	StateWin
	StateNoMoves
)
