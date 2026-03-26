// Go365 Day 86 - GO MARIO: PLATFORMER v5.0.0
// Классический 2D платформер с использованием готовых спрайтов

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

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 48
	Gravity      = 0.6
	JumpForce    = -12.0
	PlayerSpeed  = 5.0
)

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	// Player
	playerStand  *ebiten.Image
	playerWalk1  *ebiten.Image
	playerWalk2  *ebiten.Image
	playerJump   *ebiten.Image

	// Tiles
	grassTile    *ebiten.Image
	dirtTile     *ebiten.Image
	brickTile    *ebiten.Image
	platformTile *ebiten.Image

	// Decorations
	tree1        *ebiten.Image
	tree2        *ebiten.Image
	bush1        *ebiten.Image
	cloud1       *ebiten.Image
	cloud2       *ebiten.Image

	// Background
	bgSky        *ebiten.Image
	bgMountains  *ebiten.Image

	// Items
	coinSprite   *ebiten.Image
	flagSprite   *ebiten.Image

	// Font
	gameFont     font.Face
}

var gameAssets *Assets

func LoadAssets() *Assets {
	assets := &Assets{}
	var err error

	// Player (Green Alien)
	assets.playerStand, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	assets.playerWalk1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	assets.playerWalk2, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	assets.playerJump, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_jump.png")

	// Tiles
	assets.grassTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Grass/grass.png")
	assets.dirtTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickGrey.png")
	assets.brickTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/brickBrown.png")
	assets.platformTile, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/bridgeA.png")

	// Decorations
	assets.tree1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/cactus.png")
	assets.bush1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Tiles/bush.png")
	assets.cloud1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/cloud1.png")
	assets.cloud2, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/cloud2.png")

	// Items
	assets.coinSprite, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/coinGold.png")
	assets.flagSprite, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/flagGreen1.png")

	// Font
	assets.gameFont, err = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 24)
	if err != nil {
		assets.gameFont = nil
	}

	return assets
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
	health    int
	maxHealth int
	coins     int
	score     int
}

type Tile struct {
	x, y  int
	id    int
	width float32
	height float32
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

type Decoration struct {
	x, y   float64
	id     int
	animFrame int
}

type Game struct {
	player      *Player
	tiles       []*Tile
	enemies     []*Enemy
	coins       []*Coin
	decorations []*Decoration
	cameraX     float64
	state       int // 0=menu, 1=playing, 2=gameover, 3=win
	frame       int
	levelWidth  int
	flagX       float64
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	gameAssets = LoadAssets()

	g := &Game{
		player: &Player{
			x: 100,
			y: 300,
			width: 40,
			height: 56,
			facing: 1,
			maxHealth: 100,
			health: 100,
		},
		state: 0,
		tiles: make([]*Tile, 0),
		enemies: make([]*Enemy, 0),
		coins: make([]*Coin, 0),
		decorations: make([]*Decoration, 0),
	}

	g.GenerateLevel()
	return g
}

func (g *Game) GenerateLevel() {
	g.levelWidth = 100 // tiles

	// Generate ground
	for x := 0; x < g.levelWidth; x++ {
		// Ground tiles (2 rows)
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: 1}) // Grass
		g.tiles = append(g.tiles, &Tile{x: x, y: 15, id: 2}) // Dirt

		// Random platforms
		if x > 5 && rand.Float32() < 0.15 {
			platY := rand.Intn(4) + 8
			for bx := 0; bx < rand.Intn(3)+2; bx++ {
				if x+bx < g.levelWidth {
					g.tiles = append(g.tiles, &Tile{x: x+bx, y: platY, id: 3})
				}
			}
		}

		// Random coins
		if x > 5 && rand.Float32() < 0.2 {
			coinY := float64(rand.Intn(8)+4) * TileSize
			g.coins = append(g.coins, &Coin{
				x: float64(x*TileSize + 15),
				y: coinY,
			})
		}

		// Random enemies
		if x > 10 && rand.Float32() < 0.08 {
			g.enemies = append(g.enemies, &Enemy{
				x: float64(x * TileSize),
				y: float64(13 * TileSize),
				vx: -1.5,
				width: 40,
				height: 40,
				enemyType: 1,
				alive: true,
			})
		}
	}

	// Add decorations (trees, bushes, clouds)
	for x := 0; x < g.levelWidth; x++ {
		if rand.Float32() < 0.1 {
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize),
				y: float64(13*TileSize - 80),
				id: 1, // Tree
			})
		}
		if rand.Float32() < 0.15 {
			g.decorations = append(g.decorations, &Decoration{
				x: float64(x*TileSize),
				y: float64(13*TileSize - 30),
				id: 2, // Bush
			})
		}
	}

	// Clouds in background
	for i := 0; i < 20; i++ {
		g.decorations = append(g.decorations, &Decoration{
			x: float64(rand.Intn(g.levelWidth * TileSize)),
			y: float64(rand.Intn(200) + 50),
			id: 3, // Cloud
		})
	}

	// Flag at end
	g.flagX = float64((g.levelWidth - 5) * TileSize)
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frame++

	switch g.state {
	case 0: // Menu
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = 1
			g.player.health = g.player.maxHealth
			g.player.x = 100
			g.player.y = 300
			g.player.vx = 0
			g.player.vy = 0
			g.cameraX = 0
			// Reset coins
			for _, c := range g.coins {
				c.collected = false
			}
			// Reset enemies
			for _, e := range g.enemies {
				e.alive = true
			}
		}
		return nil

	case 2, 3: // GameOver/Win
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = 0
		}
		return nil
	}

	// Playing
	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.updateCoins()
	g.checkWin()

	return nil
}

