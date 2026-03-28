// Package game содержит основную игровую логику Village Platformer
package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/playgo/go90/internal/entity"
)

// State определяет текущее состояние игры
type State int

const (
	StateMenu State = iota
	StatePlaying
	StatePaused
	StateGameOver
)

// Config содержит конфигурацию игры
type Config struct {
	ScreenWidth  int
	ScreenHeight int
	TargetFPS    int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		ScreenWidth:  1280,
		ScreenHeight: 720,
		TargetFPS:    60,
	}
}

// Game представляет основную игру
type Game struct {
	config      *Config
	state       State
	player      *entity.PlatformerPlayer
	world       *entity.PlatformerWorld
	cameraX     float64
	coins       int
	lives       int
	score       int
	level       int
	particles   []*Particle
	rng         *rand.Rand
	screenShake float64 // Тряска экрана
	flashAlpha  float64 // Вспышка
	combo       int     // Комбо множитель
	comboTimer  float64
}

// Particle представляет частицу
type Particle struct {
	X      float64
	Y      float64
	VX     float64
	VY     float64
	Life   float64
	Color  color.Color
	Size   float64
}

// NewGame создаёт новую игру
func NewGame() *Game {
	cfg := DefaultConfig()
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		config:    cfg,
		state:     StateMenu,
		player:    entity.NewPlatformerPlayer(100, 500),
		world:     entity.NewPlatformerWorld(cfg.ScreenWidth, cfg.ScreenHeight),
		cameraX:   0,
		coins:     0,
		lives:     3,
		score:     0,
		level:     1,
		particles: make([]*Particle, 0),
		rng:       rng,
	}

	// Генерация уровня
	g.world.GenerateLevel(1)

	return g
}

// Update обновляет логику игры
func (g *Game) Update() error {
	// Обработка ввода
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

	// Переход из меню
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.resetGame()
		g.state = StatePlaying
	}

	// Рестарт
	if g.state == StateGameOver && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.resetGame()
		g.state = StatePlaying
	}

	// Обновление игры
	if g.state == StatePlaying {
		g.updateGame()
	}

	return nil
}

// resetGame сбрасывает игру
func (g *Game) resetGame() {
	g.player = entity.NewPlatformerPlayer(100, 500)
	g.world = entity.NewPlatformerWorld(g.config.ScreenWidth, g.config.ScreenHeight)
	g.world.GenerateLevel(1)
	g.cameraX = 0
	g.coins = 0
	g.lives = 3
	g.score = 0
	g.level = 1
	g.particles = make([]*Particle, 0)
}

// updateGame обновляет игровую логику
func (g *Game) updateGame() {
	// Обновление игрока
	g.player.Update()

	// Управление
	g.handleInput()

	// Физика игрока
	g.applyPhysics()

	// Коллизии с миром
	g.handleCollisions()

	// Обновление комбо
	if g.comboTimer > 0 {
		g.comboTimer--
	} else {
		g.combo = 1
	}

	// Обновление неуязвимости
	if g.player.Invincible > 0 {
		g.player.Invincible--
	}

	// Обновление бустов
	if g.player.SpeedBoost > 1.0 {
		g.player.SpeedBoost -= 0.001
		if g.player.SpeedBoost < 1.0 {
			g.player.SpeedBoost = 1.0
			g.player.Speed = 5.0
		}
	}

	// Тряска экрана
	if g.screenShake > 0 {
		g.screenShake *= 0.9
		if g.screenShake < 0.5 {
			g.screenShake = 0
		}
	}

	// Вспышка
	if g.flashAlpha > 0 {
		g.flashAlpha -= 0.02
	}

	// Обновление частиц игрока
	g.updatePlayerParticles()

	// Камера (слежение за игроком)
	g.cameraX = g.player.X - float64(g.config.ScreenWidth)/3
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > g.world.Width-float64(g.config.ScreenWidth) {
		g.cameraX = g.world.Width - float64(g.config.ScreenWidth)
	}

	// Сбор монеток
	g.collectCoins()

	// Обновление частиц
	g.updateParticles()

	// Проверка падения в пропасть
	if g.player.Y > float64(g.config.ScreenHeight)+100 {
		g.lives--
		g.screenShake = 10
		g.flashAlpha = 0.3
		if g.lives <= 0 {
			g.state = StateGameOver
		} else {
			g.player.X = 100
			g.player.Y = 500
			g.player.VY = 0
			g.player.Invincible = 180 // 3 секунды
		}
	}

	// Проверка победы (достиг флага)
	if g.checkWin() {
		g.level++
		g.world.GenerateLevel(g.level)
		g.player.X = 100
		g.player.Y = 500
		g.player.VY = 0
		g.cameraX = 0
		g.screenShake = 15
		g.flashAlpha = 0.8
		g.spawnParticles(g.player.X+50, g.player.Y, 0, -3, 50, color.RGBA{255, 255, 255, 255})
	}
}

