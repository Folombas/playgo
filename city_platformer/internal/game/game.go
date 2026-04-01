// Package game - основная игровая логика Cyber City Runner
// Go365 Day 92 - Киберпанк-платформер
package game

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"cyber_city/internal/entity"
	"cyber_city/internal/level"
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
	state     GameState
	player    *entity.Player
	level     *level.LevelData
	enemies   []*entity.Enemy
	items     []*entity.Item
	projectiles []*entity.Projectile
	particles []entity.Particle

	cameraX float64
	cameraY float64

	score       int
	levelNum    int
	dataCollected int
	alertLevel  float64 // Общая тревога уровня

	rng *rand.Rand

	// Input
	jumpPressed  bool
	dashPressed  bool
	shootPressed bool
	hackPressed  bool
}

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		state: StateMenu,
		rng:   rng,
	}

	return g
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.levelNum = 1
	g.score = 0
	g.dataCollected = 0
	g.startLevel()
}

// startLevel запускает уровень
func (g *Game) startLevel() {
	gen := level.NewLevelGenerator(g.rng)
	g.level = gen.GenerateLevel(g.levelNum)

	// Игрок
	g.player = entity.NewPlayer(100, 400)
	if g.levelNum > 1 {
		// Сохраняем прогресс между уровнями
		g.player.DataChunks = g.dataCollected
	}

	// Враги
	g.enemies = make([]*entity.Enemy, 0)
	for _, spawn := range g.level.Enemies {
		enemy := entity.NewEnemy(spawn.X, spawn.Y, spawn.Type)
		enemy.PatrolStart = spawn.PatrolMin
		enemy.PatrolEnd = spawn.PatrolMax
		g.enemies = append(g.enemies, enemy)
	}

	// Предметы
	g.items = make([]*entity.Item, 0)
	for _, spawn := range g.level.Items {
		item := entity.NewItem(spawn.X, spawn.Y, spawn.ItemType, 10)
		g.items = append(g.items, item)
	}

	// Снаряды и частицы
	g.projectiles = make([]*entity.Projectile, 0)
	g.particles = make([]entity.Particle, 0)

	// Камера
	g.cameraX = 0
	g.cameraY = 0
	g.alertLevel = 0
}

// Update обновляет игру
func (g *Game) Update() error {
	// ESC - пауза/меню
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

	// Меню - старт
	if g.state == StateMenu && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	// Game Over - рестарт
	if g.state == StateGameOver && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.Reset()
		g.state = StatePlaying
	}

	// Victory - следующий уровень
	if g.state == StateLevelComplete && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.levelNum++
		g.startLevel()
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
	dt := 1.0 / 60.0

	g.handleInput()
	g.player.Update(dt)
	g.applyPhysics(dt)
	g.updateCamera()
	g.collectItems()
	g.updateProjectiles(dt)
	g.updateEnemies(dt)
	g.updateParticles(dt)
	g.checkLevelExit()
	g.updateAlertLevel(dt)

	// Смерть игрока - респавн
	if g.player.Health.Dead {
		g.state = StateGameOver
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

	// Прыжок
	jumpKey := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)
	if jumpKey && !g.jumpPressed {
		g.jumpPressed = true
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{0, 255, 255, 255})
	} else if !jumpKey {
		g.jumpPressed = false
	}

	// Рывок
	dashKey := ebiten.IsKeyPressed(ebiten.KeyK)
	if dashKey && !g.dashPressed {
		g.dashPressed = true
		g.player.Dash()
		g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height/2, float64(g.player.Transform.Facing)*200, 0, 10, color.RGBA{0, 255, 255, 255})
	} else if !dashKey {
		g.dashPressed = false
	}

	// Выстрел
	shootKey := ebiten.IsKeyPressed(ebiten.KeyJ)
	if shootKey && !g.shootPressed && g.player.CanShoot() {
		g.shootPressed = true
		g.shoot()
	} else if !shootKey {
		g.shootPressed = false
	}

	// Взлом
	hackKey := ebiten.IsKeyPressed(ebiten.KeyE)
	if hackKey && !g.hackPressed {
		g.hackPressed = true
		g.tryHack()
	} else if !hackKey {
		g.hackPressed = false
	}

	// EMP
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		g.useEMP()
	}

	// Граната
	if ebiten.IsKeyPressed(ebiten.KeyL) {
		g.throwGrenade()
	}
}

// shoot стреляет
func (g *Game) shoot() {
	g.player.Shoot()

	dirX := float64(g.player.Transform.Facing)
	proj := entity.NewProjectile(
		g.player.Transform.X+g.player.Transform.Width/2,
		g.player.Transform.Y+g.player.Transform.Height/3,
		dirX*600,
		0,
		20,
		true,
	)
	g.projectiles = append(g.projectiles, proj)
}

