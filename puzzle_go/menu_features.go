// Меню с анимациями и прикольные игровые фишки
package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// ==================== МЕНЮ КНОПКА ====================
type MenuButton struct {
	X, Y, W, H   float64
	Image        *ebiten.Image
	TargetY      float64
	BobPhase     float64
	EntranceT    float64
}

func NewMenuButton(x, y, w, h float64) *MenuButton {
	return &MenuButton{X: x, Y: y - 300, W: w, H: h, TargetY: y, EntranceT: 0}
}

func (b *MenuButton) SetImage(img *ebiten.Image) { b.Image = img }

func (b *MenuButton) Update(dt float64) {
	if b.EntranceT < 1 {
		b.EntranceT += dt * 2
		if b.EntranceT > 1 { b.EntranceT = 1 }
		t := b.EntranceT
		if t < 1/2.75 {
			t = 7.5625 * t * t
		} else if t < 2/2.75 {
			t -= 1.5 / 2.75
			t = 7.5625*t*t + 0.75
		} else if t < 2.5/2.75 {
			t -= 2.25 / 2.75
			t = 7.5625*t*t + 0.9375
		} else {
			t -= 2.625 / 2.75
			t = 7.5625*t*t + 0.984375
		}
		b.Y = b.TargetY - 300*(1-t)
	}
	if b.EntranceT >= 1 {
		b.BobPhase += dt * 1.5
		b.Y = b.TargetY + math.Sin(b.BobPhase)*3
	}
}

func (b *MenuButton) Contains(mx, my float64) bool {
	return mx >= b.X-8 && mx <= b.X+b.W+8 && my >= b.Y-8 && my <= b.Y+b.H+8
}

func (b *MenuButton) Draw(screen *ebiten.Image, hover bool) {
	if b.Image != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.X+3, b.Y+3)
		op.ColorM.Scale(0.3, 0.3, 0.3, 0.5)
		screen.DrawImage(b.Image, op)
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(b.X, b.Y)
		if hover { op2.ColorM.Scale(1.15, 1.15, 1.15, 1) }
		screen.DrawImage(b.Image, op2)
		if hover {
			glow := ebiten.NewImage(int(b.W)+20, int(b.H)+20)
			glow.Fill(color.RGBA{100, 180, 255, 40})
			op3 := &ebiten.DrawImageOptions{}
			op3.GeoM.Translate(b.X-10, b.Y-10)
			screen.DrawImage(glow, op3)
		}
	}
}

// ==================== МАСКОТ ====================
type Mascot struct {
	X, Y  float64
	Phase float64
	Mood  string
}

func NewMascot(x, y float64) *Mascot {
	return &Mascot{X: x, Y: y, Mood: "happy"}
}

func (m *Mascot) Update(dt float64, combo int) {
	m.Phase += dt * 2
	if combo >= 5 { m.Mood = "shocked" }
	if combo >= 3 && combo < 5 { m.Mood = "excited" }
	if combo <= 1 { m.Mood = "happy" }
}

