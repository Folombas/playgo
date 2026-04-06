package main

// Дополнительные эффекты для Bomberman
// Go365 Day 99 - Улучшение визуала и эффектов

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

// ======================== ЧАСТИЦЫ ========================

type BombParticle struct {
	X, Y    float64
	VX, VY  float64
	Life    int
	MaxLife int
	Color   color.Color
	Size    float64
}

var bombParticles []BombParticle

func initBombParticles() {
	bombParticles = make([]BombParticle, 0)
}

func emitExplosionParticles(x, y int, count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 1 + rand.Float64()*3
		p := BombParticle{
			X:       float64(x*Tile + Tile/2),
			Y:       float64(y*Tile + Tile/2),
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    20 + rand.Intn(20),
			MaxLife: 40,
			Size:    2 + rand.Float64()*4,
		}

		c := rand.Intn(3)
		switch c {
		case 0:
			p.Color = color.RGBA{255, 200, 50, 255}
		case 1:
			p.Color = color.RGBA{255, 100, 0, 255}
		case 2:
			p.Color = color.RGBA{255, 255, 200, 255}
		}

		bombParticles = append(bombParticles, p)
	}
}

func emitSmokeParticles(x, y int) {
	for i := 0; i < 5; i++ {
		angle := rand.Float64() * math.Pi * 2
		speed := 0.5 + rand.Float64()*1
		p := BombParticle{
			X:       float64(x*Tile + Tile/2),
			Y:       float64(y*Tile + Tile/2),
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle)*speed - 1,
			Life:    30 + rand.Intn(20),
			MaxLife: 50,
			Color:   color.RGBA{150, 150, 150, 180},
			Size:    3 + rand.Float64()*3,
		}
		bombParticles = append(bombParticles, p)
	}
}

func updateBombParticles() {
	active := make([]BombParticle, 0)
	for _, p := range bombParticles {
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.05 // гравитация
		p.Life--

		if p.Life > 0 {
			active = append(active, p)
		}
	}
	bombParticles = active
}

func drawBombParticles(screen *ebiten.Image) {
	for _, p := range bombParticles {
		alpha := float64(p.Life) / float64(p.MaxLife)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-cameraX, p.Y-cameraY)
		op.GeoM.Scale(p.Size, p.Size)

		if col, ok := p.Color.(color.RGBA); ok {
			op.ColorScale.SetR(float32(col.R) / 255)
			op.ColorScale.SetG(float32(col.G) / 255)
			op.ColorScale.SetB(float32(col.B) / 255)
			op.ColorScale.SetA(float32(col.A) / 255 * alpha)
		}

		screen.DrawImage(onePx, op)
	}
}

// ======================== АНИМАЦИИ ========================

type ScreenShake struct {
	Intensity float64
	Duration  int
}

var shake ScreenShake

func triggerShake(intensity float64, duration int) {
	shake = ScreenShake{
		Intensity: intensity,
		Duration:  duration,
	}
}

func updateShake() {
	if shake.Duration > 0 {
		shake.Duration--
	}
}

func getShakeOffset() (float64, float64) {
	if shake.Duration <= 0 {
		return 0, 0
	}
	dx := (rand.Float64() - 0.5) * shake.Intensity * 2
	dy := (rand.Float64() - 0.5) * shake.Intensity * 2
	return dx, dy
}

// ======================== ЭФФЕКТ ВЗРЫВА ========================

type ExplosionEffect struct {
	X, Y    float64
	Radius  float64
	MaxRadius float64
	Life    int
	MaxLife int
}

var explosionEffects []ExplosionEffect

func initExplosions() {
	explosionEffects = make([]ExplosionEffect, 0)
}

func triggerExplosion(x, y int) {
	effect := ExplosionEffect{
		X:         float64(x * Tile),
		Y:         float64(y * Tile),
		Radius:    0,
		MaxRadius: float64(Tile) * 0.8,
		Life:      15,
		MaxLife:   15,
	}
	explosionEffects = append(explosionEffects, effect)
	triggerShake(5, 10)
	emitExplosionParticles(x, y, 30)
}

func updateExplosionEffects() {
	active := make([]ExplosionEffect, 0)
	for _, e := range explosionEffects {
		e.Radius += (e.MaxRadius - e.Radius) * 0.3
		e.Life--
		if e.Life > 0 {
			active = append(active, e)
		}
	}
	explosionEffects = active
}

func drawExplosionEffects(screen *ebiten.Image) {
	for _, e := range explosionEffects {
		alpha := float64(e.Life) / float64(e.MaxLife)
		
		// Внешнее свечение
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-cameraX+Tile/2-e.MaxRadius, e.Y-cameraY+Tile/2-e.MaxRadius)
		op.GeoM.Scale(e.MaxRadius*2, e.MaxRadius*2)
		op.ColorScale.SetR(1)
		op.ColorScale.SetG(0.6)
		op.ColorScale.SetB(0)
		op.ColorScale.SetA(0.3 * alpha)
		screen.DrawImage(onePx, op)

		// Внутренний круг
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-cameraX+Tile/2-e.Radius, e.Y-cameraY+Tile/2-e.Radius)
		op.GeoM.Scale(e.Radius*2, e.Radius*2)
		op.ColorScale.SetR(1)
		op.ColorScale.SetG(0.9)
		op.ColorScale.SetB(0.5)
		op.ColorScale.SetA(0.7 * alpha)
		screen.DrawImage(onePx, op)
	}
}

// ======================== ПЕРЕМЕННЫЕ ========================

var (
	cameraX  float64
	cameraY  float64
	onePx    *ebiten.Image
)

func initOnePx() {
	if onePx == nil {
		onePx = ebiten.NewImage(1, 1)
		onePx.Fill(color.White)
	}
}
