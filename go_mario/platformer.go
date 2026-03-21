package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	TileSize     = 40

	// Physics
	Gravity       = 0.5
	JumpForce     = -11.0
	RunSpeed      = 5.0
	WalkSpeed     = 2.5
	MaxFallSpeed  = 12.0
	Friction      = 0.8
	Acceleration  = 0.5

	// Tile types
	TileAir         = 0
	TileGround      = 1
	TileBrick       = 2
	TileQuestion    = 3
	TileHard        = 4
	TilePipeL       = 5
	TilePipeR       = 6
	TilePipeTopL    = 7
	TilePipeTopR    = 8
	TileCoin        = 9
	TileUsed        = 10

	// Enemy types
	EnemyGoomba  = 1
	EnemyKoopa   = 2
	EnemyPiranha = 3

	// Powerup types
	PowerupMushroom = 1
	PowerupFlower   = 2
	PowerupStar     = 3
	Powerup1UP      = 4

	// Game states
	StateMenu     = 0
	StatePlaying  = 1
	StateGameOver = 2
	StateWon      = 3
)

// ============================================================================
// GAME STRUCTURES
// ============================================================================

// Player - наш герой (Mario-style)
type Player struct {
	x, y        float64
	vx, vy      float64
	width       float32
	height      float32
	onGround    bool
	facing      int
	animFrame   int
	animTimer   int

	// Stats
	coins       int
	score       int
	lives       int
	world       int

	// Power state
	isBig       bool
	isFire      bool
	isInvincible bool
	powerTimer  int
}

// Enemy - враг
type Enemy struct {
	x, y      float64
	vx, vy    float64
	width     float32
	height    float32
	enemyType int
	alive     bool
	squashed  bool
	animFrame int
	facing    int
}

// Powerup - бонус
type Powerup struct {
	x, y      float64
	vy        float64
	width     float32
	height    float32
	powerType int
	alive     bool
	animFrame int
}

// Particle - частица
type Particle struct {
	x, y    float64
	vx, vy  float64
	life    int
	color   color.RGBA
	size    float32
}

// Level - уровень
type Level struct {
	width   int
	height  int
	tiles   [][]int
	coins   []Coin
	enemies []*Enemy
	powerups []*Powerup
	flagX   int
	flagY   int
}

// Coin - монета
type Coin struct {
	x, y      float64
	collected bool
	animFrame int
}

// Camera - камера
type Camera struct {
	x, y float64
}

// Game - основная игра
type Game struct {
	player     *Player
	level      *Level
	camera     *Camera
	particles  []*Particle
	state      int
	frameCount int

	// Audio
	audioEnabled bool
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		player: &Player{
			x:      100,
			y:      100,
			width:  30,
			height: 40,
			facing: 1,
			lives:  3,
		},
		camera:     &Camera{},
		state:      StateMenu,
		frameCount: 0,
		audioEnabled: true,
	}

	g.LoadLevel(1)
	return g
}

// LoadLevel загружает уровень
func (g *Game) LoadLevel(world int) {
	g.player.world = world
	g.level = GenerateLevel(world)
	g.player.x = 100
	g.player.y = 100
	g.player.vx = 0
	g.player.vy = 0
	g.camera.x = 0
	g.particles = make([]*Particle, 0)
}

