// Neon Survivor — Top-Down Shooter
// Go365 Challenge — Day 103
// Современный top-down shooter с волнами врагов
// 9 апреля 2026

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
	ScreenW         = 800
	ScreenH         = 600
	PlayerMaxHP     = 100
	BulletSpeed     = 10.0
	PlayerSpeed     = 4.0
	ShootCooldown   = 0.12
	MaxEnemies      = 30
	MaxBullets      = 100
	MaxParticles    = 500
)

// ============================================================================
// ЦВЕТА (NEON STYLE)
// ============================================================================

var (
	bgColor       = color.RGBA{10, 10, 20, 255}
	gridColor     = color.RGBA{30, 30, 60, 100}
	playerColor   = color.RGBA{0, 255, 200, 255}
	playerGlow    = color.RGBA{0, 255, 200, 80}
	bulletColor   = color.RGBA{255, 255, 100, 255}
	bulletGlow    = color.RGBA{255, 200, 50, 120}
	
	// Enemy colors by type
	enemyNormal   = color.RGBA{255, 80, 80, 255}
	enemyFast     = color.RGBA{255, 200, 60, 255}
	enemyTank     = color.RGBA{200, 80, 255, 255}
	
	// UI colors
	hudBg         = color.RGBA{10, 10, 25, 200}
	textWhite     = color.RGBA{240, 240, 255, 255}
	textNeon      = color.RGBA{0, 255, 200, 255}
	textGold      = color.RGBA{255, 210, 60, 255}
	textRed       = color.RGBA{255, 80, 80, 255}
	hpGreen       = color.RGBA{0, 255, 100, 255}
	hpYellow      = color.RGBA{255, 200, 0, 255}
	hpRed         = color.RGBA{255, 50, 50, 255}
	
	// Power-up colors
	powerupHP     = color.RGBA{255, 100, 100, 255}
	powerupDamage = color.RGBA{255, 150, 50, 255}
	powerupSpeed  = color.RGBA{100, 200, 255, 255}
)

// ============================================================================
// ТИПЫ
// ============================================================================

type EnemyType int

const (
	EnemyNormal EnemyType = iota
	EnemyFast
	EnemyTank
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

type Bullet struct {
	X, Y       float64
	VX, VY     float64
	Damage     float64
	Life       float64
	IsEnemy    bool
	Radius     float64
}

type Enemy struct {
	X, Y         float64
	VX, VY       float64
	HP           float64
	MaxHP        float64
	Type         EnemyType
	Speed        float64
	Damage       float64
	Radius       float64
	AttackTimer  float64
	AttackCD     float64
	HitTimer     float64
}

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
	Color        color.RGBA
	Size         float64
	Gravity      float64
}

type PowerUp struct {
	X, Y   float64
	Type   int // 0=HP, 1=Damage, 2=Speed
	Radius float64
	Life   float64
	BobT   float64
}

type Player struct {
	X, Y         float64
	HP           float64
	MaxHP        float64
	Speed        float64
	Damage       float64
	ShootTimer   float64
	Radius       float64
	Angle        float64
	InvulnTimer  float64
}

