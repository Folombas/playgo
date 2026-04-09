// Space Survivor — Top-Down Shooter с анимированными спрайтами
// Go365 Challenge — День 104 — 9 апреля 2026
// Продвинутый шутер с волнами врагов, разными типами оружия, бонусами и боссами
// Демонстрация мощи Ebitengine на Go

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
	ScreenW          = 960
	ScreenH          = 640
	MaxEnemies       = 50
	MaxBullets       = 200
	MaxParticles     = 1000
	MaxPowerUps      = 20
	StarCount        = 150
)

// ============================================================================
// ПЕРЕМЕННЫЕ АССЕТОВ
// ============================================================================

var (
	// Images
	playerShipImg    *ebiten.Image
	enemySmallImg    *ebiten.Image
	enemyMediumImg   *ebiten.Image
	enemyLargeImg    *ebiten.Image
	bulletPlayerImg  *ebiten.Image
	bulletEnemyImg   *ebiten.Image
	laserBlueImg     *ebiten.Image
	laserRedImg      *ebiten.Image
	powerupHPImg     *ebiten.Image
	powerupShieldImg *ebiten.Image
	powerupWeaponImg *ebiten.Image

	// Colors
	colBackground  = color.RGBA{5, 5, 15, 255}
	colNeonBlue    = color.RGBA{0, 180, 255, 255}
	colNeonGreen   = color.RGBA{0, 255, 150, 255}
	colNeonPink    = color.RGBA{255, 50, 150, 255}
	colNeonYellow  = color.RGBA{255, 220, 50, 255}
	colNeonPurple  = color.RGBA{180, 80, 255, 255}
	colNeonOrange  = color.RGBA{255, 140, 50, 255}
	colHUD         = color.RGBA{10, 15, 30, 220}
	colWhite       = color.RGBA{240, 245, 255, 255}
	colHPGreen     = color.RGBA{0, 255, 120, 255}
	colHPYellow    = color.RGBA{255, 220, 0, 255}
	colHPRed       = color.RGBA{255, 60, 60, 255}
)

// ============================================================================
// ТИПЫ
// ============================================================================

type WeaponType int

const (
	WeaponPistol WeaponType = iota
	WeaponShotgun
	WeaponRifle
	WeaponLaser
	WeaponPlasma
)

type EnemyType int

const (
	EnemyScout EnemyType = iota
	EnemyFighter
	EnemyBomber
	EnemyElite
	EnemyBoss
)

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateWaveComplete
)

type Vector2 struct {
	X, Y float64
}

// Bullet
type Bullet struct {
	X, Y      float64
	VX, VY    float64
	Damage    float64
	Life      float64
	IsEnemy   bool
	Radius    float64
	Img       *ebiten.Image
}

// Enemy
type Enemy struct {
	X, Y           float64
	VX, VY         float64
	HP             float64
	MaxHP          float64
	Type           EnemyType
	Speed          float64
	Damage         float64
	Radius         float64
	AttackTimer    float64
	AttackCD       float64
	HitTimer       float64
	Angle          float64
	ScoreValue     int
}

// Particle
type Particle struct {
	X, Y, VX, VY  float64
	Life          float64
	MaxLife       float64
	Color         color.RGBA
	Size          float64
	Gravity       float64
	Rotation      float64
	RotSpeed      float64
}

// PowerUp
type PowerUp struct {
	X, Y     float64
	Type     int // 0=HP, 1=Shield, 2=Weapon
	SubType  int // weapon type
	Radius   float64
	Life     float64
	BobT     float64
	Rotation float64
	Img      *ebiten.Image
}

// Player
type Player struct {
	X, Y            float64
	HP              float64
	MaxHP           float64
	Shield          float64
	MaxShield       float64
	Speed           float64
	ShootTimer      float64
	Radius          float64
	Angle           float64
	InvulnTimer     float64
	Weapon          WeaponType
	WeaponTimer     float64
}

// Star для фонового эффекта
type Star struct {
	X, Y            float64
	Speed           float64
	Size            float64
	Brightness      float64
	TwinkleSpeed    float64
	TwinkleOffset   float64
}

// Game
type Game struct {
	State         GameState
	Player        Player
	Enemies       []*Enemy
	Bullets       []*Bullet
	Particles     []*Particle
	PowerUps      []*PowerUp
	Stars         []Star

	Score         int
	Combo         int
	MaxCombo      int
	ComboTimer    float64
	Wave          int
	WaveTimer     float64
	EnemiesLeft   int
	SpawnTimer    float64
	BossActive    bool

	GameTime      float64
	ShakeTimer    float64
	ShakeX        float64
	ShakeY        float64
	SlowMotion    float64

	BestScore     int
}

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

func clamp(v, min, max float64) float64 {
	if v < min { return min }
	if v > max { return max }
	return v
}

func normalize(x, y float64) (float64, float64) {
	l := math.Sqrt(x*x + y*y)
	if l == 0 { return 0, 0 }
	return x / l, y / l
}

func angleDiff(a1, a2 float64) float64 {
	diff := a2 - a1
	for diff > math.Pi {
		diff -= 2 * math.Pi
	}
	for diff < -math.Pi {
		diff += 2 * math.Pi
	}
	return diff
}

