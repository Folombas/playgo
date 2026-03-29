// Package entity - игровые сущности Food Platformer
// Go365 Day 88
package entity

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Player - игровой персонаж (повар)
type Player struct {
	X, Y         float64
	Width        float64
	Height       float64
	VX, VY       float64
	Speed        float64
	JumpForce    float64
	OnGround     bool
	Facing       int
	AnimFrame    float64
	AnimTimer    float64
	Invincible   int
	IsCrouching  bool
	IsMoving     bool
	ShootCooldown int
	Score        int
	Health       int
	MaxHealth    int
}

// NewPlayer - создание нового игрока
func NewPlayer(x, y float64) *Player {
	return &Player{
		X: x, Y: y, Width: 40, Height: 50,
		Speed: 6.0, JumpForce: -14.0,
		Facing: 1,
		Health: 100,
		MaxHealth: 100,
	}
}

// Update - обновление состояния игрока
func (p *Player) Update() {
	if p.IsMoving {
		p.AnimTimer += 0.2
		if p.AnimTimer > 4 {
			p.AnimTimer = 0
		}
		p.AnimFrame = math.Floor(p.AnimTimer)
	} else {
		p.AnimFrame = 0
	}
	
	if p.Invincible > 0 {
		p.Invincible--
	}
	
	if p.ShootCooldown > 0 {
		p.ShootCooldown--
	}
	
	p.IsMoving = false
}

// MoveLeft - движение влево
func (p *Player) MoveLeft() {
	p.VX = -p.Speed
	p.Facing = -1
	p.IsMoving = true
}

// MoveRight - движение вправо
func (p *Player) MoveRight() {
	p.VX = p.Speed
	p.Facing = 1
	p.IsMoving = true
}

// Jump - прыжок
func (p *Player) Jump() {
	p.VY = p.JumpForce
	p.OnGround = false
}

// CanJump - можно ли прыгать
func (p *Player) CanJump() bool {
	return p.OnGround && !p.IsCrouching
}

// Crouch - присесть
func (p *Player) Crouch() {
	p.IsCrouching = true
	p.Height = 30
}

// Stand - встать
func (p *Player) Stand() {
	p.IsCrouching = false
	p.Height = 50
}

// Shoot - бросок ингредиентом
func (p *Player) Shoot() {
	p.ShootCooldown = 15
}

// CanShoot - можно ли бросать
func (p *Player) CanShoot() bool {
	return p.ShootCooldown <= 0
}

// TakeDamage - получение урона
func (p *Player) TakeDamage(damage int) {
	if p.Invincible > 0 {
		return
	}
	p.Health -= damage
	p.Invincible = 120
	p.VY = -8
	p.VX = float64(-p.Facing) * 5
}

// Heal - лечение
func (p *Player) Heal(amount int) {
	p.Health += amount
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
	}
}

// Draw - отрисовка игрока
func (p *Player) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	// Мигание если неуязвим
	if p.Invincible > 0 && (p.Invincible%6 < 3) {
		return
	}
	
	// Тело (повар в белом)
	bodyColor := color.RGBA{255, 255, 255, 255}
	if p.IsCrouching {
		ebitenutil.DrawRect(screen, screenX, screenY+20, p.Width, 30, bodyColor)
	} else {
		ebitenutil.DrawRect(screen, screenX, screenY, p.Width, p.Height, bodyColor)
	}
	
	// Колпак повара
	hatColor := color.RGBA{255, 255, 255, 255}
	hatX := screenX + 5
	hatY := screenY - 10
	ebitenutil.DrawRect(screen, hatX, hatY, p.Width-10, 20, hatColor)
	ebitenutil.DrawRect(screen, hatX+5, hatY-8, p.Width-20, 10, hatColor)
	
	// Глаза
	eyeX := screenX + p.Width - 15
	if p.Facing == -1 {
		eyeX = screenX + 8
	}
	ebitenutil.DrawRect(screen, eyeX, screenY+12, 8, 8, color.RGBA{0, 0, 0, 255})
	
	// Фартук
	apronColor := color.RGBA{200, 50, 50, 255}
	ebitenutil.DrawRect(screen, screenX+8, screenY+25, p.Width-16, 20, apronColor)
}

// Food - еда
type Food struct {
	X, Y      float64
	Width     float64
	Height    float64
	Type      int // FoodType
	Value     int
	Collected bool
	AnimFrame float64
	AnimTimer float64
}

