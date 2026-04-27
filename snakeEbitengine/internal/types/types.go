package types

import (
	"image/color"
)

const (
	ScreenW       = 1280
	ScreenH       = 800
	TileSize      = 32
	GridW         = ScreenW / TileSize
	GridH         = ScreenH / TileSize
	InitialLength = 5
	MaxHealth     = 100
)

type Vec struct{ X, Y int }

type GameState int

const (
	STATE_MENU GameState = iota
	STATE_PLAYING
	STATE_PAUSED
	STATE_GAMEOVER
	STATE_SETTINGS
)

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
	Glow   bool
}

type Bomb struct {
	X, Y  int
	Timer float64
}

type IceBlock struct{ X, Y int }

type Viking struct {
	X, Y   int
	Frame  int
	Timer  float64
	Active bool
}

type Gift struct {
	X, Y   int
	Color  int
	Opened bool
	Life   float64
}

type Coin struct {
	X, Y   int
	Frame  int
	Timer  float64
	Life   float64
}

type KeyOnField struct {
	X, Y   int
	Active bool
	Life   float64
}

const (
	FruitApple = iota
	FruitStrawberry
	FruitOrange
	FruitBanana
	FruitPineapple
)