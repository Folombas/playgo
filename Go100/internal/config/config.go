package config

import "github.com/hajimehoshi/ebiten/v2"

const (
	// Screen
	ScreenWidth  = 1024
	ScreenHeight = 768
	Title        = "Tower Defense - Go365 Day 100"
	FPS          = 60

	// Grid
	TileSize   = 48
	GridWidth  = 20
	GridHeight = 14
	GridOffsetX = (ScreenWidth - GridWidth*TileSize) / 2
	GridOffsetY = 60

	// Game
	InitialGold   = 200
	InitialLives  = 20
	MaxWaves      = 30
	WaveInterval  = 180 // frames between waves

	// Colors
	CGold      = 0xFFD700
	CRed       = 0xFF4444
	CBlue      = 0x4488FF
	CGreen     = 0x44FF44
	CPurple    = 0xAA44FF
	COrange    = 0xFF8844
	CWhite     = 0xFFFFFF
	CBlack     = 0x000000
	CGray      = 0x888888
	CDarkGray  = 0x444444
	CYellow    = 0xFFFF44
	CCyan      = 0x44FFFF
)

// Keys
var (
	KeyPause   = ebiten.KeyEscape
	KeySpeedUp = ebiten.KeySpace
	KeyRestart = ebiten.KeyEnter
)

// Tower types
const (
	TowerBasic = iota
	TowerSniper
	TowerSlow
	TowerSplash
	TowerLaser
)

// Enemy types
const (
	EnemyBasic = iota
	EnemyFast
	EnemyTank
	EnemyBoss
	EnemySwarm
)
