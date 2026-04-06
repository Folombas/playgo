package entity

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Common colors
var (
	ColorDarkBG = color.RGBA{20, 20, 30, 255}
)

// DamageColor represents a color for damage numbers
type DamageColor struct {
	R, G, B uint8
}

func (c DamageColor) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) * 0x101, uint32(c.G) * 0x101, uint32(c.B) * 0x101, 0xFFFF
}

// AttackHitBox represents attack hitbox for collision
type AttackHitBox struct {
	X, Y, Width, Height float64
}

func (a *AttackHitBox) GetX() float64      { return a.X }
func (a *AttackHitBox) GetY() float64      { return a.Y }
func (a *AttackHitBox) GetWidth() float64  { return a.Width }
func (a *AttackHitBox) GetHeight() float64 { return a.Height }

// Entity interface
type Entity interface {
	Update() error
	Draw(screen *ebiten.Image, offsetX, offsetY float64)
	GetX() float64
	GetY() float64
	GetWidth() float64
	GetHeight() float64
	IsActive() bool
	SetActive(bool)
}

// Base entity
type Base struct {
	X, Y     float64
	Width    float64
	Height   float64
	Active   bool
	Image    *ebiten.Image
}

func (b *Base) Update() error { return nil }

func (b *Base) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if b.Image == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(b.X-offsetX, b.Y-offsetY)
	screen.DrawImage(b.Image, op)
}

func (b *Base) GetX() float64     { return b.X }
func (b *Base) GetY() float64     { return b.Y }
func (b *Base) GetWidth() float64 { return b.Width }
func (b *Base) GetHeight() float64 { return b.Height }
func (b *Base) IsActive() bool    { return b.Active }
func (b *Base) SetActive(v bool)  { b.Active = v }

// CollidesWith checks AABB collision
func CollidesWith(a, b interface{ GetX() float64; GetY() float64; GetWidth() float64; GetHeight() float64 }) bool {
	return a.GetX() < b.GetX()+b.GetWidth() &&
		a.GetX()+a.GetWidth() > b.GetX() &&
		a.GetY() < b.GetY()+b.GetHeight() &&
		a.GetY()+a.GetHeight() > b.GetY()
}

// DamageNumber represents floating damage text
type DamageNumber struct {
	X, Y    float64
	Value   int
	Life    int
	MaxLife int
	Color   color.Color
}

func NewDamageNumber(x, y float64, value int, c color.Color) *DamageNumber {
	return &DamageNumber{
		X: x, Y: y, Value: value,
		Life: 60, MaxLife: 60, Color: c,
	}
}

func (d *DamageNumber) Update() {
	d.Y -= 0.5
	d.Life--
}

func (d *DamageNumber) IsActive() bool {
	return d.Life > 0
}
