// Go365 Day 86 - GO MARIO: SPACE ADVENTURE v4.0.0
// Космический шутер (Geometry Wars / Asteroids style)
// Корабль, враги, боссы, волны, апгрейды, частицы

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	ScreenWidth  = 1024
	ScreenHeight = 768

	PlayerSpeed     = 5.0
	PlayerTurnSpeed = 0.08
	BulletSpeed     = 10.0
	EnemySpeed      = 2.0
)

// ============================================================================
// COLORS - Neon palette
// ============================================================================

var (
	ColorBG          = color.RGBA{10, 10, 20, 255}
	ColorPlayer      = color.RGBA{0, 255, 255, 255}
	ColorPlayerDark  = color.RGBA{0, 150, 150, 255}
	ColorBullet      = color.RGBA{255, 255, 100, 255}
	ColorEnemy1      = color.RGBA{255, 50, 50, 255}
	ColorEnemy2      = color.RGBA{255, 150, 50, 255}
	ColorEnemy3      = color.RGBA{255, 50, 255, 255}
	ColorBoss        = color.RGBA{200, 50, 50, 255}
	ColorParticle    = color.RGBA{255, 200, 50, 255}
	ColorHealth      = color.RGBA{0, 255, 100, 255}
	ColorGold        = color.RGBA{255, 215, 0, 255}
	ColorWave        = color.RGBA{100, 200, 255, 255}
)

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	gameFont  font.Face
	largeFont font.Face
}

var gameAssets *Assets

func LoadAssets() *Assets {
	assets := &Assets{}
	// Используем встроенный шрифт вместо загрузки из файла
	assets.gameFont = basicfont.Face7x13
	assets.largeFont = basicfont.Face7x13
	return assets
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Vector2 struct {
	X, Y float64
}

type Player struct {
	x, y         float64
	angle        float64
	vx, vy       float64
	health       int
	maxHealth    int
	shield       int
	score        int
	coins        int
	
	// Weapons
	weaponLevel  int
	fireRate     int
	fireDelay    int
	multiShot    bool
	homing       bool
	
	// Status
	invincible   int
	dashCooldown int
	isDashing    bool
}

type Bullet struct {
	x, y      float64
	vx, vy    float64
	angle     float64
	damage    int
	isPlayer  bool
	life      int
	color     color.RGBA
	homing    bool
}

type Enemy struct {
	x, y       float64
	vx, vy     float64
	angle      float64
	enemyType  int
	health     int
	maxHealth  int
	damage     int
	score      int
	size       float64
	isAlive    bool
	shootTimer int
}

type Boss struct {
	x, y         float64
	vx, vy       float64
	health       int
	maxHealth    int
	phase        int
	attackTimer  int
	pattern      int
	size         float64
	isAlive      bool
}

type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	maxLife  int
	color    color.RGBA
	size     float64
	decay    float64
}

type PowerUp struct {
	x, y      float64
	vy        float64
	pType     int // 0: health, 1: weapon, 2: shield, 3: multishot, 4: homing
	size      float64
	isActive  bool
	angle     float64
}

type Wave struct {
	number    int
	enemies   int
	elite     bool
	bossWave  bool
}

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
)

type Game struct {
	player    *Player
	bullets   []*Bullet
	enemies   []*Enemy
	boss      *Boss
	particles []*Particle
	powerUps  []*PowerUp
	
	state     GameState
	wave      Wave
	frame     int
	difficulty float64
	
	screenShake float64
	flashAlpha  int
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	gameAssets = LoadAssets()

	g := &Game{
		player: &Player{
			x: ScreenWidth / 2,
			y: ScreenHeight / 2,
			maxHealth: 100,
			health: 100,
			weaponLevel: 1,
			fireDelay: 15,
		},
		state: StateMenu,
		wave: Wave{
			number: 1,
			enemies: 5,
		},
		bullets: make([]*Bullet, 0),
		enemies: make([]*Enemy, 0),
		particles: make([]*Particle, 0),
		powerUps: make([]*PowerUp, 0),
		difficulty: 1.0,
	}

	return g
}