func (m *Mascot) Draw(screen *ebiten.Image) {
	yOff := math.Sin(m.Phase) * 5
	bodyCol := color.RGBA{255, 200, 100, 255}
	if m.Mood == "excited" { bodyCol = color.RGBA{255, 255, 100, 255} }
	if m.Mood == "shocked" { bodyCol = color.RGBA{255, 100, 100, 255} }

	body := ebiten.NewImage(40, 40)
	body.Fill(bodyCol)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(m.X, m.Y+yOff)
	screen.DrawImage(body, op)

	eyeCol := color.RGBA{0, 0, 0, 255}
	if m.Mood == "shocked" {
		eye := ebiten.NewImage(12, 12)
		eye.Fill(color.RGBA{255, 255, 255, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(m.X+5, m.Y+yOff+8)
		screen.DrawImage(eye, op)
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(m.X+23, m.Y+yOff+8)
		screen.DrawImage(eye, op2)
		pupil := ebiten.NewImage(6, 6)
		pupil.Fill(eyeCol)
		op3 := &ebiten.DrawImageOptions{}
		op3.GeoM.Translate(m.X+8, m.Y+yOff+11)
		screen.DrawImage(pupil, op3)
		op4 := &ebiten.DrawImageOptions{}
		op4.GeoM.Translate(m.X+26, m.Y+yOff+11)
		screen.DrawImage(pupil, op4)
	} else {
		eye := ebiten.NewImage(8, 8)
		eye.Fill(eyeCol)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(m.X+8, m.Y+yOff+10)
		screen.DrawImage(eye, op)
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(m.X+24, m.Y+yOff+10)
		screen.DrawImage(eye, op2)
	}

	if m.Mood == "excited" || m.Mood == "shocked" {
		mouth := ebiten.NewImage(16, 10)
		mouth.Fill(color.RGBA{200, 50, 50, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(m.X+12, m.Y+yOff+24)
		screen.DrawImage(mouth, op)
	} else {
		ebitenutil.DrawLine(screen, m.X+12, m.Y+yOff+26, m.X+28, m.Y+yOff+26, color.RGBA{0, 0, 0, 255})
	}
}

// ==================== ФЕЙЕРВЕРКИ ====================
type Firework struct {
	X, Y   float64
	Parts  []FWParticle
	Active bool
	Delay  float64
}

type FWParticle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.RGBA
}

func NewFirework(x, y float64) *Firework {
	fw := &Firework{X: x, Y: y, Active: true, Delay: 0.3}
	num := 30 + rand.Intn(20)
	baseCol := color.RGBA{uint8(100 + rand.Intn(155)), uint8(100 + rand.Intn(155)), uint8(100 + rand.Intn(155)), 255}
	for i := 0; i < num; i++ {
		angle := float64(i) * 6.2832 / float64(num)
		speed := 2 + rand.Float64()*4
		fw.Parts = append(fw.Parts, FWParticle{X: x, Y: y, VX: math.Cos(angle) * speed, VY: math.Sin(angle) * speed, Life: 1, Color: baseCol})
	}
	return fw
}

func (fw *Firework) Update(dt float64) bool {
	if fw.Delay > 0 { fw.Delay -= dt; return true }
	done := true
	for i := range fw.Parts {
		p := &fw.Parts[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.08
		p.VX *= 0.98
		p.Life -= 0.015
		if p.Life > 0 { done = false }
	}
	fw.Active = !done
	return fw.Active
}

func (fw *Firework) Draw(screen *ebiten.Image) {
	if fw.Delay > 0 { return }
	for _, p := range fw.Parts {
		if p.Life <= 0 { continue }
		s := int(4 * p.Life)
		if s < 1 { continue }
		img := ebiten.NewImage(s, s)
		img.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, uint8(p.Life * 255)})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(s)/2, p.Y-float64(s)/2)
		screen.DrawImage(img, op)
	}
}

// ==================== SPARKLE TRAIL ====================
type SparkleTrail struct {
	points []SparklePt
}

type SparklePt struct {
	X, Y, Life float64
	Color      color.RGBA
}

func NewSparkleTrail() *SparkleTrail {
	return &SparkleTrail{points: make([]SparklePt, 0)}
}

func (st *SparkleTrail) Add(x, y float64, col color.RGBA) {
	st.points = append(st.points, SparklePt{X: x, Y: y, Life: 1, Color: col})
	if len(st.points) > 50 { st.points = st.points[1:] }
}

func (st *SparkleTrail) Update() {
	for i := len(st.points) - 1; i >= 0; i-- {
		st.points[i].Life -= 0.05
		if st.points[i].Life <= 0 {
			st.points = append(st.points[:i], st.points[i+1:]...)
		}
	}
}

func (st *SparkleTrail) Draw(screen *ebiten.Image) {
	for _, p := range st.points {
		if p.Life <= 0 { continue }
		s := int(8 * p.Life)
		if s < 1 { continue }
		img := ebiten.NewImage(s, s)
		img.Fill(color.RGBA{p.Color.R, p.Color.G, p.Color.B, uint8(p.Life * 200)})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(s)/2, p.Y-float64(s)/2)
		screen.DrawImage(img, op)
	}
}

// ==================== COMBO METER ====================
type ComboMeter struct {
	X, Y    float64
	Level   int
	MaxLvl  int
	FlashT  float64
}

func NewComboMeter(x, y float64) *ComboMeter {
	return &ComboMeter{X: x, Y: y, MaxLvl: 10}
}

func (cm *ComboMeter) Update(combo int) {
	cm.Level = combo
	if combo > 1 { cm.FlashT = 1 }
	if cm.FlashT > 0 { cm.FlashT -= 0.03 }
}

func (cm *ComboMeter) Draw(screen *ebiten.Image) {
	if cm.Level <= 1 { return }
	for i := 0; i < cm.MaxLvl && i < cm.Level; i++ {
		col := color.RGBA{255, 200, 50, 255}
		if i >= 5 { col = color.RGBA{255, 100, 50, 255} }
		if i >= 8 { col = color.RGBA{255, 50, 50, 255} }
		block := ebiten.NewImage(20, 20)
		block.Fill(col)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(cm.X+float64(i*22), cm.Y)
		if cm.FlashT > 0 { op.ColorM.Scale(1.3, 1.3, 1.3, 1) }
		screen.DrawImage(block, op)
	}
}

// ==================== LEVEL PROGRESS ====================
type LevelProgress struct {
	X, Y, W, H float64
	Current    float64
}

func NewLevelProgress(x, y, w, h float64) *LevelProgress {
	return &LevelProgress{X: x, Y: y, W: w, H: h}
}

func (lp *LevelProgress) Update(score int) {
	lp.Current = float64(score % 1000)
}

func (lp *LevelProgress) Draw(screen *ebiten.Image) {
	bg := ebiten.NewImage(int(lp.W), int(lp.H))
	bg.Fill(color.RGBA{30, 30, 60, 200})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(lp.X, lp.Y)
	screen.DrawImage(bg, op)

	progress := lp.Current / 1000
	if progress > 1 { progress = 1 }
	fillW := int(lp.W * progress)
	fill := ebiten.NewImage(fillW, int(lp.H))
	fill.Fill(color.RGBA{100, 200, 255, 255})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(lp.X, lp.Y)
	screen.DrawImage(fill, op2)

	glow := ebiten.NewImage(fillW+4, int(lp.H)+4)
	glow.Fill(color.RGBA{100, 200, 255, 80})
	op3 := &ebiten.DrawImageOptions{}
	op3.GeoM.Translate(lp.X-2, lp.Y-2)
	screen.DrawImage(glow, op3)
}