// GenerateLevel генерирует уровень
func GenerateLevel(world int) *Level {
	width := 200  // tiles
	height := 15  // tiles

	level := &Level{
		width:  width,
		height: height,
		tiles:  make([][]int, width),
		coins:  make([]Coin, 0),
		enemies: make([]*Enemy, 0),
		powerups: make([]*Powerup, 0),
	}

	// Initialize tiles
	for x := range level.tiles {
		level.tiles[x] = make([]int, height)
	}

	// Generate terrain
	for x := 0; x < width; x++ {
		// Ground
		for y := 10; y < height; y++ {
			level.tiles[x][y] = TileGround
		}

		// Gaps (pipes over pits)
		if x%50 == 45 && x > 50 {
			for y := 10; y < height; y++ {
				level.tiles[x][y] = TileAir
				if x+1 < width {
					level.tiles[x+1][y] = TileAir
				}
			}
		}

		// Random structures
		if x > 10 && rand.Float32() < 0.1 {
			structureType := rand.Intn(5)

			switch structureType {
			case 0: // Brick platform
				platY := rand.Intn(3) + 5
				for bx := 0; bx < 5; bx++ {
					if x+bx < width {
						level.tiles[x+bx][platY] = TileBrick
					}
				}
				// Add coin above
				if x+2 < width {
					level.coins = append(level.coins, Coin{
						x: float64((x+2)*TileSize),
						y: float64((platY-1)*TileSize),
					})
				}

			case 1: // Question block
				if x < width && rand.Intn(8) > 3 {
					level.tiles[x][rand.Intn(3)+5] = TileQuestion
				}

			case 2: // Pipe
				pipeHeight := rand.Intn(3) + 2
				pipeY := 10 - pipeHeight
				for py := pipeY; py < 10; py++ {
					if x < width {
						level.tiles[x][py] = TilePipeL
						if x+1 < width {
							level.tiles[x+1][py] = TilePipeR
						}
					}
				}
				// Pipe top
				if pipeY > 0 && x < width {
					level.tiles[x][pipeY-1] = TilePipeTopL
					if x+1 < width {
						level.tiles[x+1][pipeY-1] = TilePipeTopR
					}
				}

				// Piranha plant chance
				if rand.Float32() < 0.3 {
					level.enemies = append(level.enemies, &Enemy{
						x: float64(x * TileSize),
						y: float64((pipeY - 2) * TileSize),
						width: 30,
						height: 30,
						enemyType: EnemyPiranha,
						alive: true,
					})
				}

			case 3: // Enemy spawn
				enemyType := EnemyGoomba
				if world > 1 && rand.Float32() < 0.3 {
					enemyType = EnemyKoopa
				}
				level.enemies = append(level.enemies, &Enemy{
					x: float64(x * TileSize),
					y: float64(8 * TileSize),
					width: 32,
					height: 32,
					enemyType: enemyType,
					alive: true,
					facing: -1,
				})

			case 4: // Stairs
				stairHeight := rand.Intn(4) + 2
				for sy := 0; sy < stairHeight; sy++ {
					for sx := 0; sx <= sy; sx++ {
						if x+sx < width {
							level.tiles[x+sx][9-sy] = TileHard
						}
					}
				}
			}
		}
	}

	// Add flag at end
	level.flagX = (width - 5) * TileSize
	level.flagY = 6 * TileSize

	// Add coins along the level
	for i := 0; i < 100; i++ {
		cx := rand.Intn(width-10) + 5
		cy := rand.Intn(8) + 2
		if level.tiles[cx][cy+1] != TileAir {
			level.coins = append(level.coins, Coin{
				x: float64(cx * TileSize),
				y: float64(cy * TileSize),
			})
		}
	}

	return level
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frameCount++

	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StateGameOver, StateWon:
		g.updateEndScreen()
	}

	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StatePlaying
		g.player.lives = 3
		g.player.score = 0
		g.player.coins = 0
		g.LoadLevel(1)
		playSound(SoundStart)
	}
}

func (g *Game) updatePlaying() {
	// Player input
	g.updatePlayer()

	// Update camera
	g.camera.x = g.player.x - ScreenWidth/2
	if g.camera.x < 0 {
		g.camera.x = 0
	}
	if g.camera.x > float64(g.level.width*TileSize-ScreenWidth) {
		g.camera.x = float64(g.level.width*TileSize - ScreenWidth)
	}

	// Update enemies
	g.updateEnemies()

	// Update powerups
	g.updatePowerups()

	// Update particles
	g.updateParticles()

	// Check win condition
	if g.player.x >= float64(g.level.flagX) {
		g.state = StateWon
		playSound(SoundWin)
	}

	// Check death
	if g.player.y > ScreenHeight {
		g.playerDie()
	}
}