func (g *Game) StartGame() {
	g.state = StatePlaying
	g.player.health = g.player.maxHealth
	g.player.score = 0
	g.player.coins = 0
	g.player.weaponLevel = 1
	g.wave.number = 1
	g.wave.enemies = 5
	g.wave.bossWave = false
	g.difficulty = 1.0
	g.enemies = make([]*Enemy, 0)
	g.bullets = make([]*Bullet, 0)
	g.particles = make([]*Particle, 0)
	g.powerUps = make([]*PowerUp, 0)
	g.StartWave()
}

func (g *Game) StartWave() {
	w := &g.wave
	
	if w.number%5 == 0 {
		w.bossWave = true
		w.enemies = 0
		g.SpawnBoss()
	} else {
		w.bossWave = false
		w.enemies = 5 + w.number*2
		w.elite = w.number%3 == 0
		
		for i := 0; i < w.enemies; i++ {
			g.SpawnEnemy()
		}
	}
}

func (g *Game) SpawnEnemy() {
	// Spawn at edges
	angle := rand.Float64() * math.Pi * 2
	dist := float64(ScreenWidth/2 + 100)
	
	x := g.player.x + math.Cos(angle)*dist
	y := g.player.y + math.Sin(angle)*dist
	
	// Clamp to screen
	x = math.Max(50, math.Min(float64(ScreenWidth)-50, x))
	y = math.Max(50, math.Min(float64(ScreenHeight)-50, y))
	
	enemyType := rand.Intn(3)
	if g.wave.elite && rand.Float32() < 0.2 {
		enemyType = 2 // Elite
	}
	
	enemy := &Enemy{
		x: x,
		y: y,
		enemyType: enemyType,
		isAlive: true,
		size: 20,
	}
	
	switch enemyType {
	case 0: // Basic chaser
		enemy.maxHealth = 20 + g.wave.number*5
		enemy.health = enemy.maxHealth
		enemy.damage = 10
		enemy.score = 100
		enemy.size = 20
	case 1: // Shooter
		enemy.maxHealth = 15 + g.wave.number*3
		enemy.health = enemy.maxHealth
		enemy.damage = 15
		enemy.score = 150
		enemy.size = 18
		enemy.shootTimer = 120
	case 2: // Elite
		enemy.maxHealth = 80 + g.wave.number*15
		enemy.health = enemy.maxHealth
		enemy.damage = 25
		enemy.score = 500
		enemy.size = 35
	}
	
	g.enemies = append(g.enemies, enemy)
}

func (g *Game) SpawnBoss() {
	g.boss = &Boss{
		x: ScreenWidth / 2,
		y: -100,
		maxHealth: 500 + g.wave.number*100,
		health: 500 + g.wave.number*100,
		size: 80,
		isAlive: true,
		phase: 1,
	}
}

func (g *Game) SpawnPowerUp(x, y float64) {
	types := []int{0, 1, 2, 3, 4}
	pType := types[rand.Intn(len(types))]
	
	g.powerUps = append(g.powerUps, &PowerUp{
		x: x,
		y: y,
		pType: pType,
		size: 15,
		isActive: true,
	})
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frame++
	
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		g.updatePaused()
	case StateGameOver, StateVictory:
		g.updateEnd()
	}
	
	// Screen shake decay
	if g.screenShake > 0 {
		g.screenShake *= 0.9
		if g.screenShake < 0.5 {
			g.screenShake = 0
		}
	}
	
	// Flash decay
	if g.flashAlpha > 0 {
		g.flashAlpha -= 5
	}
	
	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.StartGame()
	}
}

func (g *Game) updatePlaying() {
	g.updatePlayer()
	g.updateBullets()
	g.updateEnemies()
	g.updateBoss()
	g.updateParticles()
	g.updatePowerUps()
	g.checkCollisions()
	g.checkWaveClear()
}

