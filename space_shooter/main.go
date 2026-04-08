// Space Shooter - Go365 Challenge Day 102
// Космический шутер с реальными спрайтами
// 8 апреля 2026

package main

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

//go:embed assets/sprites/*.png
var assetFS embed.FS

// ============================================================================
// КОНСТАНТЫ
// ============================================================================

const (
	ScreenW       = 800
	ScreenH       = 1000
	PlayerSpeed   = 5.0
	BulletSpeed   = 8.0
	EnemyBaseSpeed = 1.5
	FireRate      = 10
)

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func loadPNG(name string) *ebiten.Image {
	data, err := assetFS.ReadFile(name)
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(img)
}

func cropImage(src *ebiten.Image, x, y, w, h int) *ebiten.Image {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	if x < 0 || y < 0 || x+w > bounds.Dx() || y+h > bounds.Dy() {
		return src
	}
	subImg := src.SubImage(image.Rect(x, y, x+w, y+h))
	if subImg == nil {
		return src
	}
	return ebiten.NewImageFromImage(subImg)
}

// ============================================================================
// ТИПЫ
// ============================================================================

type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StateGameOver
)

type Player struct {
	X, Y      float64
	HP        int
	MaxHP     int
	FireTimer int
	Power     int
	Shield    int
	Speed     float64
}

type Bullet struct {
	X, Y, VX, VY float64
	Type         int // 0=player, 1=enemy
	Life         int
}

type Enemy struct {
	X, Y        float64
	VX, VY      float64
	HP          int
	MaxHP       int
	Type        int // 0=basic, 1=fast, 2=tank, 3=shooter
	Size        float64
	FireTimer   int
	Score       int
}

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
	Color        color.RGBA
	Size         float64
}

type Star struct {
	X, Y, Speed float64
	Size        float64
	Brightness  float64
}

type PowerUp struct {
	X, Y float64
	Type int // 0=health, 1=power, 2=shield
	Life int
}

type Game struct {
	State   GameState
	Player  *Player
	Bullets []Bullet
	Enemies []Enemy
	Particles []Particle
	Stars   []Star
	PowerUps []PowerUp

	Score      int
	Wave       int
	HighScore  int
	Kills      int

	GameTime   float64
	SpawnTimer int
	WaveTimer  int
	ShakeTimer float64

	PlayerSprite *ebiten.Image
	EnemySprites []*ebiten.Image
	BeamSprite   *ebiten.Image
	BulletSprite *ebiten.Image
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		State: StateMenu,
		Player: &Player{
			X: ScreenW / 2,
			Y: ScreenH - 100,
			HP: 100, MaxHP: 100,
			Power: 1,
			Speed: PlayerSpeed,
		},
		Bullets:   []Bullet{},
		Enemies:   []Enemy{},
		Particles: []Particle{},
	}

	// Load sprites from shooter folder
	g.loadSprites()

	// Create star field
	for i := 0; i < 150; i++ {
		g.Stars = append(g.Stars, Star{
			X: rand.Float64() * ScreenW,
			Y: rand.Float64() * ScreenH,
			Speed: rand.Float64()*2 + 0.5,
			Size: rand.Float64()*2 + 0.5,
			Brightness: rand.Float64()*0.5 + 0.5,
		})
	}

	// Load high score
	g.loadHighScore()

	return g
}