func (g *Game) updatePlayer() {
	p := g.player

	// Horizontal movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		if p.vx < RunSpeed {
			p.vx += Acceleration
		}
		p.facing = 1
		p.animFrame++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		if p.vx > -RunSpeed {
			p.vx -= Acceleration
		}
		p.facing = -1
		p.animFrame++
	} else {
		// Friction
		p.vx *= Friction
		if math.Abs(p.vx) < 0.1 {
			p.vx = 0
		}
	}

	// Jump
	if (ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeySpace)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
		playSound(SoundJump)
		g.spawnParticles(p.x+float64(p.width/2), p.y+float64(p.height), 5, color.RGBA{200, 200, 200, 255})
	}

	// Variable jump height
	if !ebiten.IsKeyPressed(ebiten.KeyArrowUp) && !ebiten.IsKeyPressed(ebiten.KeyW) && !ebiten.IsKeyPressed(ebiten.KeySpace) && p.vy < JumpForce/2 {
		p.vy *= 0.5
	}

	// Apply gravity
	p.vy += Gravity
	if p.vy > MaxFallSpeed {
		p.vy = MaxFallSpeed
	}

	// Move and collide
	p.x += p.vx
	g.collideHorizontal(p)

	p.y += p.vy
	g.collideVertical(p)

	// World bounds
	if p.x < 0 {
		p.x = 0
		p.vx = 0
	}

	// Animation timer
	if p.animFrame > 1000 {
		p.animFrame = 0
	}

	// Invincibility timer
	if p.isInvincible {
		p.powerTimer--
		if p.powerTimer <= 0 {
			p.isInvincible = false
		}
	}
}

func (g *Game) collideHorizontal(p *Player) {
	leftTile := int(p.x) / TileSize
	rightTile := int(p.x+float64(p.width)) / TileSize
	topTile := int(p.y) / TileSize
	bottomTile := int(p.y+float64(p.height)-1) / TileSize

	// Check left
	if p.vx < 0 {
		if g.isSolid(leftTile, topTile) || g.isSolid(leftTile, bottomTile) {
			p.x = float64((leftTile+1)*TileSize)
			p.vx = 0
		}
	}

	// Check right
	if p.vx > 0 {
		if g.isSolid(rightTile, topTile) || g.isSolid(rightTile, bottomTile) {
			p.x = float64(rightTile*TileSize - int(p.width))
			p.vx = 0
		}
	}
}

func (g *Game) collideVertical(p *Player) {
	p.onGround = false

	leftTile := int(p.x) / TileSize
	rightTile := int(p.x+float64(p.width)) / TileSize

	// Check falling
	if p.vy > 0 {
		bottomTile := int(p.y+float64(p.height)) / TileSize

		if g.isSolid(leftTile, bottomTile) || g.isSolid(rightTile, bottomTile) {
			p.y = float64(bottomTile*TileSize - int(p.height))
			p.vy = 0
			p.onGround = true
		}
	}

	// Check jumping
	if p.vy < 0 {
		topTile := int(p.y) / TileSize

		if g.isSolid(leftTile, topTile) || g.isSolid(rightTile, topTile) {
			p.y = float64((topTile+1)*TileSize)
			p.vy = 0

			// Hit block
			g.hitBlock(leftTile, topTile)
		}
	}
}

func (g *Game) isSolid(x, y int) bool {
	if x < 0 || x >= g.level.width || y < 0 || y >= g.level.height {
		return false
	}
	tile := g.level.tiles[x][y]
	return tile != TileAir && tile != TileCoin && tile != TileQuestion && tile != TileUsed
}

func (g *Game) hitBlock(x, y int) {
	if x < 0 || x >= g.level.width || y < 0 || y >= g.level.height {
		return
	}

	tile := g.level.tiles[x][y]

	if tile == TileQuestion {
		g.level.tiles[x][y] = TileUsed
		g.player.coins++
		g.player.score += 200
		playSound(SoundCoin)
		g.spawnParticles(float64(x*TileSize+TileSize/2), float64(y*TileSize), 10, color.RGBA{255, 215, 0, 255})

		// Chance for powerup
		if rand.Float32() < 0.1 {
			powerType := PowerupMushroom
			if g.player.isBig {
				powerType = PowerupFlower
			}
			g.level.powerups = append(g.level.powerups, &Powerup{
				x: float64(x * TileSize),
				y: float64((y - 1) * TileSize),
				powerType: powerType,
				alive: true,
				width: 30,
				height: 30,
			})
		}
	} else if tile == TileBrick {
		if g.player.isBig {
			g.level.tiles[x][y] = TileAir
			g.player.score += 50
			playSound(SoundBreak)
			g.spawnParticles(float64(x*TileSize+TileSize/2), float64(y*TileSize+TileSize/2), 8, color.RGBA{139, 69, 19, 255})
		} else {
			playSound(SoundBump)
		}
	}
}

