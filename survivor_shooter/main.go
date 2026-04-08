// Survivor Shooter - Go365 Challenge Day 102
// Vampire Survivors стиль - персонаж выживает против орд
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
	"path/filepath"
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
	ScreenW     = 1280
	ScreenH     = 800
	WorldW      = 3000
	WorldH      = 3000
	PlayerSpeed = 3.5
	FireRate    = 20
)

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func loadSprite(name string) *ebiten.Image {
	execPath, _ := os.Getwd()
	path := filepath.Join(execPath, "assets", "sprites", name)
	data, err := os.ReadFile(path)
	if err != nil {
		// Try embedded FS
		data, err = assetFS.ReadFile("assets/sprites/" + name)
		if err != nil {
			return nil
		}
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	return ebiten.NewImageFromImage(img)
}

func cropTilesheet(img *ebiten.Image, frame, cols, tileW, tileH int) *ebiten.Image {
	if img == nil {
		return nil
	}
	x := (frame % cols) * tileW
	y := (frame / cols) * tileH
	sub := img.SubImage(image.Rect(x, y, x+tileW, y+tileH))
	return ebiten.NewImageFromImage(sub)
}

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
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

type Animation struct {
	Frames    []*ebiten.Image
	Frame     int
	Tick      int
	TickSpeed int
}

func (a *Animation) Update() {
	if a == nil || len(a.Frames) == 0 {
		return
	}
	a.Tick++
	if a.Tick >= a.TickSpeed {
		a.Tick = 0
		a.Frame = (a.Frame + 1) % len(a.Frames)
	}
}

func (a *Animation) Image() *ebiten.Image {
	if a == nil || len(a.Frames) == 0 {
		return nil
	}
	return a.Frames[a.Frame]
}

type Player struct {
	X, Y      float64
	HP        int
	MaxHP     int
	Speed     float64
	Power     int
	FireTimer int
	XP        int
	Level     int
	XPToNext  int
	Anim      *Animation
	FireAnim  *Animation
	IsFiring  bool
}

type Enemy struct {
	X, Y    float64
	HP      int
	MaxHP   int
	Speed   float64
	Type    int // 0=zombie, 1=fast, 2=tank, 3=shooter
	Size    float64
	Damage  int
	Anim    *Animation
}

type Bullet struct {
	X, Y, VX, VY float64
	Damage       int
	Life         int
	FromPlayer   bool
}

type XPorb struct {
	X, Y  float64
	Value int
	Life  int
}

type DamageNumber struct {
	X, Y   float64
	Value  int
	Life   float64
	Color  color.RGBA
}

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
	Color        color.RGBA
	Size         float64
}

type Game struct {
	State       GameState
	Player      *Player
	Enemies     []Enemy
	Bullets     []Bullet
	XPOrbs      []XPorb
	Particles   []Particle
	DmgNumbers  []DamageNumber

	CameraX, CameraY float64

	Score    int
	Wave     int
	Time     int
	Kills    int
	HighScore int

	SpawnTimer int
	WaveTimer  int
	GameTime   float64

	// Sprite sheets
	IdleSheet     *ebiten.Image
	RunSheet      *ebiten.Image
	IdleFireSheet *ebiten.Image
	RunFireSheet  *ebiten.Image
	DeathSheet    *ebiten.Image
	JumpSheet     *ebiten.Image

	// Zombie animation
	ZombieImgs []*ebiten.Image
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		State: StateMenu,
		Player: &Player{
			X: WorldW / 2,
			Y: WorldH / 2,
			HP: 100, MaxHP: 100,
			Speed: PlayerSpeed,
			Power: 1,
			Level: 1,
			XPToNext: 50,
		},
		Enemies:    []Enemy{},
		Bullets:    []Bullet{},
		XPOrbs:     []XPorb{},
		Particles:  []Particle{},
		DmgNumbers: []DamageNumber{},
	}

	g.loadAssets()
	g.loadHighScore()

	// Stars for menu
	for i := 0; i < 100; i++ {
		g.Particles = append(g.Particles, Particle{
			X: rand.Float64() * ScreenW,
			Y: rand.Float64() * ScreenH,
			VX: 0, VY: 0,
			Life: 9999, MaxLife: 9999,
			Color: color.RGBA{200, 200, 255, 100},
			Size: rand.Float64()*2 + 0.5,
		})
	}

	return g
}