func (g *Game) loadSprites() {
	// Main ship
	shipImg := loadPNG("assets/sprites/ship2.png")
	if shipImg != nil {
		w := shipImg.Bounds().Dx()
		h := shipImg.Bounds().Dy()
		if w > 64 || h > 64 {
			// Scale down
			scale := 64.0 / float64(w)
			if float64(h)*scale > 64 {
				scale = 64.0 / float64(h)
			}
			resized := ebiten.NewImage(int(float64(w)*scale), int(float64(h)*scale))
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(scale, scale)
			resized.DrawImage(shipImg, op)
			g.PlayerSprite = resized
		} else {
			g.PlayerSprite = shipImg
		}
	}

	// Beam sprites
	beamsImg := loadPNG("assets/sprites/beams.png")
	if beamsImg != nil {
		// Crop beam from spritesheet
		g.BeamSprite = cropImage(beamsImg, 0, 0, 16, 32)
	}

	// Enemy sprites from spritesheet
	sheetImg := loadPNG("assets/sprites/spritesheet.png")
	if sheetImg != nil {
		w := sheetImg.Bounds().Dx()
		tileW := w / 5
		tileH := sheetImg.Bounds().Dy() / 3

		for i := 0; i < 4; i++ {
			col := i % 5
			row := i / 5
			sprite := cropImage(sheetImg, col*tileW, row*tileH, tileW, tileH)
			if sprite != nil {
				// Scale to reasonable size
				sw := sprite.Bounds().Dx()
				sh := sprite.Bounds().Dy()
				if sw > 48 || sh > 48 {
					scale := 48.0 / float64(sw)
					if float64(sh)*scale > 48 {
						scale = 48.0 / float64(sh)
					}
					resized := ebiten.NewImage(int(float64(sw)*scale), int(float64(sh)*scale))
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(scale, scale)
					resized.DrawImage(sprite, op)
					g.EnemySprites = append(g.EnemySprites, resized)
				} else {
					g.EnemySprites = append(g.EnemySprites, sprite)
				}
			}
		}
	}

	// Fallback sprites if loading failed
	if g.PlayerSprite == nil {
		g.PlayerSprite = createFallbackPlayer()
	}
	if len(g.EnemySprites) == 0 {
		g.EnemySprites = []*ebiten.Image{
			createFallbackEnemy(color.RGBA{200, 60, 60, 255}),
			createFallbackEnemy(color.RGBA{220, 150, 50, 255}),
			createFallbackEnemy(color.RGBA{150, 60, 200, 255}),
			createFallbackEnemy(color.RGBA{60, 180, 220, 255}),
		}
	}
}