// NewFood - создание еды
func NewFood(x, y float64, foodType int, value int) *Food {
	return &Food{
		X: x, Y: y, Width: 28, Height: 28,
		Type: foodType, Value: value,
	}
}

// Update - обновление еды
func (f *Food) Update() {
	f.AnimTimer += 0.1
	if f.AnimTimer > 4 {
		f.AnimTimer = 0
	}
	f.AnimFrame = math.Floor(f.AnimTimer)
}

// Draw - отрисовка еды
func (f *Food) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if f.Collected {
		return
	}
	
	screenX := f.X - cameraX
	screenY := f.Y - cameraY
	
	// Анимация парения
	offsetY := math.Sin(f.AnimTimer*0.5) * 3
	
	// Рисуем еду как цветной круг
	c := GetFoodColorByType(f.Type)
	
	// Основная часть
	vector.DrawFilledCircle(screen, float32(screenX)+14, float32(screenY)+14+float32(offsetY), 12, c, true)
	
	// Блик
	vector.DrawFilledCircle(screen, float32(screenX)+10, float32(screenY)+10+float32(offsetY), 4, color.RGBA{255, 255, 255, 200}, true)
}

// GetFoodColorByType - получение цвета по типу еды
func GetFoodColorByType(foodType int) color.RGBA {
	switch foodType {
	case 0: // Fruit
		return color.RGBA{255, 100, 100, 255}
	case 1: // Vegetable
		return color.RGBA{100, 200, 100, 255}
	case 2: // Meat
		return color.RGBA{180, 80, 80, 255}
	case 3: // Dairy
		return color.RGBA{255, 255, 200, 255}
	case 4: // Bakery
		return color.RGBA{200, 150, 80, 255}
	case 5: // Junk
		return color.RGBA{150, 100, 50, 255}
	case 6: // Drink
		return color.RGBA{100, 150, 255, 255}
	case 7: // Sweet
		return color.RGBA{255, 100, 200, 255}
	default:
		return color.RGBA{255, 200, 50, 255}
	}
}

// Enemy - враг (испорченная еда/насекомые)
type Enemy struct {
	X, Y      float64
	Width     float64
	Height    float64
	VX, VY    float64
	Type      string // "rotten", "bug"
	AnimFrame float64
	AnimTimer float64
	Health    int
	Damage    int
}

// NewEnemy - создание врага
func NewEnemy(x, y float64, enemyType string) *Enemy {
	e := &Enemy{
		X: x, Y: y, Width: 32, Height: 32,
		Type: enemyType,
		Health: 1,
		Damage: 10,
	}
	
	switch enemyType {
	case "bug":
		e.Width = 24
		e.Height = 24
	case "rotten":
		e.Width = 36
		e.Height = 36
	}
	
	return e
}

// Update - обновление врага
func (e *Enemy) Update() {
	e.AnimTimer += 0.15
	if e.AnimTimer > 4 {
		e.AnimTimer = 0
	}
	e.AnimFrame = math.Floor(e.AnimTimer)
}

// Draw - отрисовка врага
func (e *Enemy) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	var bodyColor color.RGBA
	
	switch e.Type {
	case "rotten":
		bodyColor = color.RGBA{100, 150, 50, 255}
	case "bug":
		bodyColor = color.RGBA{80, 60, 40, 255}
	default:
		bodyColor = color.RGBA{150, 150, 150, 255}
	}
	
	// Тело
	vector.DrawFilledCircle(screen, float32(screenX)+float32(e.Width)/2, float32(screenY)+float32(e.Height)/2, float32(e.Width)/2, bodyColor, true)
	
	// Глаза
	eyeY := screenY + float64(e.Height)/3
	vector.DrawFilledCircle(screen, float32(screenX)+float32(e.Width)/3, float32(eyeY), 5, color.RGBA{255, 0, 0, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+float32(e.Width)*2/3, float32(eyeY), 5, color.RGBA{255, 0, 0, 255}, true)
	
	// Для жука - усики
	if e.Type == "bug" {
		ebitenutil.DrawLine(screen, screenX+8, screenY+5, screenX+5, screenY-5, color.RGBA{80, 60, 40, 255})
		ebitenutil.DrawLine(screen, screenX+float64(e.Width)-8, screenY+5, screenX+float64(e.Width)-5, screenY-5, color.RGBA{80, 60, 40, 255})
	}
}