func (g *Game) loadAssets() {
	// Load sprite sheets
	g.IdleSheet = loadSprite("idle_spritesheet.png")
	g.RunSheet = loadSprite("run_spritesheet.png")
	g.IdleFireSheet = loadSprite("idle-fire_spritesheet.png")
	g.RunFireSheet = loadSprite("run-fire_spritesheet.png")
	g.DeathSheet = loadSprite("death_spritesheet.png")
	g.JumpSheet = loadSprite("jump_spritesheet.png")

	// Create player animations
	if g.RunSheet != nil {
		w := g.RunSheet.Bounds().Dx()
		h := g.RunSheet.Bounds().Dy()
		cols := 8
		rows := 2
		tileW := w / cols
		tileH := h / rows

		// Run animation (first row)
		for i := 0; i < cols; i++ {
			frame := cropTilesheet(g.RunSheet, i, cols, tileW, tileH)
			if frame != nil {
				// Scale up
				scaled := ebiten.NewImage(tileW*2, tileH*2)
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Scale(2, 2)
				scaled.DrawImage(frame, op)
				g.Player.Anim.Frames = append(g.Player.Anim.Frames, scaled)
			}
		}
		g.Player.Anim.TickSpeed = 6

		// Fire animation
		if g.RunFireSheet != nil {
			w2 := g.RunFireSheet.Bounds().Dx()
			cols2 := 8
			tileW2 := w2 / cols2
			tileH2 := g.RunFireSheet.Bounds().Dy() / 2

			for i := 0; i < cols2; i++ {
				frame := cropTilesheet(g.RunFireSheet, i, cols2, tileW2, tileH2)
				if frame != nil {
					scaled := ebiten.NewImage(tileW2*2, tileH2*2)
					op := &ebiten.DrawImageOptions{}
					op.GeoM.Scale(2, 2)
					scaled.DrawImage(frame, op)
					g.Player.FireAnim.Frames = append(g.Player.FireAnim.Frames, scaled)
				}
			}
			g.Player.FireAnim.TickSpeed = 6
		}
	}

	// Create zombie enemies
	colors := []color.RGBA{
		{100, 150, 80, 255},   // green zombie
		{180, 100, 50, 255},   // brown fast
		{120, 80, 140, 255},   // purple tank
		{80, 120, 160, 255},   // blue shooter
	}
	for _, c := range colors {
		g.ZombieImgs = append(g.ZombieImgs, createZombieSprite(c))
	}
}

