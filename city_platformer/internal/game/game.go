// Package game - основная игровая логика City Survivor
// Go365 Day 90 - Постапокалиптический платформер
package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
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
	tileSize     = 64
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
	levelData   *level.LevelData
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
	levelGen    *level.LevelGenerator
}

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		state: StateMenu,
		rng:   rng,
	}

	// Загрузка спрайтов
	var err error
	g.spriteSheet, err = sprite.LoadSpriteSheet()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: sprite loading error: %v\n", err)
	}

	g.renderer = render.NewRenderer(g.spriteSheet)
	g.levelGen = level.NewLevelGenerator(rng, tileSize, g.spriteSheet)

	return g
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.levelNum = 1
	g.score = 0
	g.startLevel()
}

// startLevel запускает уровень
func (g *Game) startLevel() {
	g.levelData = g.levelGen.GenerateLevel(g.levelNum)
	
	// Спавн игрока на уровне земли (groundY = data.Height - 2, tileSize = 64)
	// Y позиции платформы минус высота игрока
	spawnY := float64(g.levelData.Height-2)*float64(g.levelData.TileSize) - 60
	
	g.player = entity.NewPlayer(100, spawnY, g.spriteSheet)
	g.enemies = make([]*entity.Enemy, 0)
	g.projectiles = make([]*entity.Projectile, 0)
	g.particles = make([]render.Particle, 0)
	g.cameraX = 0
	g.cameraY = 0

	// Создание врагов из данных уровня
	for _, le := range g.levelData.Enemies {
		if le.Active {
			enemy := entity.NewEnemy(le.X, le.Y, le.Type, g.spriteSheet)
			g.enemies = append(g.enemies, enemy)
		}
	}
}

// Update обновляет игровое состояние
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
		if g.levelNum > 10 {
			g.state = StateVictory
		} else {
			g.startLevel()
			g.state = StatePlaying
		}
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

// updateGame обновляет игровую логику
func (g *Game) updateGame() {
	dt := 1.0 / 60.0 // Фиксированный timestep

	// Управление игроком
	g.handlePlayerInput()

	// Обновление игрока
	g.player.Update(dt)

	// Физика игрока
	g.applyPhysics(dt)

	// Обновление камеры
	g.updateCamera()

	// Сбор предметов
	g.collectItems()

	// Обновление пуль
	g.updateProjectiles(dt)

	// Обновление врагов
	g.updateEnemies(dt)

	// Обновление частиц
	g.updateParticles(dt)

	// Проверка победы (все враги убиты + достигнут выход)
	if len(g.enemies) == 0 && g.levelData.CheckExitReach(g.player.Transform.X, g.player.Transform.Y, g.player.Transform.Width, g.player.Transform.Height) {
		g.state = StateLevelComplete
	}

	// Проверка смерти
	if g.player.Health.Dead {
		g.state = StateGameOver
	}
}

// handlePlayerInput обрабатывает ввод игрока
func (g *Game) handlePlayerInput() {
	// Проверка на лестнице ли игрок
	onLadder := g.checkLadderCollision()

	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if onLadder {
			g.player.MoveLeft()
			g.player.Physics.VelocityY = 0
		} else {
			g.player.MoveLeft()
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		if onLadder {
			g.player.MoveRight()
			g.player.Physics.VelocityY = 0
		} else {
			g.player.MoveRight()
		}
	}

	// Прыжок (W, Up или Пробел)
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.player.Physics.OnGround && !onLadder {
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+20, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{150, 150, 150, 255})
	}

	// Лазание по лестнице
	if onLadder {
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.player.ClimbUp()
		} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.player.ClimbDown()
		}
	} else {
		// Приседание только когда не на лестнице
		if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.player.Crouch()
		} else {
			g.player.Stand()
		}
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

// shoot производит выстрел
func (g *Game) shoot() {
	g.player.Shoot()

	dirX := float64(g.player.Transform.Facing)
	dirY := 0.0

	projectile := entity.NewProjectile(
		g.player.Transform.X+g.player.Transform.Width/2,
		g.player.Transform.Y+g.player.Transform.Height/3,
		dirX*500,
		dirY,
		false,
		15,
		g.spriteSheet,
	)

	g.projectiles = append(g.projectiles, projectile)
}

