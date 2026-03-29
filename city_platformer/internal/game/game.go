// Package game - основная игровая логика City Platformer
// Go365 Day 91 - Постапокалиптический город
package game

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"city_platformer/internal/entity"
	"city_platformer/internal/level"
	"city_platformer/internal/render"
	"city_platformer/internal/sprite"
)

const (
	screenWidth  = 1280
	screenHeight = 720
)

// GameState - состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
	StateLevelComplete
)

// Game - основная игровая структура
type Game struct {
	state       GameState
	player      *entity.Player
	level       *level.Level
	enemies     []*entity.Enemy
	projectiles []*entity.Projectile
	particles   []render.Particle
	cameraX     float64
	cameraY     float64
	score       int
	levelNum    int
	spriteSheet *sprite.SpriteSheet
	renderer    *render.Renderer
	rng         *rand.Rand
}

// NewGame - создание новой игры
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		state: StateMenu,
		rng:   rng,
	}
	g.spriteSheet = sprite.LoadSpriteSheet()
	g.renderer = render.NewRenderer(g.spriteSheet)
	return g
}

// Reset - сброс игры
func (g *Game) Reset() {
	g.levelNum = 1
	g.score = 0
	g.startLevel()
}

// startLevel - запуск уровня
func (g *Game) startLevel() {
	g.level = level.GenerateLevel(g.levelNum, g.rng)
	g.player = entity.NewPlayer(100, 500, g.spriteSheet)
	g.enemies = make([]*entity.Enemy, 0)
	g.projectiles = make([]*entity.Projectile, 0)
	g.particles = make([]render.Particle, 0)
	g.cameraX = 0
	g.cameraY = 0

	// Создание врагов из данных уровня
	for _, le := range g.level.Enemies {
		if le.Active {
			enemy := entity.NewEnemy(le.X, le.Y, le.Type, g.spriteSheet)
			g.enemies = append(g.enemies, enemy)
		}
	}
}

// Update - обновление игрового состояния
func (g *Game) Update() error {
	// Обработка ESC
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused:
			g.state = StatePlaying
		case StateMenu, StateGameOver, StateVictory:
			return ebiten.Termination
		}
	}

	// Меню - старт игры
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	// Game Over - рестарт
	if g.state == StateGameOver && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	// Level Complete - следующий уровень
	if g.state == StateLevelComplete && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.levelNum++
		g.startLevel()
		g.state = StatePlaying
	}

	// Victory - рестарт
	if g.state == StateVictory && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	// Игровой процесс
	if g.state == StatePlaying {
		g.updateGame()
	}

	return nil
}

// updateGame - обновление игровой логики
func (g *Game) updateGame() {
	// Управление игроком
	g.handlePlayerInput()

	// Обновление игрока
	g.player.Update()

	// Физика игрока
	g.applyPhysics()

	// Обновление камеры
	g.updateCamera()

	// Сбор предметов
	g.collectItems()

	// Обновление пуль
	g.updateProjectiles()

	// Обновление врагов
	g.updateEnemies()

	// Обновление частиц
	g.updateParticles()

	// Проверка победы (все враги убиты + достигнут выход)
	if len(g.enemies) == 0 && g.level.CheckExitReach(g.player.X, g.player.Y, g.player.Width, g.player.Height) {
		if g.levelNum >= 10 {
			g.state = StateVictory
		} else {
			g.state = StateLevelComplete
		}
	}

	// Проверка смерти
	if g.player.Health <= 0 {
		g.state = StateGameOver
	}
}

// handlePlayerInput - обработка ввода игрока
func (g *Game) handlePlayerInput() {
	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.MoveLeft()
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.MoveRight()
	}

	// Прыжок
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp)) && g.player.CanJump() {
		g.player.Jump()
		g.spawnParticles(g.player.X+16, g.player.Y+g.player.Height, 0, -2, 8, color.RGBA{150, 150, 150, 255})
	}

	// Приседание
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Crouch()
	} else {
		g.player.Stand()
	}

	// Выстрел
	if ebiten.IsKeyPressed(ebiten.KeyJ) && g.player.CanShoot() {
		g.shoot()
	}

	// Перезарядка
	if ebiten.IsKeyPressed(ebiten.KeyK) {
		g.player.Reload()
	}
}