func createZombieSprite(c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 40))

	// Body
	for y := 10; y < 36; y++ {
		for x := 8; x < 24; x++ {
			img.Set(x, y, c)
		}
	}
	// Head
	for y := 0; y < 14; y++ {
		for x := 10; x < 22; x++ {
			dx := float64(x-16)
			dy := float64(y-7)
			if dx*dx+dy*dy <= 36 {
				img.Set(x, y, c)
			}
		}
	}
	// Eyes
	img.Set(13, 6, color.RGBA{255, 50, 50, 255})
	img.Set(18, 6, color.RGBA{255, 50, 50, 255})

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
			if g.isInButton(ScreenW/2-100, 520, 200, 55) {
				g.startGame()
			}
			if g.isInButton(ScreenW/2-100, 590, 200, 55) {
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

	p.IsFiring = false

	if dx != 0 || dy != 0 {
		length := math.Sqrt(dx*dx + dy*dy)
		p.X += (dx / length) * p.Speed
		p.Y += (dy / length) * p.Speed

		// Clamp to world
		p.X = math.Max(30, math.Min(WorldW-30, p.X))
		p.Y = math.Max(30, math.Min(WorldH-30, p.Y))

		p.Anim.Update()
	}

	// Auto-fire nearest enemy
	if p.FireTimer > 0 {
		p.FireTimer--
	}

	if len(g.Enemies) > 0 && p.FireTimer <= 0 {
		// Find nearest enemy
		var nearest *Enemy
		minDist := 400.0 // fire range

		for i := range g.Enemies {
			e := &g.Enemies[i]
			d := dist(p.X, p.Y, e.X, e.Y)
			if d < minDist {
				minDist = d
				nearest = e
			}
		}

		if nearest != nil {
			angle := math.Atan2(nearest.Y-p.Y, nearest.X-p.X)
			g.Bullets = append(g.Bullets, Bullet{
				X: p.X, Y: p.Y - 10,
				VX: math.Cos(angle) * 7,
				VY: math.Sin(angle) * 7,
				Damage: p.Power * 10,
				Life: 50,
				FromPlayer: true,
			})
			p.FireTimer = FireRate
			p.IsFiring = true

			// Muzzle flash
			for i := 0; i < 5; i++ {
				a := angle + (rand.Float64()-0.5)*0.6
				s := 2 + rand.Float64()*3
				g.Particles = append(g.Particles, Particle{
					X: p.X + math.Cos(angle)*20,
					Y: p.Y - 10 + math.Sin(angle)*20,
					VX: math.Cos(a)*s, VY: math.Sin(a)*s,
					Life: 0.2, MaxLife: 0.2,
					Color: color.RGBA{255, 220, 80, 255},
					Size: 2 + rand.Float64()*2,
				})
			}
		}
	}

	// Camera follow
	g.CameraX = p.X - ScreenW/2
	g.CameraY = p.Y - ScreenH/2
	g.CameraX = math.Max(0, math.Min(WorldW-ScreenW, g.CameraX))
	g.CameraY = math.Max(0, math.Min(WorldH-ScreenH, g.CameraY))

	// Update enemies
	for _, e := range g.Enemies {
		angle := math.Atan2(p.Y-e.Y, p.X-e.X)
		e.X += math.Cos(angle) * e.Speed
		e.Y += math.Sin(angle) * e.Speed

		// Hit player
		if dist(e.X, e.Y, p.X, p.Y) < e.Size/2+15 {
			p.HP -= e.Damage
			// Knockback
			p.X += math.Cos(angle) * 20
			p.Y += math.Sin(angle) * 20

			g.addDmgNumber(p.X, p.Y, e.Damage, color.RGBA{255, 80, 80, 255})

			if p.HP <= 0 {
				g.State = StateGameOver
				g.saveHighScore()
				g.spawnExplosion(p.X, p.Y, color.RGBA{255, 100, 100, 255}, 30)
			}
		}
	}

	// Update bullets
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := &g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life--

		if b.Life <= 0 {
			g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
			continue
		}

		if b.FromPlayer {
			// Hit enemies
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				e := &g.Enemies[j]
				if dist(b.X, b.Y, e.X, e.Y) < e.Size/2+5 {
					e.HP -= b.Damage
					g.addDmgNumber(e.X, e.Y-20, b.Damage, color.RGBA{255, 255, 100, 255})

					// Hit particles
					for k := 0; k < 5; k++ {
						a := rand.Float64() * 6.2832
						s := 2 + rand.Float64()*3
						g.Particles = append(g.Particles, Particle{
							X: b.X, Y: b.Y,
							VX: math.Cos(a)*s, VY: math.Sin(a)*s,
							Life: 0.3, MaxLife: 0.3,
							Color: color.RGBA{255, 180, 50, 255},
							Size: 2 + rand.Float64()*2,
						})
					}

					if e.HP <= 0 {
						g.Kills++
						g.Score += e.Type*100 + 50

						// Drop XP
						xpVal := 10 + e.Type*10
						g.XPOrbs = append(g.XPOrbs, XPorb{
							X: e.X, Y: e.Y,
							Value: xpVal,
							Life: 600,
						})

						// Death particles
						g.spawnExplosion(e.X, e.Y, color.RGBA{150, 200, 100, 255}, 15)

						g.Enemies = append(g.Enemies[:j], g.Enemies[j+1:]...)
					}

					g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
					break
				}
			}
		}
	}

	// Update XP orbs
	for i := len(g.XPOrbs) - 1; i >= 0; i-- {
		xp := &g.XPOrbs[i]
		xp.Life--
		if xp.Life <= 0 {
			g.XPOrbs = append(g.XPOrbs[:i], g.XPOrbs[i+1:]...)
			continue
		}

		// Magnet towards player
		d := dist(xp.X, xp.Y, p.X, p.Y)
		if d < 100 {
			angle := math.Atan2(p.Y-xp.Y, p.X-xp.X)
			speed := 5.0
			xp.X += math.Cos(angle) * speed
			xp.Y += math.Sin(angle) * speed
		}

		if d < 20 {
			p.XP += xp.Value
			if p.XP >= p.XPToNext {
				p.Level++
				p.XP -= p.XPToNext
				p.XPToNext = int(float64(p.XPToNext) * 1.3)
				p.Power++
				p.HP = p.MaxHP // heal on level up
			}
			g.XPOrbs = append(g.XPOrbs[:i], g.XPOrbs[i+1:]...)
		}
	}

	// Update particles
	g.updateParticles()

	// Update damage numbers
	for i := len(g.DmgNumbers) - 1; i >= 0; i-- {
		dn := &g.DmgNumbers[i]
		dn.Y -= 1.5
		dn.Life -= 1.0 / 60.0
		if dn.Life <= 0 {
			g.DmgNumbers = append(g.DmgNumbers[:i], g.DmgNumbers[i+1:]...)
		}
	}

	// Spawn enemies
	g.SpawnTimer++
	spawnRate := int(math.Max(8, 45-float64(g.Wave)*3))
	if g.SpawnTimer >= spawnRate && len(g.Enemies) < 50+g.Wave*5 {
		g.spawnEnemy()
		g.SpawnTimer = 0
	}

	// Wave progression
	g.WaveTimer++
	if g.WaveTimer >= 1800 { // 30 seconds per wave
		g.Wave++
		g.WaveTimer = 0
	}
}