// tryHack пытается взломать
func (g *Game) tryHack() {
	// Проверяем, есть ли терминал рядом
	for _, term := range g.level.Terminals {
		dist := math.Abs(g.player.Transform.X+16 - term.X)
		if dist < 60 && math.Abs(g.player.Transform.Y - term.Y) < 60 {
			g.player.Hack()
			break
		}
	}
}

// useEMP использует EMP
func (g *Game) useEMP() {
	if g.player.UseEMP() {
		// Оглушаем всех врагов в радиусе
		for _, enemy := range g.enemies {
			dist := math.Sqrt(math.Pow(enemy.Transform.X-g.player.Transform.X, 2) + math.Pow(enemy.Transform.Y-g.player.Transform.Y, 2))
			if dist < 300 {
				enemy.Health.Invincible = 3.0 // Оглушение на 3 секунды
				g.spawnParticles(enemy.Transform.X+16, enemy.Transform.Y+24, 0, -100, 15, color.RGBA{150, 50, 255, 255})
			}
		}
		g.alertLevel -= 0.3
		if g.alertLevel < 0 {
			g.alertLevel = 0
		}
	}
}

// throwGrenade бросает гранату
func (g *Game) throwGrenade() {
	g.player.ThrowGrenade()
	// Упрощённо - мгновенный урон в области
	targetX := g.player.Transform.X + float64(g.player.Transform.Facing)*200
	targetY := g.player.Transform.Y - 100

	for _, enemy := range g.enemies {
		dist := math.Sqrt(math.Pow(enemy.Transform.X-targetX, 2) + math.Pow(enemy.Transform.Y-targetY, 2))
		if dist < 100 {
			enemy.Health.TakeDamage(50)
		}
	}

	g.spawnParticles(targetX, targetY, 0, -50, 20, color.RGBA{255, 165, 0, 255})
}

// applyPhysics применяет физику
func (g *Game) applyPhysics(dt float64) {
	g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	if g.player.Physics.VelocityY > 1000 {
		g.player.Physics.VelocityY = 1000
	}

	oldY := g.player.Transform.Y
	oldX := g.player.Transform.X

	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.player.Transform.Y += g.player.Physics.VelocityY * dt

	g.player.Physics.OnGround = false
	g.player.Physics.OnWall = false

	// Коллизии с платформами
	for _, p := range g.level.Platforms {
		if !p.Solid {
			continue
		}

		platRect := p.Transform

		if entity.CheckCollision(g.player.Transform, platRect) {
			// Вертикальная коллизия
			if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Y+10 {
				g.player.Transform.Y = p.Y - g.player.Transform.Height
				g.player.Physics.VelocityY = 0
				g.player.Physics.OnGround = true
				g.player.ResetJump()
			} else if g.player.Physics.VelocityY < 0 && oldY >= p.Y+p.Height-10 {
				g.player.Transform.Y = p.Y + p.Height
				g.player.Physics.VelocityY = 0
			} else {
				// Горизонтальная коллизия (стены)
				if g.player.Physics.VelocityX > 0 && oldX+g.player.Transform.Width <= p.X+10 {
					g.player.Transform.X = p.X - g.player.Transform.Width
					g.player.Physics.VelocityX = 0
					g.player.Physics.OnWall = true
					g.player.Transform.Facing = -1
				} else if g.player.Physics.VelocityX < 0 && oldX >= p.X+p.Width-10 {
					g.player.Transform.X = p.X + p.Width
					g.player.Physics.VelocityX = 0
					g.player.Physics.OnWall = true
					g.player.Transform.Facing = 1
				}
			}
		}
	}

	// Границы
	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
	}

	// Падение в яму
	if g.player.Transform.Y > g.level.Height+100 {
		g.player.Health.TakeDamage(100)
	}

	// Проверка опасностей
	g.checkHazards()
}

// checkHazards проверяет опасности
func (g *Game) checkHazards() {
	for _, hazard := range g.level.Hazards {
		if !hazard.Active {
			continue
		}

		hazardRect := entity.NewTransform(hazard.X, hazard.Y, hazard.Width, hazard.Height)
		if entity.CheckCollision(g.player.Transform, hazardRect) {
			g.player.Health.TakeDamage(hazard.Damage)
			g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 0, 0, 255})
		}
	}
}