// applyPhysics применяет физику
func (g *Game) applyPhysics(dt float64) {
	// Проверка на лестнице
	onLadder := g.checkLadderCollision()

	// Сохраняем старую позицию для отката
	oldX := g.player.Transform.X
	oldY := g.player.Transform.Y

	// Гравитация не действует на лестнице
	if !onLadder {
		g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	}

	// Трение
	if !g.player.Physics.IsMoving {
		g.player.Physics.VelocityX *= g.player.Physics.Friction
	}

	// Обновляем позицию
	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.player.Transform.Y += g.player.Physics.VelocityY * dt

	// Коллизии с платформами
	g.player.Physics.OnGround = false
	for _, p := range g.levelData.Platforms {
		if !p.Solid {
			continue
		}
		if g.checkCollision(g.player.Transform, p) {
			// Простая коллизия сверху (приземление)
			if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Y+10 {
				g.player.Transform.Y = p.Y - g.player.Transform.Height
				g.player.Physics.VelocityY = 0
				g.player.Physics.OnGround = true
				g.player.State = entity.PlayerIdle
			} else if g.player.Physics.VelocityY < 0 && oldY >= p.Y+p.Height-10 {
				// Удар головой
				g.player.Transform.Y = p.Y + p.Height
				g.player.Physics.VelocityY = 0
			} else if g.player.Physics.VelocityX > 0 && oldX+g.player.Transform.Width <= p.X+10 {
				// Столкновение слева
				g.player.Transform.X = p.X - g.player.Transform.Width
				g.player.Physics.VelocityX = 0
			} else if g.player.Physics.VelocityX < 0 && oldX >= p.X+p.Width-10 {
				// Столкновение справа
				g.player.Transform.X = p.X + p.Width
				g.player.Physics.VelocityX = 0
			}
		}
	}

	// Пол (если упал с карты)
	// Смерть если игрок упал ниже уровня земли на 200 пикселей
	groundLevel := float64(g.levelData.Height-2)*float64(g.levelData.TileSize)
	if g.player.Transform.Y > groundLevel+200 {
		g.player.Health.TakeDamage(100)
	}

	// Границы уровня
	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
	}
	maxX := float64(g.levelData.Width*g.levelData.TileSize) - g.player.Transform.Width
	if g.player.Transform.X > maxX {
		g.player.Transform.X = maxX
	}
}

// updateCamera обновляет камеру
func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	targetY := g.player.Transform.Y - screenHeight/2

	// Плавное слежение
	g.cameraX += (targetX - g.cameraX) * 0.1
	g.cameraY += (targetY - g.cameraY) * 0.1

	// Ограничения камеры
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	maxCameraX := float64(g.levelData.Width*g.levelData.TileSize) - screenWidth
	if g.cameraX > maxCameraX {
		g.cameraX = maxCameraX
	}
	if g.cameraY < -100 {
		g.cameraY = -100
	}
	maxCameraY := float64(g.levelData.Height*g.levelData.TileSize) - screenHeight
	if g.cameraY > maxCameraY {
		g.cameraY = maxCameraY
	}
}

// checkLadderCollision проверяет, находится ли игрок на лестнице
func (g *Game) checkLadderCollision() bool {
	playerCenter := g.player.Transform.X + g.player.Transform.Width/2
	playerBottom := g.player.Transform.Y + g.player.Transform.Height

	for _, platform := range g.levelData.Platforms {
		if platform.Type == level.TileLadder {
			// Проверяем, находится ли игрок в зоне лестницы
			if playerCenter >= platform.X && playerCenter <= platform.X+platform.Width &&
				playerBottom >= platform.Y && playerBottom <= platform.Y+platform.Height+10 {
				return true
			}
		}
	}
	return false
}

// collectItems обрабатывает сбор предметов
func (g *Game) collectItems() {
	item := g.levelData.CheckItemCollection(
		g.player.Transform.X,
		g.player.Transform.Y,
		g.player.Transform.Width,
		g.player.Transform.Height,
	)

	if item != nil {
		g.collectItem(item)
	}
}

// collectItem обрабатывает сбор одного предмета
func (g *Game) collectItem(item *level.LevelItem) {
	switch item.Type {
	case "coinGold", "coinSilver", "coinBronze":
		g.score += item.Value
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 10, color.RGBA{255, 215, 0, 255})
	case "gemRed", "gemBlue", "gemGreen", "gemYellow":
		g.score += item.Value
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 15, color.RGBA{100, 200, 255, 255})
	case "star":
		g.score += item.Value
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 20, color.RGBA{255, 255, 255, 255})
	case "food-OCAL", "food-Glitch", "food-DCSS", "mushroomRed", "mushroomBrown":
		g.player.Health.Heal(item.Value)
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 10, color.RGBA{255, 100, 100, 255})
	default:
		g.score += item.Value
	}
}