func (g *Game) spawnEnemy() {
	// Spawn around player
	angle := rand.Float64() * 6.2832
	distFromPlayer := 400.0 + rand.Float64()*200

	x := g.Player.X + math.Cos(angle)*distFromPlayer
	y := g.Player.Y + math.Sin(angle)*distFromPlayer

	// Clamp to world
	x = math.Max(30, math.Min(WorldW-30, x))
	y = math.Max(30, math.Min(WorldH-30, y))

	eType := 0
	r := rand.Float64()
	if g.Wave >= 3 && r < 0.3 {
		eType = 1 // fast
	}
	if g.Wave >= 5 && r < 0.15 {
		eType = 2 // tank
	}

	hp := 20 + g.Wave*5
	speed := 1.0
	size := 28.0
	dmg := 5
	imgIdx := 0

	switch eType {
	case 1:
		hp = 15 + g.Wave*3
		speed = 2.2
		size = 22
		dmg = 3
		imgIdx = 1
	case 2:
		hp = 60 + g.Wave*10
		speed = 0.6
		size = 36
		dmg = 15
		imgIdx = 2
	}

	g.Enemies = append(g.Enemies, Enemy{
		X: x, Y: y,
		HP: hp, MaxHP: hp,
		Speed: speed,
		Type: eType,
		Size: size,
		Damage: dmg,
	})
	_ = imgIdx
}

func (g *Game) addDmgNumber(x, y float64, val int, c color.RGBA) {
	g.DmgNumbers = append(g.DmgNumbers, DamageNumber{
		X: x, Y: y,
		Value: val,
		Life: 1.0,
		Color: c,
	})
}