func (g *Game) updatePlayer() {
	p := g.player
	
	// Rotation
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.angle -= PlayerTurnSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.angle += PlayerTurnSpeed
	}
	
	// Thrust
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		p.vx += math.Cos(p.angle) * 0.3
		p.vy += math.Sin(p.angle) * 0.3
	}
	
	// Friction
	p.vx *= 0.98
	p.vy *= 0.98
	
	// Dash
	if p.isDashing {
		p.dashCooldown--
		if p.dashCooldown <= 0 {
			p.isDashing = false
		}
	}
	
	if (ebiten.IsKeyPressed(ebiten.KeyShift) || ebiten.IsKeyPressed(ebiten.KeyK)) && p.dashCooldown == 0 {
		p.isDashing = true
		p.dashCooldown = 60
		p.vx = math.Cos(p.angle) * 15
		p.vy = math.Sin(p.angle) * 15
		g.spawnDashParticles()
	}
	
	// Apply velocity
	p.x += p.vx
	p.y += p.vy
	
	// Screen wrap
	if p.x < 0 {
		p.x = ScreenWidth
	}
	if p.x > ScreenWidth {
		p.x = 0
	}
	if p.y < 0 {
		p.y = ScreenHeight
	}
	if p.y > ScreenHeight {
		p.y = 0
	}
	
	// Shooting
	if ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyZ) {
		if p.fireDelay <= 0 {
			g.fireBullet()
			p.fireDelay = p.fireRate
		}
	}
	if p.fireDelay > 0 {
		p.fireDelay--
	}
	
	// Invincibility
	if p.invincible > 0 {
		p.invincible--
	}
}

func (g *Game) fireBullet() {
	p := g.player
	
	// Create bullet
	bullet := &Bullet{
		x: p.x + math.Cos(p.angle)*30,
		y: p.y + math.Sin(p.angle)*30,
		angle: p.angle,
		vx: math.Cos(p.angle) * BulletSpeed,
		vy: math.Sin(p.angle) * BulletSpeed,
		damage: 10 + p.weaponLevel*5,
		isPlayer: true,
		life: 120,
		color: ColorBullet,
		homing: p.homing,
	}
	
	g.bullets = append(g.bullets, bullet)
	
	// Multi-shot
	if p.multiShot {
		for i := -1; i <= 1; i += 2 {
			angle := p.angle + float64(i)*0.2
			g.bullets = append(g.bullets, &Bullet{
				x: p.x + math.Cos(p.angle)*20,
				y: p.y + math.Sin(p.angle)*20,
				angle: angle,
				vx: math.Cos(angle) * BulletSpeed,
				vy: math.Sin(angle) * BulletSpeed,
				damage: 10 + p.weaponLevel*5,
				isPlayer: true,
				life: 120,
				color: ColorBullet,
			})
		}
	}
}

func (g *Game) updateBullets() {
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		
		// Homing
		if b.homing && b.isPlayer {
			var target *Enemy
			minDist := float64(300)
			
			for _, e := range g.enemies {
				if !e.isAlive {
					continue
				}
				dist := math.Hypot(e.x-b.x, e.y-b.y)
				if dist < minDist {
					minDist = dist
					target = e
				}
			}
			
			if target != nil {
				angle := math.Atan2(target.y-b.y, target.x-b.x)
				b.angle += (angle - b.angle) * 0.1
				b.vx = math.Cos(b.angle) * BulletSpeed
				b.vy = math.Sin(b.angle) * BulletSpeed
			}
		}
		
		b.x += b.vx
		b.y += b.vy
		b.life--
		
		// Out of bounds
		if b.life <= 0 || b.x < -50 || b.x > float64(ScreenWidth)+50 ||
			b.y < -50 || b.y > float64(ScreenHeight)+50 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}
}

func (g *Game) updateEnemies() {
	p := g.player
	
	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}
		
		// Move towards player
		angle := math.Atan2(p.y-e.y, p.x-e.x)
		
		if e.enemyType == 0 {
			// Basic chaser
			e.vx = math.Cos(angle) * EnemySpeed
			e.vy = math.Sin(angle) * EnemySpeed
		} else if e.enemyType == 1 {
			// Shooter - keep distance
			dist := math.Hypot(p.x-e.x, p.y-e.y)
			if dist > 300 {
				e.vx = math.Cos(angle) * EnemySpeed
				e.vy = math.Sin(angle) * EnemySpeed
			} else if dist < 200 {
				e.vx = -math.Cos(angle) * EnemySpeed
				e.vy = -math.Sin(angle) * EnemySpeed
			}
			
			// Shoot
			e.shootTimer--
			if e.shootTimer <= 0 {
				e.shootTimer = 120
				g.bullets = append(g.bullets, &Bullet{
					x: e.x,
					y: e.y,
					angle: angle,
					vx: math.Cos(angle) * 5,
					vy: math.Sin(angle) * 5,
					damage: e.damage,
					isPlayer: false,
					life: 180,
					color: ColorEnemy1,
				})
			}
		} else {
			// Elite - faster
			e.vx = math.Cos(angle) * EnemySpeed * 1.5
			e.vy = math.Sin(angle) * EnemySpeed * 1.5
		}
		
		e.x += e.vx
		e.y += e.vy
	}
}

