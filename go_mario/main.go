// Go365 Day 86 - go_mario v1.0.0 (Complete Rewrite)
// Simple Mario-style platformer using Go + Ebitengine
// Using Green Alien sprite as main character

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	// Screen
	ScreenWidth  = 800
	ScreenHeight = 600
	TileSize     = 40

	// Physics
	Gravity      = 0.6
	JumpForce    = -12.0
	RunSpeed     = 5.0
	WalkSpeed    = 2.5
	MaxFallSpeed = 10.0
	Friction     = 0.85
	Acceleration = 0.5

	// Tile types
	TileAir      = 0
	TileGround   = 1
	TileBrick    = 2
	TileQuestion = 3
	TileHard     = 4
	TilePipe     = 5
	TileUsed     = 6

	// Enemy types
	EnemySlimeGreen = 1
	EnemySlimeBlue  = 2

	// Game states
	StateMenu     = 0
	StatePlaying  = 1
	StateGameOver = 2
	StateWon      = 3
)

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	// Player (Green Alien)
	playerStand *ebiten.Image
	playerWalk1 *ebiten.Image
	playerWalk2 *ebiten.Image
	playerJump  *ebiten.Image

	// Enemies
	slimeGreen *ebiten.Image
	slimeBlue  *ebiten.Image

	// Tiles
	grassTile    *ebiten.Image
	brickTile    *ebiten.Image
	questionTile *ebiten.Image
	hardTile     *ebiten.Image
	usedTile     *ebiten.Image

	// Items
	coinSprite *ebiten.Image
	flagSprite *ebiten.Image

	// Font
	gameFont font.Face
}

var gameAssets *Assets

func LoadAssets() (*Assets, error) {
	assets := &Assets{}
	var err error

	// Load player sprites (Green Alien from assets/PNG/Players)
	assets.playerStand, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	if err != nil {
		fmt.Printf("Warning: Could not load player stand sprite: %v\n", err)
	}

	assets.playerWalk1, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	if err != nil {
		fmt.Printf("Warning: Could not load player walk1 sprite: %v\n", err)
	}

	assets.playerWalk2, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	if err != nil {
		fmt.Printf("Warning: Could not load player walk2 sprite: %v\n", err)
	}

	assets.playerJump, _, err = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_jump.png")
	if err != nil {
		fmt.Printf("Warning: Could not load player jump sprite: %v\n", err)
	}

	// Load enemy sprites
	assets.slimeGreen, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeGreen.png")
	if err != nil {
		fmt.Printf("Warning: Could not load slimeGreen sprite: %v\n", err)
	}

	assets.slimeBlue, _, err = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeBlue.png")
	if err != nil {
		fmt.Printf("Warning: Could not load slimeBlue sprite: %v\n", err)
	}

	// Load tile sprites
	assets.grassTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Grass/grass.png")
	if err != nil {
		fmt.Printf("Warning: Could not load grass tile: %v\n", err)
	}

	assets.brickTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickGrey.png")
	if err != nil {
		fmt.Printf("Warning: Could not load brick tile: %v\n", err)
	}

	assets.questionTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem.png")
	if err != nil {
		fmt.Printf("Warning: Could not load question tile: %v\n", err)
	}

	assets.hardTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickBrown.png")
	if err != nil {
		fmt.Printf("Warning: Could not load hard tile: %v\n", err)
	}

	assets.usedTile, _, err = ebitenutil.NewImageFromFile("assets/PNG/Tiles/boxItem_disabled.png")
	if err != nil {
		fmt.Printf("Warning: Could not load used tile: %v\n", err)
	}

	// Load items
	assets.coinSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/coinGold.png")
	if err != nil {
		fmt.Printf("Warning: Could not load coin sprite: %v\n", err)
	}

	assets.flagSprite, _, err = ebitenutil.NewImageFromFile("assets/PNG/Items/flagGreen1.png")
	if err != nil {
		fmt.Printf("Warning: Could not load flag sprite: %v\n", err)
	}

	// Load font
	assets.gameFont, err = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 28)
	if err != nil {
		fmt.Printf("Warning: Could not load font, using default: %v\n", err)
		assets.gameFont = nil
	}

	return assets, nil
}