// Создание изображения круга
func createCircleImage(size int, c color.RGBA, glow bool) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2
	radius := float64(size/2 - 1)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			d := math.Sqrt(dx*dx + dy*dy)

			if d <= radius {
				t := d / radius
				alpha := uint8(255)
				if d > radius-3 {
					alpha = uint8((radius - d) / 3 * 255)
				}

				r := uint8(lerp(float64(c.R)*0.6, float64(c.R), t))
				g := uint8(lerp(float64(c.G)*0.6, float64(c.G), t))
				b := uint8(lerp(float64(c.B)*0.6, float64(c.B), t))

				img.Set(x, y, color.RGBA{r, g, b, alpha})
			} else if glow && d <= radius+4 {
				glowT := (d - radius) / 4
				alpha := uint8((1 - glowT) * 80)
				img.Set(x, y, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// ============================================================================
// ЗАГРУЗКА АССЕТОВ
// ============================================================================

func (g *Game) loadAssets() {
	g.loadShipSprite()
	g.loadBulletSprites()
	g.loadPowerupSprites()
}

func (g *Game) loadShipSprite() {
	img := image.NewRGBA(image.Rect(0, 0, 48, 48))
	cx, cy := 24, 24

	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			absDx := math.Abs(dx)

			if dy < -15 || dy > 20 {
				continue
			}

			width := 6.0
			if dy < 0 {
				width = 6.0 + dy*0.2
			} else {
				width = 6.0 + dy*0.8
			}

			if dy > 5 {
				wingSpan := 18.0 + (dy-5)*0.5
				wingWidth := 4.0 - (dy-5)*0.1
				if absDx > width && absDx < wingSpan && math.Abs(dy-10) < 8 {
					if absDx < wingSpan && absDx > wingSpan-wingWidth {
						img.Set(x, y, color.RGBA{0, 150, 255, 200})
						continue
					}
				}
			}

			if absDx < width {
				t := absDx / width
				r := uint8(lerp(255, 100, t))
				gr := uint8(lerp(255, 200, t))
				b := uint8(255)
				img.Set(x, y, color.RGBA{r, gr, b, 255})
			}
		}
	}

	for y := 35; y < 48; y++ {
		for x := 18; x < 30; x++ {
			dx := float64(x - 24)
			dy := float64(y - 42)
			d := math.Sqrt(dx*dx + dy*dy)
			if d < 6 {
				t := 1.0 - d/6
				alpha := uint8(t * 200)
				img.Set(x, y, color.RGBA{50, 150, 255, alpha})
			}
		}
	}

	playerShipImg = ebiten.NewImageFromImage(img)

	generateEnemySprite(24, colNeonPink)
	generateEnemySprite(32, colNeonOrange)
	generateEnemySprite(48, colNeonPurple)
}

func generateEnemySprite(size int, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2
	radius := float64(size/2 - 2)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			d := math.Sqrt(dx*dx + dy*dy)

			if d <= radius {
				t := d / radius
				angle := math.Atan2(dy, dx)
				spikeFactor := 1.0 + 0.2*math.Sin(angle*8)

				if d <= radius*spikeFactor {
					r := uint8(lerp(float64(c.R)*0.5, float64(c.R), t))
					gr := uint8(lerp(float64(c.G)*0.5, float64(c.G), t))
					b := uint8(lerp(float64(c.B)*0.5, float64(c.B), t))
					alpha := uint8(255)
					if d > radius-2 {
						alpha = uint8((radius-d)/2 * 255)
					}
					img.Set(x, y, color.RGBA{r, gr, b, alpha})
				}
			}
		}
	}

	result := ebiten.NewImageFromImage(img)

	switch size {
	case 24:
		enemySmallImg = result
	case 32:
		enemyMediumImg = result
	case 48:
		enemyLargeImg = result
	}
}

func (g *Game) loadBulletSprites() {
	bulletPlayerImg = createCircleImage(10, colNeonBlue, true)
	bulletEnemyImg = createCircleImage(8, colNeonPink, true)
	laserBlueImg = createCircleImage(6, colNeonGreen, true)
	laserRedImg = createCircleImage(6, colNeonOrange, true)
}

func (g *Game) loadPowerupSprites() {
	powerupHPImg = createCircleImage(20, colHPGreen, true)
	powerupShieldImg = createCircleImage(20, colNeonBlue, true)
	powerupWeaponImg = createCircleImage(20, colNeonYellow, true)
}

// ============================================================================
// ИНИЦИАЛИЗАЦИЯ
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		State:      StateMenu,
		Enemies:    make([]*Enemy, 0, MaxEnemies),
		Bullets:    make([]*Bullet, 0, MaxBullets),
		Particles:  make([]*Particle, 0, MaxParticles),
		PowerUps:   make([]*PowerUp, 0, MaxPowerUps),
		Wave:       0,
		Score:      0,
		Combo:      0,
		MaxCombo:   0,
		BestScore:  0,
	}

	// Инициализация звёзд
	g.Stars = make([]Star, StarCount)
	for i := range g.Stars {
		g.Stars[i] = Star{
			X:             rand.Float64() * ScreenW,
			Y:             rand.Float64() * ScreenH,
			Speed:         0.2 + rand.Float64()*0.8,
			Size:          0.5 + rand.Float64()*2.5,
			Brightness:    0.5 + rand.Float64()*0.5,
			TwinkleSpeed:  1.0 + rand.Float64()*3.0,
			TwinkleOffset: rand.Float64() * math.Pi * 2,
		}
	}

	return g
}

func (g *Game) resetGame() {
	g.Player = Player{
		X:         ScreenW / 2,
		Y:         float64(ScreenH) - 100,
		HP:        100,
		MaxHP:     100,
		Shield:    50,
		MaxShield: 50,
		Speed:     5.0,
		ShootTimer: 0,
		Radius:    16,
		Angle:     -math.Pi / 2,
		Weapon:    WeaponPistol,
	}

	g.Enemies = g.Enemies[:0]
	g.Bullets = g.Bullets[:0]
	g.Particles = g.Particles[:0]
	g.PowerUps = g.PowerUps[:0]

	g.Score = 0
	g.Combo = 0
	g.MaxCombo = 0
	g.ComboTimer = 0
	g.Wave = 0
	g.WaveTimer = 2.5
	g.EnemiesLeft = 0
	g.SpawnTimer = 0
	g.ShakeTimer = 0
	g.ShakeX = 0
	g.ShakeY = 0
	g.BossActive = false
	g.SlowMotion = 1.0

	g.startNextWave()
}