func (g *Game) updatePlayer() {
	p := g.player

	// Input
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = PlayerSpeed
		p.facing = 1
		p.animFrame++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -PlayerSpeed
		p.facing = -1
		p.animFrame++
	} else {
		p.vx = 0
	}

	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
	}

	// Physics
	p.vy += Gravity
	p.x += p.vx
	p.y += p.vy

	// Collision with tiles
	g.checkTileCollision()

	// Screen boundaries
	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(g.levelWidth*TileSize)-float64(p.width) {
		p.x = float64(g.levelWidth*TileSize) - float64(p.width)
	}

	// Death by falling
	if p.y > ScreenHeight {
		p.health = 0
		if p.health <= 0 {
			g.state = 2
		}
	}
}

func (g *Game) checkTileCollision() {
	p := g.player
	p.onGround = false

	for _, tile := range g.tiles {
		tileX := float64(tile.x * TileSize)
		tileY := float64(tile.y * TileSize)

		// AABB collision
		if p.x < tileX+float64(TileSize) &&
			p.x+float64(p.width) > tileX &&
			p.y < tileY+float64(TileSize) &&
			p.y+float64(p.height) > tileY {

			// Landing on top
			if p.vy >= 0 && p.y+float64(p.height) <= tileY+20 {
				p.y = tileY - float64(p.height)
				p.vy = 0
				p.onGround = true
			}
			// Hitting ceiling
			if p.vy < 0 && p.y >= tileY {
				p.y = tileY + float64(TileSize)
				p.vy = 0
			}
			// Side collision
			if p.vx > 0 {
				p.x = tileX - float64(p.width)
			} else if p.vx < 0 {
				p.x = tileX + float64(TileSize)
			}
		}
	}
}

func (g *Game) updateCamera() {
	g.cameraX = g.player.x - ScreenWidth/3
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > float64(g.levelWidth*TileSize)-ScreenWidth {
		g.cameraX = float64(g.levelWidth*TileSize) - ScreenWidth
	}
}

func (g *Game) updateEnemies() {
	p := g.player

	for _, e := range g.enemies {
		if !e.alive {
			continue
		}

		e.x += e.vx

		// Simple patrol
		if rand.Float32() < 0.02 {
			e.vx *= -1
		}

		// Collision with player
		if p.x < e.x+float64(e.width) &&
			p.x+float64(p.width) > e.x &&
			p.y < e.y+float64(e.height) &&
			p.y+float64(p.height) > e.y {

			// Jump on enemy
			if p.vy > 0 && p.y+float64(p.height) < e.y+float64(e.height)/2 {
				e.alive = false
				p.vy = JumpForce / 2
				p.score += 100
			} else {
				p.health -= 10
				p.vy = -5
				if p.health <= 0 {
					g.state = 2
				}
			}
		}
	}
}

func (g *Game) updateCoins() {
	p := g.player

	for _, c := range g.coins {
		if c.collected {
			continue
		}

		c.animFrame++

		// Collection
		if p.x < c.x+20 &&
			p.x+float64(p.width) > c.x &&
			p.y < c.y+20 &&
			p.y+float64(p.height) > c.y {
			c.collected = true
			p.coins++
			p.score += 10
		}
	}
}