// updateCamera обновляет камеру
func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1

	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > g.level.Width-float64(screenWidth) {
		g.cameraX = g.level.Width - float64(screenWidth)
	}

	targetY := g.player.Transform.Y - screenHeight/2
	g.cameraY += (targetY - g.cameraY) * 0.1

	if g.cameraY < 0 {
		g.cameraY = 0
	}
}

// collectItems собирает предметы
func (g *Game) collectItems() {
	for _, item := range g.items {
		if item.Collected {
			continue
		}
		if entity.CheckCollision(g.player.Transform, item.Transform) {
			item.Collected = true
			g.score += 10

			switch item.ItemType {
			case entity.ItemStimpack:
				g.player.Health.Heal(25)
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{255, 50, 50, 255})
			case entity.ItemEnergy:
				g.player.Energy.Current += 50
				if g.player.Energy.Current > g.player.Energy.Max {
					g.player.Energy.Current = g.player.Energy.Max
				}
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{0, 255, 255, 255})
			case entity.ItemArmor:
				g.player.Armor += 25
				if g.player.Armor > g.player.MaxArmor {
					g.player.Armor = g.player.MaxArmor
				}
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{100, 100, 100, 255})
			case entity.ItemData:
				g.dataCollected++
				g.score += 100
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{0, 255, 0, 255})
			case entity.ItemKeycard:
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{255, 255, 0, 255})
			case entity.ItemGrenade:
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{255, 165, 0, 255})
			case entity.ItemEMP:
				g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, color.RGBA{150, 50, 255, 255})
			}
		}
	}
}

// updateProjectiles обновляет снаряды
func (g *Game) updateProjectiles(dt float64) {
	active := make([]*entity.Projectile, 0)

	for _, p := range g.projectiles {
		if !p.Active {
			continue
		}

		p.Update(dt)

		// Коллизия с врагами
		for _, enemy := range g.enemies {
			if !enemy.Health.Dead && entity.CheckCollision(p.Transform, enemy.Transform) {
				enemy.Health.TakeDamage(p.Damage)
				p.Active = false
				g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -50, 8, color.RGBA{0, 255, 255, 255})
				break
			}
		}

		// Коллизия со стенами
		for _, plat := range g.level.Platforms {
			if plat.Solid && entity.CheckCollision(p.Transform, plat.Transform) {
				p.Active = false
				g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -50, 5, color.RGBA{0, 255, 255, 255})
				break
			}
		}

		if p.Active {
			active = append(active, p)
		}
	}

	g.projectiles = active
}

// updateEnemies обновляет врагов
func (g *Game) updateEnemies(dt float64) {
	playerVisible := g.isPlayerVisible()

	for _, enemy := range g.enemies {
		if enemy.Health.Dead {
			continue
		}

		enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y, playerVisible)

		// Коллизия с игроком
		if entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 {
				damage := enemy.Damage
				if g.player.Armor > 0 {
					armorReduce := damage / 2
					if g.player.Armor >= armorReduce {
						g.player.Armor -= armorReduce
						damage = armorReduce
					} else {
						damage = g.player.Armor
						g.player.Armor = 0
					}
				}
				g.player.Health.TakeDamage(damage)
				g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
			}
		}
	}
}

// isPlayerVisible проверяет, виден ли игрок
func (g *Game) isPlayerVisible() bool {
	// Упрощённая проверка - расстояние до ближайшего врага
	for _, enemy := range g.enemies {
		if enemy.Health.Dead {
			continue
		}
		dist := math.Sqrt(math.Pow(enemy.Transform.X-g.player.Transform.X, 2) + math.Pow(enemy.Transform.Y-g.player.Transform.Y, 2))
		if dist < 300 {
			return true
		}
	}
	return false
}

// updateParticles обновляет частицы
func (g *Game) updateParticles(dt float64) {
	active := make([]entity.Particle, 0)

	for i := range g.particles {
		p := &g.particles[i]
		p.Update(dt)

		if p.Life > 0 {
			active = append(active, *p)
		}
	}

	g.particles = active
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, entity.Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life:    0.5 + g.rng.Float64()*0.3,
			MaxLife: 0.8,
			Color:   c,
			Size:    3 + g.rng.Float64()*4,
		})
	}
}

// updateAlertLevel обновляет уровень тревоги
func (g *Game) updateAlertLevel(dt float64) {
	// Тревога растёт от действий игрока
	if g.player.Stealth.Noise > 0.5 {
		g.alertLevel += dt * 0.2
	}

	// Тревога падает со временем
	g.alertLevel -= dt * 0.05

	if g.alertLevel < 0 {
		g.alertLevel = 0
	}
	if g.alertLevel > 1 {
		g.alertLevel = 1
	}
}