func (g *Game) startNextWave() {
	g.Wave++

	if g.Wave%5 == 0 {
		g.BossActive = true
		g.EnemiesLeft = 1
	} else {
		g.EnemiesLeft = 8 + g.Wave*4
	}

	g.SpawnTimer = 0
	g.State = StatePlaying
}

// ============================================================================
// СИСТЕМА ЧАСТИЦ
// ============================================================================

func (g *Game) spawnExplosion(x, y float64, c color.RGBA, count int, speed float64) {
	for i := 0; i < count && len(g.Particles) < MaxParticles; i++ {
		angle := float64(i)*6.2832/float64(count) + rand.Float64()*0.5
		s := speed * (0.5 + rand.Float64())

		g.Particles = append(g.Particles, &Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * s,
			VY: math.Sin(angle) * s,
			Life: 0.4 + rand.Float64()*0.4,
			MaxLife: 0.8,
			Color: c,
			Size: 2 + rand.Float64()*4,
			Gravity: 0.02,
			Rotation: rand.Float64() * math.Pi * 2,
			RotSpeed: (rand.Float64() - 0.5) * 0.2,
		})
	}
}

func (g *Game) spawnTrail(x, y, vx, vy float64, c color.RGBA) {
	if len(g.Particles) >= MaxParticles {
		return
	}

	g.Particles = append(g.Particles, &Particle{
		X: x, Y: y,
		VX: vx * 0.05 + (rand.Float64()-0.5)*0.5,
		VY: vy * 0.05 + (rand.Float64()-0.5)*0.5,
		Life: 0.15 + rand.Float64()*0.15,
		MaxLife: 0.3,
		Color: c,
		Size: 1 + rand.Float64()*2,
		Gravity: 0,
	})
}

func (g *Game) spawnThrust(x, y, angle float64) {
	if len(g.Particles) >= MaxParticles {
		return
	}

	thrustAngle := angle + math.Pi + (rand.Float64()-0.5)*0.5
	speed := 2 + rand.Float64()*2

	g.Particles = append(g.Particles, &Particle{
		X: x, Y: y,
		VX: math.Cos(thrustAngle) * speed,
		VY: math.Sin(thrustAngle) * speed,
		Life: 0.2 + rand.Float64()*0.2,
		MaxLife: 0.4,
		Color: colNeonBlue,
		Size: 2 + rand.Float64()*3,
		Gravity: 0,
	})
}

// ============================================================================
// ОБНОВЛЕНИЕ
// ============================================================================

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.GameTime += dt

	// Slow motion effect
	if g.SlowMotion > 0 {
		dt *= g.SlowMotion
		g.SlowMotion = lerp(g.SlowMotion, 1.0, 0.05)
	}

	// Screen shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= dt * 60
		intensity := g.ShakeTimer * 6
		g.ShakeX = (rand.Float64() - 0.5) * intensity
		g.ShakeY = (rand.Float64() - 0.5) * intensity
	} else {
		g.ShakeX = 0
		g.ShakeY = 0
	}

	// Update particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += p.Gravity
		p.Rotation += p.RotSpeed
		p.Life -= dt

		if p.Life <= 0 {
			g.Particles[i] = g.Particles[len(g.Particles)-1]
			g.Particles = g.Particles[:len(g.Particles)-1]
		}
	}

	// Update stars
	for i := range g.Stars {
		g.Stars[i].Y += g.Stars[i].Speed
		if g.Stars[i].Y > ScreenH {
			g.Stars[i].Y = 0
			g.Stars[i].X = rand.Float64() * ScreenW
		}
	}

	mx, my := ebiten.CursorPosition()
	fmx, fmy := float64(mx), float64(my)

	switch g.State {
	case StateMenu:
		g.updateMenu(fmx, fmy)
	case StatePlaying:
		g.updatePlaying(dt, fmx, fmy)
	case StatePaused:
		g.updatePaused(fmx, fmy)
	case StateWaveComplete:
		g.WaveTimer -= dt
		if g.WaveTimer <= 0 {
			g.startNextWave()
		}
	case StateGameOver:
		g.updateGameOver(fmx, fmy)
	}

	return nil
}

func (g *Game) updateMenu(fmx, fmy float64) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if fmx >= ScreenW/2-90 && fmx <= ScreenW/2+90 &&
			fmy >= 380 && fmy <= 430 {
			g.resetGame()
		}
	}
}

