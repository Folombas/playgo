package ui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Цвета
var (
	ColorDeepPurple = color.RGBA{30, 10, 60, 255}
	ColorDarkPurple = color.RGBA{40, 20, 80, 255}
	ColorDarkerPurple = color.RGBA{25, 15, 50, 255}
	ColorHotPink    = color.RGBA{255, 105, 180, 255}
	ColorCyan       = color.RGBA{0, 255, 255, 255}
	ColorGold       = color.RGBA{255, 215, 0, 255}
	ColorWhite      = color.RGBA{255, 255, 255, 255}
	ColorGray       = color.RGBA{150, 150, 150, 255}
	ColorDarkGray   = color.RGBA{80, 80, 80, 255}
	ColorGreen      = color.RGBA{50, 205, 50, 255}
	ColorOrange     = color.RGBA{255, 165, 0, 255}
	ColorRed        = color.RGBA{255, 50, 50, 255}
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

// DrawText рисует текст
func DrawText(screen *ebiten.Image, x, y int, content string, size int, col color.Color) {
	text.Draw(screen, content, basicfont.Face7x13, x, y, col)
}

// DrawCenteredText рисует текст по центру
func DrawCenteredText(screen *ebiten.Image, x, y int, content string, size int, col color.Color) {
	lines := strings.Split(content, "\n")
	lineHeight := size + 5
	startY := y - (len(lines)-1)*lineHeight/2
	
	for i, line := range lines {
		width := len(line) * size / 2
		DrawText(screen, x-width, startY+i*lineHeight, line, size, col)
	}
}

// DrawProgressBar рисует прогресс-бар
func DrawProgressBar(screen *ebiten.Image, x, y, width, height int, progress float64, bgColor, fillColor color.Color) {
	if progress > 1.0 {
		progress = 1.0
	}
	
	// Фон
	_ = bgColor
	_ = height
	_ = width
	_ = x
	_ = y
	
	// TODO: Реализовать через vector
	fmt.Sprintf("") // Заглушка
	_ = fillColor
	_ = progress
}

// FormatNumber форматирует большие числа
func FormatNumber(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}