// updateProjectiles обновляет снаряды
func (g *Game) updateProjectiles(dt float64) {
	active := make([]*entity.Projectile, 0)

	for _, p := range g.projectiles {
		p.Update(dt)

		if p.Active {
			// Коллизия с врагами (пули игрока)
			if !p.IsEnemy {
				for i, enemy := range g.enemies {
					if entity.CheckCollision(p.Transform, enemy.Transform) {
						enemy.Health.TakeDamage(p.Damage)
						p.Active = false
						g.spawnParticles(p.Transform.X, p.Transform.Y, p.VelocityX*0.2, p.VelocityY*0.2, 5, color.RGBA{255, 100, 50, 255})

						if enemy.Health.Dead {
							g.killEnemy(i, enemy)
						}
						break
					}
				}
			}

			// Коллизия с игроком (пули врагов)
			if p.IsEnemy && p.Active {
				if entity.CheckCollision(p.Transform, g.player.Transform) {
					g.player.Health.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -100, 10, color.RGBA{255, 50, 50, 255})
				}
			}

			// Коллизия с платформами
			if p.Active {
				for _, platform := range g.levelData.Platforms {
					if platform.Solid && g.checkCollision(p.Transform, platform) {
						p.Active = false
						g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -50, 3, color.RGBA{200, 200, 200, 255})
						break
					}
				}
			}

			if p.Active {
				active = append(active, p)
			}
		}
	}

	g.projectiles = active
}

// updateEnemies обновляет врагов
func (g *Game) updateEnemies(dt float64) {
	active := make([]*entity.Enemy, 0)

	for _, enemy := range g.enemies {
		enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y)
		enemy.AI(dt, g.player.Transform.X, g.player.Transform.Y)

		// Коллизия с игроком
		if entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 {
				g.player.Health.TakeDamage(enemy.Damage)
			}
		}

		if enemy.Health.IsAlive() {
			active = append(active, enemy)
		}
	}

	g.enemies = active
}

// killEnemy убивает врага
func (g *Game) killEnemy(index int, enemy *entity.Enemy) {
	g.score += 25
	g.spawnParticles(enemy.Transform.X+20, enemy.Transform.Y+20, 0, -100, 15, color.RGBA{100, 150, 50, 255})
	g.enemies = append(g.enemies[:index], g.enemies[index+1:]...)
}

// updateParticles обновляет частицы
func (g *Game) updateParticles(dt float64) {
	active := make([]render.Particle, 0)

	for _, p := range g.particles {
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt // Гравитация
		p.Life -= dt * 0.5
		p.VX *= 0.98

		if p.Life > 0 {
			active = append(active, p)
		}
	}

	g.particles = active
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, render.Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life:    1.0,
			MaxLife: 1.0,
			Color:   c,
			Size:    3 + g.rng.Float64()*4,
		})
	}
}

// checkCollision проверяет коллизию между сущностью и платформой
func (g *Game) checkCollision(transform *entity.Transform, platform *level.Platform) bool {
	return transform.X < platform.X+platform.Width &&
		transform.X+transform.Width > platform.X &&
		transform.Y < platform.Y+platform.Height &&
		transform.Y+transform.Height > platform.Y
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	g.renderer.DrawBackground(screen, g.cameraX, g.cameraY)

	// Отрисовка игровых объектов только если уровень загружен
	if g.levelData != nil {
		// Платформы
		for _, p := range g.levelData.Platforms {
			g.renderer.DrawPlatform(screen, p, g.cameraX, g.cameraY)
		}

		// Предметы
		for _, item := range g.levelData.Items {
			if !item.Collected {
				entityItem := entity.NewItem(item.X, item.Y, item.Type, item.Value, g.spriteSheet)
				g.renderer.DrawItem(screen, entityItem, g.cameraX, g.cameraY)
			}
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
		g.renderer.DrawExit(screen, g.levelData.ExitX, g.levelData.ExitY, g.cameraX, g.cameraY)
	}

	// Частицы
	g.renderer.DrawParticles(screen, g.particles, g.cameraX, g.cameraY)

	// UI
	switch g.state {
	case StateMenu:
		g.renderer.DrawMenu(screen)
	case StatePlaying:
		g.drawHUD(screen)
	case StatePaused:
		g.drawHUD(screen)
		g.renderer.DrawPause(screen)
	case StateGameOver:
		g.renderer.DrawGameOver(screen, g.score, g.levelNum)
	case StateVictory:
		g.renderer.DrawVictory(screen, g.score)
	case StateLevelComplete:
		g.renderer.DrawLevelComplete(screen, g.levelNum, g.score)
	}
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	g.renderer.DrawHUD(
		screen,
		g.player.Health.Current,
		g.player.Health.Max,
		g.player.Ammo,
		g.player.MaxAmmo,
		g.score,
		g.levelNum,
		g.levelData.Name,
	)

	// Подсказки
	ebitenutil.DebugPrintAt(screen, "[ESC] - Пауза  [J] - Огонь  [K] - Перезарядка", 10, screenHeight-30)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
