// Package game - основная игровая логика Cyber City Runner
// Go365 Day 91 - Киберпанк платформер
package game

import (
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"cyber_city_runner/internal/entity"
	"cyber_city_runner/internal/level"
	"cyber_city_runner/internal/render"
	"cyber_city_runner/internal/sprite"
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
	StateCharacterSelect
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
	StateLevelComplete
)

// Game - основная игровая структура
type Game struct {
	state           GameState
	player          *entity.Player
	charType        entity.CharacterType
	charSelectIndex int
	levelData       *level.LevelData
	enemies         []*entity.Enemy
	projectiles     []*entity.Projectile
	particles       []render.Particle
	cameraX         float64
	cameraY         float64
	score           int
	levelNum        int
	combo           int
	maxCombo        int
	spriteSheet     *sprite.SpriteSheet
	renderer        *render.Renderer
	rng             *rand.Rand
	levelGen        *level.LevelGenerator
	dashPressed     bool
	shootPressed    bool
}

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		state: StateMenu,
		rng:   rng,
	}

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
	g.combo = 0
	g.maxCombo = 0
	g.startLevel()
}

// startLevel запускает уровень
func (g *Game) startLevel() {
	g.levelData = g.levelGen.GenerateLevel(g.levelNum)

	groundLevel := float64(g.levelData.Height-2)*float64(g.levelData.TileSize)

	g.player = entity.NewPlayer(100, groundLevel, g.charType, g.spriteSheet)
	g.player.Physics.OnGround = true
	g.player.Physics.VelocityY = 0

	g.enemies = make([]*entity.Enemy, 0)
	g.projectiles = make([]*entity.Projectile, 0)
	g.particles = make([]render.Particle, 0)
	g.cameraX = 0
	g.cameraY = 0

	for _, le := range g.levelData.Enemies {
		if le.Active {
			enemy := entity.NewEnemy(le.X, le.Y, le.Type, g.spriteSheet)
			g.enemies = append(g.enemies, enemy)
		}
	}
}

// Update обновляет игровое состояние
func (g *Game) Update() error {
	// ESC - пауза/меню
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying:
			g.state = StatePaused
		case StatePaused:
			g.state = StatePlaying
		case StateCharacterSelect:
			g.state = StateMenu
		case StateMenu, StateGameOver, StateVictory:
			return ebiten.Termination
		}
	}

	// Меню - переход к выбору персонажа
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.charSelectIndex = 0
		g.charType = entity.CharHacker
		g.state = StateCharacterSelect
	}

	// Выбор персонажа
	if g.state == StateCharacterSelect {
		g.handleCharacterSelect()
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

// handleCharacterSelect обрабатывает выбор персонажа
func (g *Game) handleCharacterSelect() {
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.charSelectIndex = 0
		g.charType = entity.CharHacker
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.charSelectIndex = 1
		g.charType = entity.CharRobot
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.charSelectIndex = 2
		g.charType = entity.CharNinja
	}

	if ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}
}

// updateGame обновляет игровую логику
func (g *Game) updateGame() {
	dt := 1.0 / 60.0

	g.handlePlayerInput()
	g.player.Update(dt)
	g.applyPhysics(dt)
	g.updateCamera()
	g.collectItems()
	g.updateProjectiles(dt)
	g.updateEnemies(dt)
	g.updateParticles(dt)

	// Проверка победы на уровне
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
	onLadder := g.checkLadderCollision()

	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if onLadder {
			g.player.Physics.VelocityY = 0
		} else {
			g.player.MoveLeft()
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		if onLadder {
			g.player.Physics.VelocityY = 0
		} else {
			g.player.MoveRight()
		}
	}

	// Прыжок (с поддержкой двойного)
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)) && !g.wasKeyPressed(ebiten.KeyW, ebiten.KeyUp, ebiten.KeySpace) {
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+20, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{100, 200, 255, 255})
	}

	// Рывок (dash)
	if ebiten.IsKeyPressed(ebiten.KeyShift) && !g.dashPressed {
		g.dashPressed = true
		if g.player.Dash() {
			g.spawnParticles(g.player.Transform.X+20, g.player.Transform.Y+g.player.Transform.Height/2, g.player.Physics.VelocityX*0.3, 0, 15, color.RGBA{0, 255, 255, 255})
		}
	} else if !ebiten.IsKeyPressed(ebiten.KeyShift) {
		g.dashPressed = false
	}

	// Лазание по лестнице
	if onLadder {
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.player.Physics.VelocityY = -150
		} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.player.Physics.VelocityY = 150
		}
	}

	// Приседание
	if (ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown)) && !onLadder {
		g.player.Crouch()
	} else {
		g.player.Stand()
	}

	// Выстрел
	if ebiten.IsKeyPressed(ebiten.KeyJ) && !g.shootPressed && g.player.CanShoot() {
		g.shootPressed = true
		g.shoot()
	} else if !ebiten.IsKeyPressed(ebiten.KeyJ) {
		g.shootPressed = false
	}

	// Специальная способность
	if ebiten.IsKeyPressed(ebiten.KeyK) && !g.wasKeyPressed(ebiten.KeyK) {
		g.player.ActivateSpecial()
	}
}