func (g *Game) updatePlaying(dt, fmx, fmy float64) {
	// Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.State = StatePaused
		return
	}

	// === PLAYER MOVEMENT ===
	var moveX, moveY float64
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		moveY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		moveY += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		moveX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		moveX += 1
	}

	if moveX != 0 || moveY != 0 {
		moveX, moveY = normalize(moveX, moveY)
		g.Player.X += moveX * g.Player.Speed
		g.Player.Y += moveY * g.Player.Speed

		g.Player.X = clamp(g.Player.X, g.Player.Radius, float64(ScreenW)-g.Player.Radius)
		g.Player.Y = clamp(g.Player.Y, g.Player.Radius, float64(ScreenH)-g.Player.Radius)

		if rand.Float64() < 0.6 {
			g.spawnThrust(g.Player.X, g.Player.Y+20, g.Player.Angle)
		}
	}

	// === PLAYER AIM ===
	targetAngle := math.Atan2(fmy-g.Player.Y, fmx-g.Player.X)
	g.Player.Angle += angleDiff(g.Player.Angle, targetAngle) * 0.2

	// === SHOOTING ===
	g.Player.ShootTimer -= dt
	weaponCD := g.getWeaponCooldown()

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.Player.ShootTimer <= 0 {
		g.Player.ShootTimer = weaponCD
		g.shoot(fmx, fmy)
	}

	// === INVULNERABILITY ===
	if g.Player.InvulnTimer > 0 {
		g.Player.InvulnTimer -= dt
	}

	// === SHIELD REGEN ===
	if g.Player.Shield < g.Player.MaxShield {
		g.Player.Shield += dt * 5
		if g.Player.Shield > g.Player.MaxShield {
			g.Player.Shield = g.Player.MaxShield
		}
	}

	// === COMBO TIMER ===
	if g.ComboTimer > 0 {
		g.ComboTimer -= dt
		if g.ComboTimer <= 0 {
			g.Combo = 0
		}
	}

	// === WEAPON TIMER ===
	if g.Player.WeaponTimer > 0 {
		g.Player.WeaponTimer -= dt
		if g.Player.WeaponTimer <= 0 {
			g.Player.Weapon = WeaponPistol
		}
	}

	// === SPAWN ENEMIES ===
	if g.EnemiesLeft > 0 && !g.BossActive {
		g.SpawnTimer -= dt
		spawnRate := 0.6 - float64(g.Wave)*0.03
		if spawnRate < 0.15 {
			spawnRate = 0.15
		}

		if g.SpawnTimer <= 0 {
			g.spawnEnemy()
			g.SpawnTimer = spawnRate
		}
	} else if g.BossActive && len(g.Enemies) == 0 {
		g.spawnBoss()
		g.BossActive = false
	}

	// === UPDATE BULLETS ===
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life -= dt

		if rand.Float64() < 0.4 && !b.IsEnemy {
			g.spawnTrail(b.X, b.Y, b.VX, b.VY, colNeonBlue)
		}

		if b.Life <= 0 || b.X < -50 || b.X > float64(ScreenW)+50 ||
			b.Y < -50 || b.Y > float64(ScreenH)+50 {
			g.Bullets[i] = g.Bullets[len(g.Bullets)-1]
			g.Bullets = g.Bullets[:len(g.Bullets)-1]
			continue
		}

		// === COLLISIONS ===
		if b.IsEnemy {
			if g.Player.InvulnTimer <= 0 &&
				dist(b.X, b.Y, g.Player.X, g.Player.Y) < b.Radius+g.Player.Radius {
				
				if g.Player.Shield > 0 {
					g.Player.Shield -= b.Damage
					if g.Player.Shield < 0 {
						g.Player.HP += g.Player.Shield
						g.Player.Shield = 0
					}
				} else {
					g.Player.HP -= b.Damage
				}

				g.Player.InvulnTimer = 0.3
				g.ShakeTimer = 0.4
				g.spawnExplosion(g.Player.X, g.Player.Y, colNeonBlue, 12, 3)

				g.Bullets[i] = g.Bullets[len(g.Bullets)-1]
				g.Bullets = g.Bullets[:len(g.Bullets)-1]

				if g.Player.HP <= 0 {
					g.State = StateGameOver
					if g.Score > g.BestScore {
						g.BestScore = g.Score
					}
					g.spawnExplosion(g.Player.X, g.Player.Y, colNeonPink, 30, 5)
				}
			}
		} else {
			hit := false
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				e := g.Enemies[j]
				if dist(b.X, b.Y, e.X, e.Y) < b.Radius+e.Radius {
					e.HP -= b.Damage
					e.HitTimer = 0.15

					g.spawnExplosion(b.X, b.Y, colNeonYellow, 6, 2)

					if e.HP <= 0 {
						g.Combo++
						if g.Combo > g.MaxCombo {
							g.MaxCombo = g.Combo
						}
						g.ComboTimer = 2.5

						points := e.ScoreValue * (1 + g.Combo/5)
						g.Score += points

						var eColor color.RGBA
						switch e.Type {
						case EnemyScout:
							eColor = colNeonPink
						case EnemyFighter:
							eColor = colNeonOrange
						case EnemyBomber:
							eColor = colNeonPurple
						case EnemyElite:
							eColor = colNeonYellow
						case EnemyBoss:
							eColor = color.RGBA{255, 255, 255, 255}
							g.ShakeTimer = 1.0
							g.SlowMotion = 0.3
						}
						g.spawnExplosion(e.X, e.Y, eColor, 25, 5)
						if e.Type != EnemyBoss {
							g.ShakeTimer = 0.25
						}

						g.spawnPowerUp(e.X, e.Y)

						g.Enemies[j] = g.Enemies[len(g.Enemies)-1]
						g.Enemies = g.Enemies[:len(g.Enemies)-1]
					}

					hit = true
					break
				}
			}

			if hit {
				g.Bullets[i] = g.Bullets[len(g.Bullets)-1]
				g.Bullets = g.Bullets[:len(g.Bullets)-1]
			}
		}
	}

	// === UPDATE ENEMIES ===
	for _, e := range g.Enemies {
		dx := g.Player.X - e.X
		dy := g.Player.Y - e.Y
		d := math.Sqrt(dx*dx + dy*dy)

		if d > 0 {
			switch e.Type {
			case EnemyScout:
				e.VX = (dx / d) * e.Speed
				e.VY = (dy / d) * e.Speed
			case EnemyFighter:
				angle := math.Atan2(dy, dx)
				angle += 0.02
				e.VX = (math.Cos(angle) * e.Speed * 0.7) + (dx/d)*e.Speed*0.3
				e.VY = (math.Sin(angle) * e.Speed * 0.7) + (dy/d)*e.Speed*0.3
			case EnemyBomber:
				e.VX = (dx / d) * e.Speed
				e.VY = (dy / d) * e.Speed
			case EnemyElite:
				angle := math.Atan2(dy, dx)
				strafeAngle := angle + math.Pi/2*math.Sin(g.GameTime*2)
				e.VX = math.Cos(strafeAngle) * e.Speed
				e.VY = math.Sin(strafeAngle) * e.Speed
			}
		}

		e.X += e.VX
		e.Y += e.VY

		e.X = clamp(e.X, e.Radius, float64(ScreenW)-e.Radius)
		e.Y = clamp(e.Y, e.Radius, float64(ScreenH)-e.Radius)

		if e.HitTimer > 0 {
			e.HitTimer -= dt
		}

		e.AttackTimer -= dt
		if e.AttackTimer <= 0 && d < 500 {
			e.AttackTimer = e.AttackCD
			g.enemyShoot(e)
		}

		if dist(e.X, e.Y, g.Player.X, g.Player.Y) < e.Radius+g.Player.Radius {
			if g.Player.InvulnTimer <= 0 {
				if g.Player.Shield > 0 {
					g.Player.Shield -= e.Damage
					if g.Player.Shield < 0 {
						g.Player.HP += g.Player.Shield
						g.Player.Shield = 0
					}
				} else {
					g.Player.HP -= e.Damage
				}
				g.Player.InvulnTimer = 0.6
				g.ShakeTimer = 0.5
				g.spawnExplosion(g.Player.X, g.Player.Y, colNeonBlue, 18, 4)

				if g.Player.HP <= 0 {
					g.State = StateGameOver
					if g.Score > g.BestScore {
						g.BestScore = g.Score
					}
					g.spawnExplosion(g.Player.X, g.Player.Y, colNeonPink, 30, 5)
				}
			}
		}
	}

	// === UPDATE POWERUPS ===
	for i := len(g.PowerUps) - 1; i >= 0; i-- {
		p := g.PowerUps[i]
		p.Life -= dt
		p.BobT += dt * 4
		p.Rotation += dt * 2

		if dist(p.X, p.Y, g.Player.X, g.Player.Y) < p.Radius+g.Player.Radius+12 {
			switch p.Type {
			case 0:
				g.Player.HP = math.Min(g.Player.MaxHP, g.Player.HP+35)
			case 1:
				g.Player.Shield = math.Min(g.Player.MaxShield, g.Player.Shield+25)
			case 2:
				g.Player.Weapon = WeaponType(p.SubType)
				g.Player.WeaponTimer = 15.0
			}

			g.spawnExplosion(p.X, p.Y, colNeonGreen, 15, 3)
			g.PowerUps[i] = g.PowerUps[len(g.PowerUps)-1]
			g.PowerUps = g.PowerUps[:len(g.PowerUps)-1]
			continue
		}

		if p.Life <= 0 {
			g.PowerUps[i] = g.PowerUps[len(g.PowerUps)-1]
			g.PowerUps = g.PowerUps[:len(g.PowerUps)-1]
		}
	}

	// === CHECK WAVE COMPLETE ===
	if len(g.Enemies) == 0 && g.EnemiesLeft == 0 && !g.BossActive {
		g.State = StateWaveComplete
		g.WaveTimer = 2.5
	}
}