func loadFont(path string, size int) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}

	return opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  72,
	})
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Player struct {
	x, y      float64
	vx, vy    float64
	width     float32
	height    float32
	onGround  bool
	facing    int
	animFrame int

	// Stats
	coins int
	score int
	lives int
}

type Enemy struct {
	x, y      float64
	vx        float64
	width     float32
	height    float32
	enemyType int
	alive     bool
}

type Coin struct {
	x, y      float64
	collected bool
	animFrame int
}

type Particle struct {
	x, y   float64
	vx, vy float64
	life   int
	color  color.RGBA
	size   float32
}

type Level struct {
	width  int
	height int
	tiles  [][]int
	coins  []Coin
	enemies []*Enemy
	flagX  int
	flagY  int
}

type Camera struct {
	x, y float64
}

type Game struct {
	player    *Player
	level     *Level
	camera    *Camera
	particles []*Particle
	state     int
	frameCount int
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	// Load assets
	var err error
	gameAssets, err = LoadAssets()
	if err != nil {
		fmt.Println("Warning: Some assets failed to load, using fallback rendering")
	}

	g := &Game{
		player: &Player{
			x:      100,
			y:      100,
			width:  32,
			height: 48,
			facing: 1,
			lives:  3,
		},
		camera:     &Camera{},
		state:      StateMenu,
		frameCount: 0,
		particles:  make([]*Particle, 0),
	}

	g.LoadLevel(1)
	return g
}

func (g *Game) LoadLevel(world int) {
	g.level = GenerateLevel(world)
	g.player.x = 100
	g.player.y = 100
	g.player.vx = 0
	g.player.vy = 0
	g.player.coins = 0
	g.player.score = 0
	g.camera.x = 0
	g.particles = make([]*Particle, 0)
}

func GenerateLevel(world int) *Level {
	width := 100
	height := 15

	level := &Level{
		width:  width,
		height: height,
		tiles:  make([][]int, width),
		coins:  make([]Coin, 0),
		enemies: make([]*Enemy, 0),
	}

	// Initialize tiles
	for x := range level.tiles {
		level.tiles[x] = make([]int, height)
	}

	// Generate terrain
	for x := 0; x < width; x++ {
		// Ground layer
		for y := 10; y < height; y++ {
			level.tiles[x][y] = TileGround
		}

		// Gaps (pits)
		if x%40 == 35 && x > 50 {
			for gx := 0; gx < 3; gx++ {
				if x+gx < width {
					for y := 10; y < height; y++ {
						level.tiles[x+gx][y] = TileAir
					}
				}
			}
		}

		// Random structures
		if x > 15 && rand.Float32() < 0.15 {
			structureType := rand.Intn(4)

			switch structureType {
			case 0: // Brick platform with coins
				platY := rand.Intn(3) + 5
				for bx := 0; bx < 4; bx++ {
					if x+bx < width {
						level.tiles[x+bx][platY] = TileBrick
					}
				}
				// Add coin above
				if x+2 < width {
					level.coins = append(level.coins, Coin{
						x: float64((x+2)*TileSize + 10),
						y: float64((platY-1)*TileSize + 10),
					})
				}

			case 1: // Question block
				if x < width {
					qY := rand.Intn(3) + 5
					level.tiles[x][qY] = TileQuestion
				}

			case 2: // Pipe
				pipeHeight := rand.Intn(2) + 2
				pipeY := 10 - pipeHeight
				for py := pipeY; py < 10; py++ {
					if x < width {
						level.tiles[x][py] = TilePipe
					}
				}

			case 3: // Enemy spawn
				level.enemies = append(level.enemies, &Enemy{
					x: float64(x * TileSize),
					y: float64(8 * TileSize),
					vx: -1.5,
					width:  32,
					height: 32,
					enemyType: EnemySlimeGreen,
					alive: true,
				})
			}
		}

		// Add more enemies
		if x > 20 && rand.Float32() < 0.08 {
			enemyType := EnemySlimeGreen
			if rand.Float32() < 0.3 {
				enemyType = EnemySlimeBlue
			}
			level.enemies = append(level.enemies, &Enemy{
				x: float64(x * TileSize),
				y: float64(8 * TileSize),
				vx: -1.5,
				width:  32,
				height: 32,
				enemyType: enemyType,
				alive: true,
			})
		}
	}

	// Add coins along the level
	for i := 0; i < 50; i++ {
		cx := rand.Intn(width-10) + 5
		cy := rand.Intn(8) + 2
		level.coins = append(level.coins, Coin{
			x: float64(cx*TileSize + 10),
			y: float64(cy*TileSize + 10),
		})
	}

	// Add flag at end
	level.flagX = (width - 5) * TileSize
	level.flagY = 6 * TileSize

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
	}
}