// wasKeyPressed проверяет, была ли клавиша нажата в предыдущем кадре
func (g *Game) wasKeyPressed(keys ...ebiten.Key) bool {
	for _, key := range keys {
		if ebiten.IsKeyPressed(key) {
			return true
		}
	}
	return false
}

// shoot производит выстрел
func (g *Game) shoot() {
	dirX := float64(g.player.Transform.Facing)
	dirY := 0.0

	projectile := entity.NewProjectile(
		g.player.Transform.X+g.player.Transform.Width/2,
		g.player.Transform.Y+g.player.Transform.Height/3,
		dirX*600,
		dirY,
		false,
		15,
		g.spriteSheet,
	)

	g.projectiles = append(g.projectiles, projectile)
}

// applyPhysics применяет физику
func (g *Game) applyPhysics(dt float64) {
	onLadder := g.checkLadderCollision()
	oldX := g.player.Transform.X
	oldY := g.player.Transform.Y

	if !onLadder {
		g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	}

	if !g.player.Physics.IsMoving {
		g.player.Physics.VelocityX *= g.player.Physics.Friction
	}

	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.player.Transform.Y += g.player.Physics.VelocityY * dt

	g.player.Physics.OnGround = false

	for _, p := range g.levelData.Platforms {
		if !p.Solid {
			continue
		}

		if onLadder && g.player.Physics.VelocityY < 0 && p.Type == level.TileLadder {
			continue
		}

		if g.checkCollision(g.player.Transform, p) {
			if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Y+20 {
				g.player.Transform.Y = p.Y - g.player.Transform.Height
				g.player.Physics.VelocityY = 0
				g.player.Physics.OnGround = true
				g.player.ResetJump()
				if g.player.State != entity.PlayerDashing {
					g.player.State = entity.PlayerIdle
				}
			} else if g.player.Physics.VelocityY < 0 && oldY >= p.Y+p.Height-10 {
				if !onLadder {
					g.player.Transform.Y = p.Y + p.Height
					g.player.Physics.VelocityY = 0
				}
			} else if g.player.Physics.VelocityX > 0 && oldX+g.player.Transform.Width <= p.X+10 {
				g.player.Transform.X = p.X - g.player.Transform.Width
				g.player.Physics.VelocityX = 0
			} else if g.player.Physics.VelocityX < 0 && oldX >= p.X+p.Width-10 {
				g.player.Transform.X = p.X + p.Width
				g.player.Physics.VelocityX = 0
			}
		}
	}

	groundLevel := float64(g.levelData.Height-2)*float64(g.levelData.TileSize)
	if g.player.Transform.Y > groundLevel+500 {
		g.player.Health.TakeDamage(100)
	}

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

	g.cameraX += (targetX - g.cameraX) * 0.1
	g.cameraY += (targetY - g.cameraY) * 0.1

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
		g.combo++
		if g.combo > g.maxCombo {
			g.maxCombo = g.combo
		}
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 10, color.RGBA{255, 215, 0, 255})
	case "gemRed", "gemBlue", "gemGreen", "gemYellow":
		g.score += item.Value
		g.combo++
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 15, color.RGBA{100, 200, 255, 255})
	case "star":
		g.score += item.Value
		g.combo++
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

			if p.IsEnemy && p.Active {
				if entity.CheckCollision(p.Transform, g.player.Transform) {
					g.player.Health.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -100, 10, color.RGBA{255, 50, 50, 255})
				}
			}

			if p.Active {
				for _, platform := range g.levelData.Platforms {
					if platform.Solid && entity.CheckCollision(p.Transform, &platform.Transform) {
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

		if entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 && !g.player.Dashing {
				g.player.Health.TakeDamage(enemy.Damage)
				g.combo = 0 // Сброс комбо при получении урона
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
	g.score += 25 * (1 + g.combo/5) // Бонус за комбо
	g.spawnParticles(enemy.Transform.X+20, enemy.Transform.Y+20, 0, -100, 15, color.RGBA{100, 255, 100, 255})
	g.enemies = append(g.enemies[:index], g.enemies[index+1:]...)
}

// updateParticles обновляет частицы
func (g *Game) updateParticles(dt float64) {
	active := make([]render.Particle, 0)

	for i := range g.particles {
		p := &g.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt
		p.Life -= dt * 0.5
		p.VX *= 0.98

		if p.Life > 0 {
			active = append(active, *p)
		}
	}

	g.particles = active
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, render.Particle{
			X:      x,
			Y:      y,
			VX:     vx + (g.rng.Float64()-0.5)*100,
			VY:     vy + (g.rng.Float64()-0.5)*100,
			Life:   1.0,
			MaxLife: 1.0,
			Color:  c,
			Size:   3 + g.rng.Float64()*4,
		})
	}
}

// checkCollision проверяет коллизию
func (g *Game) checkCollision(transform *entity.Transform, platform *level.Platform) bool {
	return transform.X < platform.X+platform.Width &&
		transform.X+transform.Width > platform.X &&
		transform.Y < platform.Y+platform.Height &&
		transform.Y+transform.Height > platform.Y
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.DrawBackground(screen, g.cameraX, g.cameraY)

	if g.levelData != nil {
		for _, p := range g.levelData.Platforms {
			g.renderer.DrawPlatform(screen, p, g.cameraX, g.cameraY)
		}

		for _, item := range g.levelData.Items {
			if !item.Collected {
				entityItem := entity.NewItem(item.X, item.Y, item.Type, item.Value, g.spriteSheet)
				g.renderer.DrawItem(screen, entityItem, g.cameraX, g.cameraY)
			}
		}

		for _, enemy := range g.enemies {
			g.renderer.DrawEnemy(screen, enemy, g.cameraX, g.cameraY)
		}

		for _, p := range g.projectiles {
			g.renderer.DrawProjectile(screen, p, g.cameraX, g.cameraY)
		}

		if g.player != nil {
			g.renderer.DrawPlayer(screen, g.player, g.cameraX, g.cameraY)
		}

		g.renderer.DrawExit(screen, g.levelData.ExitX, g.levelData.ExitY, g.cameraX, g.cameraY)
	}

	g.renderer.DrawParticles(screen, g.particles, g.cameraX, g.cameraY)

	switch g.state {
	case StateMenu:
		g.renderer.DrawMenu(screen)
	case StateCharacterSelect:
		g.drawCharacterSelect(screen)
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
	config := entity.GetCharacterConfig(g.player.CharacterType)
	
	g.renderer.DrawHUD(
		screen,
		g.player.Health.Current,
		g.player.Health.Max,
		g.player.Energy,
		g.player.MaxEnergy,
		g.score,
		g.levelNum,
		g.levelData.Name,
		g.combo,
		config.Special,
	)

	ebitenutil.DebugPrintAt(screen, "[ESC] Пауза  [J] Огонь  [Shift] Рывок  [K] Способность", 10, screenHeight-30)
}

// drawCharacterSelect отрисовывает меню выбора персонажа
func (g *Game) drawCharacterSelect(screen *ebiten.Image) {
	screenW := screen.Bounds().Dx()
	screenH := screen.Bounds().Dy()

	overlay := ebiten.NewImage(screenW, screenH)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	screen.DrawImage(overlay, nil)

	title := `
╔═══════════════════════════════════════════════╗
║      🌃 ВЫБЕРИ БОЙЦА 🌃                       ║
╠═══════════════════════════════════════════════╣
║                                               ║
║   A/D/S или ←/→/↓ для выбора                 ║
║   [SPACE] или [ENTER] - Подтвердить          ║
║   [ESC] - Назад в меню                       ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrintAt(screen, title, screenW/2-240, 50)

	hackerConfig := entity.GetCharacterConfig(entity.CharHacker)
	robotConfig := entity.GetCharacterConfig(entity.CharRobot)
	ninjaConfig := entity.GetCharacterConfig(entity.CharNinja)

	// Хакер
	hackerBox := "┌"
	hackerBoxEnd := "┘"
	if g.charSelectIndex == 0 {
		hackerBox = "╔"
		hackerBoxEnd = "╚"
	}

	hackerText := fmt.Sprintf(`
%s═══════════════════════════════════════════╗
║  👨‍💻 %s  ║
║                                           ║
║  Здоровье: %d  |  Скорость: %d  |  Прыжок: %d ║
║  Энергия рывка: %d                        ║
║                                           ║
║  Сбалансированный боец с умением взлома  ║
║                                           ║
%s═══════════════════════════════════════════╝
`, hackerBox, hackerConfig.Name, hackerConfig.MaxHealth, int(hackerConfig.Speed), int(hackerConfig.JumpForce), int(hackerConfig.DashCost), hackerBoxEnd)

	ebitenutil.DebugPrintAt(screen, hackerText, screenW/2-500, 150)

	// Робот
	robotBox := "┌"
	robotBoxEnd := "┘"
	if g.charSelectIndex == 1 {
		robotBox = "╔"
		robotBoxEnd = "╚"
	}

	robotText := fmt.Sprintf(`
%s═══════════════════════════════════════════╗
║  🤖 %s  ║
║                                           ║
║  Здоровье: %d  |  Скорость: %d  |  Прыжок: %d ║
║  Энергия рывка: %d                        ║
║                                           ║
║  Танк с щитом, медленный но прочный      ║
║                                           ║
%s═══════════════════════════════════════════╝
`, robotBox, robotConfig.Name, robotConfig.MaxHealth, int(robotConfig.Speed), int(robotConfig.JumpForce), int(robotConfig.DashCost), robotBoxEnd)

	ebitenutil.DebugPrintAt(screen, robotText, screenW/2-160, 150)

	// Ниндзя
	ninjaBox := "┌"
	ninjaBoxEnd := "┘"
	if g.charSelectIndex == 2 {
		ninjaBox = "╔"
		ninjaBoxEnd = "╚"
	}

	ninjaText := fmt.Sprintf(`
%s═══════════════════════════════════════════╗
║  🥷 %s  ║
║                                           ║
║  Здоровье: %d  |  Скорость: %d  |  Прыжок: %d ║
║  Энергия рывка: %d                        ║
║                                           ║
║  Быстрый и ловкий с невидимостью         ║
║                                           ║
%s═══════════════════════════════════════════╝
`, ninjaBox, ninjaConfig.Name, ninjaConfig.MaxHealth, int(ninjaConfig.Speed), int(ninjaConfig.JumpForce), int(ninjaConfig.DashCost), ninjaBoxEnd)

	ebitenutil.DebugPrintAt(screen, ninjaText, screenW/2+180, 150)

	selectText := "[A/←] Хакер    [D/→] Робот    [S/↓] Ниндзя"
	if g.charSelectIndex == 0 {
		selectText = "[A/←] 👻 ХАКЕР    [D/→] Робот    [S/↓] Ниндзя"
	} else if g.charSelectIndex == 1 {
		selectText = "[A/←] Хакер    [D/→] 🤖 РОБОТ    [S/↓] Ниндзя"
	} else {
		selectText = "[A/←] Хакер    [D/→] Робот    [S/↓] 🥷 НИНДЗЯ"
	}
	ebitenutil.DebugPrintAt(screen, selectText, screenW/2-180, screenH-80)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