func (g *Game) updateEnemies() {
	for _, enemy := range g.level.enemies {
		if !enemy.alive || enemy.squashed {
			continue
		}

		enemy.animFrame++

		// Simple AI
		if enemy.enemyType == EnemyPiranha {
			// Move up and down
			enemy.y += math.Sin(float64(enemy.animFrame)*0.05) * 0.5
		} else {
			// Walk
			enemy.x += float64(enemy.facing) * 0.5

			// Turn at edges or walls
			leftTile := int(enemy.x) / TileSize
			rightTile := int(enemy.x+float64(enemy.width)) / TileSize
			bottomTile := int(enemy.y+float64(enemy.height)+1) / TileSize

			if enemy.facing < 0 && (!g.isSolid(leftTile, bottomTile) || g.isSolid(leftTile, int(enemy.y)/TileSize)) {
				enemy.facing = 1
			} else if enemy.facing > 0 && (!g.isSolid(rightTile, bottomTile) || g.isSolid(rightTile, int(enemy.y)/TileSize)) {
				enemy.facing = -1
			}
		}

		// Collision with player
		if g.checkCollision(g.player, enemy) {
			if enemy.enemyType == EnemyPiranha {
				g.playerHit()
			} else if g.player.vy > 0 && g.player.y+float64(g.player.height) < enemy.y+float64(enemy.height)/2 {
				// Stomp enemy
				enemy.squashed = true
				g.player.vy = -6
				g.player.score += 100
				playSound(SoundStomp)
				g.spawnParticles(enemy.x+float64(enemy.width/2), enemy.y+float64(enemy.height/2), 15, color.RGBA{139, 69, 19, 255})
			} else if !g.player.isInvincible {
				g.playerHit()
			}
		}
	}

	// Remove squashed enemies
	activeEnemies := make([]*Enemy, 0)
	for _, e := range g.level.enemies {
		if e.alive && !e.squashed {
			activeEnemies = append(activeEnemies, e)
		}
	}
	g.level.enemies = activeEnemies
}

func (g *Game) updatePowerups() {
	for _, p := range g.level.powerups {
		if !p.alive {
			continue
		}

		p.animFrame++
		p.vy += Gravity
		p.y += p.vy

		// Ground collision
		bottomTile := int(p.y+float64(p.height)) / TileSize
		leftTile := int(p.x) / TileSize
		rightTile := int(p.x+float64(p.width)) / TileSize

		if g.isSolid(leftTile, bottomTile) || g.isSolid(rightTile, bottomTile) {
			p.y = float64(bottomTile*TileSize - int(p.height))
			p.vy = 0
		}

		// Collision with player
		if g.checkPlayerPowerup(p) {
			p.alive = false
			g.applyPowerup(p.powerType)
		}
	}
}

func (g *Game) checkPlayerPowerup(p *Powerup) bool {
	return g.player.x < p.x+float64(p.width) &&
		g.player.x+float64(g.player.width) > p.x &&
		g.player.y < p.y+float64(p.height) &&
		g.player.y+float64(g.player.height) > p.y
}

func (g *Game) applyPowerup(powerType int) {
	playSound(SoundPowerup)

	switch powerType {
	case PowerupMushroom:
		g.player.isBig = true
		g.player.height = 50
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{220, 20, 60, 255})

	case PowerupFlower:
		g.player.isFire = true
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{255, 100, 0, 255})

	case PowerupStar:
		g.player.isInvincible = true
		g.player.powerTimer = 600 // 10 seconds
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 30, color.RGBA{255, 215, 0, 255})

	case Powerup1UP:
		g.player.lives++
		g.player.score += 1000
		g.spawnParticles(g.player.x+float64(g.player.width/2), g.player.y+float64(g.player.height), 20, color.RGBA{0, 255, 0, 255})
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.2
		p.life--

		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x: x,
			y: y,
			vx: float64(rand.Intn(10)-5) * 0.5,
			vy: float64(rand.Intn(10)-5) * 0.5,
			life: 30 + rand.Intn(20),
			color: c,
			size: float32(rand.Intn(4)+2),
		})
	}
}

func (g *Game) checkCollision(p *Player, e *Enemy) bool {
	return p.x < e.x+float64(e.width) &&
		p.x+float64(p.width) > e.x &&
		p.y < e.y+float64(e.height) &&
		p.y+float64(p.height) > e.y
}

func (g *Game) playerHit() {
	if g.player.isInvincible {
		return
	}

	if g.player.isBig {
		g.player.isBig = false
		g.player.height = 40
		g.player.isInvincible = true
		g.player.powerTimer = 120
		playSound(SoundHit)
	} else {
		g.playerDie()
	}
}

