// Package game содержит основную игровую логику
package game

import (
	"fmt"
	"image/color"
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
	StateUpgrade
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
	config       *Config
	state        State
	score        int
	wave         int
	waveTimer    float64
	enemies      []*entity.Enemy
	projectiles  []*entity.Projectile
	player       *entity.Player
	spawnTimer   float64
	spawnRate    float64
	rng          *rand.Rand
	waveEnemies  int
	wavesSurvived int
}

// NewGame создаёт новую игру
func NewGame() *Game {
	cfg := DefaultConfig()
	return &Game{
		config:      cfg,
		state:       StateMenu,
		score:       0,
		wave:        1,
		player:      entity.NewPlayer(640, 360),
		enemies:     make([]*entity.Enemy, 0),
		projectiles: make([]*entity.Projectile, 0),
		spawnRate:   60.0,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update обновляет логику игры (вызывается каждый кадр)
func (g *Game) Update() error {
	// Обработка ввода для смены состояний
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

	// Простой переход из меню в игру по пробелу
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.resetGame()
		g.state = StatePlaying
	}

	// Рестарт из Game Over
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

// resetGame сбрасывает игру в начальное состояние
func (g *Game) resetGame() {
	g.player = entity.NewPlayer(640, 360)
	g.enemies = make([]*entity.Enemy, 0)
	g.projectiles = make([]*entity.Projectile, 0)
	g.score = 0
	g.wave = 1
	g.waveTimer = 0
	g.spawnTimer = 0
	g.spawnRate = 60.0
	g.waveEnemies = 0
	g.wavesSurvived = 0
}

// updateGame обновляет игровую логику
func (g *Game) updateGame() {
	// Обновление игрока
	g.player.Update()

	// Управление игроком
	g.handleInput()

	// Спавн врагов
	g.spawnEnemies()

	// Обновление врагов
	for _, enemy := range g.enemies {
		enemy.Update(g.player)

		// Атака врага на игрока
		if enemy.CanAttack() {
			dist := g.player.DistanceTo(enemy.X, enemy.Y)
			if dist < 30 { // Ближний бой
				g.player.TakeDamage(enemy.Damage)
				enemy.ResetAttack()
			}
		}
	}

	// Авто-атака игрока
	g.playerAutoAttack()

	// Обновление снарядов
	for _, proj := range g.projectiles {
		proj.Update()

		// Коллизия с врагами
		for _, enemy := range g.enemies {
			if proj.IsActive && enemy.IsActive {
				if g.checkCollision(proj.Entity, enemy.Entity) {
					enemy.TakeDamage(proj.Damage)
					proj.IsActive = false

					if !enemy.IsActive {
						g.score += int(enemy.XPValue * 10)
						g.player.AddXP(enemy.XPValue)
					}
				}
			}
		}
	}

	// Очистка неактивных сущностей
	g.cleanupEntities()

	// Проверка волны
	g.updateWave()

	// Проверка проигрыша
	if g.player.HP <= 0 {
		g.state = StateGameOver
	}
}

// handleInput обрабатывает ввод игрока
func (g *Game) handleInput() {
	speed := g.player.MoveSpeed

	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
		g.player.Y -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Y += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.X -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.X += speed
	}

	// Ограничение границами экрана
	if g.player.X < g.player.Width/2 {
		g.player.X = g.player.Width / 2
	}
	if g.player.X > float64(g.config.ScreenWidth)-g.player.Width/2 {
		g.player.X = float64(g.config.ScreenWidth) - g.player.Width/2
	}
	if g.player.Y < g.player.Height/2 {
		g.player.Y = g.player.Height / 2
	}
	if g.player.Y > float64(g.config.ScreenHeight)-g.player.Height/2 {
		g.player.Y = float64(g.config.ScreenHeight) - g.player.Height/2
	}
}

// spawnEnemies спавнит врагов
func (g *Game) spawnEnemies() {
	g.spawnTimer--

	if g.spawnTimer <= 0 && g.waveEnemies < g.getWaveMaxEnemies() {
		g.spawnEnemy()
		g.spawnTimer = g.spawnRate
	}
}

// spawnEnemy спавнит одного врага
func (g *Game) spawnEnemy() {
	// Спавн за пределами экрана
	var x, y float64
	side := g.rng.Intn(4)

	switch side {
	case 0: // Сверху
		x = float64(g.rng.Intn(g.config.ScreenWidth))
		y = -30
	case 1: // Снизу
		x = float64(g.rng.Intn(g.config.ScreenWidth))
		y = float64(g.config.ScreenHeight) + 30
	case 2: // Слева
		x = -30
		y = float64(g.rng.Intn(g.config.ScreenHeight))
	case 3: // Справа
		x = float64(g.config.ScreenWidth) + 30
		y = float64(g.rng.Intn(g.config.ScreenHeight))
	}

	enemyType := entity.GetEnemyTypeForWave(g.wave)
	
	// Босс спавнится в центре
	if enemyType == entity.EnemyBoss {
		x = float64(g.config.ScreenWidth) / 2
		y = -50
	}

	enemy := entity.NewEnemy(x, y, enemyType)
	g.enemies = append(g.enemies, enemy)
	g.waveEnemies++
}

// getWaveMaxEnemies возвращает макс. количество врагов в волне
func (g *Game) getWaveMaxEnemies() int {
	return 5 + g.wave*3
}

// playerAutoAttack реализует авто-атаку игрока
func (g *Game) playerAutoAttack() {
	if g.player.CanAttack() {
		proj := entity.FromPlayerToEnemy(
			g.player,
			g.enemies,
			entity.ProjectileBullet,
			8.0,
			g.player.Damage,
			3.0,
		)

		if proj != nil {
			g.projectiles = append(g.projectiles, proj)
			g.player.ResetAttack()
		}
	}
}

// checkCollision проверяет коллизию между сущностями
func (g *Game) checkCollision(a, b *entity.Entity) bool {
	left1, right1, top1, bottom1 := a.Bounds()
	left2, right2, top2, bottom2 := b.Bounds()

	return left1 < right2 && right1 > left2 && top1 < bottom2 && bottom1 > top2
}

// cleanupEntities удаляет неактивные сущности
func (g *Game) cleanupEntities() {
	// Очистка врагов
	activeEnemies := make([]*entity.Enemy, 0)
	for _, e := range g.enemies {
		if e.IsActive {
			activeEnemies = append(activeEnemies, e)
		}
	}
	g.enemies = activeEnemies

	// Очистка снарядов
	activeProjs := make([]*entity.Projectile, 0)
	for _, p := range g.projectiles {
		if p.IsActive {
			activeProjs = append(activeProjs, p)
		}
	}
	g.projectiles = activeProjs
}

// updateWave обновляет состояние волны
func (g *Game) updateWave() {
	// Проверка завершения волны
	if g.waveEnemies >= g.getWaveMaxEnemies() && len(g.enemies) == 0 {
		g.wave++
		g.waveEnemies = 0
		g.wavesSurvived++
		g.spawnRate = max(10, 60-float64(g.wave)*2)

		// Лечение игрока между волнами
		g.player.Heal(20)
	}
}

// Draw отрисовывает игру (вызывается каждый кадр)
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана (тёмно-фиолетовый фон подземелья)
	screen.Fill(color.RGBA{20, 10, 40, 255})

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying, StatePaused:
		g.drawGame(screen)
		if g.state == StatePaused {
			g.drawPause(screen)
		}
	case StateGameOver:
		g.drawGame(screen)
		g.drawGameOver(screen)
	}
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.config.ScreenWidth, g.config.ScreenHeight
}

