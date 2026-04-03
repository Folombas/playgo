package engine

import (
	"fmt"
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"city_platformer/internal/entity"
	"city_platformer/internal/hud"
	"city_platformer/internal/input"
	"city_platformer/internal/level"
	"city_platformer/internal/sprite"
)

// State is the game state.
type State int

const (
	Menu State = iota
	Playing
	GameOver
)

// Game is the main game.
type Game struct {
	state   State
	player  *entity.Player
	gen     *level.Generator
	input   *input.Manager
	obstacles []*entity.Obstacle
	coins   []*entity.Coin
	particles []Particle
	groundY float64
	bg      *ebiten.Image
	ground  *ebiten.Image
	coinSpr *ebiten.Image
	obstacleSpr []*ebiten.Image
	playerSpr *ebiten.Image
	walkImgs  []*ebiten.Image
	spawnTimer float64
	coinTimer  float64
	distance   float64
}

// Particle is a visual effect.
type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife int
	R, G, B uint8
}

// NewGame creates the game.
func NewGame() *Game {
	g := &Game{
		state:   Menu,
		input:   input.NewManager(),
		groundY: 550,
		gen:     level.NewGenerator(42, 550),
	}

	g.loadAssets()
	g.resetGame()
	return g
}

func (g *Game) loadAssets() {
	// Try to load sprites, fallback to colored rects
	g.bg, _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/bg.png")
	g.ground, _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Tiles/grass.png")
	g.coinSpr, _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Items/coinGold.png")

	// Player sprites
	g.playerSpr, _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Player/p1_stand.png")
	g.walkImgs = make([]*ebiten.Image, 11)
	for i := 0; i < 11; i++ {
		g.walkImgs[i], _ = sprite.Load(fmt.Sprintf("assets/sprites/PlatformerComplete/Base pack/Player/p1_walk/PNG/p1_walk%02d.png", i+1))
	}

	// Obstacle sprites
	g.obstacleSpr = make([]*ebiten.Image, 3)
	g.obstacleSpr[0], _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Items/spikes.png")
	g.obstacleSpr[1], _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Enemies/flyFly1.png")
	g.obstacleSpr[2], _ = sprite.Load("assets/sprites/PlatformerComplete/Base pack/Tiles/brick.png")
}

func (g *Game) resetGame() {
	g.player = entity.NewPlayer(150, g.groundY-48)
	g.obstacles = nil
	g.coins = nil
	g.particles = nil
	g.spawnTimer = 0
	g.coinTimer = 0
	g.distance = 0
	g.gen = level.NewGenerator(42, g.groundY)
}

// Update updates the game.
func (g *Game) Update() error {
	switch g.state {
	case Menu:
		if g.input.JustEnter() {
			g.state = Playing
			g.resetGame()
		}
	case Playing:
		g.updatePlaying()
	case GameOver:
		if g.input.JustEnter() {
			g.state = Playing
			g.resetGame()
		}
	}
	return nil
}