func (g *Game) playerDie() {
	g.player.lives--
	playSound(SoundDie)

	if g.player.lives <= 0 {
		g.state = StateGameOver
	} else {
		g.LoadLevel(g.player.world)
	}
}

func (g *Game) updateEndScreen() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StatePlaying
		g.player.lives = 3
		g.player.score = 0
		g.player.coins = 0
		g.LoadLevel(1)
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateWon:
		g.drawWon(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Sky
	screen.Fill(color.RGBA{100, 150, 200, 255})

	// Title
	title := "SUPER GO MARIO"
	titleX := ScreenWidth/2 - len(title)*10
	ebitenutil.DebugPrintAt(screen, title, titleX, 150)

	// Subtitle
	subtitle := "A Classic 2D Platformer"
	subX := ScreenWidth/2 - len(subtitle)*6
	ebitenutil.DebugPrintAt(screen, subtitle, subX, 200)

	// Instructions
	instructions := []string{
		"Arrow Keys / WASD - Move",
		"Space / W / Up - Jump",
		"Stomp enemies, collect coins, reach the flag!",
		"",
		"Press ENTER or SPACE to Start",
	}

	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, ScreenWidth/2-len(line)*6, 300+i*25)
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Sky
	screen.Fill(color.RGBA{100, 150, 200, 255})

	// Draw level
	g.drawLevel(screen)

	// Draw player
	g.drawPlayer(screen)

	// Draw UI
	g.drawUI(screen)
}

func (g *Game) drawLevel(screen *ebiten.Image) {
	startX := int(g.camera.x) / TileSize
	endX := startX + ScreenWidth/TileSize + 2

	for x := startX; x < endX && x < g.level.width; x++ {
		for y := 0; y < g.level.height; y++ {
			tile := g.level.tiles[x][y]
			if tile != TileAir {
				drawX := float32(x*TileSize) - float32(g.camera.x)
				drawY := float32(y * TileSize)

				g.drawTile(screen, tile, drawX, drawY)
			}
		}
	}

	// Draw coins
	for _, coin := range g.level.coins {
		if coin.collected {
			continue
		}

		drawX := float32(coin.x) - float32(g.camera.x)
		drawY := float32(coin.y)

		if drawX > -20 && drawX < ScreenWidth+20 {
			coin.animFrame++
			offset := float32(math.Sin(float64(coin.animFrame)*0.1) * 3)

			// Coin sprite
			vector.DrawFilledCircle(screen, drawX+10, drawY+10+offset, 8, color.RGBA{255, 215, 0, 255}, false)
			vector.DrawFilledCircle(screen, drawX+10, drawY+10+offset, 5, color.RGBA{255, 235, 100, 255}, false)
		}
	}

	// Draw flag
	flagX := float32(g.level.flagX) - float32(g.camera.x)
	vector.StrokeLine(screen, flagX+10, float32(g.level.flagY), flagX+10, float32(g.level.flagY+TileSize*4), 3, color.RGBA{100, 100, 100, 255}, false)
	vector.DrawFilledRect(screen, flagX+10, float32(g.level.flagY), 40, 30, color.RGBA{0, 200, 0, 255}, false)
}