func createFallbackPlayer() *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 48, 48))
	// Simple triangle ship
	for y := 0; y < 48; y++ {
		for x := 0; x < 48; x++ {
			dx := math.Abs(float64(x-24)) / 12
			dy := float64(y-8) / 40
			if dx+dy <= 1 && dy >= 0 {
				img.Set(x, y, color.RGBA{50, 180, 220, 255})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func createFallbackEnemy(c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dx := float64(x-16)
			dy := float64(y-16)
			if dx*dx+dy*dy <= 256 {
				img.Set(x, y, c)
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func (g *Game) loadHighScore() {
	data, err := os.ReadFile("highscore.txt")
	if err == nil {
		fmt.Sscanf(string(data), "%d", &g.HighScore)
	}
}

func (g *Game) saveHighScore() {
	if g.Score > g.HighScore {
		g.HighScore = g.Score
		os.WriteFile("highscore.txt", []byte(fmt.Sprintf("%d", g.HighScore)), 0644)
	}
}

func (g *Game) Update() error {
	g.GameTime += 1.0 / 60.0

	// Update shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= 1.0 / 60.0
	}

	// Update stars (always)
	for i := range g.Stars {
		g.Stars[i].Y += g.Stars[i].Speed
		if g.Stars[i].Y > ScreenH {
			g.Stars[i].Y = 0
			g.Stars[i].X = rand.Float64() * ScreenW
		}
	}

	switch g.State {
	case StateMenu:
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.isInButton(ScreenW/2-120, 500, 240, 60) {
				g.startGame()
			}
		}

	case StatePlaying:
		g.updatePlaying()

	case StateGameOver:
		g.updateParticles()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if g.isInButton(ScreenW/2-100, 540, 200, 55) {
				g.startGame()
			}
			if g.isInButton(ScreenW/2-100, 610, 200, 55) {
				g.State = StateMenu
			}
		}
	}

	return nil
}

func (g *Game) isInButton(bx, by, bw, bh int) bool {
	mx, my := ebiten.CursorPosition()
	return float64(mx) >= float64(bx) && float64(mx) <= float64(bx+bw) &&
		float64(my) >= float64(by) && float64(my) <= float64(by+bh)
}

func (g *Game) updatePlaying() {
	p := g.Player

	// Movement
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy = 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx = -1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx = 1
	}

	if dx != 0 || dy != 0 {
		length := math.Sqrt(dx*dx + dy*dy)
		p.X += (dx / length) * p.Speed
		p.Y += (dy / length) * p.Speed
	}

	p.X = math.Max(24, math.Min(ScreenW-24, p.X))
	p.Y = math.Max(24, math.Min(ScreenH-24, p.Y))

	// Auto-fire
	if p.FireTimer > 0 {
		p.FireTimer--
	}

	if p.FireTimer <= 0 {
		g.firePlayer()
		p.FireTimer = FireRate / p.Power
		if p.FireTimer < 3 {
			p.FireTimer = 3
		}
	}

	// Update bullets
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := &g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life--

		if b.Life <= 0 || b.X < -20 || b.X > ScreenW+20 || b.Y < -20 || b.Y > ScreenH+20 {
			g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
			continue
		}

		if b.Type == 0 {
			// Player bullet hits enemy
			hit := false
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				e := &g.Enemies[j]
				if dist(b.X, b.Y, e.X, e.Y) < e.Size/2+4 {
					e.HP--
					hit = true
					g.spawnHitParticles(b.X, b.Y)

					if e.HP <= 0 {
						g.Kills++
						g.Score += e.Score
						g.spawnExplosion(e.X, e.Y, e.Type)
						g.ShakeTimer = 0.15

						// Drop powerup
						if rand.Float64() < 0.2 {
							g.PowerUps = append(g.PowerUps, PowerUp{
								X: e.X, Y: e.Y,
								Type: rand.Intn(3),
								Life: 600,
							})
						}

						g.Enemies = append(g.Enemies[:j], g.Enemies[j+1:]...)
					}
					break
				}
			}
			if hit {
				g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
			}
		} else {
			// Enemy bullet hits player
			if dist(b.X, b.Y, p.X, p.Y) < 20 {
				p.HP -= 10
				g.spawnHitParticles(p.X, p.Y)
				g.ShakeTimer = 0.2
				g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)

				if p.HP <= 0 {
					g.State = StateGameOver
					g.spawnExplosion(p.X, p.Y, 99)
					g.saveHighScore()
				}
			}
		}
	}

	// Update enemies
	for _, e := range g.Enemies {
		e.X += e.VX
		e.Y += e.VY

		// Enemy shooting
		if e.Type == 3 {
			e.FireTimer--
			if e.FireTimer <= 0 {
				angle := math.Atan2(p.Y-e.Y, p.X-e.X)
				g.Bullets = append(g.Bullets, Bullet{
					X: e.X, Y: e.Y,
					VX: math.Cos(angle) * 4,
					VY: math.Sin(angle) * 4,
					Type: 1, Life: 120,
				})
				e.FireTimer = 90
			}
		}

		// Collision with player
		if dist(e.X, e.Y, p.X, p.Y) < e.Size/2+18 {
			p.HP -= 20
			g.ShakeTimer = 0.3
			g.spawnExplosion(e.X, e.Y, e.Type)
			e.HP = 0

			if p.HP <= 0 {
				g.State = StateGameOver
				g.spawnExplosion(p.X, p.Y, 99)
				g.saveHighScore()
			}
		}
	}

	// Remove dead enemies
	for i := len(g.Enemies) - 1; i >= 0; i-- {
		if g.Enemies[i].HP <= 0 {
			g.Enemies = append(g.Enemies[:i], g.Enemies[i+1:]...)
		}
		// Off screen
		if g.Enemies[i].Y > ScreenH+50 {
			g.Enemies = append(g.Enemies[:i], g.Enemies[i+1:]...)
		}
	}

	// Update powerups
	for i := len(g.PowerUps) - 1; i >= 0; i-- {
		pu := &g.PowerUps[i]
		pu.Life--
		if pu.Life <= 0 {
			g.PowerUps = append(g.PowerUps[:i], g.PowerUps[i+1:]...)
			continue
		}

		if dist(pu.X, pu.Y, p.X, p.Y) < 30 {
			switch pu.Type {
			case 0:
				p.HP = int(math.Min(float64(p.MaxHP), float64(p.HP+30)))
			case 1:
				p.Power = int(math.Min(5, float64(p.Power+1)))
			case 2:
				p.Shield = 300 // 5 seconds
			}
			g.PowerUps = append(g.PowerUps[:i], g.PowerUps[i+1:]...)
		}
	}

	// Update particles
	g.updateParticles()

	// Spawn enemies
	g.WaveTimer++
	if g.WaveTimer >= 900 {
		g.Wave++
		g.WaveTimer = 0
	}

	g.SpawnTimer++
	spawnRate := int(math.Max(10, 50-float64(g.Wave)*4))
	if g.SpawnTimer >= spawnRate {
		g.spawnEnemy()
		g.SpawnTimer = 0
	}
}

