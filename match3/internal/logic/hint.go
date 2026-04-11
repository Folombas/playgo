package logic

import (
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// HintSystem управляет системой подсказок
// Показывает игроку возможные ходы после бездействия
type HintSystem struct {
	hintTile1    *Tile
	hintTile2    *Tile
	lastHintTime time.Time
	hintInterval time.Duration
	showHint     bool
	hintAlpha    float64
	hintTimer    float64
}

// NewHintSystem создаёт новую систему подсказок
func NewHintSystem() *HintSystem {
	return &HintSystem{
		hintInterval: 5 * time.Second,
		lastHintTime: time.Now(),
		showHint:     false,
		hintAlpha:    0,
		hintTimer:    0,
	}
}

// Update обновляет систему подсказок
func (h *HintSystem) Update(board *Board) {
	// Обновляем таймер анимации
	if h.showHint {
		h.hintTimer += 0.05
		// Пульсация прозрачности
		h.hintAlpha = 0.5 + 0.5*math.Sin(h.hintTimer*3)
		
		// Скрываем подсказку через 3 секунды
		if h.hintTimer > 3 {
			h.showHint = false
			h.hintAlpha = 0
		}
	}

	// Проверяем, пора ли показать новую подсказку
	if time.Since(h.lastHintTime) > h.hintInterval && !h.showHint {
		h.ShowHint(board)
	}
}

// ShowHint показывает подсказку
func (h *HintSystem) ShowHint(board *Board) {
	tile1, tile2 := board.FindHint()
	if tile1 != nil && tile2 != nil {
		h.hintTile1 = tile1
		h.hintTile2 = tile2
		h.showHint = true
		h.hintTimer = 0
		h.lastHintTime = time.Now()
	}
}

// HideHint скрывает подсказку
func (h *HintSystem) HideHint() {
	h.showHint = false
	h.hintAlpha = 0
}

// IsVisible проверяет, видима ли подсказка
func (h *HintSystem) IsVisible() bool {
	return h.showHint
}

// Draw отрисовывает подсказки
func (h *HintSystem) Draw(screen *ebiten.Image, offsetX, offsetY, tileSize int) {
	if !h.showHint || h.hintTile1 == nil || h.hintTile2 == nil {
		return
	}

	// Рисуем подсветку для первого камня
	h.drawHintHighlight(screen, h.hintTile1, offsetX, offsetY, tileSize)
	// Рисуем подсветку для второго камня
	h.drawHintHighlight(screen, h.hintTile2, offsetX, offsetY, tileSize)
}

// drawHintHighlight рисует подсветку для одного камня
func (h *HintSystem) drawHintHighlight(screen *ebiten.Image, tile *Tile, offsetX, offsetY, tileSize int) {
	x := offsetX + tile.Col*tileSize
	y := offsetY + tile.Row*tileSize

	// Создаём полупрозрачный квадрат с пульсацией
	alpha := uint8(h.hintAlpha * 200)
	highlight := ebiten.NewImage(tileSize, tileSize)
	highlight.Fill(color.RGBA{255, 255, 0, alpha})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(highlight, op)
}

// Reset сбрасывает таймер подсказок
func (h *HintSystem) Reset() {
	h.lastHintTime = time.Now()
	h.showHint = false
	h.hintAlpha = 0
}