func (g *Game) updatePlaying() {
	// Jump input
	if g.input.JustJump() {
		g.player.Jump(480)
		g.spawnParticles(g.player.X+16, g.player.Y+48, 5, 200, 200, 200)
	}

	// Update player
	g.player.Update()

	// Ground collision
	if g.player.Y >= g.groundY-g.player.H {
		g.player.Y = g.groundY - g.player.H
		g.player.Land()
	}

	// Spawn obstacles
	speed := g.gen.Speed()
	g.spawnTimer -= 1.0 / 60.0
	if g.spawnTimer <= 0 {
		g.spawnTimer = 1.0 + rand.Float64()*2.0
		g.obstacles = append(g.obstacles, g.gen.SpawnObstacle(1280))
	}

	// Spawn coins
	g.coinTimer -= 1.0 / 60.0
	if g.coinTimer <= 0 {
		g.coinTimer = 0.5 + rand.Float64()*1.5
		g.coins = append(g.coins, g.gen.SpawnCoin(1280))
	}

	// Update obstacles
	for i := len(g.obstacles) - 1; i >= 0; i-- {
		o := g.obstacles[i]
		o.Update(speed)
		if o.X < -100 {
			g.obstacles = append(g.obstacles[:i], g.obstacles[i+1:]...)
			g.player.Score += 10
			continue
		}
		// Collision
		if entity.AABB(g.player.X, g.player.Y, g.player.W, g.player.H, o.X, o.Y, o.W, o.H) {
			g.player.Hit(1)
			g.spawnParticles(g.player.X+16, g.player.Y+24, 10, 255, 0, 0)
			if g.player.HP <= 0 {
				g.state = GameOver
				return
			}
		}
	}

	// Update coins
	for i := len(g.coins) - 1; i >= 0; i-- {
		c := g.coins[i]
		c.Update(speed)
		if c.X < -50 {
			g.coins = append(g.coins[:i], g.coins[i+1:]...)
			continue
		}
		if !c.Collected && entity.AABB(g.player.X, g.player.Y, g.player.W, g.player.H, c.X, c.Y, c.W, c.H) {
			c.Collected = true
			g.player.Coins++
			g.player.Score += 25
			g.spawnParticles(c.X+8, c.Y+8, 8, 255, 215, 0)
			g.coins = append(g.coins[:i], g.coins[i+1:]...)
		}
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 5
		p.Life--
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Distance & difficulty
	g.distance += speed / 60.0
	if int(g.distance)%500 == 0 && g.distance > 0 {
		g.gen.IncreaseSpeed()
	}
}

func (g *Game) spawnParticles(x, y float64, count int, r, g2, b uint8) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, Particle{
			X: x, Y: y,
			VX: (rand.Float64() - 0.5) * 200,
			VY: -rand.Float64() * 150,
			Life: 20 + rand.Intn(20),
			MaxLife: 40,
			R: r, G: g2, B: b,
		})
	}
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Sky
	screen.Fill(color.RGBA{135, 206, 235, 255})

	// Background
	if g.bg != nil {
		screen.DrawImage(g.bg, nil)
	}

	switch g.state {
	case Menu:
		hud.DrawMenu(screen)
	case Playing:
		g.drawGame(screen)
	case GameOver:
		g.drawGame(screen)
		hud.DrawGameOver(screen, g.player.Score)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Ground
	if g.ground != nil {
		for x := 0; x < 1280; x += 64 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), g.groundY)
			screen.DrawImage(g.ground, op)
		}
	} else {
		groundRect := ebiten.NewImage(1280, 200)
		groundRect.Fill(color.RGBA{100, 180, 80, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, g.groundY)
		screen.DrawImage(groundRect, op)
	}

	// Coins
	for _, c := range g.coins {
		if c.Collected {
			continue
		}
		bob := entity.Clamp(entity.Lerp(-3, 3, (entity.Lerp(0, 1, c.BobTimer))), -3, 3)
		if c.Sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(c.X, c.Y+bob)
			screen.DrawImage(c.Sprite, op)
		} else {
			img := ebiten.NewImage(16, 16)
			img.Fill(color.RGBA{255, 215, 0, 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(c.X, c.Y+bob)
			screen.DrawImage(img, op)
		}
	}

	// Obstacles
	for _, o := range g.obstacles {
		if o.Sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(o.X, o.Y)
			screen.DrawImage(o.Sprite, op)
		} else {
			img := ebiten.NewImage(int(o.W), int(o.H))
			if o.Type == 0 {
				img.Fill(color.RGBA{200, 50, 50, 255})
			} else {
				img.Fill(color.RGBA{100, 100, 200, 255})
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(o.X, o.Y)
			screen.DrawImage(img, op)
		}
	}

	// Player
	if g.player.Invincible <= 0 || g.player.Invincible%4 < 2 {
		sprite := g.playerSpr
		if len(g.walkImgs) > 0 && g.player.OnGround {
			sprite = g.walkImgs[g.player.Frame]
		}
		if sprite != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.player.X, g.player.Y)
			screen.DrawImage(sprite, op)
		} else {
			img := ebiten.NewImage(32, 48)
			img.Fill(color.RGBA{0, 150, 255, 255})
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(g.player.X, g.player.Y)
			screen.DrawImage(img, op)
		}
	}

	// Particles
	for _, p := range g.particles {
		alpha := uint8(p.Life * 255 / p.MaxLife)
		img := ebiten.NewImage(4, 4)
		img.Fill(color.RGBA{p.R, p.G, p.B, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X, p.Y)
		screen.DrawImage(img, op)
	}

	// HUD
	hud.DrawHUD(screen, g.player.HP, g.player.MaxHP, g.player.Score, g.player.Coins)

	// Debug
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Speed: %.0f | Dist: %.0f", g.gen.Speed(), g.distance))
}

// Layout returns screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}
