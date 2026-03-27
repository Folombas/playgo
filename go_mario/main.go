// Go365 Day 87 - SUPER GO MARIO v3.2.0
// С оружием и исправленной анимацией!

package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 48
	Gravity      = 0.6
	JumpForce    = -14.0
	Speed        = 5.0
	MaxWorld     = 3
)

// Состояния игры
const (
	StateMenu = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateWin
	StateBoss
)

// Power-up типы
const (
	PowerNone = iota
	PowerMushroom
	PowerStar
	PowerFlower
)

// Типы оружия
const (
	WeaponNone = iota
	WeaponSword
	WeaponGun
	WeaponBlaster
)

// Кэш спрайтов
var spriteCache map[string]*ebiten.Image
var walkSprites []*ebiten.Image // Кадры анимации ходьбы

func initSpriteCache() {
	spriteCache = make(map[string]*ebiten.Image)
	walkSprites = make([]*ebiten.Image, 0)

	// Игрок
	spriteCache["player_stand"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
	spriteCache["player_jump"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")

	// Загружаем спрайт-лист ходьбы и разрезаем на кадры
	walkSheet, _, _ := ebitenutil.NewImageFromFile("assets/sprites/mario/p1_walk/p1_walk.png")
	if walkSheet != nil {
		bounds := walkSheet.Bounds()
		frameWidth := bounds.Dx() / 2 // 2 кадра
		// Вырезаем 2 кадра
		for i := 0; i < 2; i++ {
			rect := image.Rect(i*frameWidth, 0, (i+1)*frameWidth, bounds.Dy())
			frame := walkSheet.SubImage(rect).(*ebiten.Image)
			walkSprites = append(walkSprites, frame)
		}
	}

	spriteCache["player_hurt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_hurt.png")

	// Тайлы
	spriteCache["grass"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
	spriteCache["dirt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
	spriteCache["brick"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")
	spriteCache["box"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")
	spriteCache["chest"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxItem.png")

	// Враги
	spriteCache["slime1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
	spriteCache["slime2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")
	spriteCache["fly1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly1.png")
	spriteCache["fly2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly2.png")

	// Предметы
	spriteCache["coin"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
	spriteCache["mushroom"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/mushroomRed.png")
	spriteCache["star"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/star.png")
	spriteCache["flower"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/plantPurple.png")
	spriteCache["key"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/keyGold.png")
	spriteCache["bomb"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/bomb.png")

	// Оружие в сундуках
	spriteCache["sword"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemRed.png")
	spriteCache["gun"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemBlue.png")
	spriteCache["blaster"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemGreen.png")
}

func getSprite(name string) *ebiten.Image {
	if spriteCache == nil {
		initSpriteCache()
	}
	return spriteCache[name]
}

func getWalkSprite(frame int) *ebiten.Image {
	if len(walkSprites) == 0 {
		initSpriteCache()
	}
	if frame >= len(walkSprites) {
		frame = 0
	}
	return walkSprites[frame]
}

// Игрок
type Player struct {
	x, y           float64
	vx, vy         float64
	width          float64
	height         float64
	onGround       bool
	facing         int
	animFrame      int
	walkFrame      int
	walkAnimTimer  int
	health         int
	maxHealth      int
	powerLevel     int
	powerTimer     int
	invincible     int
	coins          int
	score          int
	lives          int
	combo          int
	comboTimer     int
	maxCombo       int
	fireCooldown   int
	weapon         int
	weaponCooldown int
	hasDoubleJump  bool
	jumpCount      int
}

// Тайл
type Tile struct {
	x, y    int
	id      int
	opened  bool // Для сундуков
}

// Снаряд
type Projectile struct {
	x, y, vx, vy float64
	damage       int
	life         int
	isPlayer     bool
	color        color.RGBA
	size         float32
}

// Враг
type Enemy struct {
	x, y      float64
	vx, vy    float64
	enemyType int
	alive     bool
	animFrame int
	health    int
	maxHealth int
}

// Монета
type Coin struct {
	x, y      float64
	value     int
	collected bool
	animFrame int
}

// Частица
type Particle struct {
	x, y, vx, vy float64
	life         int
	maxLife      int
	color        color.RGBA
	size         float32
}

// Power-up
type PowerUp struct {
	x, y      float64
	id        int
	alive     bool
	animFrame int
}

// Босс
type Boss struct {
	x, y        float64
	vx, vy      float64
	width       float64
	height      float64
	health      int
	maxHealth   int
	phase       int
	attackTimer int
	alive       bool
}

// Игра
type Game struct {
	player      *Player
	tiles       []*Tile
	enemies     []*Enemy
	coins       []*Coin
	projectiles []*Projectile
	particles   []*Particle
	powerUps    []*PowerUp
	boss        *Boss
	cameraX     float64
	levelW      int
	levelH      int
	frame       int
	state       int
	world       int
	level       int
	flagX       float64
	screenShake float64
	comboText   int
	highScore   int
}

// ============================================================================
// ИНИЦИАЛИЗАЦИЯ
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	if spriteCache == nil {
		initSpriteCache()
	}

	g := &Game{
		player: &Player{
			x:         100,
			y:         300,
			width:     40,
			height:    48,
			facing:    1,
			maxHealth: 3,
			health:    3,
			lives:     3,
			weapon:    WeaponNone,
		},
		state:  StateMenu,
		world:  1,
		level:  1,
		levelH: 16,
	}

	g.startLevel()
	return g
}

func (g *Game) startLevel() {
	g.levelW = 80 + g.world*20
	g.tiles = nil
	g.enemies = nil
	g.coins = nil
	g.projectiles = nil
	g.particles = nil
	g.powerUps = nil
	g.boss = nil

	g.generateLevel()

	g.player.x = 100
	g.player.y = 300
	g.player.vx = 0
	g.player.vy = 0
	g.cameraX = 0
}

func (g *Game) generateLevel() {
	// Земля
	for x := 0; x < g.levelW; x++ {
		if x > 15 && x < g.levelW-10 && x%30 >= 25 && x%30 < 28 {
			continue
		}
		g.tiles = append(g.tiles, &Tile{x: x, y: 13, id: 1})
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: 2})
		g.tiles = append(g.tiles, &Tile{x: x, y: 15, id: 2})
	}

	// Платформы и сундуки
	for x := 10; x < g.levelW-15; x++ {
		if x%12 < 4 && x%20 != 0 {
			platY := 7 + rand.Intn(3)
			for i := 0; i < 3+rand.Intn(2); i++ {
				tileID := 3
				r := rand.Float32()
				if r < 0.15 {
					tileID = 4 // box
				} else if r < 0.20 {
					tileID = 5 // chest (сундук)
				}
				g.tiles = append(g.tiles, &Tile{x: x + i, y: platY, id: tileID, opened: false})
			}
		}
	}

	// Монеты
	for x := 10; x < g.levelW-15; x++ {
		if rand.Float32() < 0.15 {
			coinY := float64(5+rand.Intn(5)) * TileSize
			g.coins = append(g.coins, &Coin{
				x:     float64(x*TileSize + 15),
				y:     coinY,
				value: 1 + rand.Intn(3),
			})
		}
	}

	// Враги
	for x := 20; x < g.levelW-15; x++ {
		if rand.Float32() < 0.08 {
			enemyType := rand.Intn(2)
			enemyY := float64(12 * TileSize)
			if enemyType == 1 && rand.Float32() < 0.5 {
				enemyY = float64(6+rand.Intn(4)) * TileSize
			}
			g.enemies = append(g.enemies, &Enemy{
				x:         float64(x * TileSize),
				y:         enemyY,
				vx:        -1.5,
				enemyType: enemyType,
				alive:     true,
				health:    1,
			})
		}
	}

	// Power-ups
	for x := 15; x < g.levelW-15; x++ {
		if rand.Float32() < 0.02 {
			g.powerUps = append(g.powerUps, &PowerUp{
				x:     float64(x*TileSize + 10),
				y:     float64(10 * TileSize),
				id:    rand.Intn(3),
				alive: true,
			})
		}
	}

	g.flagX = float64((g.levelW - 8) * TileSize)

	if g.level%3 == 0 {
		g.state = StateBoss
		g.boss = &Boss{
			x:         g.flagX - 200,
			y:         400,
			width:     80,
			height:    80,
			health:    10 + g.world*5,
			maxHealth: 10 + g.world*5,
			alive:     true,
		}
	}
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frame++

	switch g.state {
	case StateMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = StatePlaying
		}
		return nil

	case StatePaused:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = StatePlaying
		}
		return nil

	case StateGameOver, StateWin:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			if g.state == StateWin && g.world < MaxWorld {
				g.world++
				g.level = 1
				g.startLevel()
				g.state = StatePlaying
			} else {
				g.saveHighScore()
				g.world = 1
				g.level = 1
				g.player.health = g.player.maxHealth
				g.player.lives = 3
				g.player.score = 0
				g.player.coins = 0
				g.player.powerLevel = PowerNone
				g.player.weapon = WeaponNone
				g.startLevel()
				g.state = StatePlaying
			}
		}
		return nil

	case StateBoss:
		g.updatePlayer()
		g.updateCamera()
		g.updateBoss()
		g.updateProjectiles()
		g.updateParticles()
		if g.boss == nil || !g.boss.alive {
			g.player.score += 1000
			g.spawnParticles(g.boss.x+40, g.boss.y+40, 50, color.RGBA{255, 100, 100, 255})
			g.boss = nil
			g.level++
			g.startLevel()
			g.state = StatePlaying
		}
		return nil
	}

	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.updateCoins()
	g.updatePowerUps()
	g.updateProjectiles()
	g.updateParticles()
	g.checkCollisions()
	g.updateScreenShake()

	if g.player.x > g.flagX {
		g.level++
		if g.level > 3 {
			g.world++
			g.level = 1
		}
		if g.world > MaxWorld {
			g.state = StateWin
		} else {
			g.startLevel()
		}
	}

	return nil
}

func (g *Game) updatePlayer() {
	p := g.player

	// Управление движением
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = Speed
		p.facing = 1
		p.walkFrame = (p.walkFrame + 1) % 2
		p.walkAnimTimer++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -Speed
		p.facing = -1
		p.walkFrame = (p.walkFrame + 1) % 2
		p.walkAnimTimer++
	} else {
		p.vx = 0
		p.walkFrame = 0
	}

	// Прыжок
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && p.jumpCount == 0 {
		p.vy = JumpForce
		p.onGround = false
		p.jumpCount = 1
		g.spawnParticles(p.x+p.width/2, p.y+p.height, 10, color.RGBA{255, 255, 255, 255})
	}

	// Двойной прыжок (если есть звезда)
	if p.powerLevel == PowerStar && p.jumpCount < 2 {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) || inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			p.vy = JumpForce
			p.jumpCount = 2
			g.spawnParticles(p.x+p.width/2, p.y+p.height, 10, color.RGBA{255, 255, 0, 255})
		}
	}

	// Атака оружием
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) && p.weaponCooldown <= 0 {
		p.weaponCooldown = 20
		g.playerAttack()
	}
	if p.weaponCooldown > 0 {
		p.weaponCooldown--
	}

	// Стрельба (если есть flower)
	if inpututil.IsKeyJustPressed(ebiten.KeyX) && p.powerLevel == PowerFlower && p.fireCooldown <= 0 {
		p.fireCooldown = 30
		g.projectiles = append(g.projectiles, &Projectile{
			x:        p.x + p.width,
			y:        p.y + p.height/2,
			vx:       float64(p.facing) * 8,
			vy:       0,
			damage:   1,
			life:     60,
			isPlayer: true,
			color:    color.RGBA{255, 100, 0, 255},
			size:     12,
		})
	}
	if p.fireCooldown > 0 {
		p.fireCooldown--
	}

	// Физика
	p.vy += Gravity
	if p.vy > 12 {
		p.vy = 12
	}

	p.x += p.vx
	p.y += p.vy

	// Коллизии
	p.onGround = false
	p.jumpCount = 0
	for _, t := range g.tiles {
		tx := float64(t.x * TileSize)
		ty := float64(t.y * TileSize)

		if p.x < tx+TileSize-5 && p.x+p.width > tx+5 &&
			p.y < ty+TileSize && p.y+p.height > ty {

			if p.vy >= 0 && p.y+p.height <= ty+25 {
				p.y = ty - p.height
				p.vy = 0
				p.onGround = true
				p.jumpCount = 0
			} else if p.vy < 0 && p.y >= ty {
				p.y = ty + TileSize
				p.vy = 0
			}
		}
	}

	// Границы
	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(g.levelW*TileSize)-p.width {
		p.x = float64(g.levelW*TileSize) - p.width
	}

	// Смерть
	if p.y > ScreenHeight+100 {
		p.health = 0
	}

	// Таймеры
	if p.powerTimer > 0 {
		p.powerTimer--
		if p.powerTimer <= 0 && p.powerLevel == PowerStar {
			p.powerLevel = PowerNone
		}
	}
	if p.invincible > 0 {
		p.invincible--
	}
	if p.combo > 0 {
		p.comboTimer--
		if p.comboTimer <= 0 {
			p.combo = 0
			g.comboText = 0
		}
	}
	if g.comboText > 0 {
		g.comboText--
	}
}

func (g *Game) playerAttack() {
	p := g.player

	switch p.weapon {
	case WeaponNone:
		// Без оружия - ничего
		return
	case WeaponSword:
		// Меч - ближний бой
		for _, e := range g.enemies {
			if !e.alive {
				continue
			}
			dist := p.x - e.x
			if dist < 0 {
				dist = -dist
			}
			if dist < 60 && e.y > p.y-20 && e.y < p.y+40 {
				e.alive = false
				p.score += 200
				g.spawnParticles(e.x+18, e.y+18, 15, color.RGBA{255, 255, 255, 255})
				g.screenShake = 3
			}
		}
		// Босс
		if g.boss != nil && g.boss.alive {
			dist := p.x - g.boss.x
			if dist < 0 {
				dist = -dist
			}
			if dist < 100 {
				g.boss.health--
				g.spawnParticles(g.boss.x+40, g.boss.y+40, 10, color.RGBA{255, 100, 100, 255})
				g.screenShake = 3
				if g.boss.health <= 0 {
					g.boss.alive = false
				}
			}
		}
	case WeaponGun, WeaponBlaster:
		// Стрельба
		damage := 1
		if p.weapon == WeaponBlaster {
			damage = 3
		}
		g.projectiles = append(g.projectiles, &Projectile{
			x:        p.x + p.width,
			y:        p.y + p.height/2,
			vx:       float64(p.facing) * 10,
			vy:       0,
			damage:   damage,
			life:     60,
			isPlayer: true,
			color:    color.RGBA{255, 255, 0, 255},
			size:     8,
		})
	}
}

func (g *Game) updateCamera() {
	targetX := g.player.x - ScreenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	maxX := float64(g.levelW*TileSize) - ScreenWidth
	if g.cameraX > maxX {
		g.cameraX = maxX
	}
}

func (g *Game) updateEnemies() {
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		e.x += e.vx
		e.animFrame++
		if rand.Float32() < 0.02 {
			e.vx *= -1
		}
		if e.enemyType == 1 {
			e.vy += 0.2
			if e.vy > 3 {
				e.vy = -3
			}
			e.y += e.vy
		}
	}
}

func (g *Game) updateBoss() {
	b := g.boss
	if b == nil || !b.alive {
		return
	}
	p := g.player
	if p.x > b.x {
		b.vx = 2
	} else {
		b.vx = -2
	}
	if rand.Float32() < 0.03 && b.y >= 400 {
		b.vy = -12
	}
	b.vy += 0.5
	if b.vy > 10 {
		b.vy = 10
	}
	b.x += b.vx
	b.y += b.vy
	if b.y > 400 {
		b.y = 400
		b.vy = 0
	}
	b.attackTimer--
	if b.attackTimer <= 0 {
		b.attackTimer = 40
		g.projectiles = append(g.projectiles, &Projectile{
			x:        b.x + 40,
			y:        b.y + 40,
			vx:       float64(p.facing) * 5,
			vy:       -0.5,
			damage:   1,
			life:     90,
			isPlayer: false,
			color:    color.RGBA{255, 50, 50, 255},
			size:     15,
		})
	}
	if rectOverlap(p.x, p.y, p.width, p.height, b.x, b.y, 80, 80) {
		if p.vy > 0 && p.y+p.height < b.y+40 {
			b.health--
			p.vy = -10
			g.spawnParticles(b.x+40, b.y, 10, color.RGBA{255, 100, 100, 255})
			g.screenShake = 5
			if b.health <= 0 {
				b.alive = false
			}
		} else if p.invincible <= 0 {
			p.health--
			p.invincible = 120
			p.vy = -8
			if p.x < b.x {
				p.vx = -10
			} else {
				p.vx = 10
			}
			g.screenShake = 10
		}
	}
}

func (g *Game) updateProjectiles() {
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		p := g.projectiles[i]
		p.x += p.vx
		p.y += p.vy
		p.life--
		if p.life <= 0 {
			g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
		}
	}
}

func (g *Game) updateCoins() {
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		c.animFrame++
	}
}

func (g *Game) updatePowerUps() {
	for _, p := range g.powerUps {
		if !p.alive {
			continue
		}
		p.animFrame++
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.3
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) checkCollisions() {
	p := g.player

	// Монеты
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, c.x-16, c.y-16, 32, 32) {
			c.collected = true
			p.coins += c.value
			p.score += c.value * 100
			g.spawnParticles(c.x, c.y, 5, color.RGBA{255, 215, 0, 255})
		}
	}

	// Сундуки с оружием
	for _, t := range g.tiles {
		if t.id == 5 && !t.opened { // chest
			if rectOverlap(p.x, p.y, p.width, p.height, float64(t.x*TileSize), float64(t.y*TileSize), TileSize, TileSize) {
				t.opened = true
				// Даём случайное оружие
				weapon := rand.Intn(3) + 1 // 1=sword, 2=gun, 3=blaster
				p.weapon = weapon
				p.score += 500
				g.spawnParticles(float64(t.x*TileSize)+24, float64(t.y*TileSize)+24, 20, color.RGBA{255, 255, 100, 255})
			}
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, e.x, e.y, 36, 36) {
			if p.vy > 0 && p.y+p.height < e.y+20 {
				e.alive = false
				p.vy = -8
				p.score += 200
				p.combo++
				p.comboTimer = 120
				g.comboText = 60
				if p.combo > p.maxCombo {
					p.maxCombo = p.combo
				}
				g.spawnParticles(e.x+18, e.y+18, 15, color.RGBA{100, 255, 100, 255})
				g.screenShake = 3
			} else if p.invincible <= 0 {
				if p.powerLevel > 0 && p.powerLevel != PowerStar {
					p.powerLevel--
					p.invincible = 90
				} else if p.powerLevel == PowerStar {
					p.score += 100
				} else {
					p.health--
					p.invincible = 120
				}
				p.vy = -6
				if p.x < e.x {
					p.vx = -8
				} else {
					p.vx = 8
				}
				g.screenShake = 5
				if p.health <= 0 {
					p.lives--
					if p.lives > 0 {
						p.health = p.maxHealth
						p.x = 100
						p.y = 300
						p.vx = 0
						p.vy = 0
					} else {
						g.state = StateGameOver
					}
				}
			}
		}
	}

	// Power-ups
	for _, pu := range g.powerUps {
		if !pu.alive {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, pu.x, pu.y, 32, 32) {
			pu.alive = false
			switch pu.id {
			case 0:
				if p.powerLevel < PowerMushroom {
					p.powerLevel = PowerMushroom
					p.maxHealth++
					p.health = p.maxHealth
				}
			case 1:
				p.powerLevel = PowerStar
				p.powerTimer = 600
				p.hasDoubleJump = true
			case 2:
				p.powerLevel = PowerFlower
			}
			p.score += 500
			g.spawnParticles(pu.x+16, pu.y+16, 20, color.RGBA{255, 255, 255, 255})
		}
	}

	// Снаряды
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		pr := g.projectiles[i]
		if pr.isPlayer {
			// Попадание во врагов
			for _, e := range g.enemies {
				if e.alive && rectOverlap(pr.x, pr.y, float64(pr.size), float64(pr.size), e.x, e.y, 36, 36) {
					e.alive = false
					p.score += 200
					g.spawnParticles(e.x+18, e.y+18, 15, color.RGBA{255, 100, 0, 255})
					pr.life = 0
				}
			}
			// Попадание в босса
			if g.boss != nil && g.boss.alive && rectOverlap(pr.x, pr.y, float64(pr.size), float64(pr.size), g.boss.x, g.boss.y, 80, 80) {
				g.boss.health -= pr.damage
				g.spawnParticles(g.boss.x+40, g.boss.y+40, 10, color.RGBA{255, 100, 100, 255})
				g.screenShake = 3
				pr.life = 0
				if g.boss.health <= 0 {
					g.boss.alive = false
				}
			}
		} else {
			// Вражеский снаряд попал в игрока
			if p.invincible <= 0 && rectOverlap(pr.x, pr.y, float64(pr.size), float64(pr.size), p.x, p.y, p.width, p.height) {
				p.health--
				p.invincible = 120
				p.vy = -6
				g.screenShake = 5
				pr.life = 0
				if p.health <= 0 {
					p.lives--
					if p.lives <= 0 {
						g.state = StateGameOver
					} else {
						p.health = p.maxHealth
						p.x = 100
						p.y = 300
					}
				}
			}
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x:     x,
			y:     y,
			vx:    (rand.Float64() - 0.5) * 8,
			vy:    (rand.Float64() - 0.5) * 8,
			life:  30 + rand.Intn(30),
			color: c,
			size:  4 + rand.Float32()*6,
		})
	}
}

func (g *Game) updateScreenShake() {
	if g.screenShake > 0 {
		g.screenShake *= 0.9
		if g.screenShake < 0.5 {
			g.screenShake = 0
		}
	}
}

func rectOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{100, 149, 237, 255})

	var shakeX, shakeY float64
	if g.screenShake > 0 {
		shakeX = (rand.Float64() - 0.5) * g.screenShake * 2
		shakeY = (rand.Float64() - 0.5) * g.screenShake * 2
	}

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
		return
	case StateGameOver:
		g.drawGame(screen, shakeX, shakeY)
		g.drawGameOver(screen)
		return
	case StateWin:
		g.drawGame(screen, shakeX, shakeY)
		g.drawWin(screen)
		return
	}

	g.drawGame(screen, shakeX, shakeY)
}

