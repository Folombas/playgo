// Package game - основная игровая логика City Platformer
package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/playgo/city_platformer/internal/entity"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

type State int

const (
	StateMenu State = iota
	StatePlaying
	StatePaused
	StateGameOver
)

type Game struct {
	state     State
	player    *entity.Player
	platforms []entity.Platform
	coins     []entity.Coin
	enemies   []entity.Enemy
	cameraX   float64
	score     int
	lives     int
	level     int
	particles []Particle
	rng       *rand.Rand
}

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.Color
	Size   float64
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		state: StateMenu,
		rng:   rng,
	}
	g.Reset()
	return g
}

func (g *Game) Reset() {
	g.player = entity.NewPlayer(100, 500)
	g.platforms = g.generateLevel(g.level)
	g.coins = g.generateCoins()
	g.enemies = g.generateEnemies()
	g.cameraX = 0
	g.score = 0
	g.lives = 3
	g.level = 1
	g.particles = make([]Particle, 0)
}

func (g *Game) generateLevel(level int) []entity.Platform {
	platforms := make([]entity.Platform, 0)

	// Земля
	platforms = append(platforms, entity.Platform{
		X: 0, Y: 650, Width: 3000, Height: 70, Type: "grass",
	})

	// Платформы
	for i := 0; i < 15+level*3; i++ {
		x := float64(200 + i*150 + g.rng.Intn(100))
		y := float64(550 - g.rng.Intn(200))
		width := float64(80 + g.rng.Intn(80))

		platforms = append(platforms, entity.Platform{
			X: x, Y: y, Width: width, Height: 20, Type: "stone",
		})
	}

	// Здания из City Mega Pack
	for i := 0; i < 3+level; i++ {
		x := float64(400 + i*700)
		platforms = append(platforms, entity.Platform{
			X: x, Y: 650, Width: 200, Height: 120, Type: "building",
		})
	}

	return platforms
}

func (g *Game) generateCoins() []entity.Coin {
	coins := make([]entity.Coin, 0)
	for i := 0; i < 20; i++ {
		coins = append(coins, entity.Coin{
			X: float64(300 + i*200),
			Y: float64(600 - g.rng.Intn(150)),
		})
	}
	return coins
}

func (g *Game) generateEnemies() []entity.Enemy {
	enemies := make([]entity.Enemy, 0)
	for i := 0; i < 5+g.level; i++ {
		enemies = append(enemies, entity.Enemy{
			X: float64(500 + i*400),
			Y: 620,
		})
	}
	return enemies
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused:
			g.state = StatePlaying
		case StateMenu, StateGameOver:
			return ebiten.Termination
		}
	}

	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	if g.state == StateGameOver && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	if g.state == StatePlaying {
		g.updateGame()
	}

	return nil
}

func (g *Game) updateGame() {
	g.player.Update()

	// Управление
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.MoveRight()
	}
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.player.CanJump() {
		g.player.Jump()
		g.spawnParticles(g.player.X+20, g.player.Y+40, 0, -2, 10, color.RGBA{200, 200, 200, 255})
	}

	// Физика
	g.player.VY += 0.5 // Гравитация
	g.player.X += g.player.VX
	g.player.Y += g.player.VY

	// Коллизии с платформами
	g.player.OnGround = false
	for _, p := range g.platforms {
		if g.checkCollision(g.player.X+20, g.player.Y+40, p.X, p.Y, p.Width, p.Height) {
			if g.player.VY > 0 && g.player.Y+40 < p.Y+20 {
				g.player.Y = p.Y - 40
				g.player.VY = 0
				g.player.OnGround = true
			}
		}
	}

	// Земля
	if g.player.Y > 610 {
		g.player.Y = 610
		g.player.VY = 0
		g.player.OnGround = true
	}

	// Камера
	g.cameraX = g.player.X - 400
	if g.cameraX < 0 {
		g.cameraX = 0
	}

	// Сбор монет
	activeCoins := make([]entity.Coin, 0)
	for _, coin := range g.coins {
		if g.checkCollision(g.player.X+20, g.player.Y+20, coin.X, coin.Y, 30, 30) {
			g.score += 10
			g.spawnParticles(coin.X+15, coin.Y+15, 0, -2, 8, color.RGBA{255, 215, 0, 255})
		} else {
			activeCoins = append(activeCoins, coin)
		}
	}
	g.coins = activeCoins

	// Частицы
	g.updateParticles()

	// Смерть от падения
	if g.player.Y > float64(screenHeight) {
		g.lives--
		if g.lives <= 0 {
			g.state = StateGameOver
		} else {
			g.player.X = 100
			g.player.Y = 500
			g.player.VY = 0
		}
	}
}

func (g *Game) checkCollision(x1, y1, x2, y2, w2, h2 float64) bool {
	return x1 > x2 && x1 < x2+w2 && y1 > y2 && y1 < y2+h2
}

func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, Particle{
			X: x, Y: y, VX: vx + (g.rng.Float64()-0.5)*2,
			VY: vy, Life: 1.0, Color: c, Size: 4 + g.rng.Float64()*4,
		})
	}
}

