package config

import "github.com/hajimehoshi/ebiten/v2"

const (
	// Screen settings
	ScreenWidth  = 960
	ScreenHeight = 640
	Title        = "Dungeon Crawler - Go365 Day 98"
	FPS          = 60

	// Tile settings
	TileSize   = 32
	MapWidth   = 30 // tiles
	MapHeight  = 20 // tiles
	ViewWidth  = ScreenWidth / TileSize
	ViewHeight = ScreenHeight / TileSize

	// Game constants
	MaxFloors     = 10
	MinRoomSize   = 4
	MaxRoomSize   = 8
	MaxRooms      = 10
	CorridorWidth = 2

	// Player constants
	PlayerSpeed     = 3
	PlayerMaxHP     = 100
	PlayerAttackDMG = 10
	InvincibleTime  = 60 // frames

	// Enemy constants
	EnemySlimeSpeed = 1
	EnemyBeeSpeed   = 2
	EnemyFlySpeed   = 2

	// Item constants
	HealPotionValue  = 30
	CoinValue        = 10
	GemValue         = 50
	KeyRequired      = 1

	// Colors (tile types)
	TileFloor      = 0
	TileWall       = 1
	TileDoor       = 2
	TileStairs     = 3
	TileSpikes     = 4
	TileWater      = 5

	// Game states
	StateMenu      = "menu"
	StatePlaying   = "playing"
	StatePaused    = "paused"
	StateGameOver  = "game_over"
	StateVictory   = "victory"
	StateNextFloor = "next_floor"
)

// Keys maps keyboard keys
type Keys struct {
	Up      ebiten.Key
	Down    ebiten.Key
	Left    ebiten.Key
	Right   ebiten.Key
	Attack  ebiten.Key
	UseItems ebiten.Key
	Pause   ebiten.Key
	Restart ebiten.Key
}

var GameKeys = Keys{
	Up:      ebiten.KeyW,
	Down:    ebiten.KeyS,
	Left:    ebiten.KeyA,
	Right:   ebiten.KeyD,
	Attack:  ebiten.KeyJ,
	UseItems: ebiten.KeyK,
	Pause:   ebiten.KeyEscape,
	Restart: ebiten.KeyEnter,
}