func (g *Game) updateBoss() {
	if g.boss == nil || !g.boss.isAlive {
		return
	}
	
	b := g.boss
	p := g.player
	
	// Move
	b.y += 2
	if b.y > 150 {
		b.y = 150
	}
	
	// Side to side
	b.x += math.Sin(float64(g.frame)*0.02) * 3
	
	// Attack patterns
	b.attackTimer--
	if b.attackTimer <= 0 {
		b.attackTimer = 30
		
		switch b.pattern {
		case 0: // Spread shot
			for i := -2; i <= 2; i++ {
				angle := math.Atan2(p.y-b.y, p.x-b.x) + float64(i)*0.3
				g.bullets = append(g.bullets, &Bullet{
					x: b.x,
					y: b.y,
					angle: angle,
					vx: math.Cos(angle) * 4,
					vy: math.Sin(angle) * 4,
					damage: 20,
					isPlayer: false,
					life: 200,
					color: ColorBoss,
				})
			}
		case 1: // Circle shot
			for i := 0; i < 12; i++ {
				angle := float64(i) * math.Pi / 6
				g.bullets = append(g.bullets, &Bullet{
					x: b.x,
					y: b.y,
					angle: angle,
					vx: math.Cos(angle) * 3,
					vy: math.Sin(angle) * 3,
					damage: 15,
					isPlayer: false,
					life: 200,
					color: ColorBoss,
				})
			}
		case 2: // Aimed burst
			for i := 0; i < 5; i++ {
				angle := math.Atan2(p.y-b.y, p.x-b.x)
				g.bullets = append(g.bullets, &Bullet{
					x: b.x,
					y: b.y,
					angle: angle + (rand.Float64()-0.5)*0.5,
					vx: math.Cos(angle) * 6,
					vy: math.Sin(angle) * 6,
					damage: 20,
					isPlayer: false,
					life: 200,
					color: ColorBoss,
				})
			}
		}
		
		b.pattern = (b.pattern + 1) % 3
	}
	
	// Phase change
	if b.health < b.maxHealth/2 && b.phase == 1 {
		b.phase = 2
		b.attackTimer = 15
		g.screenShake = 10
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.life--
		p.size *= p.decay
		
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updatePowerUps() {
	p := g.player
	
	for i := range g.powerUps {
		pu := g.powerUps[i]
		if !pu.isActive {
			continue
		}
		
		pu.y += 2
		pu.angle += 0.05
		
		// Magnet
		dist := math.Hypot(pu.x-p.x, pu.y-p.y)
		if dist < 150 {
			pu.x += (p.x - pu.x) * 0.05
			pu.y += (p.y - pu.y) * 0.05
		}
		
		// Collect
		if dist < 30 {
			pu.isActive = false
			g.collectPowerUp(pu)
		}
	}
}

func (g *Game) collectPowerUp(pu *PowerUp) {
	p := g.player
	
	switch pu.pType {
	case 0: // Health
		p.health = min(p.health+30, p.maxHealth)
	case 1: // Weapon
		p.weaponLevel = min(p.weaponLevel+1, 5)
		p.fireRate = max(5, 15-p.weaponLevel*2)
	case 2: // Shield
		p.shield = min(p.shield+50, 100)
	case 3: // Multi-shot
		p.multiShot = true
	case 4: // Homing
		p.homing = true
	}
	
	g.spawnCollectParticles(pu.x, pu.y)
}

func (g *Game) checkCollisions() {
	p := g.player
	
	// Bullets vs Enemies/Boss/Player
	for _, b := range g.bullets {
		if b.isPlayer {
			// vs Enemies
			for _, e := range g.enemies {
				if !e.isAlive {
					continue
				}
				if math.Hypot(e.x-b.x, e.y-b.y) < e.size+5 {
					e.health -= b.damage
					g.spawnHitParticles(b.x, b.y, 5)
					
					if e.health <= 0 {
						e.isAlive = false
						p.score += e.score
						g.spawnExplosion(e.x, e.y, e.size)
						g.ScreenShake(5)
						
						// Drop powerup chance
						if rand.Float32() < 0.15 {
							g.SpawnPowerUp(e.x, e.y)
						}
					}
					
					g.bullets = append(g.bullets[:0], g.bullets[1:]...)
					break
				}
			}
			
			// vs Boss
			if g.boss != nil && g.boss.isAlive {
				if math.Hypot(g.boss.x-b.x, g.boss.y-b.y) < g.boss.size+10 {
					g.boss.health -= b.damage
					g.spawnHitParticles(b.x, b.y, 8)
					
					if g.boss.health <= 0 {
						g.boss.isAlive = false
						p.score += 5000
						g.spawnExplosion(g.boss.x, g.boss.y, g.boss.size*2)
						g.ScreenShake(20)
						g.flashAlpha = 50
						
						// Drop multiple powerups
						for i := 0; i < 5; i++ {
							g.SpawnPowerUp(g.boss.x+(rand.Float64()-0.5)*100, g.boss.y+(rand.Float64()-0.5)*100)
						}
						
						g.wave.number++
						g.StartWave()
					}
					
					g.bullets = append(g.bullets[:0], g.bullets[1:]...)
				}
			}
		} else {
			// Enemy bullet vs Player
			if p.invincible == 0 && !p.isDashing {
				if math.Hypot(p.x-b.x, p.y-b.y) < 20 {
					g.playerHit(b.damage)
					g.bullets = append(g.bullets[:0], g.bullets[1:]...)
				}
			}
		}
	}
	
	// Enemies vs Player
	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}
		if p.invincible == 0 && !p.isDashing {
			if math.Hypot(p.x-e.x, p.y-e.y) < e.size+15 {
				g.playerHit(e.damage)
				e.isAlive = false
				g.spawnExplosion(e.x, e.y, e.size)
			}
		}
	}
	
	// Boss vs Player
	if g.boss != nil && g.boss.isAlive {
		if p.invincible == 0 && !p.isDashing {
			if math.Hypot(p.x-g.boss.x, p.y-g.boss.y) < g.boss.size+15 {
				g.playerHit(30)
			}
		}
	}
}