func (g *Game) spawnExplosion(x, y float64, c color.RGBA, count int) {
	for i := 0; i < count; i++ {
		a := rand.Float64() * 6.2832
		s := 2 + rand.Float64()*5
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: math.Cos(a)*s, VY: math.Sin(a)*s - 1,
			Life: 0.5 + rand.Float64()*0.5,
			MaxLife: 1.0,
			Color: c,
			Size: 3 + rand.Float64()*5,
		})
	}
}

func (g *Game) updateParticles() {
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		if p.Life >= 9998 {
			continue // static particles (stars)
		}
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.05
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Player = &Player{
		X: WorldW / 2, Y: WorldH / 2,
		HP: 100, MaxHP: 100,
		Speed: PlayerSpeed,
		Power: 1,
		Level: 1,
		XP: 0,
		XPToNext: 50,
		Anim: &Animation{TickSpeed: 6},
		FireAnim: &Animation{TickSpeed: 6},
	}
	g.Enemies = []Enemy{}
	g.Bullets = []Bullet{}
	g.XPOrbs = []XPorb{}
	g.Particles = []Particle{}
	g.DmgNumbers = []DamageNumber{}
	g.Score = 0
	g.Wave = 1
	g.Time = 0
	g.Kills = 0
	g.SpawnTimer = 0
	g.WaveTimer = 0

	g.loadAssets()
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenW, ScreenH, color.RGBA{20, 22, 18, 255}, false)

	if g.State == StateMenu {
		g.drawMenu(screen)
		return
	}

	// Camera
	camX, camY := g.CameraX, g.CameraY

	// Ground grid
	for x := 0; x < WorldW; x += 100 {
		sx := float32(x) - float32(camX)
		if sx >= -100 && sx <= ScreenW+100 {
			vector.StrokeLine(screen, sx, 0, sx, ScreenH, 1, color.RGBA{35, 38, 30, 150}, false)
		}
	}
	for y := 0; y < WorldH; y += 100 {
		sy := float32(y) - float32(camY)
		if sy >= -100 && sy <= ScreenH+100 {
			vector.StrokeLine(screen, 0, sy, ScreenW, sy, 1, color.RGBA{35, 38, 30, 150}, false)
		}
	}

	// XP orbs
	for _, xp := range g.XPOrbs {
		sx := float32(xp.X - camX)
		sy := float32(xp.Y - camY)
		if sx < -20 || sx > ScreenW+20 || sy < -20 || sy > ScreenH+20 {
			continue
		}

		alpha := 255
		if xp.Life < 120 {
			alpha = int(float64(xp.Life) / 120 * 255)
		}

		xpImg := ebiten.NewImage(12, 12)
		for py := 0; py < 12; py++ {
			for px := 0; px < 12; px++ {
				dx := float64(px-6)
				dy := float64(py-6)
				if dx*dx+dy*dy <= 36 {
					xpImg.Set(px, py, color.RGBA{100, 200, 255, uint8(alpha)})
				}
			}
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx-6), float64(sy-6))
		screen.DrawImage(xpImg, op)
	}

	// Enemies
	for _, e := range g.Enemies {
		sx := float32(e.X - camX)
		sy := float32(e.Y - camY)
		if sx < -50 || sx > ScreenW+50 || sy < -50 || sy > ScreenH+50 {
			continue
		}

		imgIdx := e.Type
		if imgIdx >= len(g.ZombieImgs) {
			imgIdx = 0
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx)-e.Size/2, float64(sy)-e.Size)
		screen.DrawImage(g.ZombieImgs[imgIdx], op)

		// HP bar
		if e.HP < e.MaxHP {
			barW := e.Size
			ratio := float64(e.HP) / float64(e.MaxHP)
			vector.DrawFilledRect(screen, float32(sx)-float32(barW)/2, float32(sy)-float32(e.Size)-10, float32(barW), 4, color.RGBA{60, 0, 0, 200}, false)
			vector.DrawFilledRect(screen, float32(sx)-float32(barW)/2, float32(sy)-float32(e.Size)-10, float32(barW)*float32(ratio), 4, color.RGBA{255, 50, 50, 255}, false)
		}
	}

	// Bullets
	for _, b := range g.Bullets {
		sx := float32(b.X - camX)
		sy := float32(b.Y - camY)
		if sx < -20 || sx > ScreenW+20 || sy < -20 || sy > ScreenH+20 {
			continue
		}

		bImg := ebiten.NewImage(6, 6)
		bImg.Fill(color.RGBA{255, 230, 80, 255})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx-3), float64(sy-3))
		screen.DrawImage(bImg, op)
	}

	// Player
	p := g.Player
	sx := float32(p.X - camX)
	sy := float32(p.Y - camY)

	// Player sprite
	playerImg := g.getSprite()
	if playerImg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(sx-32), float64(sy-64))
		screen.DrawImage(playerImg, op)
	}

	// Damage numbers
	for _, dn := range g.DmgNumbers {
		sx := float32(dn.X - camX)
		sy := float32(dn.Y - camY)
		alpha := uint8((dn.Life / 1.0) * 255)
		c := color.RGBA{dn.Color.R, dn.Color.G, dn.Color.B, alpha}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", dn.Value), int(sx)-10, int(sy))
		_ = c
	}

	// Particles
	for _, part := range g.Particles {
		sx := float32(part.X - camX)
		sy := float32(part.Y - camY)
		size := int(part.Size * (part.Life / part.MaxLife))
		if size < 1 {
			continue
		}
		alpha := uint8((part.Life / part.MaxLife) * 255)
		c := color.RGBA{part.Color.R, part.Color.G, part.Color.B, alpha}
		pImg := ebiten.NewImage(size, size)
		pImg.Fill(c)
		pop := &ebiten.DrawImageOptions{}
		pop.GeoM.Translate(float64(sx)-float64(size)/2, float64(sy)-float64(size)/2)
		screen.DrawImage(pImg, pop)
	}

	// HUD
	g.drawHUD(screen)

	if g.State == StateGameOver {
		g.drawGameOver(screen)
	}
}