type Game struct {
	State       GameState
	Player      Player
	Enemies     []*Enemy
	Bullets     []*Bullet
	Particles   []*Particle
	PowerUps    []*PowerUp
	
	Score       int
	Combo       int
	MaxCombo    int
	ComboTimer  float64
	Wave        int
	WaveTimer   float64
	EnemiesLeft int
	SpawnTimer  float64
	
	GameTime    float64
	ShakeTimer  float64
	ShakeX      float64
	ShakeY      float64
	
	BestScore   int
	
	// Images cache
	PlayerImg   *ebiten.Image
	BulletImg   *ebiten.Image
	EnemyImgs   map[EnemyType]*ebiten.Image
	PowerupImgs map[int]*ebiten.Image
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

func createCircleImage(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2
	radius := float64(size/2 - 1)
	
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			d := math.Sqrt(dx*dx + dy*dy)
			
			if d <= radius {
				// Glow effect
				t := d / radius
				alpha := uint8(255)
				if d > radius-3 {
					alpha = uint8((radius - d) / 3 * 255)
				}
				
				r := uint8(lerp(float64(c.R)*0.6, float64(c.R), t))
				g := uint8(lerp(float64(c.G)*0.6, float64(c.G), t))
				b := uint8(lerp(float64(c.B)*0.6, float64(c.B), t))
				
				img.Set(x, y, color.RGBA{r, g, b, alpha})
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

func createGlowCircle(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	center := size / 2
	radius := float64(size / 2)
	
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - center)
			dy := float64(y - center)
			d := math.Sqrt(dx*dx + dy*dy)
			
			if d <= radius {
				t := 1.0 - d/radius
				alpha := uint8(t * t * float64(c.A))
				img.Set(x, y, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
	
	return ebiten.NewImageFromImage(img)
}

// ============================================================================
// GAME INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	
	g := &Game{
		State:     StateMenu,
		Enemies:   make([]*Enemy, 0, MaxEnemies),
		Bullets:   make([]*Bullet, 0, MaxBullets),
		Particles: make([]*Particle, 0, MaxParticles),
		PowerUps:  make([]*PowerUp, 0),
		Wave:      0,
		Score:     0,
		Combo:     0,
		MaxCombo:  0,
		BestScore: 0,
	}
	
	// Create player image
	g.PlayerImg = createCircleImage(24, playerColor)
	g.BulletImg = createCircleImage(8, bulletColor)
	
	// Create enemy images
	g.EnemyImgs[EnemyNormal] = createCircleImage(20, enemyNormal)
	g.EnemyImgs[EnemyFast] = createCircleImage(16, enemyFast)
	g.EnemyImgs[EnemyTank] = createCircleImage(28, enemyTank)
	
	// Create powerup images
	g.PowerupImgs[0] = createCircleImage(16, powerupHP)
	g.PowerupImgs[1] = createCircleImage(16, powerupDamage)
	g.PowerupImgs[2] = createCircleImage(16, powerupSpeed)
	
	return g
}

func (g *Game) resetGame() {
	g.Player = Player{
		X:       ScreenW / 2,
		Y:       ScreenH / 2,
		HP:      PlayerMaxHP,
		MaxHP:   PlayerMaxHP,
		Speed:   PlayerSpeed,
		Damage:  20,
		Radius:  12,
		Angle:   0,
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
	g.WaveTimer = 2.0
	g.EnemiesLeft = 0
	g.ShakeTimer = 0
	g.ShakeX = 0
	g.ShakeY = 0
	
	g.startNextWave()
}

func (g *Game) startNextWave() {
	g.Wave++
	g.EnemiesLeft = 5 + g.Wave*3
	g.SpawnTimer = 0
	g.State = StatePlaying
}

// ============================================================================
// SPAWN FUNCTIONS
// ============================================================================

func (g *Game) spawnEnemy() {
	if len(g.Enemies) >= MaxEnemies || g.EnemiesLeft <= 0 {
		return
	}
	
	// Spawn from edges
	var x, y float64
	side := rand.Intn(4)
	switch side {
	case 0: // top
		x = float64(rand.Intn(ScreenW))
		y = -30
	case 1: // right
		x = float64(ScreenW + 30)
		y = float64(rand.Intn(ScreenH))
	case 2: // bottom
		x = float64(rand.Intn(ScreenW))
		y = float64(ScreenH + 30)
	case 3: // left
		x = -30
		y = float64(rand.Intn(ScreenH))
	}
	
	// Determine enemy type based on wave
	var enemyType EnemyType
	r := rand.Float64()
	if g.Wave >= 5 && r < 0.15 {
		enemyType = EnemyTank
	} else if g.Wave >= 3 && r < 0.4 {
		enemyType = EnemyFast
	} else {
		enemyType = EnemyNormal
	}
	
	var enemy Enemy
	switch enemyType {
	case EnemyNormal:
		enemy = Enemy{
			X: x, Y: y,
			HP: 40 + float64(g.Wave)*5,
			MaxHP: 40 + float64(g.Wave)*5,
			Type: EnemyNormal,
			Speed: 1.5 + float64(g.Wave)*0.1,
			Damage: 10,
			Radius: 10,
			AttackCD: 1.0,
		}
	case EnemyFast:
		enemy = Enemy{
			X: x, Y: y,
			HP: 25 + float64(g.Wave)*3,
			MaxHP: 25 + float64(g.Wave)*3,
			Type: EnemyFast,
			Speed: 3.0 + float64(g.Wave)*0.15,
			Damage: 8,
			Radius: 8,
			AttackCD: 0.5,
		}
	case EnemyTank:
		enemy = Enemy{
			X: x, Y: y,
			HP: 100 + float64(g.Wave)*15,
			MaxHP: 100 + float64(g.Wave)*15,
			Type: EnemyTank,
			Speed: 0.8 + float64(g.Wave)*0.05,
			Damage: 25,
			Radius: 14,
			AttackCD: 2.0,
		}
	}
	
	g.Enemies = append(g.Enemies, &enemy)
	g.EnemiesLeft--
}

func (g *Game) spawnPowerUp(x, y float64) {
	if rand.Float64() > 0.25 { // 25% chance
		return
	}
	
	pType := rand.Intn(3)
	g.PowerUps = append(g.PowerUps, &PowerUp{
		X: x, Y: y,
		Type: pType,
		Radius: 8,
		Life: 10.0,
	})
}

// ============================================================================
// PARTICLE SYSTEM
// ============================================================================

func (g *Game) spawnExplosion(x, y float64, c color.RGBA, count int, speed float64) {
	for i := 0; i < count && len(g.Particles) < MaxParticles; i++ {
		angle := float64(i)*6.2832/float64(count) + rand.Float64()*0.5
		s := speed * (0.5 + rand.Float64())
		
		g.Particles = append(g.Particles, &Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * s,
			VY: math.Sin(angle) * s,
			Life: 0.5 + rand.Float64()*0.3,
			MaxLife: 0.8,
			Color: c,
			Size: 2 + rand.Float64()*3,
			Gravity: 0.05,
		})
	}
}

func (g *Game) spawnTrail(x, y, vx, vy float64, c color.RGBA) {
	if len(g.Particles) >= MaxParticles {
		return
	}
	
	g.Particles = append(g.Particles, &Particle{
		X: x, Y: y,
		VX: vx * 0.1,
		VY: vy * 0.1,
		Life: 0.2 + rand.Float64()*0.1,
		MaxLife: 0.3,
		Color: c,
		Size: 1 + rand.Float64()*2,
		Gravity: 0,
	})
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.GameTime += dt
	
	// Update shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= dt
		intensity := g.ShakeTimer * 8
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
		p.Life -= dt
		
		if p.Life <= 0 {
			g.Particles[i] = g.Particles[len(g.Particles)-1]
			g.Particles = g.Particles[:len(g.Particles)-1]
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
	}
	
	return nil
}

func (g *Game) updateMenu(fmx, fmy float64) {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Play button
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 350 && fmy <= 400 {
			g.State = StatePlaying
			g.resetGame()
		}
	}
}

func (g *Game) updatePlaying(dt, fmx, fmy float64) {
	// Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
		inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.State = StatePaused
		return
	}
	
	// Player movement
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
		
		// Clamp to screen
		g.Player.X = clamp(g.Player.X, g.Player.Radius, float64(ScreenW)-g.Player.Radius)
		g.Player.Y = clamp(g.Player.Y, g.Player.Radius, float64(ScreenH)-g.Player.Radius)
	}
	
	// Player aim
	g.Player.Angle = math.Atan2(fmy-g.Player.Y, fmx-g.Player.X)
	
	// Shooting
	g.Player.ShootTimer -= dt
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && g.Player.ShootTimer <= 0 {
		g.Player.ShootTimer = ShootCooldown
		
		dx, dy := normalize(fmx-g.Player.X, fmy-g.Player.Y)
		g.Bullets = append(g.Bullets, &Bullet{
			X: g.Player.X + dx*20,
			Y: g.Player.Y + dy*20,
			VX: dx * BulletSpeed,
			VY: dy * BulletSpeed,
			Damage: g.Player.Damage,
			Life: 1.5,
			Radius: 4,
		})
		
		// Muzzle flash
		g.spawnExplosion(g.Player.X+dx*20, g.Player.Y+dy*20, bulletGlow, 5, 2)
	}
	
	// Invulnerability timer
	if g.Player.InvulnTimer > 0 {
		g.Player.InvulnTimer -= dt
	}
	
	// Combo timer
	if g.ComboTimer > 0 {
		g.ComboTimer -= dt
		if g.ComboTimer <= 0 {
			g.Combo = 0
		}
	}
	
	// Spawn enemies
	if g.EnemiesLeft > 0 {
		g.SpawnTimer -= dt
		if g.SpawnTimer <= 0 {
			g.spawnEnemy()
			g.SpawnTimer = 0.5 - float64(g.Wave)*0.02
			if g.SpawnTimer < 0.1 {
				g.SpawnTimer = 0.1
			}
		}
	}
	
	// Update bullets
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life -= dt
		
		// Trail
		if rand.Float64() < 0.3 {
			g.spawnTrail(b.X, b.Y, b.VX, b.VY, bulletGlow)
		}
		
		// Remove if off screen or expired
		if b.Life <= 0 || b.X < -50 || b.X > float64(ScreenW)+50 ||
			b.Y < -50 || b.Y > float64(ScreenH)+50 {
			g.Bullets[i] = g.Bullets[len(g.Bullets)-1]
			g.Bullets = g.Bullets[:len(g.Bullets)-1]
			continue
		}
		
		// Check collisions
		if b.IsEnemy {
			// Enemy bullet hits player
			if g.Player.InvulnTimer <= 0 &&
				dist(b.X, b.Y, g.Player.X, g.Player.Y) < b.Radius+g.Player.Radius {
				g.Player.HP -= b.Damage
				g.Player.InvulnTimer = 0.3
				g.ShakeTimer = 0.3
				g.spawnExplosion(g.Player.X, g.Player.Y, playerGlow, 10, 3)
				
				g.Bullets[i] = g.Bullets[len(g.Bullets)-1]
				g.Bullets = g.Bullets[:len(g.Bullets)-1]
				
				if g.Player.HP <= 0 {
					g.State = StateGameOver
					if g.Score > g.BestScore {
						g.BestScore = g.Score
					}
				}
			}
		} else {
			// Player bullet hits enemy
			hit := false
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				e := g.Enemies[j]
				if dist(b.X, b.Y, e.X, e.Y) < b.Radius+e.Radius {
					e.HP -= b.Damage
					e.HitTimer = 0.1
					
					// Hit effect
					g.spawnExplosion(b.X, b.Y, enemyNormal, 8, 2)
					
					if e.HP <= 0 {
						// Enemy killed
						g.Combo++
						if g.Combo > g.MaxCombo {
							g.MaxCombo = g.Combo
						}
						g.ComboTimer = 2.0
						
						points := int(100 * (1 + float64(g.Combo)*0.2))
						g.Score += points
						
						// Death explosion
						var eColor color.RGBA
						switch e.Type {
						case EnemyNormal:
							eColor = enemyNormal
						case EnemyFast:
							eColor = enemyFast
						case EnemyTank:
							eColor = enemyTank
						}
						g.spawnExplosion(e.X, e.Y, eColor, 20, 4)
						g.ShakeTimer = 0.2
						
						// Spawn powerup
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
	
	// Update enemies
	for _, e := range g.Enemies {
		// Move towards player
		dx := g.Player.X - e.X
		dy := g.Player.Y - e.Y
		d := math.Sqrt(dx*dx + dy*dy)
		
		if d > 0 {
			e.VX = (dx / d) * e.Speed
			e.VY = (dy / d) * e.Speed
		}
		
		e.X += e.VX
		e.Y += e.VY
		
		// Hit timer
		if e.HitTimer > 0 {
			e.HitTimer -= dt
		}
		
		// Attack player on contact
		if dist(e.X, e.Y, g.Player.X, g.Player.Y) < e.Radius+g.Player.Radius {
			e.AttackTimer -= dt
			if e.AttackTimer <= 0 && g.Player.InvulnTimer <= 0 {
				g.Player.HP -= e.Damage
				g.Player.InvulnTimer = 0.5
				g.ShakeTimer = 0.4
				e.AttackTimer = e.AttackCD
				g.spawnExplosion(g.Player.X, g.Player.Y, playerGlow, 15, 3)
				
				if g.Player.HP <= 0 {
					g.State = StateGameOver
					if g.Score > g.BestScore {
						g.BestScore = g.Score
					}
				}
			}
		}
	}
	
	// Update powerups
	for i := len(g.PowerUps) - 1; i >= 0; i-- {
		p := g.PowerUps[i]
		p.Life -= dt
		p.BobT += dt * 3
		
		// Player collects powerup
		if dist(p.X, p.Y, g.Player.X, g.Player.Y) < p.Radius+g.Player.Radius+10 {
			switch p.Type {
			case 0: // HP
				g.Player.HP = math.Min(g.Player.MaxHP, g.Player.HP+30)
			case 1: // Damage
				g.Player.Damage += 5
			case 2: // Speed
				g.Player.Speed += 0.5
			}
			
			g.spawnExplosion(p.X, p.Y, textNeon, 12, 3)
			g.PowerUps[i] = g.PowerUps[len(g.PowerUps)-1]
			g.PowerUps = g.PowerUps[:len(g.PowerUps)-1]
			continue
		}
		
		// Expire
		if p.Life <= 0 {
			g.PowerUps[i] = g.PowerUps[len(g.PowerUps)-1]
			g.PowerUps = g.PowerUps[:len(g.PowerUps)-1]
		}
	}
	
	// Check wave complete
	if len(g.Enemies) == 0 && g.EnemiesLeft == 0 {
		g.State = StateWaveComplete
		g.WaveTimer = 2.0
	}
}

func (g *Game) updatePaused(fmx, fmy float64) {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) ||
		inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.State = StatePlaying
		return
	}
	
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Resume button
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 320 && fmy <= 370 {
			g.State = StatePlaying
		}
		
		// Menu button
		if fmx >= ScreenW/2-80 && fmx <= ScreenW/2+80 &&
			fmy >= 380 && fmy <= 430 {
			g.State = StateMenu
		}
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(bgColor)
	
	// Grid
	g.drawGrid(screen)
	
	// Apply shake
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(g.ShakeX, g.ShakeY)
	
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

func (g *Game) drawGrid(screen *ebiten.Image) {
	spacing := 50.0
	for x := 0.0; x < float64(ScreenW); x += spacing {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(ScreenH), 1, gridColor, false)
	}
	for y := 0.0; y < float64(ScreenH); y += spacing {
		vector.StrokeLine(screen, 0, float32(y), float32(ScreenW), float32(y), 1, gridColor, false)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Powerups
	for _, p := range g.PowerUps {
		bob := math.Sin(p.BobT) * 3
		alpha := 255
		if p.Life < 2.0 {
			alpha = int(p.Life / 2.0 * 255)
		}
		
		img := g.PowerupImgs[p.Type]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-8, p.Y-8+bob)
		op.ColorM.Scale(1, 1, 1, float64(alpha)/255)
		screen.DrawImage(img, op)
	}
	
	// Particles (behind)
	for _, p := range g.Particles {
		alpha := p.Life / p.MaxLife
		sz := int(p.Size * alpha)
		if sz < 1 { continue }
		
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, uint8(alpha * 255)}
		img := ebiten.NewImage(sz, sz)
		img.Fill(c)
		
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(sz)/2, p.Y-float64(sz)/2)
		screen.DrawImage(img, op)
	}
	
	// Bullets
	for _, b := range g.Bullets {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(b.X-4, b.Y-4)
		screen.DrawImage(g.BulletImg, op)
	}
	
	// Enemies
	for _, e := range g.Enemies {
		img := g.EnemyImgs[e.Type]
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-float64(img.Bounds().Dx())/2, e.Y-float64(img.Bounds().Dy())/2)
		
		// Flash white when hit
		if e.HitTimer > 0 {
			op.ColorM.Scale(2, 2, 2, 1)
		}
		
		screen.DrawImage(img, op)
		
		// HP bar
		if e.HP < e.MaxHP {
			barW := float32(e.Radius * 2)
			barH := float32(3)
			hpRatio := e.HP / e.MaxHP
			
			vector.DrawFilledRect(screen, float32(e.X-float64(e.Radius)), float32(e.Y-float64(e.Radius)-8),
				barW, barH, color.RGBA{60, 60, 60, 200}, false)
			
			var hpColor color.RGBA
			if hpRatio > 0.6 {
				hpColor = hpGreen
			} else if hpRatio > 0.3 {
				hpColor = hpYellow
			} else {
				hpColor = hpRed
			}
			
			vector.DrawFilledRect(screen, float32(e.X-float64(e.Radius)), float32(e.Y-float64(e.Radius)-8),
				barW*float32(hpRatio), barH, hpColor, false)
		}
	}
	
	// Player
	if g.Player.InvulnTimer <= 0 || int(g.GameTime*10)%2 == 0 {
		// Glow
		glow := createGlowCircle(40, playerGlow)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.Player.X-20, g.Player.Y-20)
		screen.DrawImage(glow, op)
		
		// Player body
		op2 := &ebiten.DrawImageOptions{}
		op2.GeoM.Translate(g.Player.X-12, g.Player.Y-12)
		screen.DrawImage(g.PlayerImg, op2)
		
		// Aim line
		aimLen := 25.0
		aimX := g.Player.X + math.Cos(g.Player.Angle)*aimLen
		aimY := g.Player.Y + math.Sin(g.Player.Angle)*aimLen
		vector.StrokeLine(screen, float32(g.Player.X), float32(g.Player.Y), float32(aimX), float32(aimY), 2, textNeon, false)
	}
	
	// Particles (front)
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
	
	// HUD
	g.drawHUD(screen)
	
	// Wave complete message
	if g.State == StateWaveComplete {
		alpha := int(g.WaveTimer / 2.0 * 255)
		c := color.RGBA{textNeon.R, textNeon.G, textNeon.B, uint8(alpha)}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА %d ПРОЙДЕНА!", g.Wave),
			ScreenW/2-140, ScreenH/2-30)
		_ = c
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// HUD background
	vector.DrawFilledRect(screen, 0, 0, float32(ScreenW), 50, hudBg, false)

	// HP bar
	hpBarW := float32(200)
	hpBarH := float32(20)
	hpX := float32(15)
	hpY := float32(15)
	hpRatio := g.Player.HP / g.Player.MaxHP

	vector.DrawFilledRect(screen, hpX, hpY, hpBarW, hpBarH, color.RGBA{40, 40, 40, 200}, false)

	var hpColor color.RGBA
	if hpRatio > 0.6 {
		hpColor = hpGreen
	} else if hpRatio > 0.3 {
		hpColor = hpYellow
	} else {
		hpColor = hpRed
	}

	vector.DrawFilledRect(screen, hpX, hpY, hpBarW*float32(hpRatio), hpBarH, hpColor, false)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %.0f/%.0f", g.Player.HP, g.Player.MaxHP),
		int(hpX+5), int(hpY+3))
	
	// Score
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 230, 18)
	
	// Wave
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 400, 18)
	
	// Combo
	if g.Combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("COMBO x%d", g.Combo), 560, 18)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title
	ebitenutil.DebugPrintAt(screen, "NEON SURVIVOR", ScreenW/2-130, 150)
	
	// Subtitle
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge — Day 103", ScreenW/2-120, 200)
	
	// Decorative neon circles
	for i := 0; i < 8; i++ {
		angle := float64(i) * 6.2832 / 8 + g.GameTime*0.5
		x := ScreenW/2 + int(math.Cos(angle)*120)
		y := 280 + int(math.Sin(angle)*40)
		
		var c color.RGBA
		switch i % 3 {
		case 0:
			c = playerColor
		case 1:
			c = enemyNormal
		case 2:
			c = bulletColor
		}
		
		img := createCircleImage(20, c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x-10), float64(y-10))
		screen.DrawImage(img, op)
	}
	
	// Play button
	g.drawButton(screen, "▶  ИГРАТЬ", ScreenW/2-80, 350, 160, 50)
	
	// Controls
	ebitenutil.DebugPrintAt(screen, "WASD — движение", ScreenW/2-90, 430)
	ebitenutil.DebugPrintAt(screen, "МЫШЬ — прицел", ScreenW/2-80, 455)
	ebitenutil.DebugPrintAt(screen, "ЛКМ — стрельба", ScreenW/2-80, 480)
	ebitenutil.DebugPrintAt(screen, "ESC/P — пауза", ScreenW/2-80, 505)
	
	// Best score
	if g.BestScore > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ЛУЧШИЙ: %d", g.BestScore),
			ScreenW/2-80, 545)
	}
}