func (g *Game) playerHit(damage int) {
	p := g.player
	
	if p.shield > 0 {
		p.shield -= damage
		if p.shield < 0 {
			p.health += p.shield
			p.shield = 0
		}
	} else {
		p.health -= damage
	}
	
	p.invincible = 60
	g.screenShake = 8
	g.flashAlpha = 80
	
	if p.health <= 0 {
		g.state = StateGameOver
	}
}

func (g *Game) checkWaveClear() {
	// Check if all enemies dead
	allDead := true
	for _, e := range g.enemies {
		if e.isAlive {
			allDead = false
			break
		}
	}
	
	if allDead && !g.wave.bossWave && len(g.enemies) > 0 {
		g.wave.number++
		g.difficulty += 0.2
		g.StartWave()
	}
	
	if g.boss != nil && !g.boss.isAlive && g.wave.bossWave {
		g.boss = nil
		if g.wave.number >= 5 {
			g.state = StateVictory
		}
	}
}

func (g *Game) updatePaused() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) {
		g.state = StatePlaying
	}
}

func (g *Game) updateEnd() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.StartGame()
	}
}

func (g *Game) ScreenShake(amount float64) {
	g.screenShake = amount
}

// ============================================================================
// PARTICLES
// ============================================================================

func (g *Game) spawnHitParticles(x, y float64, count int) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: (rand.Float64() - 0.5) * 8,
			vy: (rand.Float64() - 0.5) * 8,
			life: 20 + rand.Intn(10),
			color: ColorParticle,
			size: float64(3 + rand.Float32()*2),
			decay: 0.9,
		})
	}
}