func (g *Game) firePlayer() {
	p := g.Player

	// Center shot
	g.Bullets = append(g.Bullets, Bullet{
		X: p.X, Y: p.Y - 20,
		VX: 0, VY: -BulletSpeed,
		Type: 0, Life: 80,
	})

	// Multi-shot based on power
	if p.Power >= 2 {
		g.Bullets = append(g.Bullets, Bullet{
			X: p.X - 10, Y: p.Y - 15,
			VX: -1, VY: -BulletSpeed,
			Type: 0, Life: 80,
		})
		g.Bullets = append(g.Bullets, Bullet{
			X: p.X + 10, Y: p.Y - 15,
			VX: 1, VY: -BulletSpeed,
			Type: 0, Life: 80,
		})
	}
	if p.Power >= 4 {
		g.Bullets = append(g.Bullets, Bullet{
			X: p.X - 15, Y: p.Y - 10,
			VX: -2, VY: -BulletSpeed,
			Type: 0, Life: 80,
		})
		g.Bullets = append(g.Bullets, Bullet{
			X: p.X + 15, Y: p.Y - 10,
			VX: 2, VY: -BulletSpeed,
			Type: 0, Life: 80,
		})
	}
}

func (g *Game) spawnEnemy() {
	var x float64 = rand.Float64() * (ScreenW - 60) + 30
	y := -30

	eType := 0
	r := rand.Float64()
	if g.Wave >= 5 && r < 0.2 {
		eType = 3 // shooter
	} else if g.Wave >= 3 && r < 0.4 {
		eType = 2 // tank
	} else if r < 0.6 {
		eType = 1 // fast
	}

	hp := 1
	speed := EnemyBaseSpeed
	size := 30.0
	score := 100
	fireTimer := 90.0

	switch eType {
	case 1:
		hp = 1
		speed = EnemyBaseSpeed * 2
		size = 24
		score = 150
	case 2:
		hp = 3 + g.Wave/3
		speed = EnemyBaseSpeed * 0.5
		size = 40
		score = 300
	case 3:
		hp = 2
		speed = EnemyBaseSpeed * 0.7
		size = 32
		score = 250
	}

	vx := math.Sin(g.GameTime*2) * 0.5
	if eType == 1 {
		vx = math.Sin(g.GameTime*3) * 1.5
	}

	g.Enemies = append(g.Enemies, Enemy{
		X: x, Y: float64(y),
		VX: vx, VY: float64(speed),
		HP: hp, MaxHP: hp,
		Type: eType,
		Size: size,
		FireTimer: int(fireTimer),
		Score: score,
	})
}

func (g *Game) spawnHitParticles(x, y float64) {
	for i := 0; i < 8; i++ {
		angle := rand.Float64() * 6.2832
		speed := 2 + rand.Float64()*3
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * speed,
			VY: math.Sin(angle) * speed,
			Life: 0.4, MaxLife: 0.4,
			Color: color.RGBA{255, 200, 50, 255},
			Size: 2 + rand.Float64()*2,
		})
	}
}

func (g *Game) spawnExplosion(x, y float64, eType int) {
	count := 15
	if eType == 99 {
		count = 40 // player death
	}

	for i := 0; i < count; i++ {
		angle := rand.Float64() * 6.2832
		speed := 3 + rand.Float64()*6
		c := color.RGBA{255, 150, 50, 255}
		switch eType {
		case 1:
			c = color.RGBA{255, 220, 80, 255}
		case 2:
			c = color.RGBA{200, 100, 255, 255}
		case 3:
			c = color.RGBA{80, 200, 255, 255}
		case 99:
			c = color.RGBA{255, 80, 80, 255}
		}
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: math.Cos(angle) * speed,
			VY: math.Sin(angle) * speed - 1,
			Life: 0.8 + rand.Float64()*0.5,
			MaxLife: 1.3,
			Color: c,
			Size: 3 + rand.Float64()*5,
		})
	}
}