func (g *Game) drawGame(screen *ebiten.Image, shakeX, shakeY float64) {
	camX := g.cameraX

	// Тайлы
	for _, t := range g.tiles {
		tx := float64(t.x*TileSize) - camX + shakeX
		ty := float64(t.y*TileSize) + shakeY
		if tx < -TileSize || tx > ScreenWidth {
			continue
		}
		var img *ebiten.Image
		switch t.id {
		case 1:
			img = getSprite("grass")
		case 2:
			img = getSprite("dirt")
		case 3:
			img = getSprite("brick")
		case 4:
			img = getSprite("box")
		case 5:
			if t.opened {
				img = getSprite("box") // opened chest
			} else {
				img = getSprite("chest")
			}
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, ty)
			screen.DrawImage(img, op)
		}
	}

	// Монеты
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		cx := c.x - camX + shakeX
		if cx < -32 || cx > ScreenWidth+32 {
			continue
		}
		img := getSprite("coin")
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cx, c.y+shakeY)
			screen.DrawImage(img, op)
		}
	}

	// Power-ups
	for _, pu := range g.powerUps {
		if !pu.alive {
			continue
		}
		px := pu.x - camX + shakeX
		if px < -32 || px > ScreenWidth+32 {
			continue
		}
		var imgName string
		switch pu.id {
		case 0:
			imgName = "mushroom"
		case 1:
			imgName = "star"
		case 2:
			imgName = "flower"
		}
		img := getSprite(imgName)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(px, pu.y+shakeY)
			screen.DrawImage(img, op)
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		ex := e.x - camX + shakeX
		if ex < -50 || ex > ScreenWidth+50 {
			continue
		}
		var img *ebiten.Image
		if e.enemyType == 0 {
			if (e.animFrame/10)%2 == 0 {
				img = getSprite("slime1")
			} else {
				img = getSprite("slime2")
			}
		} else {
			if (e.animFrame/10)%2 == 0 {
				img = getSprite("fly1")
			} else {
				img = getSprite("fly2")
			}
		}
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(ex, e.y+shakeY)
			if e.vx > 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(36, 0)
			}
			screen.DrawImage(img, op)
		}
	}

	// Босс
	if g.boss != nil && g.boss.alive {
		bx := float32(g.boss.x - camX + shakeX)
		by := float32(g.boss.y + shakeY)
		vector.DrawFilledRect(screen, bx, by, 80, 80, color.RGBA{255, 50, 50, 255}, false)
		vector.DrawFilledRect(screen, bx+15, by+20, 15, 15, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledRect(screen, bx+50, by+20, 15, 15, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledRect(screen, bx+20, by+25, 5, 5, color.RGBA{0, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, bx+55, by+25, 5, 5, color.RGBA{0, 0, 0, 255}, false)
		barW := float32(70 * g.boss.health / g.boss.maxHealth)
		vector.DrawFilledRect(screen, bx+5, by-15, 70, 8, color.RGBA{100, 100, 100, 255}, false)
		vector.DrawFilledRect(screen, bx+5, by-15, barW, 8, color.RGBA{255, 50, 50, 255}, false)
	}

	// Игрок
	p := g.player
	px := p.x - camX + shakeX
	py := p.y + shakeY

	if p.invincible > 0 && (g.frame/4)%2 == 0 {
		// Пропускаем отрисовку
	} else {
		var img *ebiten.Image
		if !p.onGround {
			img = getSprite("player_jump")
		} else if p.vx != 0 {
			// Используем правильно нарезанные кадры ходьбы!
			img = getWalkSprite(p.walkFrame)
		} else {
			img = getSprite("player_stand")
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			if p.facing < 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(p.width, 0)
			}
			op.GeoM.Translate(px, py)
			if p.powerLevel == PowerStar {
				op.ColorM.Scale(1.5, 1.5, 0.5, 1)
			}
			screen.DrawImage(img, op)
		}

		// Рисуем оружие
		if p.weapon != WeaponNone {
			weaponX := float32(px)
			if p.facing > 0 {
				weaponX += float32(p.width)
			} else {
				weaponX -= 10
			}
			weaponY := float32(py + p.height/2)

			var weaponColor color.RGBA
			switch p.weapon {
			case WeaponSword:
				weaponColor = color.RGBA{200, 200, 200, 255}
			case WeaponGun:
				weaponColor = color.RGBA{100, 100, 100, 255}
			case WeaponBlaster:
				weaponColor = color.RGBA{0, 255, 255, 255}
			}
			vector.DrawFilledRect(screen, weaponX, weaponY, 20, 6, weaponColor, false)
		}
	}

	// Снаряды
	for _, pr := range g.projectiles {
		prx := float32(pr.x - camX + shakeX)
		pry := float32(pr.y + shakeY)
		vector.DrawFilledCircle(screen, prx, pry, pr.size, pr.color, true)
	}

	// Частицы
	for _, pt := range g.particles {
		ptx := float32(pt.x - camX + shakeX)
		pty := float32(pt.y + shakeY)
		vector.DrawFilledCircle(screen, ptx, pty, pt.size, pt.color, true)
	}

	// Флаг
	fx := g.flagX - camX + shakeX
	if fx > -50 && fx < ScreenWidth+50 {
		flagColor := color.RGBA{0, 255, 0, 255}
		vector.DrawFilledRect(screen, float32(fx), float32(g.flagX+shakeY)-200, 5, 200, color.RGBA{139, 90, 43, 255}, false)
		vector.DrawFilledRect(screen, float32(fx)+5, float32(g.flagX+shakeY)-200, 40, 30, flagColor, false)
	}

	g.drawUI(screen)

	if g.comboText > 0 && p.combo > 1 {
		comboStr := fmt.Sprintf("COMBO x%d!", p.combo)
		text.Draw(screen, comboStr, basicfont.Face7x13, ScreenWidth/2-30, 200, color.RGBA{255, 215, 0, 255})
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 45, color.RGBA{0, 0, 0, 200}, false)
	p := g.player
	text.Draw(screen, fmt.Sprintf("SCORE: %06d", p.score), basicfont.Face7x13, 15, 30, color.White)
	text.Draw(screen, fmt.Sprintf("COINS: %d", p.coins), basicfont.Face7x13, 250, 30, color.RGBA{255, 215, 0, 255})
	hearts := ""
	for i := 0; i < p.health; i++ {
		hearts += "❤"
	}
	text.Draw(screen, "LIVES: "+hearts, basicfont.Face7x13, 450, 30, color.RGBA{255, 100, 100, 255})
	text.Draw(screen, fmt.Sprintf("WORLD %d-%d", g.world, g.level), basicfont.Face7x13, 700, 30, color.White)

	weaponStr := ""
	switch p.weapon {
	case WeaponNone:
		weaponStr = ""
	case WeaponSword:
		weaponStr = "⚔️ SWORD"
	case WeaponGun:
		weaponStr = "🔫 GUN"
	case WeaponBlaster:
		weaponStr = "🔬 BLASTER"
	}
	if weaponStr != "" {
		text.Draw(screen, weaponStr, basicfont.Face7x13, 850, 30, color.RGBA{255, 255, 100, 255})
	}

	if p.powerLevel > 0 {
		powerStr := ""
		switch p.powerLevel {
		case PowerMushroom:
			powerStr = "🍄"
		case PowerStar:
			powerStr = "⭐"
		case PowerFlower:
			powerStr = "🌸"
		}
		text.Draw(screen, powerStr, basicfont.Face7x13, 980, 30, color.RGBA{255, 255, 100, 255})
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	text.Draw(screen, "SUPER GO MARIO", basicfont.Face7x13, ScreenWidth/2-70, 150, color.RGBA{255, 215, 0, 255})
	text.Draw(screen, "v3.2.0 - Weapons Edition", basicfont.Face7x13, ScreenWidth/2-75, 190, color.White)
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-65, 300, color.White)
	controls := []string{
		"CONTROLS:",
		"Arrow Keys / WASD - Move",
		"Space / W - Jump (double with Star)",
		"Z - Attack with weapon",
		"X - Shoot fireball (Flower)",
		"ESC - Pause",
	}
	for i, line := range controls {
		text.Draw(screen, line, basicfont.Face7x13, ScreenWidth/2-80, 380+i*25, color.White)
	}
	text.Draw(screen, fmt.Sprintf("HIGH SCORE: %06d", g.highScore), basicfont.Face7x13, ScreenWidth/2-70, 580, color.RGBA{255, 215, 0, 255})
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 200}, false)
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, ScreenWidth/2-45, 300, color.RGBA{255, 50, 50, 255})
	text.Draw(screen, fmt.Sprintf("Final Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-60, 400, color.White)
	text.Draw(screen, "Press ENTER to restart", basicfont.Face7x13, ScreenWidth/2-65, 500, color.White)
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 200}, false)
	text.Draw(screen, "YOU WIN!", basicfont.Face7x13, ScreenWidth/2-40, 300, color.RGBA{255, 215, 0, 255})
	text.Draw(screen, fmt.Sprintf("Final Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-60, 400, color.White)
	text.Draw(screen, "Press ENTER to continue", basicfont.Face7x13, ScreenWidth/2-70, 500, color.White)
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func (g *Game) saveHighScore() {
	if g.player.score > g.highScore {
		g.highScore = g.player.score
	}
}

func (g *Game) loadHighScore() {
	g.highScore = 0
}

func main() {
	log.Println("🍄 SUPER GO MARIO v3.2.0 - Weapons Edition")
	log.Println("🎮 Загрузка спрайтов...")
	initSpriteCache()
	log.Println("✅ Спрайты загружены!")
	log.Println("🎮 Управление: WASD - движение, Пробел - прыжок, Z - оружие, X - огонь")
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🍄 SUPER GO MARIO v3.2.0 | Go365 Day 87")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
