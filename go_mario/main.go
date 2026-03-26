// Go365 Day 86 - GO MARIO: SPACE SHOOTER v4.0.2
// Простой и стабильный космический шутер

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
)

var (
	ColorBG       = color.RGBA{10, 10, 20, 255}
	ColorPlayer   = color.RGBA{0, 255, 255, 255}
	ColorBullet   = color.RGBA{255, 255, 0, 255}
	ColorEnemy    = color.RGBA{255, 50, 50, 255}
	ColorHealth   = color.RGBA{0, 255, 100, 255}
	ColorGold     = color.RGBA{255, 215, 0, 255}
)

type Player struct {
	x, y    float64
	angle   float64
	vx, vy  float64
	health  int
	maxHP   int
	score   int
}

type Bullet struct {
	x, y, vx, vy float64
	isPlayer     bool
	life         int
}

type Enemy struct {
	x, y, vx, vy float64
	health       int
	maxHP        int
	size         float64
}

type Particle struct {
	x, y, vx, vy float64
	life         int
	color        color.RGBA
	size         float64
}

type Game struct {
	player    *Player
	bullets   []*Bullet
	enemies   []*Enemy
	particles []*Particle
	state     int // 0=menu, 1=playing, 2=gameover
	wave      int
	frame     int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		player: &Player{
			x:      ScreenWidth / 2,
			y:      ScreenHeight / 2,
			maxHP:  100,
			health: 100,
		},
		state:     0,
		bullets:   []*Bullet{},
		enemies:   []*Enemy{},
		particles: []*Particle{},
	}
}

func (g *Game) StartGame() {
	g.state = 1
	g.player.x = ScreenWidth / 2
	g.player.y = ScreenHeight / 2
	g.player.vx = 0
	g.player.vy = 0
	g.player.health = g.player.maxHP
	g.player.score = 0
	g.player.angle = 0
	g.wave = 1
	g.bullets = []*Bullet{}
	g.enemies = []*Enemy{}
	g.particles = []*Particle{}
	g.spawnEnemies(5)
}

func (g *Game) spawnEnemies(count int) {
	for i := 0; i < count; i++ {
		angle := rand.Float64() * math.Pi * 2
		dist := 400.0
		g.enemies = append(g.enemies, &Enemy{
			x:      g.player.x + math.Cos(angle)*dist,
			y:      g.player.y + math.Sin(angle)*dist,
			vx:     (rand.Float64() - 0.5) * 2,
			vy:     (rand.Float64() - 0.5) * 2,
			health: 20 + g.wave*5,
			maxHP:  20 + g.wave*5,
			size:   15 + rand.Float64()*10,
		})
	}
}

func (g *Game) Update() error {
	g.frame++

	if g.state == 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.StartGame()
		}
		return nil
	}

	if g.state == 2 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = 0
		}
		return nil
	}

	// Player rotation
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.angle -= 0.08
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.angle += 0.08
	}

	// Thrust
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		g.player.vx += math.Cos(g.player.angle) * 0.3
		g.player.vy += math.Sin(g.player.angle) * 0.3
	}

	// Friction
	g.player.vx *= 0.98
	g.player.vy *= 0.98
	g.player.x += g.player.vx
	g.player.y += g.player.vy

	// Screen wrap
	if g.player.x < 0 {
		g.player.x = ScreenWidth
	}
	if g.player.x > ScreenWidth {
		g.player.x = 0
	}
	if g.player.y < 0 {
		g.player.y = ScreenHeight
	}
	if g.player.y > ScreenHeight {
		g.player.y = 0
	}

	// Shooting
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyZ)) && g.frame%15 == 0 {
		g.bullets = append(g.bullets, &Bullet{
			x:        g.player.x + math.Cos(g.player.angle)*25,
			y:        g.player.y + math.Sin(g.player.angle)*25,
			vx:       math.Cos(g.player.angle) * 10,
			vy:       math.Sin(g.player.angle) * 10,
			isPlayer: true,
			life:     100,
		})
	}

	// Update bullets
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		b.x += b.vx
		b.y += b.vy
		b.life--
		if b.life <= 0 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
		}
	}

	// Update enemies
	for _, e := range g.enemies {
		angle := math.Atan2(g.player.y-e.y, g.player.x-e.x)
		e.vx = math.Cos(angle) * 1.5
		e.vy = math.Sin(angle) * 1.5
		e.x += e.vx
		e.y += e.vy

		// Collision with player
		dist := math.Hypot(g.player.x-e.x, g.player.y-e.y)
		if dist < e.size+15 {
			g.player.health -= 20
			e.health = 0
			g.spawnExplosion(e.x, e.y)
			if g.player.health <= 0 {
				g.state = 2
			}
		}
	}

	// Bullet-enemy collision
	for i := len(g.bullets) - 1; i >= 0; i-- {
		b := g.bullets[i]
		if !b.isPlayer {
			continue
		}
		for j := len(g.enemies) - 1; j >= 0; j-- {
			e := g.enemies[j]
			if math.Hypot(b.x-e.x, b.y-e.y) < e.size+5 {
				e.health -= 10
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
				g.spawnHit(b.x, b.y)
				if e.health <= 0 {
					g.player.score += 100
					g.spawnExplosion(e.x, e.y)
					g.enemies = append(g.enemies[:j], g.enemies[j+1:]...)
				}
				break
			}
		}
	}

	// Update particles
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}

	// Wave management
	allDead := true
	for _, e := range g.enemies {
		if e.health > 0 {
			allDead = false
			break
		}
	}
	if allDead && len(g.enemies) == 0 {
		g.wave++
		g.spawnEnemies(5 + g.wave)
	}

	return nil
}