// drawMenu отрисовывает главное меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, `
╔══════════════════════════════════════════╗
║      🎮 DUNGEON SURVIVOR - GO90 🎮       ║
║           Roguelike Adventure            ║
╠══════════════════════════════════════════╣
║                                          ║
║         [SPACE] - Начать игру            ║
║         [ESC] - Выход                    ║
║                                          ║
║   Управление: WASD или Стрелки           ║
║                                          ║
║   Выживай как можно дольше!              ║
║   Собирай опыт, получай улучшения!       ║
║                                          ║
║   🗡️ Авто-атака                          ║
║   👾 Орды врагов                         ║
║   ⬆️ Улучшения                           ║
║   💀 Permadeath                          ║
║                                          ║
╚══════════════════════════════════════════╝
`)
}

// drawGame отрисовывает игровой процесс
func (g *Game) drawGame(screen *ebiten.Image) {
	// Отрисовка сущностей
	for _, enemy := range g.enemies {
		enemy.Draw(screen)
	}

	for _, proj := range g.projectiles {
		proj.Draw(screen)
	}

	g.player.Draw(screen)

	// UI
	g.drawUI(screen)
}

// drawUI отрисовывает интерфейс
func (g *Game) drawUI(screen *ebiten.Image) {
	// Фон UI
	vector.DrawFilledRect(screen, 0, 0, 220, 140, color.RGBA{0, 0, 0, 180}, true)

	hpPercent := g.player.HP / g.player.MaxHP
	xpPercent := g.player.XP / g.player.XPToLevel

	uiText := fmt.Sprintf(`┌─────────────────────────┐
│  DUNGEON SURVIVOR       │
├─────────────────────────┤
│  ❤️  HP: %3d/%3d %.0f%%   │
│  ⭐ Level: %-3d          │
│  📊 XP: %.0f%%            │
│  🌊 Wave: %-3d           │
│  💀 Killed: %-4d         │
│  🏆 Score: %-6d          │
└─────────────────────────┘

[ESC] - Пауза
`, int(g.player.HP), int(g.player.MaxHP), hpPercent*100,
		g.player.Level,
		xpPercent*100,
		g.wave,
		len(g.enemies),
		g.score)

	ebitenutil.DebugPrint(screen, uiText)
}

// drawPause отрисовывает меню паузы
func (g *Game) drawPause(screen *ebiten.Image) {
	// Полупрозрачный оверлей
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 128})

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	ebitenutil.DebugPrint(screen, `
╔══════════════════════════════════════════╗
║                ⏸️ ПАУЗА                   ║
╠══════════════════════════════════════════╣
║                                          ║
║      [ESC] - Продолжить                  ║
║                                          ║
╚══════════════════════════════════════════╝
`)
}

// drawGameOver отрисовывает экран проигрыша
func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Полупрозрачный оверлей
	overlay := ebiten.NewImage(g.config.ScreenWidth, g.config.ScreenHeight)
	overlay.Fill(color.RGBA{50, 0, 0, 180})

	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	gameOverText := fmt.Sprintf(`
╔══════════════════════════════════════════╗
║            💀 GAME OVER 💀               ║
╠══════════════════════════════════════════╣
║                                          ║
║         Волн пережито: %-3d              ║
║         Уровень: %-3d                    ║
║         Счёт: %-6d                       ║
║                                          ║
║      [SPACE] - Начать заново             ║
║      [ESC] - Выход в меню                ║
║                                          ║
╚══════════════════════════════════════════╝
`, g.wavesSurvived, g.player.Level, g.score)

	ebitenutil.DebugPrint(screen, gameOverText)
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
