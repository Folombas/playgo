package engine

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"city_platformer/internal/entity"
	"city_platformer/internal/hud"
	"city_platformer/internal/input"
	"city_platformer/internal/level"
	"city_platformer/internal/sprite"
)

// GameState represents the current state of the game.
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// Game is the main game struct.
type Game struct {
	state    GameState
	player   *entity.Player
	chunks   []*level.Chunk
	gen      *level.Generator
	input    *input.Manager
	hud      *hud.HUD
	loader   *sprite.Loader
	cameraX  float64
	score    int
	distance float64
	maxDist  float64

	// Background
	bgSky    *ebiten.Image
	bgClouds []*ebiten.Image

	// Sprites
	heartFull  *ebiten.Image
	heartEmpty *ebiten.Image
	coinSprite *ebiten.Image

	// Particles
	particles []Particle
}

// Particle represents a visual effect.
type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   int
	MaxLife int
	Color  color.Color
}

// NewGame creates a new game instance.
func NewGame() *Game {
	g := &Game{
		state:   StateMenu,
		input:   input.NewManager(),
		gen:     level.NewGenerator(42),
		chunks:  make([]*level.Chunk, 0),
		particles: make([]Particle, 0),
	}

	// Initialize loader
	g.loader = sprite.NewLoader("assets/sprites/PlatformerComplete")

	// Create player
	g.player = entity.NewPlayer(100, 400)

	// Load essential sprites
	g.loadSprites()

	// Generate initial chunks
	for i := 0; i < 5; i++ {
		chunk := g.gen.Generate(float64(i) * 800)
		g.chunks = append(g.chunks, chunk)
	}

	// Create HUD
	g.hud = hud.NewHUD(g.heartFull, g.heartEmpty, g.coinSprite)

	return g
}

func (g *Game) loadSprites() {
	// Load HUD sprites
	g.heartFull, _ = g.loader.Load("Base pack/HUD/hud_heartFull.png")
	g.heartEmpty, _ = g.loader.Load("Base pack/HUD/hud_heartEmpty.png")
	g.coinSprite, _ = g.loader.Load("Base pack/Items/coinGold.png")

	// Load background
	g.bgSky, _ = g.loader.Load("Base pack/bg.png")

	// Load player sprites
	if img, err := g.loader.Load("Base pack/Player/p1_stand.png"); err == nil {
		g.player.Sprite = img
		g.player.IdleSprite = img
	}
	g.player.JumpSprite, _ = g.loader.Load("Base pack/Player/p1_jump.png")
	g.player.HurtSprite, _ = g.loader.Load("Base pack/Player/p1_hurt.png")

	// Load walk animation
	g.player.WalkSprites = make([]*ebiten.Image, 11)
	for i := 0; i < 11; i++ {
		path := fmt.Sprintf("Base pack/Player/p1_walk/PNG/p1_walk%02d.png", i+1)
		if img, err := g.loader.Load(path); err == nil {
			g.player.WalkSprites[i] = img
		}
	}

	// Load enemy sprites (for reference, actual enemies load their own)
	// Load cloud sprites (disabled for now - no cloud sprites found)
	g.bgClouds = make([]*ebiten.Image, 0)
}

// Update updates the game state.
func (g *Game) Update() error {
	g.input.Update()

	switch g.state {
	case StateMenu:
		if g.input.IsJustPressed(input.ActionJump) || g.input.IsJustPressed(input.ActionRestart) {
			g.state = StatePlaying
		}
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		if g.input.IsJustPressed(input.ActionPause) {
			g.state = StatePlaying
		}
	case StateGameOver:
		if g.input.IsJustPressed(input.ActionRestart) {
			g.resetGame()
		}
	}

	return nil
}

