// 2D Shooter — Go365 Challenge Day 102
// Шутер вид сверху с волнами врагов
// 8 апреля 2026

package main

import (
	"fmt"
	"image"
	"image/color"
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

// ============================================================================
// КОНСТАНТЫ
// ============================================================================

const (
	ScreenW       = 900
	ScreenH       = 700
	PlayerSpeed   = 4.0
	BulletSpeed   = 9.0
	EnemyBaseSpd  = 1.2
	FireRate      = 8
)

// ============================================================================
// УТИЛИТЫ
// ============================================================================

func dist(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

func createPlayerSprite(size int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2

	// Корпус — зелёный танк
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			r := math.Sqrt(dx*dx + dy*dy)

			if r <= float64(size/2-2) {
				// Основной цвет
				c := color.RGBA{50, 170, 80, 255}
				// Градиент
				if r < float64(size/2-8) {
					c = color.RGBA{70, 210, 100, 255}
				}
				img.Set(x, y, c)
			}
			// Пушка вверх
			if x >= cx-3 && x <= cx+3 && y >= 2 && y <= cy {
				img.Set(x, y, color.RGBA{80, 80, 80, 255})
			}
			// Блик
			if dx*dx+dy*dy < 36 && x < cx && y < cy {
				img.Set(x, y, color.RGBA{180, 255, 200, 200})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func createEnemySprite(size int, t int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2

	var body, accent color.RGBA
	switch t {
	case 0: // красный зомби
		body = color.RGBA{180, 50, 50, 255}
		accent = color.RGBA{220, 80, 80, 255}
	case 1: // оранжевый быстрый
		body = color.RGBA{210, 140, 40, 255}
		accent = color.RGBA{255, 190, 80, 255}
	case 2: // фиолетовый танк
		body = color.RGBA{140, 50, 190, 255}
		accent = color.RGBA{180, 90, 230, 255}
	case 3: // синий стрелок
		body = color.RGBA{50, 110, 190, 255}
		accent = color.RGBA{90, 160, 240, 255}
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			r := math.Sqrt(dx*dx + dy*dy)

			if r <= float64(size/2-2) {
				c := body
				if r < float64(size/2-6) {
					c = accent
				}
				img.Set(x, y, c)
			}
			// Глаза
			eyeOff := size / 5
			eyeR := size / 10
			if (dx+float64(eyeOff))*(dx+float64(eyeOff))+(dy-float64(eyeOff/2))*(dy-float64(eyeOff/2)) <= float64(eyeR*eyeR) ||
				(dx-float64(eyeOff))*(dx-float64(eyeOff))+(dy-float64(eyeOff/2))*(dy-float64(eyeOff/2)) <= float64(eyeR*eyeR) {
				img.Set(x, y, color.RGBA{255, 50, 50, 255})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func createBulletSprite(size int) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			if dx*dx+dy*dy <= float64(cx*cx) {
				img.Set(x, y, color.RGBA{255, 240, 100, 255})
			}
			if dx*dx+dy*dy <= 4 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
}

func createExplosionParticle(size int, c color.RGBA) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	cx, cy := size/2, size/2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x - cx)
			dy := float64(y - cy)
			r := math.Sqrt(dx*dx + dy*dy)
			if r <= float64(size/2) {
				t := r / float64(size/2)
				alpha := uint8(float64(c.A) * (1 - t))
				img.Set(x, y, color.RGBA{c.R, c.G, c.B, alpha})
			}
		}
	}
	return ebiten.NewImageFromImage(img)
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
	Angle     float64
	HP        int
	MaxHP     int
	Speed     float64
	Power     int
	FireTimer int
	Invincible int
}

type Bullet struct {
	X, Y, VX, VY float64
	Damage       int
	Life         int
	FromPlayer   bool
}

type Enemy struct {
	X, Y      float64
	Type      int
	HP        int
	MaxHP     int
	Speed     float64
	Size      float64
	Damage    int
	FireTimer int
	Score     int
}

type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
	Color        color.RGBA
	Size         float64
}

type Pickup struct {
	X, Y   float64
	Type   int
	Life   int
}

type Game struct {
	State      GameState
	Player     *Player
	Bullets    []Bullet
	Enemies    []Enemy
	Particles  []Particle
	Pickups    []Pickup

	Score      int
	Wave       int
	Kills      int
	HighScore  int
	MouseX, MouseY float64

	PlayerImg  *ebiten.Image
	EnemyImgs  []*ebiten.Image
	BulletImg  *ebiten.Image
	Stars      []Star

	GameTime   float64
	ShakeTimer float64
	WaveTimer  int
	SpawnTimer int
}

type Star struct{ X, Y, S, A, Spd float64 }

// ============================================================================
// NEW GAME
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())

	g := &Game{
		State: StateMenu,
		Player: &Player{
			X: ScreenW / 2, Y: ScreenH / 2,
			HP: 100, MaxHP: 100, Speed: PlayerSpeed, Power: 1,
		},
		Bullets:   []Bullet{},
		Enemies:   []Enemy{},
		Particles: []Particle{},
		Pickups:   []Pickup{},
		Wave:      1,
	}

	// Sprites
	g.PlayerImg = createPlayerSprite(44)
	g.BulletImg = createBulletSprite(10)
	g.EnemyImgs = []*ebiten.Image{
		createEnemySprite(32, 0),
		createEnemySprite(26, 1),
		createEnemySprite(38, 2),
		createEnemySprite(32, 3),
	}

	// Stars
	for i := 0; i < 100; i++ {
		g.Stars = append(g.Stars, Star{
			X: rand.Float64() * ScreenW,
			Y: rand.Float64() * ScreenH,
			S: rand.Float64()*2 + 0.5,
			A: rand.Float64(),
			Spd: rand.Float64()*2 + 0.5,
		})
	}

	// High score
	data, _ := os.ReadFile("hiscore.txt")
	fmt.Sscanf(string(data), "%d", &g.HighScore)

	return g
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.GameTime += 1.0 / 60.0

	mx, my := ebiten.CursorPosition()
	g.MouseX, g.MouseY = float64(mx), float64(my)

	// Shake
	if g.ShakeTimer > 0 {
		g.ShakeTimer -= 1.0 / 60.0
	}

	// Stars
	for i := range g.Stars {
		g.Stars[i].A = 0.4 + 0.6*math.Sin(g.GameTime*g.Stars[i].Spd)
	}

	// Particles
	for i := len(g.Particles) - 1; i >= 0; i-- {
		p := &g.Particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.08
		p.Life -= 1.0 / 60.0
		if p.Life <= 0 {
			g.Particles = append(g.Particles[:i], g.Particles[i+1:]...)
		}
	}

	// Click handling
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		switch g.State {
		case StateMenu:
			if g.inBtn(ScreenW/2-110, 480, 220, 55) {
				g.startGame()
			}
		case StateGameOver:
			if g.inBtn(ScreenW/2-90, 500, 180, 50) {
				g.startGame()
			}
			if g.inBtn(ScreenW/2-90, 565, 180, 50) {
				g.State = StateMenu
			}
		}
	}

	switch g.State {
	case StatePlaying:
		g.updatePlaying()
	}

	return nil
}

