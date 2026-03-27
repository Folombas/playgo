// Package player предоставляет сущность игрока и управление им
package player

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	// Физика
	Gravity      = 0.6
	JumpForce    = -14.0
	MoveSpeed    = 5.0
	MaxFallSpeed = 10.0

	// Размеры игрока (в пикселях)
	PlayerWidth  = 32
	PlayerHeight = 32
)

// AnimationState представляет состояние анимации игрока
type AnimationState int

const (
	AnimIdle AnimationState = iota
	AnimWalk
	AnimJump
	AnimDuck
	AnimHurt
)

// Player представляет игрока (Марио)
type Player struct {
	// Позиция
	X, Y float64

	// Скорость
	VX, VY float64

	// Размеры
	Width  float32
	Height float32

	// Состояние
	OnGround      bool
	FacingRight   bool
	AnimState     AnimationState
	AnimFrame     int
	AnimTimer     int
	Invincible    bool
	InvincibleTimer int

	// Статистика
	Health    int
	MaxHealth int
	Coins     int
	Score     int
	Lives     int
}

// New создаёт нового игрока
func New(x, y float64) *Player {
	return &Player{
		X:         x,
		Y:         y,
		Width:     PlayerWidth,
		Height:    PlayerHeight,
		FacingRight: true,
		MaxHealth: 3,
		Health:    3,
		Lives:     3,
	}
}

// Update обновляет состояние игрока
func (p *Player) Update() {
	// Применяем гравитацию
	p.VY += Gravity
	if p.VY > MaxFallSpeed {
		p.VY = MaxFallSpeed
	}

	// Обновляем позицию
	p.X += p.VX
	p.Y += p.VY

	// Обновляем анимацию
	p.updateAnimation()

	// Обновляем таймер неуязвимости
	if p.Invincible {
		p.InvincibleTimer--
		if p.InvincibleTimer <= 0 {
			p.Invincible = false
		}
	}
}

// updateAnimation обновляет кадр анимации
func (p *Player) updateAnimation() {
	p.AnimTimer++

	switch p.AnimState {
	case AnimWalk:
		if p.AnimTimer >= 8 { // Скорость анимации ходьбы
			p.AnimFrame = (p.AnimFrame + 1) % 2
			p.AnimTimer = 0
		}
	case AnimIdle:
		p.AnimFrame = 0
	case AnimJump:
		p.AnimFrame = 0
	case AnimHurt:
		if p.AnimTimer >= 4 {
			p.AnimFrame = (p.AnimFrame + 1) % 2
			p.AnimTimer = 0
		}
	}
}

// HandleInput обрабатывает ввод игрока
func (p *Player) HandleInput() {
	// Движение влево/вправо
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.VX = MoveSpeed
		p.FacingRight = true
		if p.OnGround {
			p.AnimState = AnimWalk
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.VX = -MoveSpeed
		p.FacingRight = false
		if p.OnGround {
			p.AnimState = AnimWalk
		}
	} else {
		p.VX = 0
		if p.OnGround {
			p.AnimState = AnimIdle
		}
	}

	// Прыжок
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if p.OnGround {
			p.VY = JumpForce
			p.OnGround = false
			p.AnimState = AnimJump
		}
	}

	// Приседание
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		if p.OnGround {
			p.AnimState = AnimDuck
			p.VX = 0
		}
	}
}

// SetOnGround устанавливает состояние "на земле"
func (p *Player) SetOnGround(onGround bool) {
	p.OnGround = onGround
	if onGround && p.AnimState == AnimJump {
		p.AnimState = AnimIdle
	} else if !onGround {
		p.AnimState = AnimJump
	}
}

// TakeDamage наносит урон игроку
func (p *Player) TakeDamage(amount int) {
	if p.Invincible {
		return
	}

	p.Health -= amount
	p.Invincible = true
	p.InvincibleTimer = 120 // 2 секунды при 60 FPS
	p.AnimState = AnimHurt

	if p.Health <= 0 {
		p.Lives--
		if p.Lives > 0 {
			p.Health = p.MaxHealth
		}
	}
}

// CollectCoin добавляет монету
func (p *Player) CollectCoin(value int) {
	p.Coins += value
	p.Score += value * 10
}

// GetSprite возвращает текущий спрайт игрока
func (p *Player) GetSprite(assets interface{}) *ebiten.Image {
	// Приводим к правильному типу
	a := assets.(*struct {
		PlayerStand  *ebiten.Image
		PlayerWalk1  *ebiten.Image
		PlayerWalk2  *ebiten.Image
		PlayerJump   *ebiten.Image
		PlayerDuck   *ebiten.Image
		PlayerHurt   *ebiten.Image
	})

	if p.Invincible && p.AnimTimer%4 < 2 {
		// Мигание при неуязвимости
		return nil
	}

	switch p.AnimState {
	case AnimIdle:
		return a.PlayerStand
	case AnimWalk:
		if p.AnimFrame == 0 {
			return a.PlayerWalk1
		}
		return a.PlayerWalk2
	case AnimJump:
		return a.PlayerJump
	case AnimDuck:
		return a.PlayerDuck
	case AnimHurt:
		return a.PlayerHurt
	default:
		return a.PlayerStand
	}
}

// Draw рисует игрока
func (p *Player) Draw(screen *ebiten.Image, assets interface{}) {
	sprite := p.GetSprite(assets)
	if sprite == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Отражение по горизонтали если смотрим влево
	if !p.FacingRight {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(p.Width), 0)
	}

	op.GeoM.Translate(p.X, p.Y)

	// Применяем сканирование для отладки (можно убрать)
	ebitenutil.DebugPrintAt(screen, "", int(p.X), int(p.Y))

	screen.DrawImage(sprite, op)
}

// Bounds возвращает границы игрока для коллизий
func (p *Player) Bounds() (x, y, w, h float64) {
	return p.X, p.Y, float64(p.Width), float64(p.Height)
}

// IsDead возвращает true если игрок мёртв
func (p *Player) IsDead() bool {
	return p.Lives <= 0
}

// Reset сбрасывает игрока к начальным параметрам
func (p *Player) Reset(x, y float64) {
	p.X = x
	p.Y = y
	p.VX = 0
	p.VY = 0
	p.OnGround = false
	p.AnimState = AnimIdle
	p.AnimFrame = 0
	p.Invincible = false
	p.Health = p.MaxHealth
}