func (g *Game) getWeaponCooldown() float64 {
	switch g.Player.Weapon {
	case WeaponPistol:
		return 0.18
	case WeaponShotgun:
		return 0.5
	case WeaponRifle:
		return 0.08
	case WeaponLaser:
		return 0.05
	case WeaponPlasma:
		return 0.35
	default:
		return 0.18
	}
}

func (g *Game) shoot(fmx, fmy float64) {
	dx, dy := normalize(fmx-g.Player.X, fmy-g.Player.Y)

	switch g.Player.Weapon {
	case WeaponPistol:
		g.Bullets = append(g.Bullets, &Bullet{
			X: g.Player.X + dx*20,
			Y: g.Player.Y + dy*20,
			VX: dx * 12,
			VY: dy * 12,
			Damage: 20,
			Life: 1.5,
			Radius: 5,
			Img: bulletPlayerImg,
		})

	case WeaponShotgun:
		for i := -2; i <= 2; i++ {
			angle := math.Atan2(dy, dx) + float64(i)*0.12
			vx := math.Cos(angle) * 11
			vy := math.Sin(angle) * 11
			g.Bullets = append(g.Bullets, &Bullet{
				X: g.Player.X + dx*20,
				Y: g.Player.Y + dy*20,
				VX: vx,
				VY: vy,
				Damage: 12,
				Life: 0.8,
				Radius: 4,
				Img: bulletPlayerImg,
			})
		}
		g.ShakeTimer = 0.15

	case WeaponRifle:
		for i := 0; i < 3; i++ {
			spread := (rand.Float64() - 0.5) * 0.08
			angle := math.Atan2(dy, dx) + spread
			vx := math.Cos(angle) * 14
			vy := math.Sin(angle) * 14
			g.Bullets = append(g.Bullets, &Bullet{
				X: g.Player.X + dx*20,
				Y: g.Player.Y + dy*20,
				VX: vx,
				VY: vy,
				Damage: 10,
				Life: 1.2,
				Radius: 3,
				Img: laserBlueImg,
			})
		}

	case WeaponLaser:
		g.Bullets = append(g.Bullets, &Bullet{
			X: g.Player.X + dx*20,
			Y: g.Player.Y + dy*20,
			VX: dx * 18,
			VY: dy * 18,
			Damage: 6,
			Life: 1.0,
			Radius: 3,
			Img: laserRedImg,
		})

	case WeaponPlasma:
		g.Bullets = append(g.Bullets, &Bullet{
			X: g.Player.X + dx*20,
			Y: g.Player.Y + dy*20,
			VX: dx * 8,
			VY: dy * 8,
			Damage: 30,
			Life: 2.0,
			Radius: 8,
			Img: createCircleImage(16, colNeonYellow, true),
		})
	}
}