func (g *Game) spawnExplosion(x, y float64) {
	for i := 0; i < 20; i++ {
		g.particles = append(g.particles, &Particle{
			x:      x,
			y:      y,
			vx:     (rand.Float64() - 0.5) * 8,
			vy:     (rand.Float64() - 0.5) * 8,
			life:   30,
			color:  color.RGBA{255, uint8(100+rand.Intn(155)), 0, 255},
			size:   3 + rand.Float64()*3,
		})
	}
}

func (g *Game) spawnHit(x, y float64) {
	for i := 0; i < 5; i++ {
		g.particles = append(g.particles, &Particle{
			x:      x,
			y:      y,
			vx:     (rand.Float64() - 0.5) * 5,
			vy:     (rand.Float64() - 0.5) * 5,
			life:   15,
			color:  ColorGold,
			size:   2,
		})
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(ColorBG)

	if g.state == 0 {
		g.drawMenu(screen)
		return
	}

	// Draw player
	if g.state == 1 {
		g.drawPlayer(screen)
	}

	// Draw bullets
	for _, b := range g.bullets {
		vector.DrawFilledCircle(screen, float32(b.x), float32(b.y), 4, ColorBullet, true)
	}

	// Draw enemies
	for _, e := range g.enemies {
		vector.DrawFilledCircle(screen, float32(e.x), float32(e.y), float32(e.size), ColorEnemy, true)
		// Health bar
		if e.health < e.maxHP {
			w := float32(e.size * 2)
			vector.DrawFilledRect(screen, float32(e.x)-w/2, float32(e.y)-float32(e.size)-8, w, 3, color.RGBA{80, 0, 0, 255}, true)
			if e.maxHP > 0 {
				vector.DrawFilledRect(screen, float32(e.x)-w/2, float32(e.y)-float32(e.size)-8, w*float32(e.health)/float32(e.maxHP), 3, ColorHealth, true)
			}
		}
	}

	// Draw particles
	for _, p := range g.particles {
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), float32(p.size), p.color, true)
	}

	// Draw HUD
	g.drawHUD(screen)

	if g.state == 2 {
		g.drawGameOver(screen)
	}
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	p := g.player
	// Triangle
	p1x := float32(p.x + math.Cos(p.angle)*20)
	p1y := float32(p.y + math.Sin(p.angle)*20)
	p2x := float32(p.x + math.Cos(p.angle+2.5)*12)
	p2y := float32(p.y + math.Sin(p.angle+2.5)*12)
	p3x := float32(p.x + math.Cos(p.angle-2.5)*12)
	p3y := float32(p.y + math.Sin(p.angle-2.5)*12)

	vector.StrokeLine(screen, p1x, p1y, p2x, p2y, 2, ColorPlayer, true)
	vector.StrokeLine(screen, p2x, p2y, p3x, p3y, 2, ColorPlayer, true)
	vector.StrokeLine(screen, p3x, p3y, p1x, p1y, 2, ColorPlayer, true)

	// Thruster
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		fx := p.x - math.Cos(p.angle)*15
		fy := p.y - math.Sin(p.angle)*15
		vector.DrawFilledCircle(screen, float32(fx), float32(fy), 6, color.RGBA{255, 150, 50, 255}, true)
	}
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Background
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 150}, true)

	// Health bar
	vector.DrawFilledRect(screen, 20, 12, 150, 12, color.RGBA{80, 0, 0, 255}, true)
	if g.player.maxHP > 0 {
		vector.DrawFilledRect(screen, 20, 12, 150*float32(g.player.health)/float32(g.player.maxHP), 12, ColorHealth, true)
	}

	text.Draw(screen, fmt.Sprintf("HP:%d", g.player.health), basicfont.Face7x13, 180, 22, color.White)
	text.Draw(screen, fmt.Sprintf("SCORE:%d", g.player.score), basicfont.Face7x13, 300, 22, ColorGold)
	text.Draw(screen, fmt.Sprintf("WAVE:%d", g.wave), basicfont.Face7x13, 500, 22, ColorGold)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	text.Draw(screen, "GO MARIO: SPACE", basicfont.Face7x13, ScreenWidth/2-90, 200, ColorPlayer)
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-80, 300, color.White)
	text.Draw(screen, "Arrow keys/WASD - Move", basicfont.Face7x13, ScreenWidth/2-80, 350, color.White)
	text.Draw(screen, "Space/Z - Fire", basicfont.Face7x13, ScreenWidth/2-60, 375, color.White)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, ScreenWidth/2-50, ScreenHeight/2, ColorEnemy)
	text.Draw(screen, fmt.Sprintf("Score: %d", g.player.score), basicfont.Face7x13, ScreenWidth/2-50, ScreenHeight/2+30, color.White)
	text.Draw(screen, "Press ENTER", basicfont.Face7x13, ScreenWidth/2-50, ScreenHeight/2+60, color.White)
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GO MARIO: SPACE SHOOTER - Go365 Day 86")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
