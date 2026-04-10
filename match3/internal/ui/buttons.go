package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Button представляет интерактивную кнопку
type Button struct {
	X         int
	Y         int
	Width     int
	Height    int
	Text      string
	Action    func()
	Hovered   bool
	Pressed   bool
	Enabled   bool
	Color     color.Color
	TextColor color.Color
}

// NewButton создаёт новую кнопку
func NewButton(x, y, width, height int, text string, action func()) *Button {
	return &Button{
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		Text:      text,
		Action:    action,
		Enabled:   true,
		Color:     color.RGBA{100, 150, 255, 255},
		TextColor: color.White,
	}
}

// Update обновляет состояние кнопки
func (b *Button) Update(mx, my int) {
	if !b.Enabled {
		b.Hovered = false
		b.Pressed = false
		return
	}

	// Проверка наведения
	b.Hovered = mx >= b.X && mx <= b.X+b.Width && my >= b.Y && my <= b.Y+b.Height
	
	// Проверка клика
	if b.Hovered && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.Pressed = true
	}
	
	if b.Pressed && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if b.Hovered && b.Action != nil {
			b.Action()
		}
		b.Pressed = false
	}
}

// Draw отрисовывает кнопку
func (b *Button) Draw(screen *ebiten.Image) {
	// Определяем цвет
	baseColor := b.Color
	if !b.Enabled {
		baseColor = color.RGBA{100, 100, 100, 255}
	} else if b.Pressed {
		// Затемняем при нажатии
		r, g, b_, a := b.Color.RGBA()
		baseColor = color.RGBA{
			R: uint8(uint32(r) * 3 / 4),
			G: uint8(uint32(g) * 3 / 4),
			B: uint8(uint32(b_) * 3 / 4),
			A: uint8(a),
		}
	} else if b.Hovered {
		// Осветляем при наведении
		r, g, b_, a := b.Color.RGBA()
		baseColor = color.RGBA{
			R: uint8(min(uint32(r)*5/4, 255)),
			G: uint8(min(uint32(g)*5/4, 255)),
			B: uint8(min(uint32(b_)*5/4, 255)),
			A: uint8(a),
		}
	}

	// Фон кнопки с закруглёнными углами
	drawRoundedRect(screen, b.X, b.Y, b.Width, b.Height, 10, baseColor)
	
	// Текст
	textWidth := len(b.Text) * 7
	textHeight := 13
	textX := b.X + (b.Width-textWidth)/2
	textY := b.Y + (b.Height-textHeight)/2 + 10
	
	text.Draw(screen, b.Text, basicfont.Face7x13, textX, textY, b.TextColor)
}

// Contains проверяет, содержится ли точка в кнопке
func (b *Button) Contains(x, y int) bool {
	return x >= b.X && x <= b.X+b.Width && y >= b.Y && y <= b.Y+b.Height
}

// ProgressBar представляет полосу прогресса
type ProgressBar struct {
	X      int
	Y      int
	Width  int
	Height int
	Value  float64 // 0.0 - 1.0
	Max    float64
	Color  color.Color
}

// NewProgressBar создаёт новую полосу прогресса
func NewProgressBar(x, y, width, height int, color color.Color) *ProgressBar {
	return &ProgressBar{
		X:      x,
		Y:      y,
		Width:  width,
		Height: height,
		Value:  0,
		Max:    1.0,
		Color:  color,
	}
}

// Draw отрисовывает полосу прогресса
func (pb *ProgressBar) Draw(screen *ebiten.Image) {
	// Фон
	drawRoundedRect(screen, pb.X, pb.Y, pb.Width, pb.Height, 5, color.RGBA{50, 50, 50, 255})
	
	// Заполнение
	if pb.Value > 0 {
		fillWidth := int(float64(pb.Width) * (pb.Value / pb.Max))
		if fillWidth > 0 {
			drawRoundedRect(screen, pb.X, pb.Y, fillWidth, pb.Height, 5, pb.Color)
		}
	}
}

// SetValue устанавливает значение прогресса
func (pb *ProgressBar) SetValue(value, max float64) {
	pb.Value = value
	pb.Max = max
}

// drawRoundedRect рисует прямоугольник с закруглёнными углами
func drawRoundedRect(screen *ebiten.Image, x, y, width, height, radius int, color color.Color) {
	// Основной прямоугольник
	rect := ebiten.NewImage(width, height)
	rect.Fill(color)
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(rect, op)
	
	// Закругления (круги по углам)
	if radius > 0 {
		circle := ebiten.NewImage(radius*2, radius*2)
		circle.Fill(color)
		
		// Маска для закругления - просто рисуем кружки
		positions := []struct{ dx, dy int }{
			{0, 0},
			{width - radius*2, 0},
			{0, height - radius*2},
			{width - radius*2, height - radius*2},
		}
		
		for _, pos := range positions {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x+pos.dx), float64(y+pos.dy))
			screen.DrawImage(circle, op)
		}
	}
}

// AnimatedNumber представляет анимированное число
type AnimatedNumber struct {
	Value       int
	TargetValue int
	Animation   float64
}

// NewAnimatedNumber создаёт анимированное число
func NewAnimatedNumber(initialValue int) *AnimatedNumber {
	return &AnimatedNumber{
		Value:       initialValue,
		TargetValue: initialValue,
		Animation:   0,
	}
}

// SetTarget устанавливает целевое значение
func (an *AnimatedNumber) SetTarget(value int) {
	an.TargetValue = value
	an.Animation = 1.0
}

// Update обновляет анимацию
func (an *AnimatedNumber) Update() {
	if an.Animation > 0 {
		an.Animation -= 0.05
		if an.Animation <= 0 {
			an.Value = an.TargetValue
			an.Animation = 0
		} else {
			// Плавная интерполяция
			t := 1.0 - an.Animation
			eased := easeOutCubic(t)
			an.Value = int(float64(an.Value) + (float64(an.TargetValue)-float64(an.Value))*eased*0.1)
		}
	}
}

// easeOutCubic функция плавности
func easeOutCubic(t float64) float64 {
	return 1.0 - math.Pow(1.0-t, 3.0)
}

// DrawAnimatedNumber отрисовывает анимированное число
func DrawAnimatedNumber(screen *ebiten.Image, x, y int, number *AnimatedNumber, size int, color color.Color) {
	// Эффект "пульсации" при изменении
	// scale := 1.0
	// if number.Animation > 0 {
	// 	scale = 1.0 + number.Animation*0.2
	// }
	
	textStr := fmt.Sprintf("%d", number.Value)
	
	// Простая отрисовка (можно улучшить с масштабированием)
	text.Draw(screen, textStr, basicfont.Face7x13, x, y, color)
}

// min возвращает минимальное значение
func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}