func (g *Game) drawTile(screen *ebiten.Image, tile int, x, y float32) {
	switch tile {
	case TileGround:
		// Ground with grass top
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{139, 69, 19, 255}, false)
		vector.DrawFilledRect(screen, x, y, TileSize, 8, color.RGBA{34, 139, 34, 255}, false)

	case TileBrick:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{178, 34, 34, 255}, false)
		// Brick pattern
		vector.StrokeLine(screen, x, y+TileSize/2, x+TileSize, y+TileSize/2, 2, color.RGBA{100, 20, 20, 255}, false)
		vector.StrokeLine(screen, x+TileSize/2, y, x+TileSize/2, y+TileSize/2, 2, color.RGBA{100, 20, 20, 255}, false)

	case TileQuestion:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{255, 215, 0, 255}, false)
		ebitenutil.DebugPrintAt(screen, "?", int(x)+12, int(y)+10)

	case TileHard:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{128, 128, 128, 255}, false)

	case TileUsed:
		vector.DrawFilledRect(screen, x, y, TileSize, TileSize, color.RGBA{100, 80, 60, 255}, false)

	case TilePipeL, TilePipeTopL:
		vector.DrawFilledRect(screen, x, y, TileSize/2, TileSize, color.RGBA{0, 180, 0, 255}, false)
		vector.DrawFilledRect(screen, x+2, y, 4, TileSize, color.RGBA{0, 220, 0, 255}, false)

	case TilePipeR, TilePipeTopR:
		vector.DrawFilledRect(screen, x, y, TileSize/2, TileSize, color.RGBA{0, 160, 0, 255}, false)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := g.player
	drawX := float32(p.x) - float32(g.camera.x)
	drawY := float32(p.y)

	// Flicker when invincible
	if p.isInvincible && g.frameCount%4 < 2 {
		return
	}

	// Body color based on power state
	bodyColor := color.RGBA{220, 20, 20, 255} // Red (Mario)
	if p.isFire {
		bodyColor = color.RGBA{255, 200, 0, 255} // White/Fire
	}

	// Body
	vector.DrawFilledRect(screen, drawX, drawY, p.width, p.height, bodyColor, false)

	// Overalls (blue)
	vector.DrawFilledRect(screen, drawX+5, drawY+p.height-15, p.width-10, 10, color.RGBA{0, 0, 180, 255}, false)

	// Face
	faceX := drawX + p.width/2
	faceY := drawY + 10
	vector.DrawFilledCircle(screen, faceX, faceY, 10, color.RGBA{255, 220, 180, 255}, false)

	// Hat
	vector.DrawFilledRect(screen, drawX+2, drawY, p.width-4, 8, bodyColor, false)

	// Eyes
	eyeOffset := p.facing * 2
	vector.DrawFilledCircle(screen, faceX+float32(eyeOffset)-2, faceY+2, 3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, faceX+float32(eyeOffset)+2, faceY+2, 3, color.RGBA{0, 0, 0, 255}, false)

	// Mustache
	vector.DrawFilledRect(screen, faceX-5+float32(eyeOffset), faceY+6, 10, 3, color.RGBA{50, 30, 20, 255}, false)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 150}, false)

	// Score
	scoreText := fmt.Sprintf("MARIO\n%06d", g.player.score)
	ebitenutil.DebugPrintAt(screen, scoreText, 20, 5)

	// Coins
	coinText := fmt.Sprintf("COINS\nx%02d", g.player.coins)
	ebitenutil.DebugPrintAt(screen, coinText, 150, 5)

	// World
	worldText := fmt.Sprintf("WORLD\n%d-1", g.player.world)
	ebitenutil.DebugPrintAt(screen, worldText, 280, 5)

	// Lives
	livesText := fmt.Sprintf("LIVES\nx%d", g.player.lives)
	ebitenutil.DebugPrintAt(screen, livesText, 410, 5)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})
	text := "GAME OVER"
	ebitenutil.DebugPrintAt(screen, text, ScreenWidth/2-len(text)*8, ScreenHeight/2)
	ebitenutil.DebugPrintAt(screen, "Press ENTER to restart", ScreenWidth/2-80, ScreenHeight/2+50)
}

func (g *Game) drawWon(screen *ebiten.Image) {
	screen.Fill(color.RGBA{100, 150, 200, 255})
	text := "COURSE CLEAR!"
	ebitenutil.DebugPrintAt(screen, text, ScreenWidth/2-len(text)*8, ScreenHeight/2-30)

	scoreText := fmt.Sprintf("Score: %06d", g.player.score)
	ebitenutil.DebugPrintAt(screen, scoreText, ScreenWidth/2-50, ScreenHeight/2+20)

	ebitenutil.DebugPrintAt(screen, "Press ENTER to continue", ScreenWidth/2-80, ScreenHeight/2+70)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// AUDIO (Simple beep system)
// ============================================================================

type SoundType int

const (
	SoundJump SoundType = iota
	SoundCoin
	SoundStomp
	SoundHit
	SoundDie
	SoundPowerup
	SoundBump
	SoundBreak
	SoundStart
	SoundWin
)

var audioCtx *audio.Context

func initAudio() {
	audioCtx = audio.NewContext(44100)
}

func playSound(sound SoundType) {
	// Simplified - would need actual audio implementation
	_ = sound
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	initAudio()

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("🍄 Super Go Mario - Classic 2D Platformer")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