// shoot - выстрел игрока
func (g *Game) shoot() {
	g.player.Shoot()

	dirX := float64(g.player.Facing)
	dirY := 0.0

	projectile := entity.NewProjectile(
		g.player.X+g.player.Width,
		g.player.Y+g.player.Height/2,
		dirX*12,
		dirY,
		false,
		10,
		g.spriteSheet,
	)

	g.projectiles = append(g.projectiles, projectile)
}

// applyPhysics - применение физики
func (g *Game) applyPhysics() {
	g.player.VY += 0.6 // Гравитация

	if !g.player.IsMoving {
		g.player.VX *= 0.8
	}

	g.player.X += g.player.VX
	g.player.Y += g.player.VY

	// Коллизии с платформами
	g.player.OnGround = false
	for _, p := range g.level.Platforms {
		if g.checkCollision(g.player.X+4, g.player.Y+4, g.player.Width-8, g.player.Height-8, p.X, p.Y, p.Width, p.Height) {
			// Простая коллизия сверху
			if g.player.VY > 0 && g.player.Y+g.player.Height < p.Y+30 {
				g.player.Y = p.Y - g.player.Height
				g.player.VY = 0
				g.player.OnGround = true
			}
		}
	}

	// Пол (если упал с карты)
	if g.player.Y > float64(screenHeight)+100 {
		g.player.TakeDamage(100)
	}

	// Границы уровня
	if g.player.X < 0 {
		g.player.X = 0
	}
	if g.player.X > g.level.Width-g.player.Width {
		g.player.X = g.level.Width - g.player.Width
	}
}

// updateCamera - обновление камеры
func (g *Game) updateCamera() {
	targetX := g.player.X - screenWidth/2
	targetY := g.player.Y - screenHeight/2

	g.cameraX += (targetX - g.cameraX) * 0.1
	g.cameraY += (targetY - g.cameraY) * 0.1

	// Ограничения камеры
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > g.level.Width-screenWidth {
		g.cameraX = g.level.Width - screenWidth
	}
	if g.cameraY < -100 {
		g.cameraY = -100
	}
}

// collectItems - сбор предметов
func (g *Game) collectItems() {
	for _, item := range g.level.Items {
		if !item.Collected {
			if g.checkCollision(
				g.player.X, g.player.Y, g.player.Width, g.player.Height,
				item.X, item.Y, item.Width, item.Height,
			) {
				item.Collected = true
				g.collectItem(item)
			}
		}
	}
}

// collectItem - обработка сбора предмета
func (g *Game) collectItem(item *level.LevelItem) {
	switch item.Type {
	case "medkit":
		g.player.Heal(item.Value)
		g.spawnParticles(g.player.X+16, g.player.Y+24, 0, -2, 10, color.RGBA{255, 100, 100, 255})
	case "ammo":
		g.player.AddAmmo(item.Value)
		g.spawnParticles(g.player.X+16, g.player.Y+24, 0, -2, 10, color.RGBA{255, 200, 50, 255})
	case "food":
		g.player.Heal(item.Value)
		g.spawnParticles(g.player.X+16, g.player.Y+24, 0, -2, 10, color.RGBA{200, 150, 50, 255})
	case "parts":
		g.score += item.Value
		g.spawnParticles(g.player.X+16, g.player.Y+24, 0, -2, 10, color.RGBA{150, 150, 200, 255})
	}
}

// updateProjectiles - обновление снарядов
func (g *Game) updateProjectiles() {
	active := make([]*entity.Projectile, 0)

	for _, p := range g.projectiles {
		p.Update()

		if p.Active {
			// Коллизия с врагами (пули игрока)
			if !p.IsEnemy {
				for i, enemy := range g.enemies {
					if g.checkCollision(p.X, p.Y, p.Width, p.Height, enemy.X, enemy.Y, enemy.Width, enemy.Height) {
						enemy.TakeDamage(p.Damage)
						p.Active = false
						g.spawnParticles(p.X, p.Y, p.VX, p.VY, 5, color.RGBA{255, 100, 50, 255})

						if !enemy.IsAlive() {
							g.killEnemy(i)
						}
						break
					}
				}
			}

			// Коллизия с игроком (пули врагов)
			if p.IsEnemy && p.Active {
				if g.checkCollision(p.X, p.Y, p.Width, p.Height, g.player.X, g.player.Y, g.player.Width, g.player.Height) {
					g.player.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.X, p.Y, 0, -2, 10, color.RGBA{255, 50, 50, 255})
				}
			}

			if p.Active {
				active = append(active, p)
			}
		}
	}

	g.projectiles = active
}