func (g *Game) checkWin() {
	if g.player.x >= g.flagX {
		g.state = 3
		g.player.score += 1000
	}
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case 0:
		g.drawMenu(screen)
	case 1, 2, 3:
		g.drawGame(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Sky gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(135 - y/10)
		g := uint8(206 - y/10)
		b := uint8(235 - y/10)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	if gameAssets.gameFont != nil {
		text.Draw(screen, "GO MARIO PLATFORMER", gameAssets.gameFont, ScreenWidth/2-180, 200, color.White)
		text.Draw(screen, "Press ENTER to start", gameAssets.gameFont, ScreenWidth/2-140, 350, color.White)

		controls := []string{
			"Arrow Keys / WASD - Move",
			"Space / W / Up - Jump",
			"Jump on enemies to defeat them!",
			"Collect coins and reach the flag!",
		}
		y := 420
		for _, line := range controls {
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-150, y, color.RGBA{200, 200, 200, 255})
			y += 30
		}
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Sky background
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(135 - y/10)
		g := uint8(206 - y/10)
		b := uint8(235 - y/10)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	camX := g.cameraX

	// Draw decorations (background - clouds)
	for _, d := range g.decorations {
		if d.id == 3 { // Cloud
			img := gameAssets.cloud1
			if g.frame%120 < 60 {
				img = gameAssets.cloud2
			}
			if img != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(d.x-camX*0.3, d.y) // Parallax
				screen.DrawImage(img, op)
			}
		}
	}

	// Draw tiles
	for _, tile := range g.tiles {
		if float64(tile.x*TileSize) < camX-100 || float64(tile.x*TileSize) > camX+ScreenWidth+100 {
			continue
		}

		var img *ebiten.Image
		switch tile.id {
		case 1:
			img = gameAssets.grassTile
		case 2:
			img = gameAssets.dirtTile
		case 3:
			img = gameAssets.brickTile
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(tile.x*TileSize)-camX, float64(tile.y*TileSize))
			screen.DrawImage(img, op)
		}
	}

	// Draw decorations (foreground - trees, bushes)
	for _, d := range g.decorations {
		if d.id == 1 || d.id == 2 {
			var img *ebiten.Image
			if d.id == 1 {
				img = gameAssets.tree1
			} else {
				img = gameAssets.bush1
			}
			if img != nil && gameAssets != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(d.x-camX, d.y)
				screen.DrawImage(img, op)
			}
		}
	}

	// Draw coins
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		if gameAssets.coinSprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(c.x-camX, c.y+math.Sin(float64(c.animFrame)*0.1)*5)
			screen.DrawImage(gameAssets.coinSprite, op)
		}
	}

	// Draw enemies
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		// Draw enemy as red circle for now
		vector.DrawFilledCircle(screen, float32(e.x-camX+20), float32(e.y+20), 20, color.RGBA{255, 50, 50, 255}, true)
	}

	// Draw flag
	if gameAssets.flagSprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(g.flagX-camX, g.flagX)
		screen.DrawImage(gameAssets.flagSprite, op)
	}

	// Draw player
	g.drawPlayer(screen, camX)

	// Draw HUD
	g.drawHUD(screen)

	// Game Over / Win screens
	if g.state == 2 {
		g.drawGameOver(screen)
	} else if g.state == 3 {
		g.drawWin(screen)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image, camX float64) {
	p := g.player

	var img *ebiten.Image
	if !p.onGround {
		img = gameAssets.playerJump
	} else if math.Abs(p.vx) > 0.5 {
		if (p.animFrame/8)%2 == 0 {
			img = gameAssets.playerWalk1
		} else {
			img = gameAssets.playerWalk2
		}
	} else {
		img = gameAssets.playerStand
	}

	if img != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.x-camX, p.y)
		if p.facing == -1 {
			op.GeoM.Scale(-0.5, 0.5)
			op.GeoM.Translate(float64(p.width), 0)
		} else {
			op.GeoM.Scale(0.5, 0.5)
		}
		screen.DrawImage(img, op)
	} else {
		// Fallback
		vector.DrawFilledRect(screen, float32(p.x-camX), float32(p.y), float32(p.width), float32(p.height), color.RGBA{0, 255, 100, 255}, true)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 50, color.RGBA{0, 0, 0, 150}, true)

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("HP: %d/%d", g.player.health, g.player.maxHealth), gameAssets.gameFont, 20, 15, color.RGBA{255, 100, 100, 255})
		text.Draw(screen, fmt.Sprintf("Coins: %d", g.player.coins), gameAssets.gameFont, 250, 15, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), gameAssets.gameFont, 450, 15, color.White)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, true)
	if gameAssets.gameFont != nil {
		text.Draw(screen, "GAME OVER", gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2, color.RGBA{255, 50, 50, 255})
		text.Draw(screen, "Press ENTER", gameAssets.gameFont, ScreenWidth/2-70, ScreenHeight/2+50, color.White)
	}
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 100, 0, 180}, true)
	if gameAssets.gameFont != nil {
		text.Draw(screen, "YOU WIN!", gameAssets.gameFont, ScreenWidth/2-60, ScreenHeight/2, color.RGBA{255, 215, 0, 255})
		text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), gameAssets.gameFont, ScreenWidth/2-80, ScreenHeight/2+50, color.White)
		text.Draw(screen, "Press ENTER", gameAssets.gameFont, ScreenWidth/2-70, ScreenHeight/2+100, color.White)
	}
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GO MARIO: PLATFORMER - Go365 Day 86")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