func (g *Game) spawnExplosion(x, y float64, size float64) {
	count := int(size)
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: (rand.Float64() - 0.5) * 12,
			vy: (rand.Float64() - 0.5) * 12,
			life: 30 + rand.Intn(20),
			color: color.RGBA{255, uint8(100+rand.Intn(155)), 0, 255},
			size: float64(4 + rand.Float32()*4),
			decay: 0.88,
		})
	}
}

func (g *Game) spawnDashParticles() {
	p := g.player
	for i := 0; i < 15; i++ {
		g.particles = append(g.particles, &Particle{
			x: p.x - math.Cos(p.angle)*20,
			y: p.y - math.Sin(p.angle)*20,
			vx: (rand.Float64() - 0.5) * 5,
			vy: (rand.Float64() - 0.5) * 5,
			life: 15 + rand.Intn(10),
			color: ColorPlayer,
			size: float64(4 + rand.Float32()*2),
			decay: 0.85,
		})
	}
}

func (g *Game) spawnCollectParticles(x, y float64) {
	for i := 0; i < 10; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: (rand.Float64() - 0.5) * 6,
			vy: (rand.Float64() - 0.5) * 6,
			life: 25 + rand.Intn(10),
			color: ColorGold,
			size: float64(3 + rand.Float32()*2),
			decay: 0.9,
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	// Apply screen shake
	op := &ebiten.DrawImageOptions{}
	if g.screenShake > 0 {
		op.GeoM.Translate((rand.Float64()-0.5)*g.screenShake, (rand.Float64()-0.5)*g.screenShake)
	}
	
	// Create game surface
	gameSurface := ebiten.NewImage(ScreenWidth, ScreenHeight)
	
	switch g.state {
	case StateMenu:
		g.drawMenu(gameSurface)
	case StatePlaying, StatePaused:
		g.drawPlaying(gameSurface)
	case StateGameOver:
		g.drawGameOver(gameSurface)
	case StateVictory:
		g.drawVictory(gameSurface)
	}
	
	screen.DrawImage(gameSurface, op)
	
	// Flash overlay
	if g.flashAlpha > 0 {
		flash := ebiten.NewImage(ScreenWidth, ScreenHeight)
		flash.Fill(color.RGBA{255, 255, 255, uint8(g.flashAlpha)})
		screen.DrawImage(flash, &ebiten.DrawImageOptions{})
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	screen.Fill(ColorBG)

	// Title - центрируем вручную
	title := "GO MARIO: SPACE ADVENTURE"
	if gameAssets.largeFont != nil {
		text.Draw(screen, title, gameAssets.largeFont, ScreenWidth/2-150, 200, ColorPlayer)
		text.Draw(screen, "Geometry Wars Style Shooter", gameAssets.gameFont, ScreenWidth/2-140, 250, ColorWave)
	}

	// Instructions
	instructions := []string{
		"LEFT/RIGHT or A/D - Rotate",
		"UP or W - Thrust",
		"SPACE or J/Z - Fire",
		"SHIFT or K - Dash",
		"",
		"Press ENTER to start",
	}

	y := 350
	for _, line := range instructions {
		if gameAssets.gameFont != nil {
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-120, y, color.White)
		}
		y += 25
	}

	// Features
	features := []string{
		"[+] Waves of enemies",
		"[!] Boss battles",
		"[^] Power-ups",
		"[*] Weapon upgrades",
	}

	y = 520
	for _, line := range features {
		if gameAssets.gameFont != nil {
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-100, y, ColorGold)
		}
		y += 22
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	screen.Fill(ColorBG)
	
	// Draw powerups
	for _, pu := range g.powerUps {
		if !pu.isActive {
			continue
		}
		
		puColor := ColorHealth
		switch pu.pType {
		case 1: puColor = ColorBullet
		case 2: puColor = ColorWave
		case 3: puColor = ColorEnemy2
		case 4: puColor = ColorEnemy3
		}
		
		vector.DrawFilledCircle(screen, float32(pu.x), float32(pu.y), float32(pu.size), puColor, true)
		
		// Icon
		icon := "?"
		switch pu.pType {
		case 0: icon = "+"
		case 1: icon = "W"
		case 2: icon = "S"
		case 3: icon = "M"
		case 4: icon = "H"
		}
		if gameAssets.gameFont != nil {
			text.Draw(screen, icon, gameAssets.gameFont, int(pu.x)-5, int(pu.y)+7, color.White)
		}
	}
	
	// Draw player
	p := g.player
	if p.invincible == 0 || g.frame%4 < 2 {
		g.drawPlayerShip(screen, p.x, p.y, p.angle)
	}
	
	// Draw bullets
	for _, b := range g.bullets {
		vector.DrawFilledCircle(screen, float32(b.x), float32(b.y), 5, b.color, true)
	}
	
	// Draw enemies
	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}
		
		eColor := ColorEnemy1
		if e.enemyType == 1 {
			eColor = ColorEnemy2
		} else if e.enemyType == 2 {
			eColor = ColorEnemy3
		}
		
		// Draw enemy shape
		g.drawEnemyShape(screen, e.x, e.y, e.size, e.angle, eColor)
		
		// Health bar
		if e.health < e.maxHealth {
			barW := float32(e.size * 2)
			vector.DrawFilledRect(screen, float32(e.x)-barW/2, float32(e.y)-float32(e.size)-10, barW, 4, color.RGBA{80, 0, 0, 255}, true)
			vector.DrawFilledRect(screen, float32(e.x)-barW/2, float32(e.y)-float32(e.size)-10, barW*float32(e.health)/float32(e.maxHealth), 4, ColorHealth, true)
		}
	}
	
	// Draw boss
	if g.boss != nil && g.boss.isAlive {
		b := g.boss
		vector.DrawFilledCircle(screen, float32(b.x), float32(b.y), float32(b.size), ColorBoss, true)
		vector.StrokeCircle(screen, float32(b.x), float32(b.y), float32(b.size), 5, color.RGBA{255, 100, 100, 200}, true)
		
		// Boss health bar
		barW := float32(400)
		vector.DrawFilledRect(screen, ScreenWidth/2-barW/2, 20, barW, 20, color.RGBA{80, 0, 0, 255}, true)
		vector.DrawFilledRect(screen, ScreenWidth/2-barW/2, 20, barW*float32(b.health)/float32(b.maxHealth), 20, ColorBoss, true)
	}
	
	// Draw particles
	for _, part := range g.particles {
		alpha := uint8(255 * part.life / part.maxLife)
		c := color.RGBA{part.color.R, part.color.G, part.color.B, alpha}
		vector.DrawFilledCircle(screen, float32(part.x), float32(part.y), float32(part.size), c, true)
	}
	
	// Draw HUD
	g.drawHUD(screen)
}

