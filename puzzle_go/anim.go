// Анимации: gween (tween library) + визуальные эффекты
package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

// ==================== TWEEN SYSTEM ====================
type Tween struct {
	tween    *gween.Tween
	value    float32
	target   float32
	duration float32
	easing   ease.TweenFunc
	onDone   func()
	active   bool
}

func NewTween(from, to, duration float32, easeFn ease.TweenFunc) *Tween {
	return &Tween{
		value:    from,
		target:   to,
		duration: duration,
		easing:   easeFn,
		active:   true,
	}
}

func (t *Tween) OnDone(fn func()) *Tween {
	t.onDone = fn
	return t
}

func (t *Tween) Update(dt float32) {
	if !t.active { return }
	t.tween = gween.New(t.value, t.target, t.duration, t.easing)
	val, done := t.tween.Update(dt)
	t.value = val
	if done { t.active = false; if t.onDone != nil { t.onDone() } }
}

func (t *Tween) Value() float64 { return float64(t.value) }
func (t *Tween) Active() bool   { return t.active }

// ==================== ANIMATION MANAGER ====================
type AnimManager struct { tweens []*Tween }
func NewAnimManager() *AnimManager { return &AnimManager{tweens: make([]*Tween, 0)} }
func (am *AnimManager) Update(dt float32) {
	for i := len(am.tweens)-1; i >= 0; i-- {
		am.tweens[i].Update(dt)
		if !am.tweens[i].Active() { am.tweens = append(am.tweens[:i], am.tweens[i+1:]...) }
	}
}
func (am *AnimManager) Draw(*ebiten.Image) {}
func (am *AnimManager) AddTween(t *Tween)  { am.tweens = append(am.tweens, t) }

// ==================== EASING PRESETS ====================
var Easing = struct {
	OutBounce   ease.TweenFunc
	OutElastic  ease.TweenFunc
	OutBack     ease.TweenFunc
	OutCubic    ease.TweenFunc
	InOutQuad   ease.TweenFunc
	OutExpo     ease.TweenFunc
}{
	OutBounce:  ease.OutBounce,
	OutElastic: ease.OutElastic,
	OutBack:    ease.OutBack,
	OutCubic:   ease.OutCubic,
	InOutQuad:  ease.InOutQuad,
	OutExpo:    ease.OutExpo,
}

// ==================== SCREEN SHAKE ====================
type ScreenShake struct {
	intensity float64
	duration  float64
	time      float64
	active    bool
}

func NewScreenShake(intensity, duration float64) *ScreenShake {
	return &ScreenShake{intensity: intensity, duration: duration, active: true}
}

func (ss *ScreenShake) Update(dt float64) (float64, float64) {
	if !ss.active {
		return 0, 0
	}
	ss.time += dt
	if ss.time >= ss.duration {
		ss.active = false
		return 0, 0
	}
	progress := 1 - ss.time/ss.duration
	shake := ss.intensity * progress
	x := math.Sin(ss.time*50) * shake
	y := math.Cos(ss.time*37) * shake
	return x, y
}

// ==================== MENU BUTTON ANIMATION ====================
type MenuButtonAnim struct {
	tween  *gween.Tween
	y      float32
	baseY  float32
	active bool
}

func NewMenuButtonAnim(baseY float32, delay float32) *MenuButtonAnim {
	mb := &MenuButtonAnim{
		y:     baseY - 200,
		baseY: baseY,
	}
	mb.tween = gween.New(mb.y, baseY, 0.6, ease.OutBounce)
	mb.active = true
	return mb
}

func (mb *MenuButtonAnim) Update(dt float32) {
	if mb.active && mb.tween != nil {
		val, done := mb.tween.Update(dt)
		mb.y = val
		if done { mb.active = false }
	}
}

func (mb *MenuButtonAnim) Y() float64 { return float64(mb.y) }

// ==================== FLOATING HEART ====================
type FloatingEmoji struct {
	x, y, vy float64
	life     float64
	emoji    string
}

func NewFloatingEmoji(x, y float64, emoji string) *FloatingEmoji {
	return &FloatingEmoji{x: x, y: y, vy: -2, life: 1, emoji: emoji}
}

func (fe *FloatingEmoji) Update(dt float64) bool {
	fe.y += fe.vy
	fe.vy *= 0.95
	fe.life -= dt * 0.8
	return fe.life > 0
}

func (fe *FloatingEmoji) Draw(screen *ebiten.Image) {
	if fe.life <= 0 {
		return
	}
	alpha := uint8(fe.life * 255)
	// Рисуем цветной кружок вместо эмодзи
	img := ebiten.NewImage(20, 20)
	img.Fill(color.RGBA{255, 100, 150, alpha})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(fe.x-10, fe.y-10)
	op.ColorM.Scale(1, 1, 1, fe.life)
	screen.DrawImage(img, op)
}
