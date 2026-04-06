// Package ui — кнопки меню и экраны.
package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"

	"github.com/playgo/puzzle_go/internal/entity"
)

// LoadMenuButtons загружает кнопки главного меню.
func LoadMenuButtons() []*entity.MenuButton {
	sprPlay, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/play button.png")
	sprOpt, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Options Button.png")
	sprExit, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Exit Button.png")
	return []*entity.MenuButton{
		{X: 256, Y: 300, W: 120, H: 40, Label: "PLAY", Spr: sprPlay},
		{X: 256, Y: 360, W: 120, H: 40, Label: "OPTIONS", Spr: sprOpt},
		{X: 256, Y: 420, W: 120, H: 40, Label: "EXIT", Spr: sprExit},
	}
}

// LoadPauseButtons загружает кнопки паузы.
func LoadPauseButtons() []*entity.MenuButton {
	sprResume, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/play button.png")
	sprBack, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Back Button.png")
	return []*entity.MenuButton{
		{X: 256, Y: 280, W: 120, H: 40, Label: "RESUME", Spr: sprResume},
		{X: 256, Y: 340, W: 120, H: 40, Label: "RESTART", Spr: sprBack},
	}
}

// LoadOptionsButtons загружает кнопки настроек.
func LoadOptionsButtons() []*entity.MenuButton {
	sprBack, _, _ := ebitenutil.NewImageFromFile("assets/sprites/menu/Back Button.png")
	return []*entity.MenuButton{
		{X: 256, Y: 400, W: 120, H: 40, Label: "BACK", Spr: sprBack},
	}
}

// UpdateHover обновляет hover-состояние всех кнопок.
func UpdateHover(buttons []*entity.MenuButton, mx, my int) {
	for _, b := range buttons {
		b.Hover = b.Contains(mx, my)
	}
}

// DrawButtons рисует все кнопки.
func DrawButtons(screen *ebiten.Image, buttons []*entity.MenuButton) {
	for _, b := range buttons {
		DrawButton(screen, b)
	}
}

// DrawButton рисует одну кнопку.
func DrawButton(s *ebiten.Image, b *entity.MenuButton) {
	if b.Spr != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(b.X), float64(b.Y))
		if b.Hover {
			op.ColorScale.SetR(1.2)
			op.ColorScale.SetG(1.2)
			op.ColorScale.SetB(1.2)
		}
		s.DrawImage(b.Spr, op)
	} else {
		c := color.RGBA{60, 60, 120, 200}
		if b.Hover { c = color.RGBA{80, 80, 160, 230} }
		// Rect fallback
		img := ebiten.NewImage(b.W, b.H)
		img.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(b.X), float64(b.Y))
		s.DrawImage(img, op)
		bw := text.BoundString(basicfont.Face7x13, b.Label)
		text.Draw(s, b.Label, basicfont.Face7x13, b.X+b.W/2-bw.Dx()/2, b.Y+b.H/2+5, color.White)
	}
}