func (g *Game) inBtn(bx, by, bw, bh int) bool {
	return g.MouseX >= float64(bx) && g.MouseX <= float64(bx+bw) &&
		g.MouseY >= float64(by) && g.MouseY <= float64(by+bh)
}

func (g *Game) updatePlaying() {
	p := g.Player

	// Movement
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) { dy = -1 }
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) { dy = 1 }
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) { dx = -1 }
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) { dx = 1 }

	if dx != 0 || dy != 0 {
		l := math.Sqrt(dx*dx + dy*dy)
		p.X += (dx / l) * p.Speed
		p.Y += (dy / l) * p.Speed
	}

	p.X = math.Max(22, math.Min(ScreenW-22, p.X))
	p.Y = math.Max(22, math.Min(ScreenH-22, p.Y))

	// Aim
	p.Angle = math.Atan2(g.MouseY-p.Y, g.MouseX-p.X)

	// Auto-fire
	if p.FireTimer > 0 { p.FireTimer-- }
	if p.FireTimer <= 0 {
		g.firePlayer()
		p.FireTimer = FireRate / p.Power
		if p.FireTimer < 3 { p.FireTimer = 3 }
	}

	if p.Invincible > 0 { p.Invincible-- }

	// Bullets
	for i := len(g.Bullets) - 1; i >= 0; i-- {
		b := &g.Bullets[i]
		b.X += b.VX
		b.Y += b.VY
		b.Life--

		if b.Life <= 0 || b.X < -20 || b.X > ScreenW+20 || b.Y < -20 || b.Y > ScreenH+20 {
			g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
			continue
		}

		if b.FromPlayer {
			hit := false
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				e := &g.Enemies[j]
				if dist(b.X, b.Y, e.X, e.Y) < e.Size/2+4 {
					e.HP -= b.Damage
					hit = true
					g.spawnHit(b.X, b.Y)

					if e.HP <= 0 {
						g.Kills++
						g.Score += e.Score
						g.boom(e.X, e.Y, e.Type)
						g.ShakeTimer = 0.12

						// Pickup drop
						if rand.Float64() < 0.18 {
							g.Pickups = append(g.Pickups, Pickup{
								X: e.X, Y: e.Y,
								Type: rand.Intn(3),
								Life: 500,
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
			if dist(b.X, b.Y, p.X, p.Y) < 18 {
				p.HP -= 10
				g.spawnHit(p.X, p.Y)
				g.ShakeTimer = 0.2
				g.Bullets = append(g.Bullets[:i], g.Bullets[i+1:]...)
				if p.HP <= 0 {
					g.State = StateGameOver
					g.boom(p.X, p.Y, 99)
					g.saveHi()
				}
			}
		}
	}

	// Enemies
	for _, e := range g.Enemies {
		a := math.Atan2(p.Y-e.Y, p.X-e.X)
		e.X += math.Cos(a) * e.Speed
		e.Y += math.Sin(a) * e.Speed

		// Shooter enemy
		if e.Type == 3 {
			e.FireTimer--
			if e.FireTimer <= 0 {
				a2 := math.Atan2(p.Y-e.Y, p.X-e.X)
				g.Bullets = append(g.Bullets, Bullet{
					X: e.X, Y: e.Y,
					VX: math.Cos(a2)*4, VY: math.Sin(a2)*4,
					Life: 100, FromPlayer: false, Damage: 10,
				})
				e.FireTimer = 80
			}
		}

		// Contact damage
		if dist(e.X, e.Y, p.X, p.Y) < e.Size/2+18 {
			dmg := e.Damage
			if p.Invincible <= 0 {
				p.HP -= dmg
				p.Invincible = 25
				g.spawnHit(p.X, p.Y)
				g.ShakeTimer = 0.25
			}
			// Push enemy back
			e.X -= math.Cos(a) * 30
			e.Y -= math.Sin(a) * 30

			if p.HP <= 0 {
				g.State = StateGameOver
				g.boom(p.X, p.Y, 99)
				g.saveHi()
			}
		}
	}

	// Pickups
	for i := len(g.Pickups) - 1; i >= 0; i-- {
		pu := &g.Pickups[i]
		pu.Life--
		if pu.Life <= 0 {
			g.Pickups = append(g.Pickups[:i], g.Pickups[i+1:]...)
			continue
		}
		if dist(pu.X, pu.Y, p.X, p.Y) < 28 {
			switch pu.Type {
			case 0:
				p.HP = int(math.Min(float64(p.MaxHP), float64(p.HP+30)))
			case 1:
				p.Power = int(math.Min(5, float64(p.Power+1)))
			case 2:
				p.Invincible = 300
			}
			g.Pickups = append(g.Pickups[:i], g.Pickups[i+1:]...)
		}
	}

	// Wave system
	g.WaveTimer++
	if g.WaveTimer >= 900 {
		g.Wave++
		g.WaveTimer = 0
	}
	g.SpawnTimer++
	rate := int(math.Max(8, 45-float64(g.Wave)*4))
	if g.SpawnTimer >= rate && len(g.Enemies) < 40+g.Wave*5 {
		g.spawnEnemy()
		g.SpawnTimer = 0
	}
}

func (g *Game) firePlayer() {
	p := g.Player
	bx := p.X + math.Cos(p.Angle)*20
	by := p.Y + math.Sin(p.Angle)*20

	g.Bullets = append(g.Bullets, Bullet{
		X: bx, Y: by,
		VX: math.Cos(p.Angle) * BulletSpeed,
		VY: math.Sin(p.Angle) * BulletSpeed,
		Damage: p.Power * 8,
		Life: 70, FromPlayer: true,
	})

	if p.Power >= 2 {
		for _, off := range []float64{-0.15, 0.15} {
			a := p.Angle + off
			g.Bullets = append(g.Bullets, Bullet{
				X: bx, Y: by,
				VX: math.Cos(a)*BulletSpeed, VY: math.Sin(a)*BulletSpeed,
				Damage: p.Power * 6, Life: 60, FromPlayer: true,
			})
		}
	}

	// Muzzle flash
	for i := 0; i < 5; i++ {
		a := p.Angle + (rand.Float64()-0.5)*0.6
		s := 2 + rand.Float64()*3
		g.Particles = append(g.Particles, Particle{
			X: bx, Y: by,
			VX: math.Cos(a)*s, VY: math.Sin(a)*s,
			Life: 0.2, MaxLife: 0.2,
			Color: color.RGBA{255, 220, 60, 255},
			Size: 2 + rand.Float64()*2,
		})
	}
}

func (g *Game) spawnEnemy() {
	x := rand.Float64()*(ScreenW-60) + 30
	y := -30

	t := 0
	r := rand.Float64()
	if g.Wave >= 3 && r < 0.3 { t = 1 }
	if g.Wave >= 5 && r < 0.15 { t = 2 }
	if g.Wave >= 4 && r < 0.1 { t = 3 }

	hp := 15 + g.Wave*5
	sp := EnemyBaseSpd
	sz := 28.0
	dmg := 8
	sc := 100

	switch t {
	case 1:
		hp = 10 + g.Wave*3; sp = EnemyBaseSpd*2; sz = 22; dmg = 5; sc = 150
	case 2:
		hp = 50 + g.Wave*10; sp = EnemyBaseSpd*0.5; sz = 38; dmg = 18; sc = 300
	case 3:
		hp = 20 + g.Wave*5; sp = EnemyBaseSpd*0.7; sz = 30; dmg = 10; sc = 250
	}

	g.Enemies = append(g.Enemies, Enemy{
		X: x, Y: float64(y),
		HP: hp, MaxHP: hp,
		Speed: sp, Type: t, Size: sz,
		Damage: dmg, Score: sc, FireTimer: 60,
	})
}

func (g *Game) spawnHit(x, y float64) {
	for i := 0; i < 6; i++ {
		a := rand.Float64() * 6.28
		s := 2 + rand.Float64()*3
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: math.Cos(a)*s, VY: math.Sin(a)*s,
			Life: 0.35, MaxLife: 0.35,
			Color: color.RGBA{255, 200, 50, 255},
			Size: 2 + rand.Float64()*2,
		})
	}
}

func (g *Game) boom(x, y float64, t int) {
	n := 15
	if t == 99 { n = 40 }
	for i := 0; i < n; i++ {
		a := rand.Float64() * 6.28
		s := 3 + rand.Float64()*5
		c := color.RGBA{255, 150, 50, 255}
		if t == 1 { c = color.RGBA{255, 220, 80, 255} }
		if t == 2 { c = color.RGBA{200, 100, 255, 255} }
		if t == 3 { c = color.RGBA{80, 200, 255, 255} }
		if t == 99 { c = color.RGBA{255, 60, 60, 255} }
		g.Particles = append(g.Particles, Particle{
			X: x, Y: y,
			VX: math.Cos(a)*s, VY: math.Sin(a)*s - 1,
			Life: 0.7 + rand.Float64()*0.5,
			MaxLife: 1.2,
			Color: c, Size: 3 + rand.Float64()*5,
		})
	}
}

func (g *Game) saveHi() {
	if g.Score > g.HighScore {
		g.HighScore = g.Score
		os.WriteFile("hiscore.txt", []byte(fmt.Sprintf("%d", g.HighScore)), 0644)
	}
}

func (g *Game) startGame() {
	g.State = StatePlaying
	g.Player = &Player{
		X: ScreenW/2, Y: ScreenH/2, Speed: PlayerSpeed,
		HP: 100, MaxHP: 100, Power: 1,
	}
	g.Bullets = []Bullet{}
	g.Enemies = []Enemy{}
	g.Particles = []Particle{}
	g.Pickups = []Pickup{}
	g.Score = 0
	g.Wave = 1
	g.Kills = 0
	g.WaveTimer = 0
	g.SpawnTimer = 0
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	// BG gradient
	for y := 0; y < ScreenH; y++ {
		t := float64(y) / float64(ScreenH)
		r := uint8(lerpF(10, 18, t))
		gr := uint8(lerpF(12, 10, t))
		b := uint8(lerpF(25, 8, t))
		vector.DrawFilledRect(screen, 0, float32(y), ScreenW, 1, color.RGBA{r, gr, b, 255}, false)
	}

	// Stars
	for _, s := range g.Stars {
		a := uint8(s.A * 180)
		sz := int(s.S)
		if sz < 1 { sz = 1 }
		img := ebiten.NewImage(sz*2, sz*2)
		img.Fill(color.RGBA{200, 200, 240, a})
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(s.X, s.Y)
		screen.DrawImage(img, op)
	}

	// Grid
	for x := 0; x < ScreenW; x += 60 {
		vector.StrokeLine(screen, float32(x), 0, float32(x), ScreenH, 1, color.RGBA{25, 25, 40, 100}, false)
	}
	for y := 0; y < ScreenH; y += 60 {
		vector.StrokeLine(screen, 0, float32(y), ScreenW, float32(y), 1, color.RGBA{25, 25, 40, 100}, false)
	}

	// Shake offset
	sx, sy := 0.0, 0.0
	if g.ShakeTimer > 0 {
		sx = (rand.Float64() - 0.5) * 5
		sy = (rand.Float64() - 0.5) * 5
	}

	if g.State == StateMenu {
		g.drawMenu(screen)
		return
	}

	g.drawGame(screen, sx, sy)

	// Particles
	for _, p := range g.Particles {
		sz := int(p.Size * (p.Life / p.MaxLife))
		if sz < 1 { continue }
		a := uint8((p.Life / p.MaxLife) * 255)
		c := color.RGBA{p.Color.R, p.Color.G, p.Color.B, a}
		img := ebiten.NewImage(sz, sz)
		img.Fill(c)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(p.X-float64(sz)/2, p.Y-float64(sz)/2)
		screen.DrawImage(img, op)
	}

	if g.State == StateGameOver {
		g.drawGO(screen)
	}
}

func (g *Game) drawGame(screen *ebiten.Image, skx, sky float64) {
	// Pickups
	for _, pu := range g.Pickups {
		a := 255
		if pu.Life < 100 { a = int(float64(pu.Life)/100*255) }
		var c color.RGBA
		switch pu.Type {
		case 0: c = color.RGBA{50, 255, 80, uint8(a)}
		case 1: c = color.RGBA{255, 200, 50, uint8(a)}
		case 2: c = color.RGBA{100, 180, 255, uint8(a)}
		}
		puImg := ebiten.NewImage(18, 18)
		for y := 0; y < 18; y++ {
			for x := 0; x < 18; x++ {
				dx := float64(x-9); dy := float64(y-9)
				if dx*dx+dy*dy <= 81 {
					puImg.Set(x, y, c)
				}
			}
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pu.X-9+skx, pu.Y-9+sky)
		screen.DrawImage(puImg, op)
	}

	// Bullets
	for _, b := range g.Bullets {
		if b.FromPlayer {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(b.X-5+skx, b.Y-5+sky)
			screen.DrawImage(g.BulletImg, op)
		} else {
			// Enemy bullet — red
			img := ebiten.NewImage(8, 8)
			for y := 0; y < 8; y++ {
				for x := 0; x < 8; x++ {
					dx := float64(x-4); dy := float64(y-4)
					if dx*dx+dy*dy <= 16 {
						img.Set(x, y, color.RGBA{255, 70, 70, 255})
					}
				}
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(b.X-4+skx, b.Y-4+sky)
			screen.DrawImage(img, op)
		}
	}

	// Enemies
	for _, e := range g.Enemies {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(e.X-float64(e.Size)/2+skx, e.Y-float64(e.Size)+sky)
		if e.Type < len(g.EnemyImgs) {
			screen.DrawImage(g.EnemyImgs[e.Type], op)
		}

		// HP bar
		if e.HP < e.MaxHP {
			ratio := float64(e.HP) / float64(e.MaxHP)
			vector.DrawFilledRect(screen, float32(e.X-e.Size/2+skx), float32(e.Y-e.Size-8+sky), float32(e.Size), 4, color.RGBA{60,0,0,200}, false)
			vector.DrawFilledRect(screen, float32(e.X-e.Size/2+skx), float32(e.Y-e.Size-8+sky), float32(e.Size)*float32(ratio), 4, color.RGBA{255,50,50,255}, false)
		}
	}

	// Player
	p := g.Player
	if p.Invincible <= 0 || int(g.GameTime*10)%2 == 0 {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(-22, -22)
		op.GeoM.Rotate(p.Angle + math.Pi/2)
		op.GeoM.Translate(p.X+skx, p.Y+sky)
		screen.DrawImage(g.PlayerImg, op)
	}

	// Crosshair
	vector.StrokeCircle(screen, float32(g.MouseX), float32(g.MouseY), 10, 1.5, color.RGBA{255,255,255,180}, false)
	vector.StrokeLine(screen, float32(g.MouseX)-14, float32(g.MouseY), float32(g.MouseX)+14, float32(g.MouseY), 1, color.RGBA{255,255,255,180}, false)
	vector.StrokeLine(screen, float32(g.MouseX), float32(g.MouseY)-14, float32(g.MouseX), float32(g.MouseY)+14, 1, color.RGBA{255,255,255,180}, false)

	// HUD
	vector.DrawFilledRect(screen, 0, 0, ScreenW, 55, color.RGBA{15,18,35,220}, false)
	vector.StrokeLine(screen, 0, 55, ScreenW, 55, 2, color.RGBA{80,160,240,255}, false)

	// HP bar
	hw := 180
	ratio := float64(p.HP)/float64(p.MaxHP)
	vector.DrawFilledRect(screen, 18, 14, float32(hw), 18, color.RGBA{60,0,0,200}, false)
	vector.DrawFilledRect(screen, 18, 14, float32(hw)*float32(ratio), 18, color.RGBA{50,220,80,255}, false)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("HP: %d/%d", p.HP, p.MaxHP), 22, 16)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СЧЁТ: %d", g.Score), 220, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("ВОЛНА: %d", g.Wave), 400, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("СИЛА: %d", p.Power), 540, 16)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("УБИТО: %d", g.Kills), 680, 16)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "⚔  2D SHOOTER  ⚔", ScreenW/2-140, 260)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge — Day 102", ScreenW/2-130, 310)
	vector.StrokeLine(screen, ScreenW/2-180, 350, ScreenW/2+180, 350, 2, color.RGBA{100,180,255,255}, false)

	// Player preview
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ScreenW/2-22, 380)
	screen.DrawImage(g.PlayerImg, op)

	g.drawBtn(screen, "▶  НАЧАТЬ ИГРУ", ScreenW/2-110, 480, 220, 55)
	ebitenutil.DebugPrintAt(screen, "WASD / Стрелки — движение", ScreenW/2-120, 570)
	ebitenutil.DebugPrintAt(screen, "Автострельба по курсору", ScreenW/2-110, 600)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Рекорд: %d", g.HighScore), ScreenW/2-65, 640)
}