// updateEnemies - обновление врагов
func (g *Game) updateEnemies() {
	active := make([]*entity.Enemy, 0)

	for _, enemy := range g.enemies {
		enemy.Update(g.player.X, g.player.Y)

		// ИИ врагов
		g.updateEnemyAI(enemy)

		// Коллизия с игроком
		if g.checkCollision(
			g.player.X+8, g.player.Y+8,
			g.player.Width-16, g.player.Height-16,
			enemy.X+8, enemy.Y+8,
			enemy.Width-16, enemy.Height-16,
		) {
			if g.player.Invincible <= 0 {
				g.player.TakeDamage(enemy.Damage)
			}
		}

		if enemy.IsAlive() {
			active = append(active, enemy)
		}
	}

	g.enemies = active
}

// updateEnemyAI - ИИ врагов
func (g *Game) updateEnemyAI(enemy *entity.Enemy) {
	distToPlayer := math.Abs(g.player.X - enemy.X)

	switch enemy.Type {
	case "mutant":
		// Мутант медленно идёт к игроку
		if distToPlayer < 400 {
			if g.player.X > enemy.X {
				enemy.X += 1.5
				enemy.Facing = 1
			} else {
				enemy.X -= 1.5
				enemy.Facing = -1
			}
		}

	case "robot":
		// Робот стреляет если игрок в зоне видимости
		if distToPlayer < 300 && enemy.CanShoot() {
			enemy.Shoot()
			dirX := float64(enemy.Facing)
			projectile := entity.NewProjectile(
				enemy.X+enemy.Width,
				enemy.Y+enemy.Height/2,
				dirX*8,
				0,
				true,
				15,
				g.spriteSheet,
			)
			g.projectiles = append(g.projectiles, projectile)
		}

	case "zombie":
		// Зомби медленно преследует
		if distToPlayer < 300 {
			if g.player.X > enemy.X {
				enemy.X += 0.8
				enemy.Facing = 1
			} else {
				enemy.X -= 0.8
				enemy.Facing = -1
			}
		}
	}
}

// killEnemy - убийство врага
func (g *Game) killEnemy(index int) {
	enemy := g.enemies[index]
	g.score += 25
	g.spawnParticles(enemy.X+20, enemy.Y+20, 0, -3, 15, color.RGBA{100, 150, 50, 255})
	g.enemies = append(g.enemies[:index], g.enemies[index+1:]...)
}

// updateParticles - обновление частиц
func (g *Game) updateParticles() {
	active := make([]render.Particle, 0)

	for _, p := range g.particles {
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.1
		p.Life -= 0.02
		p.VX *= 0.98

		if p.Life > 0 {
			active = append(active, p)
		}
	}

	g.particles = active
}

// spawnParticles - создание частиц
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, render.Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*4,
			VY: vy + (g.rng.Float64()-0.5)*4,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   c,
			Size:    3 + g.rng.Float64()*3,
		})
	}
}

// checkCollision - проверка коллизии AABB
func (g *Game) checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// Draw - отрисовка игры
func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	g.renderer.DrawBackground(screen, g.cameraX, g.cameraY)

	// Отрисовка игровых объектов только если уровень загружен
	if g.level != nil {
		// Платформы
		for _, p := range g.level.Platforms {
			g.renderer.DrawPlatform(screen, p, g.cameraX, g.cameraY)
		}

		// Предметы
		for _, item := range g.level.Items {
			g.renderer.DrawItem(screen, item, g.cameraX, g.cameraY)
		}

		// Враги
		for _, enemy := range g.enemies {
			g.renderer.DrawEnemy(screen, enemy, g.cameraX, g.cameraY)
		}

		// Снаряды
		for _, p := range g.projectiles {
			g.renderer.DrawProjectile(screen, p, g.cameraX, g.cameraY)
		}

		// Игрок
		if g.player != nil {
			g.renderer.DrawPlayer(screen, g.player, g.cameraX, g.cameraY)
		}

		// Выход
		g.renderer.DrawExit(screen, g.level.ExitX, 590, g.cameraX, g.cameraY)
	}

	// Частицы
	g.renderer.DrawParticles(screen, g.particles, g.cameraX, g.cameraY)

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
	case StateVictory:
		g.drawVictory(screen)
	case StateLevelComplete:
		g.drawLevelComplete(screen)
	}
}