// handleInput обрабатывает ввод
func (g *Game) handleInput() {
	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.MoveRight()
	}

	// Прыжок (двойной прыжок если доступен)
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.player.CanJump() {
		g.player.Jump()
		g.spawnParticles(g.player.X+g.player.Width/2, g.player.Y+g.player.Height, -1, -2, 10, color.RGBA{200, 200, 200, 255})
	} else if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)) && g.player.DoubleJump && !g.player.OnGround && g.player.VY > -5 {
		g.player.VY = g.player.JumpForce * 0.8
		g.player.DoubleJump = false
		g.spawnParticles(g.player.X+g.player.Width/2, g.player.Y+g.player.Height, 0, -3, 15, color.RGBA{100, 200, 255, 255})
		g.screenShake = 3
	}
}

// updatePlayerParticles обновляет частицы игрока
func (g *Game) updatePlayerParticles() {
	// Шлейф при движении
	if math.Abs(g.player.VX) > 3 {
		g.player.ParticleTrail = append(g.player.ParticleTrail, entity.ParticleEffect{
			X:      g.player.X + g.player.Width/2,
			Y:      g.player.Y + g.player.Height,
			VX:     -g.player.VX * 0.5,
			VY:     g.rng.Float64() - 0.5,
			Life:   1.0,
			MaxLife: 1.0,
			Color:  color.RGBA{100, 200, 255, 150},
			Size:   5 + g.rng.Float64()*5,
			Type:   "dust",
		})
	}

	// Обновление частиц
	activeParticles := make([]entity.ParticleEffect, 0)
	for _, p := range g.player.ParticleTrail {
		p.X += p.VX
		p.Y += p.VY
		p.Life -= 0.05
		if p.Life > 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	g.player.ParticleTrail = activeParticles
}

// applyPhysics применяет физику
func (g *Game) applyPhysics() {
	// Гравитация
	g.player.VY += 0.5

	// Трение
	if g.player.OnGround {
		g.player.VX *= 0.8
	} else {
		g.player.VX *= 0.95
	}

	// Ограничение скорости
	maxSpeed := 8.0
	if g.player.VX > maxSpeed {
		g.player.VX = maxSpeed
	}
	if g.player.VX < -maxSpeed {
		g.player.VX = -maxSpeed
	}

	// Применение скорости
	g.player.X += g.player.VX
	g.player.Y += g.player.VY

	// Сброс флага земли
	g.player.OnGround = false
}

// handleCollisions обрабатывает коллизии
func (g *Game) handleCollisions() {
	// Коллизии с платформами
	for _, platform := range g.world.Platforms {
		if g.checkPlatformCollision(platform) {
			g.resolvePlatformCollision(platform)
		}
	}

	// Коллизии с землёй
	groundY := float64(g.config.ScreenHeight) - 50
	if g.player.Y+g.player.Height > groundY {
		g.player.Y = groundY - g.player.Height
		g.player.VY = 0
		g.player.OnGround = true
	}
}

// checkPlatformCollision проверяет коллизию с платформой
func (g *Game) checkPlatformCollision(platform entity.Platform) bool {
	return g.player.X < platform.X+platform.Width &&
		g.player.X+g.player.Width > platform.X &&
		g.player.Y < platform.Y+platform.Height &&
		g.player.Y+g.player.Height > platform.Y
}

// resolvePlatformCollision разрешает коллизию с платформой
func (g *Game) resolvePlatformCollision(platform entity.Platform) {
	// Определяем направление коллизии
	playerCenterX := g.player.X + g.player.Width/2
	playerCenterY := g.player.Y + g.player.Height/2
	platformCenterX := platform.X + platform.Width/2
	platformCenterY := platform.Y + platform.Height/2

	dx := playerCenterX - platformCenterX
	dy := playerCenterY - platformCenterY

	overlapX := (g.player.Width/2 + platform.Width/2) - math.Abs(dx)
	overlapY := (g.player.Height/2 + platform.Height/2) - math.Abs(dy)

	if overlapX < overlapY {
		// Горизонтальная коллизия
		if dx > 0 {
			g.player.X = platform.X + platform.Width
		} else {
			g.player.X = platform.X - g.player.Width
		}
		g.player.VX = 0
	} else {
		// Вертикальная коллизия
		if dy > 0 {
			g.player.Y = platform.Y + platform.Height
			g.player.VY = 0
		} else {
			g.player.Y = platform.Y - g.player.Height
			g.player.VY = 0
			g.player.OnGround = true
		}
	}
}

// collectCoins собирает монетки
func (g *Game) collectCoins() {
	activeCoins := make([]entity.Coin, 0)
	for _, coin := range g.world.Coins {
		if g.checkCoinCollision(coin) {
			// Разные эффекты для разных типов монеток
			if coin.Type == "gem" {
				g.score += 50
				g.screenShake = 5
				g.spawnParticles(coin.X+coin.Size/2, coin.Y+coin.Size/2, 0, -2, 20, color.RGBA{255, 0, 255, 255})
			} else if coin.Type == "powerup" {
				g.activatePowerup(coin.Powerup)
				g.flashAlpha = 0.5
				g.spawnParticles(coin.X+coin.Size/2, coin.Y+coin.Size/2, 0, -1, 30, color.RGBA{0, 255, 255, 255})
			} else {
				g.coins++
				g.score += 10 * g.combo
				g.combo++
				g.comboTimer = 120 // 2 секунды на 60 FPS
				g.spawnParticles(coin.X+coin.Size/2, coin.Y+coin.Size/2, 0, -1, 8, color.RGBA{255, 215, 0, 255})
			}
		} else {
			activeCoins = append(activeCoins, coin)
		}
	}
	g.world.Coins = activeCoins
}

// activatePowerup активирует бонус
func (g *Game) activatePowerup(powerup string) {
	switch powerup {
	case "doublejump":
		g.player.DoubleJump = true
		g.score += 100
	case "speed":
		g.player.SpeedBoost = 1.5
		g.player.Speed = 7.5
		g.score += 100
	case "invincible":
		g.player.Invincible = 300 // 5 секунд
		g.score += 100
	}
}

// checkCoinCollision проверяет коллизию с монеткой
func (g *Game) checkCoinCollision(coin entity.Coin) bool {
	return g.player.X < coin.X+coin.Size &&
		g.player.X+g.player.Width > coin.X &&
		g.player.Y < coin.Y+coin.Size &&
		g.player.Y+g.player.Height > coin.Y
}

// checkWin проверяет победу
func (g *Game) checkWin() bool {
	flag := g.world.Flag
	return g.player.X < flag.X+flag.Width &&
		g.player.X+g.player.Width > flag.X &&
		g.player.Y < flag.Y+flag.Height &&
		g.player.Y+g.player.Height > flag.Y
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, &Particle{
			X:     x,
			Y:     y,
			VX:    vx + (g.rng.Float64()-0.5)*2,
			VY:    vy + (g.rng.Float64()-0.5)*2,
			Life:  1.0,
			Color: c,
			Size:  3 + g.rng.Float64()*4,
		})
	}
}