func (g *Game) spawnEnemy() {
	if len(g.Enemies) >= MaxEnemies || g.EnemiesLeft <= 0 {
		return
	}

	var x, y float64
	side := rand.Intn(4)
	switch side {
	case 0: x = float64(rand.Intn(ScreenW)); y = -40
	case 1: x = float64(ScreenW) + 40; y = float64(rand.Intn(ScreenH))
	case 2: x = float64(rand.Intn(ScreenW)); y = float64(ScreenH) + 40
	case 3: x = -40; y = float64(rand.Intn(ScreenH))
	}

	var enemyType EnemyType
	r := rand.Float64()
	if g.Wave >= 8 && r < 0.1 {
		enemyType = EnemyElite
	} else if g.Wave >= 5 && r < 0.25 {
		enemyType = EnemyBomber
	} else if g.Wave >= 3 && r < 0.5 {
		enemyType = EnemyFighter
	} else {
		enemyType = EnemyScout
	}

	var enemy Enemy
	switch enemyType {
	case EnemyScout:
		enemy = Enemy{
			X: x, Y: y,
			HP: 30 + float64(g.Wave)*5,
			MaxHP: 30 + float64(g.Wave)*5,
			Type: EnemyScout,
			Speed: 2.5 + float64(g.Wave)*0.15,
			Damage: 8,
			Radius: 12,
			AttackCD: 1.5,
			ScoreValue: 100,
		}
	case EnemyFighter:
		enemy = Enemy{
			X: x, Y: y,
			HP: 50 + float64(g.Wave)*8,
			MaxHP: 50 + float64(g.Wave)*8,
			Type: EnemyFighter,
			Speed: 1.8 + float64(g.Wave)*0.1,
			Damage: 12,
			Radius: 16,
			AttackCD: 1.0,
			ScoreValue: 150,
		}
	case EnemyBomber:
		enemy = Enemy{
			X: x, Y: y,
			HP: 120 + float64(g.Wave)*15,
			MaxHP: 120 + float64(g.Wave)*15,
			Type: EnemyBomber,
			Speed: 1.0 + float64(g.Wave)*0.05,
			Damage: 20,
			Radius: 24,
			AttackCD: 2.0,
			ScoreValue: 250,
		}
	case EnemyElite:
		enemy = Enemy{
			X: x, Y: y,
			HP: 80 + float64(g.Wave)*12,
			MaxHP: 80 + float64(g.Wave)*12,
			Type: EnemyElite,
			Speed: 2.2 + float64(g.Wave)*0.12,
			Damage: 15,
			Radius: 18,
			AttackCD: 0.6,
			ScoreValue: 300,
		}
	}

	g.Enemies = append(g.Enemies, &enemy)
	g.EnemiesLeft--
}

func (g *Game) spawnBoss() {
	boss := Enemy{
		X: ScreenW / 2,
		Y: -60,
		HP: 500 + float64(g.Wave)*100,
		MaxHP: 500 + float64(g.Wave)*100,
		Type: EnemyBoss,
		Speed: 1.0,
		Damage: 30,
		Radius: 40,
		AttackCD: 0.4,
		ScoreValue: 2000,
	}
	g.Enemies = append(g.Enemies, &boss)
}

func (g *Game) enemyShoot(e *Enemy) {
	dx := g.Player.X - e.X
	dy := g.Player.Y - e.Y
	d := math.Sqrt(dx*dx + dy*dy)

	if d == 0 {
		return
	}

	dx /= d
	dy /= d

	switch e.Type {
	case EnemyScout:
		g.Bullets = append(g.Bullets, &Bullet{
			X: e.X + dx*e.Radius,
			Y: e.Y + dy*e.Radius,
			VX: dx * 6,
			VY: dy * 6,
			Damage: e.Damage,
			Life: 2.0,
			Radius: 4,
			Img: bulletEnemyImg,
			IsEnemy: true,
		})

	case EnemyFighter:
		for i := -1; i <= 1; i++ {
			angle := math.Atan2(dy, dx) + float64(i)*0.2
			vx := math.Cos(angle) * 5
			vy := math.Sin(angle) * 5
			g.Bullets = append(g.Bullets, &Bullet{
				X: e.X + dx*e.Radius,
				Y: e.Y + dy*e.Radius,
				VX: vx,
				VY: vy,
				Damage: e.Damage * 0.7,
				Life: 2.5,
				Radius: 4,
				Img: bulletEnemyImg,
				IsEnemy: true,
			})
		}

	case EnemyBomber:
		for i := 0; i < 8; i++ {
			angle := float64(i) * math.Pi * 2 / 8
			vx := math.Cos(angle) * 4
			vy := math.Sin(angle) * 4
			g.Bullets = append(g.Bullets, &Bullet{
				X: e.X,
				Y: e.Y,
				VX: vx,
				VY: vy,
				Damage: e.Damage * 0.5,
				Life: 3.0,
				Radius: 5,
				Img: bulletEnemyImg,
				IsEnemy: true,
			})
		}

	case EnemyElite:
		for i := -2; i <= 2; i++ {
			angle := math.Atan2(dy, dx) + float64(i)*0.15
			vx := math.Cos(angle) * 7
			vy := math.Sin(angle) * 7
			g.Bullets = append(g.Bullets, &Bullet{
				X: e.X + dx*e.Radius,
				Y: e.Y + dy*e.Radius,
				VX: vx,
				VY: vy,
				Damage: e.Damage * 0.6,
				Life: 1.8,
				Radius: 3,
				Img: laserRedImg,
				IsEnemy: true,
			})
		}

	case EnemyBoss:
		for i := 0; i < 12; i++ {
			angle := g.GameTime*2 + float64(i)*math.Pi*2/12
			vx := math.Cos(angle) * 5
			vy := math.Sin(angle) * 5
			g.Bullets = append(g.Bullets, &Bullet{
				X: e.X,
				Y: e.Y,
				VX: vx,
				VY: vy,
				Damage: e.Damage * 0.4,
				Life: 4.0,
				Radius: 5,
				Img: bulletEnemyImg,
				IsEnemy: true,
			})
		}
	}
}