// drawMenu - отрисовка меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := `
╔═══════════════════════════════════════════════╗
║     🏙️ CITY PLATFORMER 🎖️                    ║
║        LAST SURVIVOR                          ║
╠═══════════════════════════════════════════════╣
║                                               ║
║           [SPACE] - Начать игру               ║
║           [ESC] - Выход                       ║
║                                               ║
║  🎮 Управление:                               ║
║     A/D или ←/→ - Бег                         ║
║     W/↑ - Прыжок                              ║
║     S/↓ - Присесть                            ║
║     J - Выстрел                               ║
║     K - Перезарядка                           ║
║                                               ║
║  🎯 Цель: Доберись до точки эвакуации!        ║
║  💀 Остерегайся мутантов, роботов и зомби!    ║
║  📦 Собирай припасы: аптечки, патроны, еду    ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrint(screen, title)
}

// drawHUD - отрисовка интерфейса
func (g *Game) drawHUD(screen *ebiten.Image) {
	g.renderer.DrawHUD(
		screen,
		g.player.Health,
		g.player.MaxHealth,
		g.player.Ammo,
		g.player.MaxAmmo,
		g.score,
		g.levelNum,
		g.level.Name,
	)

	// Подсказки
	ebitenutil.DebugPrintAt(screen, "[ESC] - Пауза  [J] - Огонь  [K] - Перезарядка", 10, screenHeight-30)

	// Полоска здоровья босса (если есть)
	// TODO: добавить боссов
}

// drawPause - отрисовка паузы
func (g *Game) drawPause(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	pauseText := `
╔═══════════════════════════════════════╗
║              ⏸️ ПАУЗА                  ║
╠═══════════════════════════════════════╣
║     [ESC] - Продолжить                ║
║     [SPACE] - Выйти в меню            ║
╚═══════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, pauseText, screenWidth/2-180, screenHeight/2-100)
}

// drawGameOver - отрисовка Game Over
func (g *Game) drawGameOver(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{80, 0, 0, 200})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	gameOverText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║       💀 ВЫ ПОГИБЛИ 💀                ║
╠═══════════════════════════════════════╣
║     Финальный счёт: %6d                ║
║     Уровень: %2d                        ║
║                                       ║
║     [SPACE] - Новая попытка           ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, g.score, g.levelNum)

	ebitenutil.DebugPrintAt(screen, gameOverText, screenWidth/2-180, screenHeight/2-120)
}

// drawVictory - отрисовка победы
func (g *Game) drawVictory(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 80, 0, 150})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	victoryText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║     🚁 ЭВАКУАЦИЯ УСПЕШНА! 🚁          ║
╠═══════════════════════════════════════╣
║     Вы прошли все уровни!             ║
║     Финальный счёт: %6d                ║
║                                       ║
║     🎉 ПОЗДРАВЛЯЕМ! 🎉                ║
║                                       ║
║     [SPACE] - Играть снова            ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, g.score)

	ebitenutil.DebugPrintAt(screen, victoryText, screenWidth/2-180, screenHeight/2-140)
}

// drawLevelComplete - отрисовка завершения уровня
func (g *Game) drawLevelComplete(screen *ebiten.Image) {
	overlay := ebiten.NewImage(screenWidth, screenHeight)
	overlay.Fill(color.RGBA{0, 80, 0, 150})
	op := &ebiten.DrawImageOptions{}
	screen.DrawImage(overlay, op)

	levelCompleteText := fmt.Sprintf(`
╔═══════════════════════════════════════╗
║     ✅ УРОВЕНЬ %d ПРОЙДЕН! ✅          ║
╠═══════════════════════════════════════╣
║     Счёт: %6d                          ║
║                                       ║
║     [SPACE] - Следующий уровень       ║
║     [ESC] - Пауза                     ║
╚═══════════════════════════════════════╝
`, g.levelNum, g.score)

	ebitenutil.DebugPrintAt(screen, levelCompleteText, screenWidth/2-180, screenHeight/2-100)
}

// Layout - размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
