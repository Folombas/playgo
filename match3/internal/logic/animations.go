package logic

import (
	"math"
)

// SwapAnimation представляет анимацию обмена двух камней
type SwapAnimation struct {
	Tile1      *Tile
	Tile2      *Tile
	Progress   float64 // 0.0 - 1.0
	Duration   float64 // в секундах
	StartX1    float64
	StartY1    float64
	StartX2    float64
	StartY2    float64
	TargetX1   float64
	TargetY1   float64
	TargetX2   float64
	TargetY2   float64
	Complete   bool
}

// NewSwapAnimation создаёт новую анимацию обмена
func NewSwapAnimation(tile1, tile2 *Tile, tileSize int) *SwapAnimation {
	// Рассчитываем позиции
	startX1 := float64(tile1.Col * tileSize)
	startY1 := float64(tile1.Row * tileSize)
	startX2 := float64(tile2.Col * tileSize)
	startY2 := float64(tile2.Row * tileSize)
	
	// Целевые позиции (поменянные местами)
	targetX1 := startX2
	targetY1 := startY2
	targetX2 := startX1
	targetY2 := startY1
	
	return &SwapAnimation{
		Tile1:    tile1,
		Tile2:    tile2,
		Progress: 0,
		Duration: 0.3, // 300ms
		StartX1:  startX1,
		StartY1:  startY1,
		StartX2:  startX2,
		StartY2:  startY2,
		TargetX1: targetX1,
		TargetY1: targetY1,
		TargetX2: targetX2,
		TargetY2: targetY2,
		Complete: false,
	}
}

// Update обновляет анимацию
func (sa *SwapAnimation) Update(deltaTime float64) {
	if sa.Complete {
		return
	}
	
	sa.Progress += deltaTime / sa.Duration
	
	if sa.Progress >= 1.0 {
		sa.Progress = 1.0
		sa.Complete = true
	}
}

// GetOffset1 возвращает смещение для первого камня
func (sa *SwapAnimation) GetOffset1() (float64, float64) {
	t := easeInOutCubic(sa.Progress)
	x := sa.StartX1 + (sa.TargetX1-sa.StartX1)*t - sa.StartX1
	y := sa.StartY1 + (sa.TargetY1-sa.StartY1)*t - sa.StartY1
	return x, y
}

// GetOffset2 возвращает смещение для второго камня
func (sa *SwapAnimation) GetOffset2() (float64, float64) {
	t := easeInOutCubic(sa.Progress)
	x := sa.StartX2 + (sa.TargetX2-sa.StartX2)*t - sa.StartX2
	y := sa.StartY2 + (sa.TargetY2-sa.StartY2)*t - sa.StartY2
	return x, y
}

// easeInOutCubic функция плавности
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// IsComplete проверяет, завершена ли анимация
func (sa *SwapAnimation) IsComplete() bool {
	return sa.Complete
}