func (g *Game) updatePlaying() {
	// Pause
	if g.input.IsJustPressed(input.ActionPause) {
		g.state = StatePaused
		return
	}

	// Player input
	moveSpeed := 400.0
	if g.input.IsPressed(input.ActionMoveLeft) {
		g.player.VX -= moveSpeed * 0.15
		g.player.FacingRight = false
	}
	if g.input.IsPressed(input.ActionMoveRight) {
		g.player.VX += moveSpeed * 0.15
		g.player.FacingRight = true
	}
	if g.input.IsJustPressed(input.ActionJump) {
		g.player.Jump(450)
		// Spawn jump particles
		g.spawnParticles(g.player.X+g.player.Width/2, g.player.Y+g.player.Height, 5, color.RGBA{200, 200, 200, 255})
	}

	// Update player
	g.player.Update(1000, 0.85)

	// Platform collision
	g.player.OnGround = false
	for _, chunk := range g.chunks {
		for _, plat := range chunk.Platforms {
			if entity.CheckAABB(
				g.player.X, g.player.Y, g.player.Width, g.player.Height,
				plat.X, plat.Y, plat.Width, plat.Height,
			) {
				// Resolve collision
				overlapLeft := (g.player.X + g.player.Width) - plat.X
				overlapRight := (plat.X + plat.Width) - g.player.X
				overlapTop := (g.player.Y + g.player.Height) - plat.Y
				overlapBottom := (plat.Y + plat.Height) - g.player.Y

				minOverlap := math.Min(math.Min(overlapLeft, overlapRight), math.Min(overlapTop, overlapBottom))

				switch {
				case minOverlap == overlapTop && g.player.VY >= 0:
					g.player.Y = plat.Y - g.player.Height
					g.player.VY = 0
					g.player.ResetGround()
				case minOverlap == overlapBottom && g.player.VY < 0:
					g.player.Y = plat.Y + plat.Height
					g.player.VY = 0
				case minOverlap == overlapLeft:
					g.player.X = plat.X - g.player.Width
					g.player.VX = 0
				case minOverlap == overlapRight:
					g.player.X = plat.X + plat.Width
					g.player.VX = 0
				}
			}
		}

		// Enemy collision and update
		for _, enemy := range chunk.Enemies {
			if !enemy.Alive {
				continue
			}
			enemy.Update(800)

			// Enemy-platform collision
			for _, plat := range chunk.Platforms {
				if entity.CheckAABB(
					enemy.X, enemy.Y, enemy.Width, enemy.Height,
					plat.X, plat.Y, plat.Width, plat.Height,
				) {
					if enemy.Y+enemy.Height > plat.Y && enemy.Y < plat.Y {
						enemy.Y = plat.Y - enemy.Height
						enemy.VY = 0
					}
				}
			}

			// Player-enemy collision
			if entity.CheckAABB(
				g.player.X, g.player.Y, g.player.Width, g.player.Height,
				enemy.X, enemy.Y, enemy.Width, enemy.Height,
			) {
				// Check if stomping from above
				if g.player.VY > 0 && g.player.Y+g.player.Height < enemy.Y+enemy.Height/2 {
					enemy.TakeDamage(1)
					g.player.VY = -300 // Bounce
					g.player.Score += 50
					g.spawnParticles(enemy.X+enemy.Width/2, enemy.Y, 8, color.RGBA{255, 255, 0, 255})
				} else {
					g.player.TakeDamage(1)
					g.spawnParticles(g.player.X+g.player.Width/2, g.player.Y+g.player.Height/2, 10, color.RGBA{255, 0, 0, 255})
				}
			}
		}

		// Item collection
		for _, item := range chunk.Items {
			if item.Collected {
				continue
			}
			item.Update(1.0 / 60.0)
			if entity.CheckAABB(
				g.player.X, g.player.Y, g.player.Width, g.player.Height,
				item.X, item.Y+item.FloatOffset, item.Width, item.Height,
			) {
				item.Collected = true
				g.player.Score += item.Value
				g.player.Coins += item.Value / 10
				if item.Type == entity.ItemHeart && g.player.HP < g.player.MaxHP {
					g.player.HP++
				}
				g.spawnParticles(item.X+item.Width/2, item.Y, 6, color.RGBA{255, 255, 100, 255})
			}
		}
	}

	// Update distance
	g.distance = g.player.X / 100
	if g.distance > g.maxDist {
		g.maxDist = g.distance
	}

	// Check death
	if g.player.HP <= 0 {
		g.state = StateGameOver
		return
	}

	// Fall into pit
	if g.player.Y > 800 {
		g.player.HP = 0
		g.state = StateGameOver
		return
	}

	// Camera follow
	targetCamX := g.player.X - 400
	g.cameraX += (targetCamX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}

	// Generate new chunks as needed
	rightmostChunk := g.chunks[len(g.chunks)-1]
	if g.cameraX+1280 > rightmostChunk.X {
		newX := rightmostChunk.X + 800
		chunk := g.gen.Generate(newX)
		g.chunks = append(g.chunks, chunk)

		// Remove old chunks to save memory
		if len(g.chunks) > 10 {
			g.chunks = g.chunks[1:]
		}

		// Increase difficulty over time
		g.gen.IncreaseDifficulty()
	}

	// Update particles
	g.updateParticles()
}

func (g *Game) spawnParticles(x, y float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		p := Particle{
			X:       x,
			Y:       y,
			VX:      (float64(i) - float64(count)/2) * 2,
			VY:      -float64(i) * 1.5,
			Life:    30,
			MaxLife: 30,
			Color:   c,
		}
		g.particles = append(g.particles, p)
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 5 // gravity
		p.Life--
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) resetGame() {
	g.player = entity.NewPlayer(100, 400)
	g.chunks = make([]*level.Chunk, 0)
	g.gen = level.NewGenerator(42)
	g.cameraX = 0
	g.score = 0
	g.distance = 0
	g.maxDist = 0
	g.particles = make([]Particle, 0)

	// Reload sprites
	g.loadSprites()
	g.hud = hud.NewHUD(g.heartFull, g.heartEmpty, g.coinSprite)

	for i := 0; i < 5; i++ {
		chunk := g.gen.Generate(float64(i) * 800)
		g.chunks = append(g.chunks, chunk)
	}

	g.state = StatePlaying
}

