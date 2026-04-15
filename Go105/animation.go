package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// LineAnim представляет анимацию удаления линии
type LineAnim struct {
	Row    int
	Frame  int
	MaxFrame int
}

// Update обновляет анимацию
func (a *LineAnim) Update() bool {
	a.Frame++
	return a.Frame >= a.MaxFrame
}

// Draw рисует анимацию вспышки
func (a *LineAnim) Draw(screen *ebiten.Image) {
	progress := float64(a.Frame) / float64(a.MaxFrame)
	alpha := uint8(255 * (1 - progress))

	// Вспышка на линии
	c := color.RGBA{255, 255, 255, alpha}
	vector.DrawFilledRect(screen,
		float32(boardOffsetX),
		float32(boardOffsetY+a.Row*cellSize),
		float32(boardCols*cellSize),
		float32(cellSize),
		c,
		false,
	)

	// Пульсация
	scale := float32(1.0 + 0.3*math.Sin(progress*math.Pi))
	vector.StrokeRect(screen,
		float32(boardOffsetX),
		float32(boardOffsetY+a.Row*cellSize),
		float32(boardCols*cellSize),
		float32(cellSize),
		2*scale,
		color.RGBA{255, 255, 100, alpha},
		false,
	)
}

// BackgroundStars представляет анимированный фон
type BackgroundStars struct {
	Stars []Star
}

// Star — одна звезда
type Star struct {
	X, Y   float64
	Speed  float64
	Size   float64
	Brightness float64
}

// NewBackgroundStars создаёт звёздный фон
func NewBackgroundStars() *BackgroundStars {
	bg := &BackgroundStars{}
	for i := 0; i < 50; i++ {
		bg.Stars = append(bg.Stars, Star{
			X:          float64(randomInt(0, screenWidth)),
			Y:          float64(randomInt(0, screenHeight)),
			Speed:      0.2 + float64(randomInt(1, 5))/10.0,
			Size:       1 + float64(randomInt(0, 2)),
			Brightness: float64(randomInt(100, 255)),
		})
	}
	return bg
}

// Update обновляет звёзды
func (bg *BackgroundStars) Update() {
	for i := range bg.Stars {
		bg.Stars[i].Y += bg.Stars[i].Speed
		if bg.Stars[i].Y > screenHeight {
			bg.Stars[i].Y = 0
			bg.Stars[i].X = float64(randomInt(0, screenWidth))
		}
	}
}

// Draw рисует звёзды
func (bg *BackgroundStars) Draw(screen *ebiten.Image) {
	for _, star := range bg.Stars {
		twinkle := star.Brightness * (0.7 + 0.3*math.Sin(float64(ebiten.ActualFPS()*60)+star.X))
		alpha := uint8(twinkle)
		c := color.RGBA{255, 255, 255, alpha}
		vector.DrawFilledCircle(screen,
			float32(star.X),
			float32(star.Y),
			float32(star.Size),
			c,
			false,
		)
	}
}

// PulseEffect — эффект пульсации при установке
type PulseEffect struct {
	X, Y  float64
	Frame int
	Max   int
}

// Update обновляет пульсацию
func (p *PulseEffect) Update() bool {
	p.Frame++
	return p.Frame >= p.Max
}

// Draw рисует пульсацию
func (p *PulseEffect) Draw(screen *ebiten.Image) {
	progress := float64(p.Frame) / float64(p.Max)
	radius := 20 + 40*progress
	alpha := uint8(200 * (1 - progress))

	c := color.RGBA{255, 255, 255, alpha}
	vector.StrokeCircle(screen,
		float32(p.X),
		float32(p.Y),
		float32(radius),
		2,
		c,
		false,
	)
}