func (g *Game) updatePlaying() {
	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.updateCoins()
	g.updateParticles()
	g.checkWin()
	g.checkDeath()
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
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
	}

	// Apply gravity
	p.vy += Gravity
	if p.vy > MaxFallSpeed {
		p.vy = MaxFallSpeed
	}

	// Apply velocity
	p.x += p.vx
	p.y += p.vy

	// Collision with tiles
	g.checkTileCollision()

	// Screen boundaries
	if p.x < 0 {
		p.x = 0
		p.vx = 0
	}
}

func (g *Game) checkTileCollision() {
	p := g.player
	level := g.level

	// Check horizontal collision
	tileX := int(p.x) / TileSize
	tileY := int(p.y) / TileSize

	if tileX >= 0 && tileX < level.width && tileY >= 0 && tileY < level.height {
		tile := level.tiles[tileX][tileY]
		if tile != TileAir && tile != 0 {
			if p.vx > 0 {
				p.x = float64(tileX*TileSize) - float64(p.width)
				p.vx = 0
			} else if p.vx < 0 {
				p.x = float64((tileX+1)*TileSize)
				p.vx = 0
			}
		}
	}

	// Check vertical collision
	tileX = int((p.x + float64(p.width)/2) / TileSize)
	tileYBottom := int((p.y + float64(p.height)) / TileSize)
	tileYTop := int(p.y / TileSize)

	if tileX >= 0 && tileX < level.width {
		// Landing on ground
		if p.vy > 0 && tileYBottom < level.height {
			tile := level.tiles[tileX][tileYBottom]
			if tile != TileAir && tile != 0 {
				p.y = float64(tileYBottom*TileSize) - float64(p.height)
				p.vy = 0
				p.onGround = true
			}
		}
		// Hitting ceiling
		if p.vy < 0 && tileYTop >= 0 {
			tile := level.tiles[tileX][tileYTop]
			if tile != TileAir && tile != 0 {
				p.y = float64((tileYTop+1)*TileSize)
				p.vy = 0
				// Hit question block
				if tile == TileQuestion {
					level.tiles[tileX][tileYTop] = TileUsed
					p.score += 50
					p.coins++
					// Spawn coin particle
					g.spawnParticles(p.x+float64(p.width)/2, p.y, 5, color.RGBA{255, 215, 0, 255})
				}
			}
		}
	}
}

func (g *Game) updateCamera() {
	g.camera.x = g.player.x - ScreenWidth/2
	if g.camera.x < 0 {
		g.camera.x = 0
	}
	if g.camera.x > float64(g.level.width*TileSize-ScreenWidth) {
		g.camera.x = float64(g.level.width*TileSize - ScreenWidth)
	}
}