func (g *Game) updateParticles() {
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.03
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Player = &Player{
		X: ScreenW / 2, Y: ScreenH - 100,
		HP: 100, MaxHP: 100,
		Power: 1, Speed: PlayerSpeed,
	}
	g.Bullets = []Bullet{}
	g.Enemies = []Enemy{}
	g.Particles = []Particle{}
	g.PowerUps = []PowerUp{}
	g.Score = 0
	g.Wave = 1
	g.Kills = 0
	g.WaveTimer = 0
	g.SpawnTimer = 0
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background - dark space
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{5, 5, 15, 255}, false)

	// Stars
	for _, star := range g.Stars {
		alpha := uint8(star.Brightness * 200)
		s := int(star.Size)
		if s < 1 { s = 1 }
		starImg := ebiten.NewImage(s*2, s*2)
		starImg.Fill(color.RGBA{220, 220, 255, alpha})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(star.X, star.Y)
		screen.DrawImage(starImg, op)
	}

	// Screen shake offset
	var shakeX, shakeY float64
	if g.ShakeTimer > 0 {
		shakeX = (rand.Float64() - 0.5) * 6
		shakeY = (rand.Float64() - 0.5) * 6
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(shakeX, shakeY)

	switch g.State {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawGame(screen, shakeX, shakeY)
	case StateGameOver:
		g.drawGame(screen, shakeX, shakeY)
		g.drawGameOver(screen)
	}

	// Particles (on top)
	for _, p := range g.Particles {
		size := int(p.Size * (p.Life / p.MaxLife))
		if size < 1 { continue }
		alpha := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, alpha}
		pImg := ebiten.NewImage(size, size)
		pImg.Fill(c)
		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(p.X-float64(size)/2, p.Y-float64(size)/2)
		screen.DrawImage(pImg, pop)
	}
}