func (g *Game) drawPause(screen *ebiten.Image) {
	// Overlay
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 150}, false)
	
	// Title
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", ScreenW/2-60, 250)
	
	// Buttons
	g.drawButton(screen, "ПРОДОЛЖИТЬ", ScreenW/2-80, 320, 160, 50)
	g.drawButton(screen, "В МЕНЮ", ScreenW/2-80, 380, 160, 50)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Overlay
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{0, 0, 0, 180}, false)
	
	// Title
	ebitenutil.DebugPrintAt(screen, "GAME OVER", ScreenW/2-100, 200)
	
	// Stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), ScreenW/2-70, 260)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), ScreenW/2-60, 290)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("MAX COMBO: %d", g.MaxCombo), ScreenW/2-80, 320)
	
	if g.Score >= g.BestScore && g.Score > 0 {
		ebitenutil.DebugPrintAt(screen, "НОВЫЙ РЕКОРД!", ScreenW/2-90, 350)
	}
	
	// Buttons
	g.drawButton(screen, "ЗАНОВО", ScreenW/2-80, 390, 160, 50)
	g.drawButton(screen, "В МЕНЮ", ScreenW/2-80, 450, 160, 50)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	mx, my := ebiten.CursorPosition()
	hover := float64(mx) >= float64(x) && float64(mx) <= float64(x+w) &&
		float64(my) >= float64(y) && float64(my) <= float64(y+h)
	
	var bgColor color.RGBA
	if hover {
		bgColor = color.RGBA{40, 60, 80, 255}
	} else {
		bgColor = color.RGBA{25, 35, 50, 255}
	}
	
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bgColor, false)
	vector.StrokeRect(screen, float32(x), float32(y), float32(w), float32(h), 2, textNeon, false)
	
	ebitenutil.DebugPrintAt(screen, text, x+20, y+h/2-10)
}

// ============================================================================
// LAYOUT & MAIN
// ============================================================================

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Neon Survivor — Go365 Day 103")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	
	game := NewGame()
	
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