// Draw renders the game.
func (g *Game) Draw(screen *ebiten.Image) {
	// Clear
	screen.Fill(color.RGBA{135, 206, 235, 255}) // Sky blue

	switch g.state {
	case StateMenu:
		hud.DrawMenu(screen)
	case StatePlaying, StatePaused:
		g.drawGame(screen)
		if g.state == StatePaused {
			hud.DrawPaused(screen)
		}
	case StateGameOver:
		g.drawGame(screen)
		hud.DrawGameOver(screen, g.player.Score)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	// Draw background
	if g.bgSky != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, 0)
		screen.DrawImage(g.bgSky, op)
	}

	// Draw clouds (parallax)
	for i, cloud := range g.bgClouds {
		if cloud == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		parallaxX := float64(i*300) - (g.cameraX * 0.3)
		op.GeoM.Translate(parallaxX, 50+float64(i)*30)
		screen.DrawImage(cloud, op)
	}

	// Apply camera transform
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(-g.cameraX, 0)

	// Draw platforms
	for _, chunk := range g.chunks {
		for _, plat := range chunk.Platforms {
			if plat.Sprite != nil {
				// Tile the sprite
				for tx := 0; tx < int(plat.Width)/plat.TileWidth; tx++ {
					for ty := 0; ty < int(plat.Height)/plat.TileHeight; ty++ {
						p := &ebiten.DrawImageOptions{}
						p.GeoM.Translate(plat.X+float64(tx*plat.TileWidth), plat.Y+float64(ty*plat.TileHeight))
						screen.DrawImage(plat.Sprite, p)
					}
				}
			} else {
				// Draw colored rectangle
				p := &ebiten.DrawImageOptions{}
				p.GeoM.Translate(plat.X, plat.Y)
				// Create a simple tile from ground sprite
				groundTile := ebiten.NewImage(int(plat.Width), int(plat.Height))
				groundTile.Fill(color.RGBA{100, 180, 80, 255}) // Green grass
				screen.DrawImage(groundTile, p)
			}
		}

		// Draw items
		for _, item := range chunk.Items {
			if item.Collected {
				continue
			}
			if item.Sprite != nil {
				p := &ebiten.DrawImageOptions{}
				p.GeoM.Translate(item.X, item.Y+item.FloatOffset)
				screen.DrawImage(item.Sprite, p)
			} else {
				// Fallback: draw colored rect
				rect := ebiten.NewImage(int(item.Width), int(item.Height))
				switch item.Type {
				case entity.ItemCoin:
					rect.Fill(color.RGBA{255, 215, 0, 255})
				case entity.ItemGem:
					rect.Fill(color.RGBA{0, 200, 255, 255})
				case entity.ItemStar:
					rect.Fill(color.RGBA{255, 255, 0, 255})
				case entity.ItemHeart:
					rect.Fill(color.RGBA{255, 50, 50, 255})
				}
				p := &ebiten.DrawImageOptions{}
				p.GeoM.Translate(item.X, item.Y+item.FloatOffset)
				screen.DrawImage(rect, p)
			}
		}

		// Draw enemies
		for _, enemy := range chunk.Enemies {
			if !enemy.Alive {
				continue
			}
			sprite := enemy.CurrentSprite()
			if sprite != nil {
				p := &ebiten.DrawImageOptions{}
				if !enemy.FacingRight {
					p.GeoM.Scale(-1, 1)
					p.GeoM.Translate(enemy.X+enemy.Width, enemy.Y)
				} else {
					p.GeoM.Translate(enemy.X, enemy.Y)
				}
				screen.DrawImage(sprite, p)
			}
		}
	}

	// Draw player
	if g.player.InvincibleTime <= 0 || g.player.InvincibleTime%4 < 2 {
		sprite := g.player.CurrentSprite()
		if sprite != nil {
			p := &ebiten.DrawImageOptions{}
			if !g.player.FacingRight {
				p.GeoM.Scale(-1, 1)
				p.GeoM.Translate(g.player.X+g.player.Width, g.player.Y)
			} else {
				p.GeoM.Translate(g.player.X, g.player.Y)
			}
			screen.DrawImage(sprite, p)
		}
	}

	// Draw particles
	for _, p := range g.particles {
		alpha := uint8(float64(p.Life) / float64(p.MaxLife) * 255)
		rect := ebiten.NewImage(4, 4)
		c := p.Color
		rect.Fill(color.RGBA{c.(color.RGBA).R, c.(color.RGBA).G, c.(color.RGBA).B, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-g.cameraX, p.Y)
		screen.DrawImage(rect, op)
	}

	// Draw HUD
	g.hud.Draw(screen, g.player.HP, g.player.MaxHP, g.player.Score, g.player.Coins)

	// Debug info
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Pos: (%.0f, %.0f) State: %v", g.player.X, g.player.Y, g.player.State))
}

// Layout returns the screen size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}