func (g *Game) getSprite() *ebiten.Image {
	if g.Player == nil {
		return nil
	}

	if g.Player.IsFiring && g.Player.FireAnim != nil && len(g.Player.FireAnim.Frames) > 0 {
		return g.Player.FireAnim.Image()
	}
	if g.Player.Anim != nil && len(g.Player.Anim.Frames) > 0 {
		return g.Player.Anim.Image()
	}
	return nil
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "☠ SURVIVOR SHOOTER ☠", ScreenW/2-200, 280)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 102", ScreenW/2-140, 330)

	vector.StrokeLine(screen, ScreenW/2-200, 370, ScreenW/2+200, 370, 2, color.RGBA{150, 80, 80, 255}, false)

	g.drawButton(screen, "▶  НАЧАТЬ ИГРУ", ScreenW/2-120, 500, 240, 60)

	ebitenutil.DebugPrintAt(screen, "WASD / Стрелки - Движение", ScreenW/2-130, 600)
	ebitenutil.DebugPrintAt(screen, "Автострельба по ближайшему врагу", ScreenW/2-160, 630)
	ebitenutil.DebugPrintAt(screen, "Выживай! Собирай XP! Качайся!", ScreenW/2-150, 670)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	p := g.Player

	// Top bar
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 60, color.RGBA{15, 15, 12, 230}, false)
	vector.StrokeLine(screen, 0, 60, ScreenW, 60, 2, color.RGBA{120, 80, 60, 255}, false)

	// HP bar
	hpW := 200
	ratio := float64(p.HP) / float64(p.MaxHP)
	vector.DrawFilledRect(screen, 20, 15, float32(hpW), 22, color.RGBA{60, 0, 0, 200}, false)
	vector.DrawFilledRect(screen, 20, 15, float32(hpW)*float32(ratio), 22, color.RGBA{200, 50, 50, 255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", p.HP, p.MaxHP), 25, 17)

	// XP bar
	xpW := 200
	xpRatio := float64(p.XP) / float64(p.XPToNext)
	vector.DrawFilledRect(screen, 20, 40, float32(xpW), 12, color.RGBA{0, 0, 60, 200}, false)
	vector.DrawFilledRect(screen, 20, 40, float32(xpW)*float32(xpRatio), 12, color.RGBA{50, 150, 255, 255}, false)

	// Stats
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("LVL: %d", p.Level), 240, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СИЛА: %d", p.Power), 240, 38)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 420, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 600, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("УБИТО: %d", g.Kills), 760, 17)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("РЕКОРД: %d", g.HighScore), 950, 17)

	// Minimap
	mmW, mmH := 150, 150
	mmX := ScreenW - mmW - 15
	mmY := ScreenH - mmH - 15
	vector.DrawFilledRect(screen, float32(mmX), float32(mmY), float32(mmW), float32(mmH), color.RGBA{10, 10, 8, 200}, false)
	vector.StrokeLine(screen, float32(mmX), float32(mmY), float32(mmX+mmW), float32(mmY), 1, color.RGBA{100, 80, 60, 255}, false)
	vector.StrokeLine(screen, float32(mmX+mmW), float32(mmY), float32(mmX+mmW), float32(mmY+mmH), 1, color.RGBA{100, 80, 60, 255}, false)
	vector.StrokeLine(screen, float32(mmX+mmW), float32(mmY+mmH), float32(mmX), float32(mmY+mmH), 1, color.RGBA{100, 80, 60, 255}, false)
	vector.StrokeLine(screen, float32(mmX), float32(mmY+mmH), float32(mmX), float32(mmY), 1, color.RGBA{100, 80, 60, 255}, false)

	// Player on minimap
	px := mmX + int(float64(p.X)/float64(WorldW)*float64(mmW))
	py := mmY + int(float64(p.Y)/float64(WorldH)*float64(mmH))
	vector.DrawFilledRect(screen, float32(px-2), float32(py-2), 4, 4, color.RGBA{100, 255, 100, 255}, false)

	// Enemies on minimap
	for _, e := range g.Enemies {
		ex := mmX + int(float64(e.X)/float64(WorldW)*float64(mmW))
		ey := mmY + int(float64(e.Y)/float64(WorldH)*float64(mmH))
		vector.DrawFilledRect(screen, float32(ex), float32(ey), 2, 2, color.RGBA{255, 80, 80, 200}, false)
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, ScreenH/2-130, ScreenW, 260, color.RGBA{10, 10, 8, 220}, false)

	ebitenutil.DebugPrintAt(screen, "☠ ВЫ ПОГИБЛИ ☠", ScreenW/2-130, ScreenH/2-100)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %d | Волна: %d | Убито: %d | Уровень: %d",
		g.Score, g.Wave, g.Kills, g.Player.Level), ScreenW/2-220, ScreenH/2-45)

	if g.Score >= g.HighScore && g.Score > 0 {
		ebitenutil.DebugPrintAt(screen, "🏆 НОВЫЙ РЕКОРД! 🏆", ScreenW/2-120, ScreenH/2-10)
	}

	g.drawButton(screen, "🔄  ЗАНОВО", ScreenW/2-100, ScreenH/2+30, 200, 55)
	g.drawButton(screen, "←  МЕНЮ", ScreenW/2-100, ScreenH/2+100, 200, 55)
}

func (g *Game) drawButton(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	hover := g.isInButton(x, y, w, h)

	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{80, 50, 40, 255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{40, 30, 25, 255}, false)
	}

	border := ebiten.NewImage(w, 3)
	border.Fill(color.RGBA{150, 100, 60, 255})

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)

	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(border, op2)

	ebitenutil.DebugPrintAt(screen, text, x+25, y+h/2-10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenW, ScreenH
}

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Survivor Shooter - Go365 Day 102")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