func (g *Game) updateParticles() {
	active := make([]Particle, 0)
	for _, p := range g.particles {
		p.X += p.VX
		p.Y += p.VY
		p.Life -= 0.02
		if p.Life > 0 {
			active = append(active, p)
		}
	}
	g.particles = active
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Небо
	for y := 0; y < screenHeight; y++ {
		ratio := float64(y) / float64(screenHeight)
		r := uint8(135 - ratio*35)
		gr := uint8(206 - ratio*86)
		b := uint8(235)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screenWidth), 1, color.RGBA{r, gr, b, 255}, true)
	}

	// Городской фон из CITY_MEGA
	g.drawCityBackground(screen)

	// Платформы
	for _, p := range g.platforms {
		g.drawPlatform(screen, p)
	}

	// Монеты
	for _, coin := range g.coins {
		g.drawCoin(screen, coin)
	}

	// Игрок
	g.player.Draw(screen, g.cameraX)

	// Частицы
	for _, p := range g.particles {
		screenX := p.X - g.cameraX
		vector.DrawFilledRect(screen, float32(screenX-p.Size/2), float32(p.Y-p.Size/2), float32(p.Size), float32(p.Size), p.Color, true)
	}

	// UI
	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawHUD(screen)
	case StatePaused:
		g.drawHUD(screen)
		g.drawPause(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	}
}

func (g *Game) drawCityBackground(screen *ebiten.Image) {
	// Рисуем силуэт города
	for i := 0; i < 20; i++ {
		height := 100 + g.rng.Float64()*150
		x := float64(i*80) - (g.cameraX * 0.3)
		vector.DrawFilledRect(screen, float32(x), float32(screenHeight-int(height)), 60, float32(height), color.RGBA{100, 100, 150, 200}, true)
	}
}

func (g *Game) drawPlatform(screen *ebiten.Image, p entity.Platform) {
	screenX := p.X - g.cameraX

	if p.Type == "building" {
		// Здание
		vector.DrawFilledRect(screen, float32(screenX), float32(p.Y-100), float32(p.Width), 120, color.RGBA{150, 150, 150, 255}, true)
		// Окна
		for wy := 0; wy < 3; wy++ {
			for wx := 0; wx < int(p.Width)/30; wx++ {
				vector.DrawFilledRect(screen, float32(screenX)+float32(10+wx*30), float32(p.Y)-90+float32(wy*35), 20, 25, color.RGBA{200, 220, 255, 255}, true)
			}
		}
	} else {
		// Обычная платформа
		vector.DrawFilledRect(screen, float32(screenX), float32(p.Y), float32(p.Width), float32(p.Height), color.RGBA{100, 200, 100, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX), float32(p.Y+8), float32(p.Width), float32(p.Height-8), color.RGBA{140, 100, 60, 255}, true)
	}
}

func (g *Game) drawCoin(screen *ebiten.Image, coin entity.Coin) {
	screenX := coin.X - g.cameraX
	vector.DrawFilledCircle(screen, float32(screenX+15), float32(coin.Y+15), 12, color.RGBA{255, 215, 0, 255}, true)
	vector.DrawFilledCircle(screen, float32(screenX+15), float32(coin.Y+15), 8, color.RGBA{255, 255, 150, 255}, true)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	title := `
╔═══════════════════════════════════════════╗
║      🏙️ CITY PLATFORMER 🎖️               ║
║         PMC Contractor Mission            ║
╠═══════════════════════════════════════════╣
║                                           ║
║         [SPACE] - Начать миссию           ║
║         [ESC] - Выход                     ║
║                                           ║
║  🎮 Управление:                           ║
║     A/D или ←/→ - Бег                     ║
║     W/↑/SPACE - Прыжок                    ║
║                                           ║
║  🎯 Цель: Собери все монеты!              ║
║  ⚠️ Остерегайся падения!                  ║
║                                           ║
╚═══════════════════════════════════════════╝
`
	ebitenutil.DebugPrint(screen, title)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	hudText := fmt.Sprintf(`┌─────────────────────────┐
│  🏙️ CITY PLATFORMER     │
├─────────────────────────┤
│  💰 Счёт: %5d            │
│  ❤️  Жизни: %3d           │
│  📍 Уровень: %2d          │
└─────────────────────────┘

[ESC] - Пауза
`, g.score, g.lives, g.level)

	ebitenutil.DebugPrint(screen, hudText)
}

func (g *Game) drawPause(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 128})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	pauseText := `
╔═══════════════════════════════════════╗
║              ⏸️ ПАУЗА                  ║
╠═══════════════════════════════════════╣
║     [ESC] - Продолжить                ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, screenWidth/2-180, screenHeight/2-100)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{50, 0, 0, 180})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	gameOverText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║          💀 МИССИЯ ПРОВАЛЕНА 💀       ║
╠═══════════════════════════════════════╣
║     Счёт: %5d                          ║
║     Уровень: %2d                        ║
║                                       ║
║     [SPACE] - Новая миссия            ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, g.score, g.level)

	ebitenutil.DebugPrintAt(screen, gameOverText, screenWidth/2-180, screenHeight/2-120)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
