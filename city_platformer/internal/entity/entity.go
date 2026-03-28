// Package entity - игровые сущности
package entity

import (
	"image/color"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var PlayerSprite *ebiten.Image

func LoadPlayerSprite() {
	if PlayerSprite != nil {
		return
	}
	img, _, err := ebitenutil.NewImageFromFile("assets/PMC_Stand.bmp")
	if err == nil {
		PlayerSprite = img
	}
}

type Player struct {
	X, Y      float64
	Width     float64
	Height    float64
	VX, VY    float64
	Speed     float64
	JumpForce float64
	OnGround  bool
	Facing    int
	AnimFrame float64
}

func NewPlayer(x, y float64) *Player {
	LoadPlayerSprite()
	return &Player{
		X: x, Y: y, Width: 32, Height: 48,
		Speed: 5.0, JumpForce: -12.0,
		Facing: 1,
	}
}

func (p *Player) Update() {
	p.AnimFrame += 0.15
}

func (p *Player) MoveLeft() {
	p.VX = -p.Speed
	p.Facing = -1
}

func (p *Player) MoveRight() {
	p.VX = p.Speed
	p.Facing = 1
}

func (p *Player) Jump() {
	p.VY = p.JumpForce
	p.OnGround = false
}

func (p *Player) CanJump() bool {
	return p.OnGround
}

func (p *Player) Draw(screen *ebiten.Image, cameraX float64) {
	screenX := p.X - cameraX

	if PlayerSprite != nil {
		opts := &ebiten.DrawImageOptions{}

		if p.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}

		frame := int(p.AnimFrame) % 2
		frameWidth := 32.0
		frameHeight := 48.0
		frameX := float64(frame) * frameWidth

		opts.GeoM.Translate(screenX, p.Y)
		opts.GeoM.Scale(1.5, 1.5)

		sprite := PlayerSprite.SubImage(image.Rect(int(frameX), 0, int(frameX)+int(frameWidth), int(frameHeight))).(*ebiten.Image)
		screen.DrawImage(sprite, opts)
	} else {
		// Резерв
		vector.DrawFilledRect(screen, float32(screenX), float32(p.Y), float32(p.Width), float32(p.Height), color.RGBA{50, 100, 50, 255}, true)
	}
}

type Platform struct {
	X, Y, Width, Height float64
	Type                string
}

type Coin struct {
	X, Y float64
}

type Enemy struct {
	X, Y, VX float64
}