func (g *Game) updateEnemies() {
	for _, enemy := range g.level.enemies {
		if !enemy.alive {
			continue
		}

		// Move enemy
		enemy.x += enemy.vx

		// Simple AI: turn around at edges or walls
		tileX := int(enemy.x) / TileSize
		tileY := int(enemy.y) / TileSize

		// Check for wall
		if tileX >= 0 && tileX < g.level.width && tileY >= 0 && tileY < g.level.height {
			tile := g.level.tiles[tileX][tileY]
			if tile != TileAir && tile != 0 {
				enemy.vx *= -1
				enemy.x += enemy.vx * 2
			}
		}

		// Check collision with player
		if g.checkCollision(g.player, enemy) {
			// Player jumps on enemy
			if g.player.vy > 0 && g.player.y+float64(g.player.height) < enemy.y+float64(enemy.height)/2 {
				enemy.alive = false
				g.player.vy = JumpForce / 2
				g.player.score += 100
				g.spawnParticles(enemy.x+float64(enemy.width)/2, enemy.y+float64(enemy.height)/2, 8, color.RGBA{100, 255, 100, 255})
			} else {
				// Player takes damage
				g.player.lives--
				g.player.vy = -5
				g.player.vx = float64(g.player.facing) * 3
				if g.player.lives <= 0 {
					g.state = StateGameOver
				}
			}
		}
	}
}

func (g *Game) updateCoins() {
	for i := range g.level.coins {
		if g.level.coins[i].collected {
			continue
		}

		// Check collision with player
		if g.player.x < g.level.coins[i].x+20 &&
			g.player.x+float64(g.player.width) > g.level.coins[i].x &&
			g.player.y < g.level.coins[i].y+20 &&
			g.player.y+float64(g.player.height) > g.level.coins[i].y {
			g.level.coins[i].collected = true
			g.player.coins++
			g.player.score += 10
			g.spawnParticles(g.level.coins[i].x+10, g.level.coins[i].y+10, 5, color.RGBA{255, 215, 0, 255})
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.3 // gravity
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.RGBA) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			x:     x,
			y:     y,
			vx:    (rand.Float64() - 0.5) * 4,
			vy:    (rand.Float64() - 0.5) * 4,
			life:  30 + rand.Intn(20),
			color: c,
			size:  3 + rand.Float32()*3,
		})
	}
}

func (g *Game) checkCollision(player *Player, enemy *Enemy) bool {
	return player.x < enemy.x+float64(enemy.width) &&
		player.x+float64(player.width) > enemy.x &&
		player.y < enemy.y+float64(enemy.height) &&
		player.y+float64(player.height) > enemy.y
}

func (g *Game) checkWin() {
	if g.player.x >= float64(g.level.flagX) {
		g.state = StateWon
		g.player.score += 1000
	}
}

func (g *Game) checkDeath() {
	if g.player.y > ScreenHeight {
		g.player.lives--
		if g.player.lives <= 0 {
			g.state = StateGameOver
		} else {
			g.player.x = 100
			g.player.y = 100
			g.player.vx = 0
			g.player.vy = 0
		}
	}
}

func (g *Game) updateEndScreen() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.state = StateMenu
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
	// Clear screen
	screen.Fill(color.RGBA{135, 206, 235, 255}) // Sky blue

	// Title
	title := "GO MARIO"
	if gameAssets.gameFont != nil {
		text.Draw(screen, title, gameAssets.gameFont, ScreenWidth/2-80, 200, color.White)
		text.Draw(screen, "Press ENTER or SPACE to start", gameAssets.gameFont, ScreenWidth/2-150, 300, color.White)
	} else {
		ebitenutil.DebugPrint(screen, "GO MARIO - Press ENTER to start")
	}

	// Draw player preview
	if gameAssets.playerStand != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(ScreenWidth/2-32), 220)
		op.GeoM.Scale(0.5, 0.5)
		screen.DrawImage(gameAssets.playerStand, op)
	}

	// Instructions
	instructions := []string{
		"Controls:",
		"Arrow Keys / WASD - Move",
		"Space / Up / W - Jump",
		"Jump on enemies to defeat them!",
		"Collect coins and reach the flag!",
	}

	y := 400
	for _, line := range instructions {
		if gameAssets.gameFont != nil {
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-120, y, color.White)
		}
		y += 30
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Clear screen with sky color
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Apply camera transform
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.camera.x, -g.camera.y)

	// Draw tiles
	g.drawTiles(screen, op)

	// Draw coins
	g.drawCoins(screen, op)

	// Draw flag
	g.drawFlag(screen, op)

	// Draw enemies
	g.drawEnemies(screen, op)

	// Draw player
	g.drawPlayer(screen, op)

	// Draw particles
	g.drawParticles(screen, op)

	// Draw UI (HUD)
	g.drawUI(screen)
}

