// Package entity содержит все игровые сущности и анимации.
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// ======================== PARTICLE ========================

// Particle — визуальная частица для эффектов.
type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife int
	Clr           color.Color
	Sz            int
	Rotation      float64
	RotSpeed      float64
}

// Update обновляет позицию частицы. Возвращает true если жива.
func (p *Particle) Update() bool {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.1
	p.Life--
	p.Rotation += p.RotSpeed
	return p.Life > 0
}

// Alpha возвращает текущую прозрачность.
func (p *Particle) Alpha() float32 {
	return float32(p.Life) / float32(p.MaxLife)
}

// ======================== SELECT ANIM ========================

// SelectAnim — пульсирующее выделение выбранной клетки.
type SelectAnim struct {
	R, C int
	T    int
}

// Update инкрементирует кадр.
func (s *SelectAnim) Update() { s.T++ }

// Pulse возвращает значение пульсации 0..1.
func (s *SelectAnim) Pulse() float64 {
	return 0.5 + 0.5*math.Sin(float64(s.T)*0.1)
}

// ======================== REMOVE ANIM ========================

// RemoveAnim — анимация удаления гема с вспышкой.
type RemoveAnim struct {
	R, C         int
	X, Y         float64
	T, Total     int
	TypeID       int
	FirstFrame   bool
}

// Update обновляет анимацию. Возвращает true когда завершена.
func (a *RemoveAnim) Update() bool {
	if a.T == 0 { a.FirstFrame = true }
	a.T++
	return a.T >= a.Total
}

// Progress возвращает прогресс 0..1.
func (a *RemoveAnim) Progress() float64 {
	return float64(a.T) / float64(a.Total)
}

// Scale возвращает масштаб для анимации удаления.
func (a *RemoveAnim) Scale() float64 {
	return 1.0 + a.Progress()*0.5
}

// Alpha возвращает прозрачность.
func (a *RemoveAnim) Alpha() float64 {
	return 1.0 - a.Progress()
}

// ======================== SLIDE ANIM ========================

// SlideAnim — анимация перемещения гема (swap + gravity + fill).
type SlideAnim struct {
	R, C         int
	SX, SY       float64 // start
	TX, TY       float64 // target
	T, Total     int
	Spr          *ebiten.Image
	TypeID       int
}

// Update обновляет анимацию. Возвращает true когда завершена.
func (a *SlideAnim) Update() bool {
	a.T++
	return a.T >= a.Total
}

// Progress возвращает прогресс 0..1.
func (a *SlideAnim) Progress() float64 {
	return float64(a.T) / float64(a.Total)
}

// Eased возвращает eased прогресс (ease-out cubic).
func (a *SlideAnim) Eased() float64 {
	p := a.Progress()
	return 1.0 - math.Pow(1.0-p, 3)
}

// X возвращает текущую X позицию.
func (a *SlideAnim) X() float64 {
	return a.SX + (a.TX-a.SX)*a.Eased()
}

// Y возвращает текущую Y позицию.
func (a *SlideAnim) Y() float64 {
	return a.SY + (a.TY-a.SY)*a.Eased()
}

// ======================== MENU BUTTON ========================

// MenuButton — кнопка в меню.
type MenuButton struct {
	X, Y, W, H int
	Label      string
	Spr        *ebiten.Image
	Hover      bool
}

// Contains проверяет попадание курсора.
func (b *MenuButton) Contains(mx, my int) bool {
	return mx >= b.X && mx < b.X+b.W && my >= b.Y && my < b.Y+b.H
}
