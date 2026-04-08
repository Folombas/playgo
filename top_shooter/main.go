// Top-Down Shooter - Go365 Challenge Day 102
// Шутер вид сверху с волнами врагов
// 8 апреля 2026

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ============================================================================
// КОНСТАНТЫ
// ============================================================================

const (
	ScreenW = 1200
	ScreenH = 800
	PlayerSpeed = 4
	BulletSpeed = 10
	EnemyBaseSpeed = 1.5
	FireRate = 8 // кадров между выстрелами
)

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func angleTo(x1, y1, x2, y2 float64) float64 {
	return math.Atan2(y2-y1, x2-x1)
}

func createCircleImage(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2
	radius := size / 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			if dx*dx+dy*dy <= float64(radius*radius) {
				img.Set(x, y, c)
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

func createPlayerSprite(size int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2

	// Корпус танка
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(size/2-2) {
				// Зелёный корпус
				img.Set(x, y, color.RGBA{50, 180, 80, 255})
			}
			if dist <= float64(size/2-6) {
				img.Set(x, y, color.RGBA{60, 200, 90, 255})
			}

			// Пушка (направление вверх)
			if x >= center-3 && x <= center+3 && y <= center+5 && y >= 2 {
				img.Set(x, y, color.RGBA{100, 100, 100, 255})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

func createEnemySprite(size int, enemyType int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2

	var bodyColor, accentColor color.RGBA
	switch enemyType {
	case 0: // Обычный - красный
		bodyColor = color.RGBA{200, 60, 60, 255}
		accentColor = color.RGBA{220, 80, 80, 255}
	case 1: // Быстрый - оранжевый
		bodyColor = color.RGBA{220, 150, 50, 255}
		accentColor = color.RGBA{240, 170, 70, 255}
	case 2: // Танк - фиолетовый
		bodyColor = color.RGBA{150, 60, 200, 255}
		accentColor = color.RGBA{170, 80, 220, 255}
	default:
		bodyColor = color.RGBA{200, 60, 60, 255}
		accentColor = color.RGBA{220, 80, 80, 255}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(size/2-2) {
				img.Set(x, y, bodyColor)
			}
			if dist <= float64(size/2-6) {
				img.Set(x, y, accentColor)
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

func createBulletImage(size int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= float64(size/2-1) {
				img.Set(x, y, color.RGBA{255, 255, 100, 255})
			}
			if dist <= 2 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// ============================================================================
// ТИПЫ
// ============================================================================

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateGameOver
)

type Player struct {
	X, Y        float64
	Angle       float64
	HP          int
	MaxHP       int
	Speed       float64
	FireTimer   int
	Invincible  int
}

type Bullet struct {
	X, Y   float64
	VX, VY float64
	Life   int
}

type Enemy struct {
	X, Y     float64
	Type     int // 0=normal, 1=fast, 2=tank
	HP       int
	MaxHP    int
	Speed    float64
	Size     float64
}

type Particle struct {
	X, Y       float64
	VX, VY     float64
	Life       float64
	MaxLife    float64
	Color      color.RGBA
	Size       float64
}

type Pickup struct {
	X, Y float64
	Type int // 0=health, 1=rapid fire
	Life int
}

type Game struct {
	State     GameState
	Player    *Player
	Bullets   []Bullet
	Enemies   []Enemy
	Particles []Particle
	Pickups   []Pickup

	Score     int
	Wave      int
	EnemiesKilled int
	WaveTimer int
	SpawnTimer int

	Keys     map[ebiten.Key]bool
	MouseX   float64
	MouseY   float64

	PlayerImg  *ebiten.Image
	EnemyImgs  []*ebiten.Image
	BulletImg  *ebiten.Image
	StarField  []struct{ X, Y, S float64 }

	GameTime float64
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		State:  StateMenu,
		Player: &Player{
			X:       ScreenW / 2,
			Y:       ScreenH / 2,
			HP:      100,
			MaxHP:   100,
			Speed:   PlayerSpeed,
		},
		Bullets:   []Bullet{},
		Enemies:   []Enemy{},
		Particles: []Particle{},
		Pickups:   []Pickup{},
		Wave:      1,
		Keys:      make(map[ebiten.Key]bool),
	}

	// Create sprites
	g.PlayerImg = createPlayerSprite(40)
	g.EnemyImgs = []*ebiten.Image{
		createEnemySprite(30, 0),
		createEnemySprite(24, 1),
		createEnemySprite(36, 2),
	}
	g.BulletImg = createBulletImage(8)

	// Star field
	for i := 0; i < 100; i++ {
		g.StarField = append(g.StarField, struct{ X, Y, S float64 }{
			X: rand.Float64() * ScreenW,
			Y: rand.Float64() * ScreenH,
			S: rand.Float64()*2 + 0.5,
		})
	}

	return g
}

func (g *Game) Update() error {
	g.GameTime += 1.0 / 60.0

	mx, my := ebiten.CursorPosition()
	g.MouseX = float64(mx)
	g.MouseY = float64(my)

	// Key states
	for _, key := range []ebiten.Key{
		ebiten.KeyW, ebiten.KeyA, ebiten.KeyS, ebiten.KeyD,
		ebiten.KeyArrowUp, ebiten.KeyArrowLeft, ebiten.KeyArrowDown, ebiten.KeyArrowRight,
	} {
		g.Keys[key] = ebiten.IsKeyPressed(key)
	}

	switch g.State {
	case StateMenu:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			// Start button
			if g.MouseX >= ScreenW/2-100 && g.MouseX <= ScreenW/2+100 &&
				g.MouseY >= 480 && g.MouseY <= 540 {
				g.startGame()
			}
		}

	case StatePlaying:
		g.updatePlaying()

	case StateGameOver:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.MouseX >= ScreenW/2-100 && g.MouseX <= ScreenW/2+100 &&
				g.MouseY >= 520 && g.MouseY <= 580 {
				g.startGame()
			}
			if g.MouseX >= ScreenW/2-100 && g.MouseX <= ScreenW/2+100 &&
				g.MouseY >= 600 && g.MouseY <= 660 {
				g.State = StateMenu
			}
		}
	}

	return nil
}

func (g *Game) updatePlaying() {
	p := g.Player

	// Movement
	dx, dy := 0.0, 0.0
	if g.Keys[ebiten.KeyW] || g.Keys[ebiten.KeyArrowUp] {
		dy = -1
	}
	if g.Keys[ebiten.KeyS] || g.Keys[ebiten.KeyArrowDown] {
		dy = 1
	}
	if g.Keys[ebiten.KeyA] || g.Keys[ebiten.KeyArrowLeft] {
		dx = -1
	}
	if g.Keys[ebiten.KeyD] || g.Keys[ebiten.KeyArrowRight] {
		dx = 1
	}

	if dx != 0 || dy != 0 {
		length := math.Sqrt(dx*dx + dy*dy)
		p.X += (dx / length) * p.Speed
		p.Y += (dy / length) * p.Speed
	}

	// Clamp to screen
	p.X = math.Max(20, math.Min(ScreenW-20, p.X))
	p.Y = math.Max(20, math.Min(ScreenH-20, p.Y))

	// Aim angle
	p.Angle = angleTo(p.X, p.Y, g.MouseX, g.MouseY)

	// Shooting
	if p.FireTimer > 0 {
		p.FireTimer--
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && p.FireTimer <= 0 {
		bx := p.X + math.Cos(p.Angle)*20
		by := p.Y + math.Sin(p.Angle)*20

		g.Bullets = append(g.Bullets, Bullet{
			X:  bx,
			Y:  by,
			VX: math.Cos(p.Angle) * BulletSpeed,
			VY: math.Sin(p.Angle) * BulletSpeed,
			Life: 60,
		})

		p.FireTimer = FireRate

		// Muzzle flash particles
		for i := 0; i < 5; i++ {
			angle := p.Angle + (rand.Float64()-0.5)*0.5
			speed := 3 + rand.Float64()*3
			g.Particles = append(g.Particles, Particle{
				X: bx, Y: by,
				VX: math.Cos(angle) * speed,
				VY: math.Sin(angle) * speed,
				Life: 0.3, MaxLife: 0.3,
				Color: color.RGBA{255, 200, 50, 255},
				Size: 2 + rand.Float64()*2,
			})
		}
	}

	// Invincibility timer
	if p.Invincible > 0 {
		p.Invincible--
	}

	// Update bullets
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := &g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life--

		if b.Life <= 0 || b.X < 0 || b.X > ScreenW || b.Y < 0 || b.Y > ScreenH {
			g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
			continue
		}

		// Check enemy hits
		hit := false
		for j := len(g.Enemies) - 1; j >= 0; j-- {
			e := &g.Enemies[j]
			if dist(b.X, b.Y, e.X, e.Y) < e.Size/2+4 {
				e.HP--
				hit = true

				// Hit particles
				for k := 0; k < 8; k++ {
					angle := rand.Float64() * 6.2832
					speed := 2 + rand.Float64()*3
					g.Particles = append(g.Particles, Particle{
						X: b.X, Y: b.Y,
						VX: math.Cos(angle) * speed,
						VY: math.Sin(angle) * speed,
						Life: 0.5, MaxLife: 0.5,
						Color: color.RGBA{255, 100, 50, 255},
						Size: 3 + rand.Float64()*3,
					})
				}

				if e.HP <= 0 {
					g.EnemiesKilled++
					g.Score += (e.Type + 1) * 100

					// Death explosion
					for k := 0; k < 20; k++ {
						angle := rand.Float64() * 6.2832
						speed := 3 + rand.Float64()*5
						c := color.RGBA{255, 150, 50, 255}
						if e.Type == 1 {
							c = color.RGBA{255, 200, 50, 255}
						} else if e.Type == 2 {
							c = color.RGBA{180, 100, 255, 255}
						}
						g.Particles = append(g.Particles, Particle{
							X: e.X, Y: e.Y,
							VX: math.Cos(angle) * speed,
							VY: math.Sin(angle) * speed,
							Life: 0.8 + rand.Float64()*0.4,
							MaxLife: 1.2,
							Color: c,
							Size: 4 + rand.Float64()*5,
						})
					}

					// Drop pickup chance
					if rand.Float64() < 0.15 {
						pickType := 0
						if rand.Float64() < 0.5 {
							pickType = 1
						}
						g.Pickups = append(g.Pickups, Pickup{
							X: e.X, Y: e.Y,
							Type: pickType,
							Life: 600,
						})
					}

					g.Enemies = append(g.Enemies[:j], g.Enemies[j+1:]...)
				}
				break
			}
		}
		if hit {
			g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
		}
	}

	// Update enemies
	for _, e := range g.Enemies {
		angle := angleTo(e.X, e.Y, p.X, p.Y)
		e.X += math.Cos(angle) * e.Speed
		e.Y += math.Sin(angle) * e.Speed

		// Damage player on contact
		if dist(e.X, e.Y, p.X, p.Y) < e.Size/2+15 && p.Invincible <= 0 {
			damage := 10
			if e.Type == 2 {
				damage = 20
			}
			p.HP -= damage
			p.Invincible = 30

			// Hit particles
			for i := 0; i < 10; i++ {
				angle := rand.Float64() * 6.2832
				speed := 3 + rand.Float64()*4
				g.Particles = append(g.Particles, Particle{
					X: p.X, Y: p.Y,
					VX: math.Cos(angle) * speed,
					VY: math.Sin(angle) * speed,
					Life: 0.5, MaxLife: 0.5,
					Color: color.RGBA{255, 50, 50, 255},
					Size: 3 + rand.Float64()*3,
				})
			}

			if p.HP <= 0 {
				g.State = StateGameOver
			}
		}
	}

	// Update pickups
	for i := len(g.Pickups) - 1; i >= 0; i-- {
		pick := &g.Pickups[i]
		pick.Life--

		if pick.Life <= 0 {
			g.Pickups = append(g.Pickups[:i], g.Pickups[i+1:]...)
			continue
		}

		if dist(pick.X, pick.Y, p.X, p.Y) < 30 {
			if pick.Type == 0 {
				p.HP = int(math.Min(float64(p.MaxHP), float64(p.HP+30)))
			} else {
				p.FireTimer = -60 // Rapid fire for 1 sec
			}
			g.Pickups = append(g.Pickups[:i], g.Pickups[i+1:]...)
		}
	}

	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.05
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	// Wave system
	g.WaveTimer++
	if g.WaveTimer >= 600 { // New wave every 10 seconds
		g.Wave++
		g.WaveTimer = 0
	}

	g.SpawnTimer++
	spawnRate := math.Max(15, 60-float64(g.Wave)*5)
	if g.SpawnTimer >= int(spawnRate) {
		g.spawnEnemy()
		g.SpawnTimer = 0
	}
}

func (g *Game) spawnEnemy() {
	// Spawn from edges
	var x, y float64
	side := rand.Intn(4)
	switch side {
	case 0: // top
		x = rand.Float64() * ScreenW
		y = -30
	case 1: // right
		x = ScreenW + 30
		y = rand.Float64() * ScreenH
	case 2: // bottom
		x = rand.Float64() * ScreenW
		y = ScreenH + 30
	case 3: // left
		x = -30
		y = rand.Float64() * ScreenH
	}

	// Enemy type based on wave
	var eType int
	r := rand.Float64()
	if g.Wave < 3 {
		eType = 0
	} else if g.Wave < 6 {
		if r < 0.7 {
			eType = 0
		} else {
			eType = 1
		}
	} else {
		if r < 0.5 {
			eType = 0
		} else if r < 0.8 {
			eType = 1
		} else {
			eType = 2
		}
	}

	hp := 1
	speed := EnemyBaseSpeed
	size := 30.0

	switch eType {
	case 1: // fast
		hp = 1
		speed = EnemyBaseSpeed * 2
		size = 24
	case 2: // tank
		hp = 3
		speed = EnemyBaseSpeed * 0.6
		size = 36
	}

	// Scale with wave
	hp += g.Wave / 5

	g.Enemies = append(g.Enemies, Enemy{
		X: x, Y: y,
		Type:  eType,
		HP:    hp,
		MaxHP: hp,
		Speed: speed,
		Size:  size,
	})
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Player = &Player{
		X: ScreenW / 2,
		Y: ScreenH / 2,
		HP: 100, MaxHP: 100,
		Speed: PlayerSpeed,
	}
	g.Bullets = []Bullet{}
	g.Enemies = []Enemy{}
	g.Particles = []Particle{}
	g.Pickups = []Pickup{}
	g.Score = 0
	g.Wave = 1
	g.EnemiesKilled = 0
	g.WaveTimer = 0
	g.SpawnTimer = 0
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{12, 14, 25, 255}, false)

	// Stars
	for _, star := range g.StarField {
		alpha := uint8(100 + 80*math.Sin(g.GameTime*2+star.X))
		s := int(star.S)
		if s < 1 {
			s = 1
		}
		starImg := ebiten.NewImage(s*2, s*2)
		starImg.Fill(color.RGBA{200, 200, 255, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(star.X, star.Y)
		screen.DrawImage(starImg, op)
	}

	// Grid lines (subtle)
	for x := 0; x < ScreenW; x += 80 {
		vector.StrokeLine(screen, float32(x), 0, float32(x), ScreenH, 1, color.RGBA{30, 35, 55, 80}, false)
	}
	for y := 0; y < ScreenH; y += 80 {
		vector.StrokeLine(screen, 0, float32(y), ScreenW, float32(y), 1, color.RGBA{30, 35, 55, 80}, false)
	}

	switch g.State {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawGame(screen)
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	}

	// Particles (always)
	for _, p := range g.Particles {
		size := int(p.Size * (p.Life / p.MaxLife))
		if size < 1 {
			continue
		}
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		pImg := ebiten.NewImage(size, size)
		pImg.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
		screen.DrawImage(pImg, op)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title
	ebitenutil.DebugPrintAt(screen, "⚔ TOP-DOWN SHOOTER ⚔", ScreenW/2-180, 250)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 102", ScreenW/2-130, 300)

	// Decorative line
	vector.StrokeLine(screen, ScreenW/2-200, 340, ScreenW/2+200, 340, 3, color.RGBA{100, 200, 255, 255}, false)

	// Crosshair
	crossImg := createCircleImage(60, color.RGBA{100, 200, 255, 100})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenW/2-30, 370)
	screen.DrawImage(crossImg, op)

	// Start button
	g.drawButton(screen, "▶  НАЧАТЬ ИГРУ", ScreenW/2-120, 480, 240, 60)

	// Controls
	ebitenutil.DebugPrintAt(screen, "WASD / Стрелки - Движение", ScreenW/2-130, 580)
	ebitenutil.DebugPrintAt(screen, "ЛКМ - Стрельба", ScreenW/2-90, 610)
	ebitenutil.DebugPrintAt(screen, "Выживай как можно дольше!", ScreenW/2-130, 650)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	p := g.Player

	// Pickups
	for _, pick := range g.Pickups {
		alpha := 255
		if pick.Life < 120 {
			alpha = int(float64(pick.Life) / 120 * 255)
		}

		var c color.RGBA
		if pick.Type == 0 {
			c = color.RGBA{50, 255, 50, uint8(alpha)}
		} else {
			c = color.RGBA{255, 255, 50, uint8(alpha)}
		}

		pickImg := createCircleImage(20, c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pick.X-10, pick.Y-10)
		screen.DrawImage(pickImg, op)

		// Icon
		icon := "+"
		if pick.Type == 1 {
			icon = "⚡"
		}
		ebitenutil.DebugPrintAt(screen, icon, int(pick.X)-5, int(pick.Y)-6)
	}

	// Bullets
	for _, b := range g.Bullets {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.X-4, b.Y-4)
		screen.DrawImage(g.BulletImg, op)
	}

	// Enemies
	for _, e := range g.Enemies {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-e.Size/2, e.Y-e.Size/2)
		screen.DrawImage(g.EnemyImgs[e.Type], op)

		// HP bar
		if e.HP < e.MaxHP {
			barW := e.Size
			barH := 4
			hpRatio := float64(e.HP) / float64(e.MaxHP)

			vector.DrawFilledRect(screen, float32(e.X-barW/2), float32(e.Y-e.Size/2-10), float32(barW), float32(barH), color.RGBA{80, 0, 0, 200}, false)
			vector.DrawFilledRect(screen, float32(e.X-barW/2), float32(e.Y-e.Size/2-10), float32(barW)*float32(hpRatio), float32(barH), color.RGBA{255, 50, 50, 255}, false)
		}
	}

	// Player
	if p.Invincible <= 0 || int(g.GameTime*10)%2 == 0 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-20, -20)
		op.GeoM.Rotate(p.Angle + math.Pi/2)
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(g.PlayerImg, op)
	}

	// HUD - Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 50, color.RGBA{20, 25, 45, 220}, false)
	vector.StrokeLine(screen, 0, 50, ScreenW, 50, 2, color.RGBA{80, 160, 240, 255}, false)

	// HP bar
	hpBarW := 200
	hpRatio := float64(p.HP) / float64(p.MaxHP)
	vector.DrawFilledRect(screen, 20, 15, float32(hpBarW), 20, color.RGBA{60, 0, 0, 200}, false)
	vector.DrawFilledRect(screen, 20, 15, float32(hpBarW)*float32(hpRatio), 20, color.RGBA{50, 200, 80, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", p.HP, p.MaxHP), 25, 17)

	// Score & Wave
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 250, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 450, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("УБИТО: %d", g.EnemiesKilled), 600, 17)

	// Crosshair at mouse
	crossImg := createCircleImage(16, color.RGBA{255, 255, 255, 150})
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(g.MouseX-8, g.MouseY-8)
	screen.DrawImage(crossImg, op2)
	vector.StrokeLine(screen, float32(g.MouseX)-10, float32(g.MouseY), float32(g.MouseX)+10, float32(g.MouseY), 1, color.RGBA{255, 255, 255, 150}, false)
	vector.StrokeLine(screen, float32(g.MouseX), float32(g.MouseY)-10, float32(g.MouseX), float32(g.MouseY)+10, 1, color.RGBA{255, 255, 255, 150}, false)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Overlay
	vector.DrawFilledRect(screen, 0, ScreenH/2-150, ScreenW, 300, color.RGBA{10, 10, 20, 220}, false)

	ebitenutil.DebugPrintAt(screen, "💀 GAME OVER 💀", ScreenW/2-130, ScreenH/2-120)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %d | Волна: %d | Убито: %d",
		g.Score, g.Wave, g.EnemiesKilled), ScreenW/2-180, ScreenH/2-60)

	g.drawButton(screen, "🔄  ЗАНОВО", ScreenW/2-100, ScreenH/2, 200, 60)
	g.drawButton(screen, "←  МЕНЮ", ScreenW/2-100, ScreenH/2+80, 200, 60)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	hover := g.MouseX >= float64(x) && g.MouseX <= float64(x+w) &&
		g.MouseY >= float64(y) && g.MouseY <= float64(y+h)

	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{50, 80, 120, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{30, 45, 70, 255}, false)
	}

	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 180, 255, 255})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)

	ebitenutil.DebugPrintAt(screen, text, x+20, y+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Top-Down Shooter - Go365 Day 102")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