func (g *Game) drawTiles(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	level := g.level
	startX := int(g.camera.x / TileSize)
	endX := startX + ScreenWidth/TileSize + 2

	for x := startX; x <= endX && x < level.width; x++ {
		for y := 0; y < level.height; y++ {
			tile := level.tiles[x][y]
			if tile == TileAir {
				continue
			}

			var tileImg *ebiten.Image
			switch tile {
			case TileGround:
				tileImg = gameAssets.grassTile
			case TileBrick:
				tileImg = gameAssets.brickTile
			case TileQuestion:
				tileImg = gameAssets.questionTile
			case TileHard:
				tileImg = gameAssets.hardTile
			case TileUsed:
				tileImg = gameAssets.usedTile
			}

			if tileImg != nil {
				tileOp := &ebiten.DrawImageOptions{}
				tileOp.GeoM.Translate(float64(x*TileSize), float64(y*TileSize))
				tileOp.GeoM.Concat(op.GeoM)
				screen.DrawImage(tileImg, tileOp)
			} else {
				// Fallback: draw colored rectangle
				tileRect := color.RGBA{139, 90, 43, 255}
				vector.DrawFilledRect(screen,
					float32(float64(x*TileSize)-g.camera.x),
					float32(float64(y*TileSize)-g.camera.y),
					TileSize, TileSize, tileRect, true)
			}
		}
	}
}

func (g *Game) drawCoins(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, coin := range g.level.coins {
		if coin.collected {
			continue
		}

		if gameAssets.coinSprite != nil {
			coinOp := &ebiten.DrawImageOptions{}
			coinOp.GeoM.Translate(coin.x, coin.y)
			coinOp.GeoM.Concat(op.GeoM)
			screen.DrawImage(gameAssets.coinSprite, coinOp)
		} else {
			// Fallback: draw yellow circle
			vector.DrawFilledCircle(screen,
				float32(coin.x-g.camera.x+10),
				float32(coin.y-g.camera.y+10),
				8, color.RGBA{255, 215, 0, 255}, true)
		}
	}
}

func (g *Game) drawFlag(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if gameAssets.flagSprite != nil {
		flagOp := &ebiten.DrawImageOptions{}
		flagOp.GeoM.Translate(float64(g.level.flagX), float64(g.level.flagY))
		flagOp.GeoM.Concat(op.GeoM)
		screen.DrawImage(gameAssets.flagSprite, flagOp)
	} else {
		// Fallback: draw green flag
		vector.DrawFilledRect(screen,
			float32(float64(g.level.flagX)-g.camera.x),
			float32(float64(g.level.flagY)-g.camera.y),
			10, 80, color.RGBA{0, 128, 0, 255}, true)
	}
}