func (g *Game) spawnPowerUp(x, y float64) {
	if len(g.PowerUps) >= MaxPowerUps || rand.Float64() > 0.3 {
		return
	}

	pType := rand.Intn(3)
	var img *ebiten.Image
	subType := 0

	switch pType {
	case 0:
		img = powerupHPImg
	case 1:
		img = powerupShieldImg
	case 2:
		img = powerupWeaponImg
		subType = rand.Intn(5)
	}

	g.PowerUps = append(g.PowerUps, &PowerUp{
		X: x, Y: y,
		Type: pType,
		SubType: subType,
		Radius: 10,
		Life: 12.0,
		Img: img,
	})
}

func (g *Game) updatePaused(fmx, fmy float64) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.State = StatePlaying
		return
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 340 && fmy <= 390 {
			g.State = StatePlaying
		}
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 400 && fmy <= 450 {
			g.State = StateMenu
		}
	}
}

func (g *Game) updateGameOver(fmx, fmy float64) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 420 && fmy <= 470 {
			g.State = StateMenu
		}
	}
}

// ============================================================================
// ОТРИСОВКА
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(colBackground)

	g.drawStars(screen)

	switch g.State {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StateWaveComplete:
		g.drawGame(screen)
	case StatePaused:
		g.drawGame(screen)
		g.drawPause(screen)
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	}
}