func (g *Game) drawGame(screen *ebiten.Image, shakeX, shakeY float64) {
	p := g.Player

	// Powerups
	for _, pu := range g.PowerUps {
		alpha := 255
		if pu.Life < 120 {
			alpha = int(float64(pu.Life) / 120 * 255)
		}

		var c color.RGBA
		switch pu.Type {
		case 0: c = color.RGBA{50, 255, 80, uint8(alpha)}
		case 1: c = color.RGBA{255, 200, 50, uint8(alpha)}
		case 2: c = color.RGBA{100, 180, 255, uint8(alpha)}
		}

		puImg := ebiten.NewImage(20, 20)
		for y := 0; y < 20; y++ {
			for x := 0; x < 20; x++ {
				dx := float64(x-10)
				dy := float64(y-10)
				if dx*dx+dy*dy <= 100 {
					puImg.Set(x, y, c)
				}
			}
		}

		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(pu.X-10+shakeX, pu.Y-10+shakeY)
		screen.DrawImage(puImg, pop)
	}

	// Bullets
	for _, b := range g.Bullets {
		if b.Type == 0 {
			// Player bullet - yellow
			bImg := ebiten.NewImage(4, 12)
			bImg.Fill(color.RGBA{255, 255, 100, 255})
			bop := &ebiten.DrawImageOptions{}
			bop.GeoM.Translate(b.X-2+shakeX, b.Y-6+shakeY)
			screen.DrawImage(bImg, bop)
		} else {
			// Enemy bullet - red
			bImg := ebiten.NewImage(6, 6)
			bImg.Fill(color.RGBA{255, 80, 80, 255})
			bop := &ebiten.DrawImageOptions{}
			bop.GeoM.Translate(b.X-3+shakeX, b.Y-3+shakeY)
			screen.DrawImage(bImg, bop)
		}
	}

	// Enemies
	for _, e := range g.Enemies {
		spriteIdx := e.Type
		if spriteIdx >= len(g.EnemySprites) {
			spriteIdx = 0
		}

		eop := &ebiten.DrawImageOptions{}
		eop.GeoM.Translate(e.X-float64(e.Size)/2+shakeX, e.Y-float64(e.Size)/2+shakeY)
		screen.DrawImage(g.EnemySprites[spriteIdx], eop)

		// HP bar
		if e.HP < e.MaxHP {
			barW := e.Size
			ratio := float64(e.HP) / float64(e.MaxHP)
			vector.DrawFilledRect(screen, float32(e.X-barW/2+shakeX), float32(e.Y-e.Size/2-8+shakeY), float32(barW), 4, color.RGBA{80, 0, 0, 200}, false)
			vector.DrawFilledRect(screen, float32(e.X-barW/2+shakeX), float32(e.Y-e.Size/2-8+shakeY), float32(barW)*float32(ratio), 4, color.RGBA{255, 50, 50, 255}, false)
		}
	}

	// Player
	if g.PlayerSprite != nil {
		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(p.X-24+shakeX, p.Y-24+shakeY)
		screen.DrawImage(g.PlayerSprite, pop)
	}

	// Shield effect
	if p.Shield > 0 {
		p.Shield--
		shieldAlpha := uint8(80 + 40*math.Sin(g.GameTime*5))
		vector.StrokeCircle(screen, float32(p.X+shakeX), float32(p.Y+shakeY), 28, 2, color.RGBA{100, 180, 255, shieldAlpha}, false)
	}

	// HUD
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 55, color.RGBA{15, 18, 35, 230}, false)
	vector.StrokeLine(screen, 0, 55, ScreenW, 55, 2, color.RGBA{80, 180, 255, 255}, false)

	// HP bar
	hpW := 180
	ratio := float64(p.HP) / float64(p.MaxHP)
	vector.DrawFilledRect(screen, 15, 12, float32(hpW), 18, color.RGBA{60, 0, 0, 200}, false)
	vector.DrawFilledRect(screen, 15, 12, float32(hpW)*float32(ratio), 18, color.RGBA{50, 220, 80, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", p.HP, p.MaxHP), 20, 14)

	// Stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 220, 14)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 400, 14)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СИЛА: %d", p.Power), 540, 14)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("РЕКОРД: %d", g.HighScore), 640, 14)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Title
	ebitenutil.DebugPrintAt(screen, "🚀 SPACE SHOOTER 🚀", ScreenW/2-160, 280)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 102", ScreenW/2-130, 330)

	vector.StrokeLine(screen, ScreenW/2-180, 370, ScreenW/2+180, 370, 2, color.RGBA{100, 180, 255, 255}, false)

	// Ship preview
	if g.PlayerSprite != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(ScreenW/2-32, 400)
		screen.DrawImage(g.PlayerSprite, op)
	}

	g.drawButton(screen, "▶  НАЧАТЬ ИГРУ", ScreenW/2-120, 500, 240, 60)

	ebitenutil.DebugPrintAt(screen, "WASD / Стрелки - Движение", ScreenW/2-120, 600)
	ebitenutil.DebugPrintAt(screen, "Автострельба", ScreenW/2-70, 630)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Рекорд: %d", g.HighScore), ScreenW/2-70, 670)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, ScreenH/2-140, ScreenW, 280, color.RGBA{10, 10, 20, 220}, false)

	ebitenutil.DebugPrintAt(screen, "💀 GAME OVER 💀", ScreenW/2-120, ScreenH/2-110)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %d | Волна: %d | Убито: %d",
		g.Score, g.Wave, g.Kills), ScreenW/2-170, ScreenH/2-50)
	if g.Score >= g.HighScore && g.Score > 0 {
		ebitenutil.DebugPrintAt(screen, "🏆 НОВЫЙ РЕКОРД! 🏆", ScreenW/2-110, ScreenH/2-15)
	}

	g.drawButton(screen, "🔄  ЗАНОВО", ScreenW/2-100, ScreenH/2+20, 200, 55)
	g.drawButton(screen, "←  МЕНЮ", ScreenW/2-100, ScreenH/2+90, 200, 55)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	hover := g.isInButton(int(x), int(y), w, h)

	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{50, 80, 130, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{25, 40, 65, 255}, false)
	}

	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{100, 180, 255, 255})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)

	ebitenutil.DebugPrintAt(screen, text, int(x)+20, int(y)+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Space Shooter - Go365 Day 102")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