func (g *Game) drawEnemies(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, enemy := range g.level.enemies {
		if !enemy.alive {
			continue
		}

		var enemyImg *ebiten.Image
		switch enemy.enemyType {
		case EnemySlimeGreen:
			enemyImg = gameAssets.slimeGreen
		case EnemySlimeBlue:
			enemyImg = gameAssets.slimeBlue
		}

		if enemyImg != nil {
			enemyOp := &ebiten.DrawImageOptions{}
			enemyOp.GeoM.Translate(enemy.x, enemy.y)
			enemyOp.GeoM.Concat(op.GeoM)
			screen.DrawImage(enemyImg, enemyOp)
		} else {
			// Fallback: draw colored rectangle
			enemyColor := color.RGBA{100, 255, 100, 255}
			if enemy.enemyType == EnemySlimeBlue {
				enemyColor = color.RGBA{100, 100, 255, 255}
			}
			vector.DrawFilledRect(screen,
				float32(enemy.x-g.camera.x),
				float32(enemy.y-g.camera.y),
				float32(enemy.width), float32(enemy.height), enemyColor, true)
		}
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	p := g.player

	var playerImg *ebiten.Image
	if !p.onGround {
		playerImg = gameAssets.playerJump
	} else if math.Abs(p.vx) > 0.5 {
		// Walking animation
		if (p.animFrame / 8) % 2 == 0 {
			playerImg = gameAssets.playerWalk1
		} else {
			playerImg = gameAssets.playerWalk2
		}
	} else {
		playerImg = gameAssets.playerStand
	}

	if playerImg != nil {
		playerOp := &ebiten.DrawImageOptions{}
		playerOp.GeoM.Translate(p.x, p.y)
		if p.facing == -1 {
			playerOp.GeoM.Scale(-0.5, 0.5)
			playerOp.GeoM.Translate(float64(p.width), 0)
		} else {
			playerOp.GeoM.Scale(0.5, 0.5)
		}
		playerOp.GeoM.Concat(op.GeoM)
		screen.DrawImage(playerImg, playerOp)
	} else {
		// Fallback: draw green rectangle
		playerColor := color.RGBA{0, 255, 0, 255}
		vector.DrawFilledRect(screen,
			float32(p.x-g.camera.x),
			float32(p.y-g.camera.y),
			float32(p.width), float32(p.height), playerColor, true)
	}
}

func (g *Game) drawParticles(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen,
			float32(p.x-g.camera.x),
			float32(p.y-g.camera.y),
			p.size, p.color, true)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	// Draw HUD background
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 50, color.RGBA{0, 0, 0, 128}, true)

	// Draw stats
	if gameAssets.gameFont != nil {
		scoreText := fmt.Sprintf("SCORE: %d", g.player.score)
		coinText := fmt.Sprintf("COINS: %d", g.player.coins)
		livesText := fmt.Sprintf("LIVES: %d", g.player.lives)

		text.Draw(screen, scoreText, gameAssets.gameFont, 10, 35, color.White)
		text.Draw(screen, coinText, gameAssets.gameFont, 250, 35, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, livesText, gameAssets.gameFont, 450, 35, color.RGBA{255, 100, 100, 255})
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d | Coins: %d | Lives: %d",
			g.player.score, g.player.coins, g.player.lives))
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 255})

	if gameAssets.gameFont != nil {
		text.Draw(screen, "GAME OVER", gameAssets.gameFont, ScreenWidth/2-100, ScreenHeight/2-50, color.RGBA{255, 0, 0, 255})
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-100, ScreenHeight/2, color.White)
		text.Draw(screen, "Press ENTER to continue", gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2+50, color.White)
	} else {
		ebitenutil.DebugPrint(screen, "GAME OVER - Press ENTER")
	}
}

func (g *Game) drawWon(screen *ebiten.Image) {
	screen.Fill(color.RGBA{100, 200, 100, 255})

	if gameAssets.gameFont != nil {
		text.Draw(screen, "YOU WIN!", gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2-50, color.RGBA{255, 255, 0, 255})
		text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2, color.White)
		text.Draw(screen, fmt.Sprintf("Coins: %d", g.player.coins), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+30, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, "Press ENTER to continue", gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2+80, color.White)
	} else {
		ebitenutil.DebugPrint(screen, "YOU WIN! - Press ENTER")
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// SOUND SYSTEM (Placeholder for future implementation)
// ============================================================================

type SoundType int

const (
	SoundJump SoundType = iota
	SoundCoin
	SoundStomp
	SoundPowerup
	SoundDie
	SoundWin
)

func playSound(sound SoundType) {
	// Placeholder: Sound system will be implemented later
	// For now, sounds are silent
	_ = sound
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Go Mario - Go365 Day 86")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