// updateParticles обновляет частицы
func (g *Game) updateParticles() {
	activeParticles := make([]*Particle, 0)
	for _, p := range g.particles {
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.1 // Гравитация
		p.Life -= 0.02

		if p.Life > 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	g.particles = activeParticles
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Тряска экрана
	shakeX := 0.0
	shakeY := 0.0
	if g.screenShake > 0 {
		shakeX = (g.rng.Float64() - 0.5) * g.screenShake * 2
		shakeY = (g.rng.Float64() - 0.5) * g.screenShake * 2
	}

	// Небо (градиент)
	for iy := 0; iy < g.config.ScreenHeight; iy++ {
		ratio := float64(iy) / float64(g.config.ScreenHeight)
		r := uint8(135 - ratio*35)
		gr := uint8(206 - ratio*86)
		b := uint8(235)
		vector.DrawFilledRect(screen, float32(shakeX), float32(float64(iy)+shakeY), float32(g.config.ScreenWidth), 1, color.RGBA{r, gr, b, 255}, true)
	}

	// Отрисовка мира
	g.world.Draw(screen, g.cameraX+shakeX)

	// Отрисовка частиц игрока
	for _, p := range g.player.ParticleTrail {
		screenX := p.X - g.cameraX
		alpha := uint8(p.Life * 255)
		vector.DrawFilledRect(screen, float32(screenX-p.Size/2), float32(p.Y-p.Size/2), float32(p.Size), float32(p.Size), color.RGBA{100, 200, 255, alpha}, true)
	}

	// Отрисовка игрока (мигание если неуязвим)
	if g.player.Invincible <= 0 || int(g.player.Invincible)%10 < 5 {
		g.player.Draw(screen, g.cameraX+shakeX)
	}

	// Отрисовка частиц
	for _, p := range g.particles {
		screenX := p.X - g.cameraX
		vector.DrawFilledRect(screen, float32(screenX-p.Size/2), float32(p.Y-p.Size/2), float32(p.Size), float32(p.Size), p.Color, true)
	}

	// Вспышка
	if g.flashAlpha > 0 {
		overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
		overlay.Fill(color.RGBA{255, 255, 255, uint8(g.flashAlpha * 255)})
		op := &ebiten.DrawImageOptions{}
		screen.DrawImage(overlay, op)
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

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.config.ScreenWidth, g.config.ScreenHeight
}

// drawMenu отрисовывает меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := `
╔═══════════════════════════════════════════╗
║      🏠 VILLAGE PLATFORMER 🌲             ║
║         Деревенский Платформер            ║
╠═══════════════════════════════════════════╣
║                                           ║
║      [SPACE] - Начать игру                ║
║      [ESC] - Выход                        ║
║                                           ║
║  🎮 Управление:                           ║
║     A / D или ← / → - Бег                 ║
║     W / ↑ / SPACE - Прыжок                ║
║                                           ║
║  🏃 Исследуй деревню!                     ║
║  💰 Собирай монетки!                      ║
║  🚩 Достигни флага!                       ║
║                                           ║
╚═══════════════════════════════════════════╝
`
	ebitenutil.DebugPrint(screen, title)
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	hudText := fmt.Sprintf(`┌─────────────────────────┐
│  🏠 VILLAGE PLATFORMER  │
├─────────────────────────┤
│  💰 Монеты: %3d          │
│  ❤️  Жизни: %3d           │
│  ⭐ Счёт: %5d            │
│  📍 Уровень: %2d          │
│  🔥 Комбо: x%d           │
└─────────────────────────┘
`, g.coins, g.lives, g.score, g.level, g.combo)

	ebitenutil.DebugPrint(screen, hudText)

	// Индикаторы бонусов
	y := 150
	if g.player.DoubleJump {
		vector.DrawFilledRect(screen, 10, float32(y), 180, 25, color.RGBA{0, 100, 255, 200}, true)
		ebitenutil.DebugPrintAt(screen, "🦘 Двойной прыжок!", 20, y)
		y += 30
	}
	if g.player.SpeedBoost > 1.0 {
		vector.DrawFilledRect(screen, 10, float32(y), 180, 25, color.RGBA{255, 150, 0, 200}, true)
		ebitenutil.DebugPrintAt(screen, "⚡ Скорость!", 20, y)
		y += 30
	}
	if g.player.Invincible > 0 {
		vector.DrawFilledRect(screen, 10, float32(y), 180, 25, color.RGBA{255, 255, 0, 200}, true)
		ebitenutil.DebugPrintAt(screen, "✨ Неуязвимость!", 20, y)
	}
}

// drawPause отрисовывает паузу
func (g *Game) drawPause(screen *ebiten.Image) {
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 128})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	pauseText := `
╔═══════════════════════════════════════╗
║              ⏸️ ПАУЗА                  ║
╠═══════════════════════════════════════╣
║                                       ║
║     [ESC] - Продолжить                ║
║                                       ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, g.config.ScreenWidth/2-180, g.config.ScreenHeight/2-100)
}

// drawGameOver отрисовывает проигрыш
func (g *Game) drawGameOver(screen *ebiten.Image) {
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{50, 0, 0, 180})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	gameOverText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║          💀 GAME OVER 💀              ║
╠═══════════════════════════════════════╣
║                                       ║
║     Счёт: %5d                          ║
║     Уровень: %2d                        ║
║                                       ║
║     [SPACE] - Начать заново           ║
║     [ESC] - Выход                     ║
║                                       ║
╚═══════════════════════════════════════╝
`, g.score, g.level)

	ebitenutil.DebugPrintAt(screen, gameOverText, g.config.ScreenWidth/2-180, g.config.ScreenHeight/2-120)
}
