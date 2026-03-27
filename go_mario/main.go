// Go365 Day 87 - SUPER GO MARIO v3.0.0
// Крутой 2D-платформер с боссами, power-ups, частицами и звуками!

package main

import (
	"fmt"
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

// Игрок
type Player struct {
	x, y         float64
	vx, vy       float64
	width        float64
	height       float64
	onGround     bool
	facing       int
	animFrame    int
	animTimer    int
	health       int
	maxHealth    int
	powerLevel   int
	powerTimer   int
	invincible   int
	coins        int
	score        int
	lives        int
	combo        int
	comboTimer   int
	maxCombo     int
	fireCooldown int
}

// Тайл
type Tile struct {
	x, y int
	id   int
}

// Враг
type Enemy struct {
	x, y      float64
	vx, vy    float64
	enemyType int // 0=slime, 1=fly, 2=boss
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
	id        int // 0=mushroom, 1=star, 2=flower
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
	player     *Player
	tiles      []*Tile
	enemies    []*Enemy
	coins      []*Coin
	particles  []*Particle
	powerUps   []*PowerUp
	boss       *Boss
	cameraX    float64
	levelW     int
	levelH     int
	frame      int
	state      int
	world      int
	level      int
	flagX      float64
	screenShake float64
	comboText  int
	highScore  int
}

// ============================================================================
// ИНИЦИАЛИЗАЦИЯ
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		player: &Player{
			x:         100,
			y:         300,
			width:     40,
			height:     48,
			facing:    1,
			maxHealth: 3,
			health:    3,
			lives:     3,
		},
		state:   StateMenu,
		world:   1,
		level:   1,
		levelH:  16,
	}

	g.loadHighScore()
	g.startLevel()
	return g
}