func (g *Game) drawGO(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, ScreenH/2-120, ScreenW, 240, color.RGBA{10,10,20,220}, false)
	ebitenutil.DebugPrintAt(screen, "💀 GAME OVER 💀", ScreenW/2-120, ScreenH/2-90)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Счёт: %d | Волна: %d | Убито: %d", g.Score, g.Wave, g.Kills), ScreenW/2-170, ScreenH/2-35)
	if g.Score >= g.HighScore && g.Score > 0 {
		ebitenutil.DebugPrintAt(screen, "🏆 НОВЫЙ РЕКОРД!", ScreenW/2-100, ScreenH/2)
	}
	g.drawBtn(screen, "🔄  ЗАНОВО", ScreenW/2-90, ScreenH/2+40, 180, 50)
	g.drawBtn(screen, "←  МЕНЮ", ScreenW/2-90, ScreenH/2+105, 180, 50)
}

func (g *Game) drawBtn(screen *ebiten.Image, text string, x, y, w, h int) {
	btn := ebiten.NewImage(w, h)
	hover := g.inBtn(x, y, w, h)
	if hover {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{50,75,115,255}, false)
	} else {
		vector.DrawFilledRect(btn, 0, 0, float32(w), float32(h), color.RGBA{30,42,70,255}, false)
	}
	brd := ebiten.NewImage(w, 3)
	brd.Fill(color.RGBA{100,180,255,255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(btn, op)
	op2 := &ebiten.DrawImageOptions{}
	op2.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(brd, op2)
	ebitenutil.DebugPrintAt(screen, text, x+22, y+h/2-10)
}

func lerpF(a, b, t float64) float64 { return a + (b-a)*t }

func (g *Game) Layout(ow, oh int) (int, int) { return ScreenW, ScreenH }

func main() {
	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("2D Shooter — Go365 Day 102")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	game := NewGame()
	if err := ebiten.RunGame(game); err != nil { log.Fatal(err) }
}