func (g *Game) drawPlayerShip(screen *ebiten.Image, x, y, angle float64) {
	// Draw triangle ship
	p1x := float32(x + math.Cos(angle)*25)
	p1y := float32(y + math.Sin(angle)*25)
	p2x := float32(x + math.Cos(angle+2.5)*15)
	p2y := float32(y + math.Sin(angle+2.5)*15)
	p3x := float32(x + math.Cos(angle-2.5)*15)
	p3y := float32(y + math.Sin(angle-2.5)*15)
	
	// Fill triangle using lines
	vector.StrokeLine(screen, p1x, p1y, p2x, p2y, 3, ColorPlayer, true)
	vector.StrokeLine(screen, p2x, p2y, p3x, p3y, 3, ColorPlayer, true)
	vector.StrokeLine(screen, p3x, p3y, p1x, p1y, 3, ColorPlayer, true)
	
	// Thruster flame
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		fx := x - math.Cos(angle)*20
		fy := y - math.Sin(angle)*20
		vector.DrawFilledCircle(screen, float32(fx), float32(fy), 8+rand.Float32()*4, color.RGBA{255, 150, 50, 255}, true)
	}
}

func (g *Game) drawEnemyShape(screen *ebiten.Image, x, y, size, angle float64, c color.RGBA) {
	// Draw diamond shape using lines
	s := float32(size)
	cx, cy := float32(x), float32(y)
	
	// Four points of diamond
	p1x := cx + float32(math.Cos(angle))*s
	p1y := cy + float32(math.Sin(angle))*s
	p2x := cx + float32(math.Cos(angle+math.Pi/2))*s*0.7
	p2y := cy + float32(math.Sin(angle+math.Pi/2))*s*0.7
	p3x := cx + float32(math.Cos(angle+math.Pi))*s
	p3y := cy + float32(math.Sin(angle+math.Pi))*s
	p4x := cx + float32(math.Cos(angle-math.Pi/2))*s*0.7
	p4y := cy + float32(math.Sin(angle-math.Pi/2))*s*0.7
	
	// Draw diamond outline
	vector.StrokeLine(screen, p1x, p1y, p2x, p2y, 2, c, true)
	vector.StrokeLine(screen, p2x, p2y, p3x, p3y, 2, c, true)
	vector.StrokeLine(screen, p3x, p3y, p4x, p4y, 2, c, true)
	vector.StrokeLine(screen, p4x, p4y, p1x, p1y, 2, c, true)
	
	// Fill center
	vector.DrawFilledCircle(screen, cx, cy, s*0.5, c, true)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	p := g.player

	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 50, color.RGBA{0, 0, 0, 150}, true)

	// Health
	vector.DrawFilledRect(screen, 20, 15, 200, 15, color.RGBA{80, 0, 0, 255}, true)
	if p.maxHealth > 0 {
		vector.DrawFilledRect(screen, 20, 15, 200*float32(p.health)/float32(p.maxHealth), 15, ColorHealth, true)
	}

	// Shield
	if p.shield > 0 {
		vector.DrawFilledRect(screen, 20, 33, 150, 8, color.RGBA{0, 0, 80, 255}, true)
		vector.DrawFilledRect(screen, 20, 33, 150*float32(p.shield)/100, 8, ColorWave, true)
	}

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("HP %d/%d", p.health, p.maxHealth), gameAssets.gameFont, 230, 28, color.White)
		text.Draw(screen, fmt.Sprintf("SCORE: %d", p.score), gameAssets.gameFont, 400, 28, ColorGold)
		text.Draw(screen, fmt.Sprintf("WAVE: %d", g.wave.number), gameAssets.gameFont, 600, 28, ColorWave)
		text.Draw(screen, fmt.Sprintf("WP: %d", p.weaponLevel), gameAssets.gameFont, 750, 28, ColorBullet)

		if p.multiShot {
			text.Draw(screen, "[MULTI]", gameAssets.gameFont, 850, 28, ColorEnemy2)
		}
		if p.homing {
			text.Draw(screen, "[HOMING]", gameAssets.gameFont, 940, 28, ColorEnemy3)
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 20, 20, 255})

	if gameAssets.largeFont != nil {
		text.Draw(screen, "GAME OVER", gameAssets.largeFont, ScreenWidth/2-80, ScreenHeight/2-50, ColorEnemy1)
	}

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+20, color.White)
		text.Draw(screen, fmt.Sprintf("Wave: %d", g.wave.number), gameAssets.gameFont, ScreenWidth/2-50, ScreenHeight/2+45, color.White)
		text.Draw(screen, "Press ENTER to restart", gameAssets.gameFont, ScreenWidth/2-90, ScreenHeight/2+90, color.White)
	}
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 50, 20, 255})

	if gameAssets.largeFont != nil {
		text.Draw(screen, "VICTORY!", gameAssets.largeFont, ScreenWidth/2-60, ScreenHeight/2-50, ColorHealth)
	}

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+20, color.White)
		text.Draw(screen, "You defeated all waves!", gameAssets.gameFont, ScreenWidth/2-90, ScreenHeight/2+45, color.White)
		text.Draw(screen, "Press ENTER to play again", gameAssets.gameFont, ScreenWidth/2-95, ScreenHeight/2+90, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GO MARIO: SPACE ADVENTURE - Go365 Day 86 | Space Shooter")
	ebiten.SetVsyncEnabled(true)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