// checkLevelExit проверяет выход
func (g *Game) checkLevelExit() {
	if g.player.Transform.X > g.level.ExitX-50 && g.player.Transform.Y > g.level.ExitY-50 {
		g.state = StateLevelComplete
	}
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	g.drawBackground(screen)

	if g.state == StateMenu {
		g.drawMenu(screen)
		return
	}

	// Платформы
	for _, p := range g.level.Platforms {
		g.drawPlatform(screen, p)
	}

	// Предметы
	for _, item := range g.items {
		item.Draw(screen, g.cameraX, g.cameraY)
	}

	// Враги
	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.cameraX, g.cameraY)
	}

	// Снаряды
	for _, p := range g.projectiles {
		p.Draw(screen, g.cameraX, g.cameraY)
	}

	// Игрок
	if g.player != nil {
		g.player.Draw(screen, g.cameraX, g.cameraY)
	}

	// Частицы
	for _, p := range g.particles {
		p.Draw(screen, g.cameraX, g.cameraY)
	}

	// HUD
	if g.state == StatePlaying {
		g.drawHUD(screen)
	}

	// Меню/пауза/конец
	if g.state == StatePaused {
		g.drawPause(screen)
	}
	if g.state == StateGameOver {
		g.drawGameOver(screen)
	}
	if g.state == StateLevelComplete {
		g.drawLevelComplete(screen)
	}
}

// drawBackground отрисовывает фон
func (g *Game) drawBackground(screen *ebiten.Image) {
	// Тёмный градиент (ночной город)
	for y := 0; y < screenHeight; y++ {
		percent := float64(y) / float64(screenHeight)
		r := uint8(20 - percent*10)
		g_ := uint8(20 - percent*5)
		b := uint8(40 - percent*20)
		vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g_, b, 255}, false)
	}

	// Здания на фоне (параллакс)
	g.drawCityBackground(screen)

	// Сетка (киберпанк стиль)
	g.drawGrid(screen)
}

// drawCityBackground рисует город на фоне
func (g *Game) drawCityBackground(screen *ebiten.Image) {
	buildingColor := color.RGBA{30, 30, 50, 200}
	windowColor := color.RGBA{0, 255, 255, 150}

	for i := 0; i < 15; i++ {
		buildingX := float32(i*100 - int(g.cameraX*0.2)%100)
		buildingHeight := float32(200 + (i*37)%300)
		buildingWidth := float32(80 + (i*23)%40)

		vector.DrawFilledRect(screen, buildingX, float32(screenHeight)-buildingHeight, buildingWidth, buildingHeight, buildingColor, false)

		// Окна
		for wy := float32(20); wy < buildingHeight-20; wy += 30 {
			for wx := float32(10); wx < buildingWidth-10; wx += 20 {
				if (i+int(wy)+int(wx))%3 == 0 {
					vector.DrawFilledRect(screen, buildingX+wx, float32(screenHeight)-buildingHeight+wy, 12, 18, windowColor, false)
				}
			}
		}
	}
}

// drawGrid рисует сетку
func (g *Game) drawGrid(screen *ebiten.Image) {
	gridColor := color.RGBA{0, 255, 255, 30}
	gridSize := 50.0

	offsetX := g.cameraX * 0.5
	startX := offsetX - float64(int(offsetX)%int(gridSize))

	for x := startX; x < screenWidth; x += gridSize {
		vector.StrokeLine(screen, float32(x), 0, float32(x), float32(screenHeight), 1, gridColor, false)
	}

	for y := 0.0; y < screenHeight; y += gridSize {
		vector.StrokeLine(screen, 0, float32(y), screenWidth, float32(y), 1, gridColor, false)
	}
}