// Boss - босс (Гнилой Шеф)
type Boss struct {
	X, Y        float64
	Width       float64
	Height      float64
	VX, VY      float64
	Health      int
	MaxHealth   int
	AnimFrame   float64
	AnimTimer   float64
	AttackPhase int
}

// NewBoss - создание босса
func NewBoss(x, y float64) *Boss {
	return &Boss{
		X: x, Y: y, Width: 100, Height: 120,
		Health: 100,
		MaxHealth: 100,
	}
}

// Update - обновление босса
func (b *Boss) Update() {
	b.AnimTimer += 0.1
	if b.AnimTimer > 8 {
		b.AnimTimer = 0
	}
	b.AnimFrame = math.Floor(b.AnimTimer / 2)
	
	// Лёгкое подпрыгивание
	b.Y += math.Sin(float64(b.AnimTimer)*0.3) * 0.5
}

// TakeDamage - получение урона
func (b *Boss) TakeDamage(damage int) {
	b.Health -= damage
	if b.Health < 0 {
		b.Health = 0
	}
}

// Draw - отрисовка босса
func (b *Boss) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	screenX := b.X - cameraX
	screenY := b.Y - cameraY
	
	// Тело босса (огромный испорченный повар)
	bodyColor := color.RGBA{80, 120, 60, 255}
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(b.Width), float32(b.Height), bodyColor, true)
	
	// Грязный колпак
	hatColor := color.RGBA{200, 200, 180, 255}
	vector.DrawFilledRect(screen, float32(screenX)+10, float32(screenY)-20, float32(b.Width-20), 30, hatColor, true)
	
	// Злые красные глаза
	vector.DrawFilledCircle(screen, float32(screenX)+30, float32(screenY)+35, 15, color.RGBA{255, 0, 0, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+float32(b.Width)-30, float32(screenY)+35, 15, color.RGBA{255, 0, 0, 255}, true)
	
	// Зрачки
	vector.DrawFilledCircle(screen, float32(screenX)+35, float32(screenY)+38, 7, color.RGBA{0, 0, 0, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+float32(b.Width)-35, float32(screenY)+38, 7, color.RGBA{0, 0, 0, 255}, true)
	
	// Рот с клыками
	vector.DrawFilledRect(screen, float32(screenX)+25, float32(screenY)+70, float32(b.Width-50), 20, color.RGBA{50, 30, 30, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+35, float32(screenY)+75, 5, color.RGBA{255, 255, 255, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+float32(b.Width)-35, float32(screenY)+75, 5, color.RGBA{255, 255, 255, 255}, true)
	
	// Пятна гнили
	vector.DrawFilledCircle(screen, float32(screenX)+20, float32(screenY)+50, 10, color.RGBA{60, 100, 40, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX)+float32(b.Width)-25, float32(screenY)+60, 12, color.RGBA{60, 100, 40, 255}, true)
}

// Projectile - снаряд (ингредиент)
type Projectile struct {
	X, Y      float64
	VX, VY    float64
	Width     float64
	Height    float64
	Life      int
	Active    bool
	IsEnemy   bool
	FoodType  int
	Damage    int
}

// NewProjectile - создание снаряда
func NewProjectile(x, y, vx, vy float64, foodType int, isEnemy bool) *Projectile {
	damage := 25
	if isEnemy {
		damage = 15
	}
	
	return &Projectile{
		X: x, Y: y, VX: vx, VY: vy,
		Width: 20, Height: 20,
		Life: 120,
		Active: true,
		FoodType: foodType,
		IsEnemy: isEnemy,
		Damage: damage,
	}
}

// Update - обновление снаряда
func (p *Projectile) Update() {
	p.X += p.VX
	p.Y += p.VY
	p.VY += 0.3 // Гравитация
	p.Life--
	
	if p.Life <= 0 {
		p.Active = false
	}
}

// Draw - отрисовка снаряда
func (p *Projectile) Draw(screen *ebiten.Image, cameraX, cameraY float64) {
	if !p.Active {
		return
	}
	
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	var c color.RGBA
	if p.IsEnemy {
		c = color.RGBA{100, 150, 50, 255}
	} else {
		c = GetFoodColorByType(p.FoodType)
	}
	
	vector.DrawFilledCircle(screen, float32(screenX)+10, float32(screenY)+10, 10, c, true)
	vector.DrawFilledCircle(screen, float32(screenX)+7, float32(screenY)+7, 4, color.RGBA{255, 255, 255, 200}, true)
}