func (g *Game) drawStars(screen *ebiten.Image) {
	for _, star := range g.Stars {
		twinkle := 0.5 + 0.5*math.Sin(g.GameTime*star.TwinkleSpeed+star.TwinkleOffset)
		alpha := uint8(star.Brightness * twinkle * 255)
		
		c := color.RGBA{200, 220, 255, alpha}
		sz := int(star.Size)
		if sz < 1 { sz = 1 }
		
		img := ebiten.NewImage(sz, sz)
		img.Fill(c)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(star.X, star.Y)
		screen.DrawImage(img, op)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Screen shake
	if g.ShakeTimer > 0 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.ShakeX, g.ShakeY)
		screen.DrawImage(screen, op)
	}

	// === POWERUPS ===
	for _, p := range g.PowerUps {
		bob := math.Sin(p.BobT) * 4
		alpha := 1.0
		if p.Life < 3.0 {
			alpha = p.Life / 3.0
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(p.Img.Bounds().Dx())/2, p.Y-float64(p.Img.Bounds().Dy())/2+bob)
		op.ColorM.Scale(1, 1, 1, alpha)
		screen.DrawImage(p.Img, op)
	}

	// === PARTICLES ===
	for _, p := range g.Particles {
		alpha := p.Life / p.MaxLife
		sz := int(p.Size * alpha)
		if sz < 1 { continue }

		img := ebiten.NewImage(sz, sz)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, uint8(alpha * 255)}
		img.Fill(c)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(sz)/2, p.Y-float64(sz)/2)
		screen.DrawImage(img, op)
	}

	// === BULLETS ===
	for _, b := range g.Bullets {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.X-float64(b.Img.Bounds().Dx())/2, b.Y-float64(b.Img.Bounds().Dy())/2)
		screen.DrawImage(b.Img, op)
	}

	// === ENEMIES ===
	for _, e := range g.Enemies {
		var img *ebiten.Image
		switch e.Type {
		case EnemyScout:
			img = enemySmallImg
		case EnemyFighter:
			img = enemyMediumImg
		case EnemyBomber:
			img = enemyLargeImg
		case EnemyElite:
			img = enemyMediumImg
		case EnemyBoss:
			img = enemyLargeImg
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-float64(img.Bounds().Dx())/2, e.Y-float64(img.Bounds().Dy())/2)

		if e.HitTimer > 0 {
			op.ColorM.Scale(2, 2, 2, 1)
		}

		screen.DrawImage(img, op)

		if e.Type == EnemyBoss {
			glow := createCircleImage(100, color.RGBA{255, 50, 150, 60}, true)
			op2 := &ebiten.DrawImageOptions{}
			op2.GeoM.Translate(e.X-50, e.Y-50)
			screen.DrawImage(glow, op2)
		}

		// HP bar
		if e.HP < e.MaxHP {
			barW := float32(e.Radius * 2.5)
			barH := float32(4)
			hpRatio := e.HP / e.MaxHP

			vector.DrawFilledRect(screen, float32(e.X-float64(e.Radius)*1.25), float32(e.Y-float64(e.Radius)-10),
				barW, barH, color.RGBA{40, 40, 40, 200}, false)

			var hpColor color.RGBA
			if hpRatio > 0.6 {
				hpColor = colHPGreen
			} else if hpRatio > 0.3 {
				hpColor = colHPYellow
			} else {
				hpColor = colHPRed
			}

			vector.DrawFilledRect(screen, float32(e.X-float64(e.Radius)*1.25), float32(e.Y-float64(e.Radius)-10),
				barW*float32(hpRatio), barH, hpColor, false)
		}
	}

	// === PLAYER ===
	if g.Player.InvulnTimer <= 0 || int(g.GameTime*10)%2 == 0 {
		glow := createCircleImage(50, color.RGBA{0, 180, 255, 60}, true)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.Player.X-25, g.Player.Y-25)
		screen.DrawImage(glow, op)

		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(g.Player.X-24, g.Player.Y-24)
		screen.DrawImage(playerShipImg, op2)
	}

	// === HUD ===
	g.drawHUD(screen)

	// === WAVE COMPLETE ===
	if g.State == StateWaveComplete {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА %d ПРОЙДЕНА!", g.Wave),
			ScreenW/2-150, ScreenH/2-30)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, float32(ScreenW), 60, colHUD, false)

	// HP bar
	hpBarW := float32(180)
	hpBarH := float32(18)
	hpX := float32(15)
	hpY := float32(22)
	hpRatio := g.Player.HP / g.Player.MaxHP

	vector.DrawFilledRect(screen, hpX, hpY, hpBarW, hpBarH, color.RGBA{40, 40, 40, 200}, false)

	var hpColor color.RGBA
	if hpRatio > 0.6 {
		hpColor = colHPGreen
	} else if hpRatio > 0.3 {
		hpColor = colHPYellow
	} else {
		hpColor = colHPRed
	}

	vector.DrawFilledRect(screen, hpX, hpY, hpBarW*float32(hpRatio), hpBarH, hpColor, false)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %.0f", g.Player.HP), int(hpX+5), int(hpY+3))

	// Shield bar
	shieldX := hpX + hpBarW + 15
	shieldBarW := float32(120)
	shieldRatio := g.Player.Shield / g.Player.MaxShield

	vector.DrawFilledRect(screen, shieldX, hpY, shieldBarW, hpBarH, color.RGBA{40, 40, 40, 200}, false)
	vector.DrawFilledRect(screen, shieldX, hpY, shieldBarW*float32(shieldRatio), hpBarH, colNeonBlue, false)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("SH: %.0f", g.Player.Shield), int(shieldX+5), int(hpY+3))

	// Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 400, 15)

	// Wave
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 400, 38)

	// Combo
	if g.Combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d", g.Combo), 600, 15)
	}

	// Weapon
	weaponNames := []string{"ПИСТОЛЕТ", "ДРОБОВИК", "АВТОМАТ", "ЛАЗЕР", "ПЛАЗМА"}
	ebitenutil.DebugPrintAt(screen, weaponNames[g.Player.Weapon], 600, 38)

	// Boss HP
	for _, e := range g.Enemies {
		if e.Type == EnemyBoss {
			bossBarW := float32(ScreenW - 100)
			bossBarH := float32(20)
			bossX := float32(50)
			bossY := float32(ScreenH - 40)
			bossRatio := e.HP / e.MaxHP

			vector.DrawFilledRect(screen, bossX, bossY, bossBarW, bossBarH, color.RGBA{40, 40, 40, 200}, false)
			vector.DrawFilledRect(screen, bossX, bossY, bossBarW*float32(bossRatio), bossBarH, colNeonPink, false)

			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("БОСС: %.0f / %.0f", e.HP, e.MaxHP), int(bossX+5), int(bossY+3))
		}
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "SPACE SURVIVOR", ScreenW/2-180, 150)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge — День 104", ScreenW/2-140, 210)

	// Animated neon circles
	for i := 0; i < 10; i++ {
		angle := float64(i) * 6.2832 / 10 + g.GameTime*0.4
		x := ScreenW/2 + int(math.Cos(angle)*150)
		y := 300 + int(math.Sin(angle)*50)

		var c color.RGBA
		switch i % 4 {
		case 0: c = colNeonBlue
		case 1: c = colNeonPink
		case 2: c = colNeonGreen
		case 3: c = colNeonYellow
		}

		img := createCircleImage(24, c, true)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x-12), float64(y-12))
		screen.DrawImage(img, op)
	}

	g.drawButton(screen, "▶  ИГРАТЬ", ScreenW/2-90, 380, 180, 50, colNeonGreen)

	ebitenutil.DebugPrintAt(screen, "WASD / Стрелки — движение", ScreenW/2-110, 460)
	ebitenutil.DebugPrintAt(screen, "Мышь — прицел", ScreenW/2-80, 485)
	ebitenutil.DebugPrintAt(screen, "ЛКМ — стрельба", ScreenW/2-80, 510)
	ebitenutil.DebugPrintAt(screen, "ESC / P — пауза", ScreenW/2-80, 535)

	if g.BestScore > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ЛУЧШИЙ РЕЗУЛЬТАТ: %d", g.BestScore),
			ScreenW/2-110, 580)
	}
}

func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 180}, false)
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", ScreenW/2-60, 260)

	g.drawButton(screen, "ПРОДОЛЖИТЬ", ScreenW/2-90, 340, 180, 50, colNeonBlue)
	g.drawButton(screen, "В МЕНЮ", ScreenW/2-90, 400, 180, 50, colNeonPink)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenW/2-110, 220)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), ScreenW/2-60, 280)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), ScreenW/2-50, 315)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("MAX COMBO: %d", g.MaxCombo), ScreenW/2-70, 350)

	if g.Score >= g.BestScore && g.Score > 0 {
		ebitenutil.DebugPrintAt(screen, "НОВЫЙ РЕКОРД!", ScreenW/2-85, 385)
	}

	g.drawButton(screen, "В МЕНЮ", ScreenW/2-80, 420, 160, 50, colNeonPink)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int, c color.RGBA) {
	mx, my := ebiten.CursorPosition()
	hover := float64(mx) >= float64(x) && float64(mx) <= float64(x+w) &&
		float64(my) >= float64(y) && float64(my) <= float64(y+h)

	var bgColor color.RGBA
	if hover {
		bgColor = color.RGBA{c.R, c.G, c.B, 80}
	} else {
		bgColor = color.RGBA{30, 40, 60, 255}
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bgColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, c, false)

	ebitenutil.DebugPrintAt(screen, text, x+25, y+h/2-10)
}

// ============================================================================
// LAYOUT & MAIN
// ============================================================================

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Space Survivor — Go365")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