func (g *Game) startLevel() {
	g.levelW = 80 + g.world*20
	g.tiles = nil
	g.enemies = nil
	g.coins = nil
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
		// Пропуски для ям
		if x > 15 && x < g.levelW-10 && x%30 >= 25 && x%30 < 28 {
			continue
		}

		g.tiles = append(g.tiles, &Tile{x: x, y: 13, id: 1}) // grass
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: 2}) // dirt
		g.tiles = append(g.tiles, &Tile{x: x, y: 15, id: 2}) // dirt
	}

	// Платформы
	for x := 10; x < g.levelW-15; x++ {
		if x%12 < 4 && x%20 != 0 {
			platY := 7 + rand.Intn(3)
			for i := 0; i < 3+rand.Intn(2); i++ {
				tileID := 3 // brick
				if rand.Float32() < 0.3 {
					tileID = 4 // box
				}
				g.tiles = append(g.tiles, &Tile{x: x + i, y: platY, id: tileID})
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
			enemyY := float64(12*TileSize)
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
				y:     float64(10*TileSize),
				id:    rand.Intn(3),
				alive: true,
			})
		}
	}

	// Флаг
	g.flagX = float64((g.levelW - 8) * TileSize)

	// Босс в конце каждого 3 уровня
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
				g.startLevel()
				g.state = StatePlaying
			}
		}
		return nil

	case StateBoss:
		g.updatePlayer()
		g.updateCamera()
		g.updateBoss()
		g.updateParticles()
		if g.boss == nil || !g.boss.alive {
			g.player.score += 1000
			g.spawnParticles(g.boss.x, g.boss.y, 50, color.RGBA{255, 100, 100, 255})
			g.boss = nil
			g.level++
			g.startLevel()
			g.state = StatePlaying
		}
		return nil
	}

	// Playing
	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.updateCoins()
	g.updatePowerUps()
	g.updateParticles()
	g.checkCollisions()
	g.updateScreenShake()

	// Проверка победы
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

	// Управление
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = Speed
		p.facing = 1
		p.animFrame++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -Speed
		p.facing = -1
		p.animFrame++
	} else {
		p.vx = 0
	}

	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
		g.spawnParticles(p.x+p.width/2, p.y+p.height, 10, color.RGBA{255, 255, 255, 255})
	}

	// Стрельба (если есть flower power)
	if inpututil.IsKeyJustPressed(ebiten.KeyZ) && p.powerLevel == PowerFlower && p.fireCooldown <= 0 {
		p.fireCooldown = 30
		// Создаём огненный шар
		g.particles = append(g.particles, &Particle{
			x:     p.x + p.width,
			y:     p.y + p.height/2,
			vx:    float64(p.facing) * 8,
			vy:    0,
			life:  60,
			color: color.RGBA{255, 100, 0, 255},
			size:  12,
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

	// Коллизии с тайлами
	p.onGround = false
	for _, t := range g.tiles {
		tx := float64(t.x * TileSize)
		ty := float64(t.y * TileSize)

		if p.x < tx+TileSize-5 && p.x+p.width > tx+5 &&
			p.y < ty+TileSize && p.y+p.height > ty {

			if p.vy >= 0 && p.y+p.height <= ty+25 {
				p.y = ty - p.height
				p.vy = 0
				p.onGround = true
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

	// Смерть от падения
	if p.y > ScreenHeight+100 {
		p.health = 0
	}

	// Обновление power timer
	if p.powerTimer > 0 {
		p.powerTimer--
		if p.powerTimer <= 0 && p.powerLevel == PowerStar {
			p.powerLevel = PowerNone
		}
	}

	// Неуязвимость
	if p.invincible > 0 {
		p.invincible--
	}

	// Combo
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

		// Разворот случайно
		if rand.Float32() < 0.02 {
			e.vx *= -1
		}

		// Гравитация для летающих
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

	// Движение к игроку
	if p.x > b.x {
		b.vx = 2
	} else {
		b.vx = -2
	}

	// Прыжок
	if rand.Float32() < 0.03 && b.y >= 400 {
		b.vy = -12
	}

	// Гравитация
	b.vy += 0.5
	if b.vy > 10 {
		b.vy = 10
	}

	b.x += b.vx
	b.y += b.vy

	// Пол
	if b.y > 400 {
		b.y = 400
		b.vy = 0
	}

	// Атаки
	b.attackTimer--
	if b.attackTimer <= 0 {
		b.attackTimer = 40
		// Создаём снаряд
		angle := -0.5
		g.particles = append(g.particles, &Particle{
			x:     b.x + 40,
			y:     b.y + 40,
			vx:    float64(p.facing) * 5,
			vy:    angle,
			life:  90,
			color: color.RGBA{255, 50, 50, 255},
			size:  15,
		})
	}

	// Коллизия с игроком
	if rectOverlap(p.x, p.y, p.width, p.height, b.x, b.y, 80, 80) {
		if p.vy > 0 && p.y+p.height < b.y+40 {
			b.health -= 1
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

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, e.x, e.y, 36, 36) {
			if p.vy > 0 && p.y+p.height < e.y+20 {
				// Прыгнул сверху
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
				// Получил урон
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
			case 0: // Mushroom
				if p.powerLevel < PowerMushroom {
					p.powerLevel = PowerMushroom
					p.maxHealth++
					p.health = p.maxHealth
				}
			case 1: // Star
				p.powerLevel = PowerStar
				p.powerTimer = 600
			case 2: // Flower
				p.powerLevel = PowerFlower
			}
			p.score += 500
			g.spawnParticles(pu.x+16, pu.y+16, 20, color.RGBA{255, 255, 255, 255})
		}
	}

	// Частицы-снаряды (огненные шары)
	for i := len(g.particles) - 1; i >= 0; i-- {
		pt := g.particles[i]
		if pt.color.R == 255 && pt.color.G == 100 && pt.color.B == 0 {
			// Проверяем попадание во врагов
			for _, e := range g.enemies {
				if e.alive && rectOverlap(pt.x, pt.y, float64(pt.size), float64(pt.size), e.x, e.y, 36, 36) {
					e.alive = false
					p.score += 200
					g.spawnParticles(e.x+18, e.y+18, 15, color.RGBA{255, 100, 0, 255})
					pt.life = 0
				}
			}
			// Проверяем попадание в босса
			if g.boss != nil && g.boss.alive && rectOverlap(pt.x, pt.y, float64(pt.size), float64(pt.size), g.boss.x, g.boss.y, 80, 80) {
				g.boss.health--
				g.spawnParticles(g.boss.x+40, g.boss.y+40, 10, color.RGBA{255, 100, 100, 255})
				g.screenShake = 3
				pt.life = 0
				if g.boss.health <= 0 {
					g.boss.alive = false
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
	// Очистка
	screen.Fill(color.RGBA{100, 149, 237, 255})

	// Screen shake
	if g.screenShake > 0 {
		dx := (rand.Float64() - 0.5) * g.screenShake * 2
		dy := (rand.Float64() - 0.5) * g.screenShake * 2
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(dx, dy)
		screen = screen.SubImage(screen.Bounds()).(*ebiten.Image)
		_ = op
	}

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
		return
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
		return
	case StateWin:
		g.drawGame(screen)
		g.drawWin(screen)
		return
	}

	g.drawGame(screen)
}

func (g *Game) drawGame(screen *ebiten.Image) {
	camX := g.cameraX

	// Тайлы
	for _, t := range g.tiles {
		tx := float64(t.x*TileSize) - camX
		ty := float64(t.y * TileSize)

		if tx < -TileSize || tx > ScreenWidth {
			continue
		}

		var img *ebiten.Image
		switch t.id {
		case 1:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
		case 2:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
		case 3:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")
		case 4:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")
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
		cx := c.x - camX
		if cx < -32 || cx > ScreenWidth {
			continue
		}

		img, _, _ := ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cx, c.y)
			screen.DrawImage(img, op)
		}
	}

	// Power-ups
	for _, pu := range g.powerUps {
		if !pu.alive {
			continue
		}
		px := pu.x - camX
		if px < -32 || px > ScreenWidth {
			continue
		}

		var imgName string
		switch pu.id {
		case 0:
			imgName = "assets/sprites/items/mushroomRed.png"
		case 1:
			imgName = "assets/sprites/items/star.png"
		case 2:
			imgName = "assets/sprites/items/plantPurple.png"
		}

		img, _, _ := ebitenutil.NewImageFromFile(imgName)
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(px, pu.y)
			screen.DrawImage(img, op)
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		ex := e.x - camX
		if ex < -50 || ex > ScreenWidth+50 {
			continue
		}

		var img *ebiten.Image
		if e.enemyType == 0 {
			if (e.animFrame/10)%2 == 0 {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
			} else {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")
			}
		} else {
			if (e.animFrame/10)%2 == 0 {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly1.png")
			} else {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly2.png")
			}
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(ex, e.y)
			if e.vx > 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(36, 0)
			}
			screen.DrawImage(img, op)
		}
	}

	// Босс
	if g.boss != nil && g.boss.alive {
		bx := g.boss.x - camX
		// Рисуем босса как красный квадрат с глазами
		vector.DrawFilledRect(screen, float32(bx), float32(g.boss.y), 80, 80, color.RGBA{255, 50, 50, 255}, false)
		vector.DrawFilledRect(screen, float32(bx)+15, float32(g.boss.y)+20, 15, 15, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledRect(screen, float32(bx)+50, float32(g.boss.y)+20, 15, 15, color.RGBA{255, 255, 255, 255}, false)
		vector.DrawFilledRect(screen, float32(bx)+20, float32(g.boss.y)+25, 5, 5, color.RGBA{0, 0, 0, 255}, false)
		vector.DrawFilledRect(screen, float32(bx)+55, float32(g.boss.y)+25, 5, 5, color.RGBA{0, 0, 0, 255}, false)

		// Health bar
		barW := float32(70 * g.boss.health / g.boss.maxHealth)
		vector.DrawFilledRect(screen, float32(bx)+5, float32(g.boss.y)-15, 70, 8, color.RGBA{100, 100, 100, 255}, false)
		vector.DrawFilledRect(screen, float32(bx)+5, float32(g.boss.y)-15, barW, 8, color.RGBA{255, 50, 50, 255}, false)
	}

	// Игрок
	p := g.player
	px := p.x - camX

	// Мигание при неуязвимости
	if p.invincible > 0 && (g.frame/4)%2 == 0 {
		// Пропускаем отрисовку
	} else {
		var img *ebiten.Image
		if !p.onGround {
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")
		} else if p.vx != 0 {
			if (p.animFrame/8)%2 == 0 {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
			} else {
				img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_walk/p1_walk.png")
			}
		} else {
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			if p.facing < 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(p.width, 0)
			}
			op.GeoM.Translate(px, p.y)

			// Звёздный эффект
			if p.powerLevel == PowerStar {
				op.ColorM.Scale(1.5, 1.5, 0.5, 1)
			}

			screen.DrawImage(img, op)
		}
	}

	// Частицы
	for _, pt := range g.particles {
		ptx := float32(pt.x - camX)
		pty := float32(pt.y)
		vector.DrawFilledCircle(screen, ptx, pty, pt.size, pt.color, true)
	}

	// Флаг
	fx := g.flagX - camX
	if fx > -50 && fx < ScreenWidth+50 {
		flagColor := color.RGBA{0, 255, 0, 255}
		vector.DrawFilledRect(screen, float32(fx), float32(g.flagX)-200, 5, 200, color.RGBA{139, 90, 43, 255}, false)
		vector.DrawFilledRect(screen, float32(fx)+5, float32(g.flagX)-200, 40, 30, flagColor, false)
	}

	// UI
	g.drawUI(screen)

	// Combo текст
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

	// Hearts
	hearts := ""
	for i := 0; i < p.health; i++ {
		hearts += "❤"
	}
	text.Draw(screen, "LIVES: "+hearts, basicfont.Face7x13, 450, 30, color.RGBA{255, 100, 100, 255})

	text.Draw(screen, fmt.Sprintf("WORLD %d-%d", g.world, g.level), basicfont.Face7x13, 700, 30, color.White)

	// Power level
	if p.powerLevel > 0 {
		powerStr := ""
		switch p.powerLevel {
		case PowerMushroom:
			powerStr = "POWER: 🍄"
		case PowerStar:
			powerStr = "POWER: ⭐"
		case PowerFlower:
			powerStr = "POWER: 🌸"
		}
		text.Draw(screen, powerStr, basicfont.Face7x13, 900, 30, color.RGBA{255, 255, 100, 255})
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Заголовок
	text.Draw(screen, "SUPER GO MARIO", basicfont.Face7x13, ScreenWidth/2-70, 150, color.RGBA{255, 215, 0, 255})

	// Подзаголовок
	text.Draw(screen, "v3.0.0 - Ultimate Edition", basicfont.Face7x13, ScreenWidth/2-80, 190, color.White)

	// Инструкция
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-65, 300, color.White)

	// Управление
	controls := []string{
		"CONTROLS:",
		"Arrow Keys / WASD - Move",
		"Space / W / Up - Jump",
		"Z - Shoot fireball (when powered)",
		"ESC - Pause",
	}
	for i, line := range controls {
		text.Draw(screen, line, basicfont.Face7x13, ScreenWidth/2-80, 380+i*25, color.White)
	}

	// High score
	text.Draw(screen, fmt.Sprintf("HIGH SCORE: %06d", g.highScore), basicfont.Face7x13, ScreenWidth/2-70, 550, color.RGBA{255, 215, 0, 255})
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

// ============================================================================
// SAVE SYSTEM
// ============================================================================

func (g *Game) saveHighScore() {
	if g.player.score > g.highScore {
		g.highScore = g.player.score
		// В реальной игре нужно сохранять в файл
	}
}

func (g *Game) loadHighScore() {
	// В реальной игре нужно загружать из файла
	g.highScore = 0
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	log.Println("🍄 SUPER GO MARIO v3.0.0 - Запуск...")
	log.Println("🎮 Управление: WASD/Стрелки - движение, Пробел - прыжок, Z - огонь")

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🍄 SUPER GO MARIO v3.0.0 | Go365 Day 87")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