// drawPlatform отрисовывает платформу
func (g *Game) drawPlatform(screen *ebiten.Image, p *level.Platform) {
	x := p.X - g.cameraX
	y := p.Y - g.cameraY

	var platColor color.Color
	switch p.Color {
	case 0: // Ground
		platColor = color.RGBA{50, 50, 60, 255}
	case 1: // Platform
		platColor = color.RGBA{60, 60, 80, 255}
	case 2: // Wall
		platColor = color.RGBA{40, 40, 50, 255}
	case 3: // Moving
		platColor = color.RGBA{80, 60, 100, 255}
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(p.Width), float32(p.Height), platColor, false)

	// Неоновая обводка
	neonColor := color.RGBA{0, 255, 255, 200}
	if p.Color == 3 {
		neonColor = color.RGBA{255, 0, 255, 200}
	}
	vector.StrokeRect(screen, float32(x), float32(y), float32(p.Width), float32(p.Height), 2, neonColor, false)
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	y := 10

	// Здоровье
	vector.DrawFilledRect(screen, 10, float32(y), 200, 20, color.RGBA{50, 50, 50, 255}, false)
	hpPercent := float32(g.player.Health.Current) / float32(g.player.Health.Max)
	hpColor := color.RGBA{255, 50, 50, 255}
	if hpPercent > 0.5 {
		hpColor = color.RGBA{50, 255, 50, 255}
	}
	vector.DrawFilledRect(screen, 10, float32(y), 200*hpPercent, 20, hpColor, false)
	vector.StrokeRect(screen, 10, float32(y), 200, 20, 2, color.RGBA{0, 255, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "HP", 220, y)

	y += 25

	// Энергия
	vector.DrawFilledRect(screen, 10, float32(y), 200, 15, color.RGBA{50, 50, 50, 255}, false)
	energyPercent := float32(g.player.Energy.Current) / float32(g.player.Energy.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*energyPercent, 15, color.RGBA{0, 255, 255, 255}, false)
	vector.StrokeRect(screen, 10, float32(y), 200, 15, 2, color.RGBA{0, 255, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "NRG", 220, y)

	y += 25

	// Броня
	vector.DrawFilledRect(screen, 10, float32(y), 150, 12, color.RGBA{50, 50, 50, 255}, false)
	armorPercent := float32(g.player.Armor) / float32(g.player.MaxArmor)
	vector.DrawFilledRect(screen, 10, float32(y), 150*armorPercent, 12, color.RGBA{100, 100, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "ARM", 170, y)

	y += 20

	// Счёт
	ebitenutil.DebugPrintAt(screen, "SCORE: "+itoa(g.score), 10, y)
	y += 18

	// Данные
	ebitenutil.DebugPrintAt(screen, "DATA: "+itoa(g.dataCollected), 10, y)
	y += 18

	// Уровень
	ebitenutil.DebugPrintAt(screen, "LEVEL "+itoa(g.levelNum)+": "+g.level.Name, screenWidth/2-100, 10)

	// Тревога
	alertWidth := int(150.0 * g.alertLevel)
	vector.DrawFilledRect(screen, float32(screenWidth-170), 10, 150, 15, color.RGBA{50, 50, 50, 255}, false)
	vector.DrawFilledRect(screen, float32(screenWidth-170), 10, float32(alertWidth), 15, color.RGBA{255, 0, 0, 255}, false)
	ebitenutil.DebugPrintAt(screen, "ALERT", screenWidth-160, 12)

	// Подсказки
	ebitenutil.DebugPrintAt(screen, "[A/D] Move [W] Jump [K] Dash [J] Shoot [E] Hack [Q] EMP [ESC] Pause", 10, screenHeight-30)
}

// drawMenu отрисовывает меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	// Фон
	g.drawBackground(screen)

	title := "🌃 CYBER CITY RUNNER"
	ebitenutil.DebugPrintAt(screen, title, screenWidth/2-150, 150)

	subtitle := "KAI's Escape from Neo-Tokyo"
	ebitenutil.DebugPrintAt(screen, subtitle, screenWidth/2-120, 200)

	instructions := []string{
		"",
		"[SPACE] Start Game",
		"",
		"Controls:",
		"A/D - Move",
		"W/Space - Jump (Double!)",
		"K - Dash",
		"J - Shoot",
		"E - Hack",
		"Q - EMP",
		"",
		"Objective: Escape the city!",
		"Collect data chunks.",
		"Avoid or eliminate enemies.",
	}

	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-120, 260+i*22)
	}
}

// drawPause отрисовывает паузу
func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 180}, false)
	ebitenutil.DebugPrintAt(screen, "PAUSED", screenWidth/2-50, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "[ESC] Resume", screenWidth/2-70, screenHeight/2)
}

// drawGameOver отрисовывает Game Over
func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{150, 0, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-80, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+itoa(g.score), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "DATA: "+itoa(g.dataCollected), screenWidth/2-60, screenHeight/2+30)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Retry", screenWidth/2-70, screenHeight/2+80)
}

// drawLevelComplete отрисовывает завершение уровня
func (g *Game) drawLevelComplete(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 100, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "LEVEL COMPLETE!", screenWidth/2-100, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+itoa(g.score), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "DATA: "+itoa(g.dataCollected), screenWidth/2-60, screenHeight/2+30)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Next Level", screenWidth/2-90, screenHeight/2+80)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// itoa - конвертация int в string
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}

	if negative {
		digits = append(digits, '-')
	}

	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}

	return string(digits)
}
