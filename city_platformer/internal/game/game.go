// Package game - основная игровая логика City Platformer
// Go365 Day 93 - Neon Runner: Cyber Escape
package game

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"city_platformer/internal/entity"
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
)

// Game - основная игровая структура
type Game struct {
	state      GameState
	player     *entity.Player
	platforms  []*entity.Platform
	enemies    []*entity.Enemy
	items      []*entity.Item
	projectiles []*entity.Projectile
	particles  []Particle

	cameraX    float64
	cameraY    float64
	score      int
	level      int
	dataTotal  int
	dataCollected int

	rng         *rand.Rand
	spriteSheet *sprite.SpriteSheet

	jumpPressed  bool
	shootPressed bool
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
		println("Warning: sprite loading error:", err.Error())
	}

	return g
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.level = 1
	g.score = 0
	g.startLevel()
}

// startLevel запускает уровень
func (g *Game) startLevel() {
	g.generateLevel(g.level)
}

// generateLevel генерирует уровень
func (g *Game) generateLevel(levelNum int) {
	levelWidth := 2000 + levelNum*300
	levelHeight := 600
	groundY := float64(levelHeight) - 32

	// Игрок
	g.player = entity.NewPlayer(100, groundY-64, g.spriteSheet)
	g.player.Physics.OnGround = true

	// Платформы
	g.platforms = make([]*entity.Platform, 0)

	// Пол (земля)
	g.platforms = append(g.platforms, entity.NewPlatform(
		0, groundY, float64(levelWidth), 32, "grass", g.spriteSheet,
	))

	// Платформы
	numPlatforms := 5 + levelNum*2
	for i := 0; i < numPlatforms; i++ {
		px := 250.0 + float64(i)*200 + g.rng.Float64()*80
		py := groundY - 80.0 - g.rng.Float64()*100
		width := 80.0 + g.rng.Float64()*60

		tileType := "stone"
		if levelNum > 3 {
			tileType = "brickWall"
		}
		if levelNum > 6 {
			tileType = "castle"
		}

		g.platforms = append(g.platforms, entity.NewPlatform(px, py, width, 32, tileType, g.spriteSheet))
	}

	// Враги
	g.enemies = make([]*entity.Enemy, 0)
	enemyTypes := []string{"slime", "snake", "spider", "fly", "bat", "ghost"}
	numEnemies := 3 + levelNum
	for i := 0; i < numEnemies; i++ {
		x := 400.0 + float64(i)*250 + g.rng.Float64()*100
		y := groundY - 40
		enemyType := enemyTypes[g.rng.Intn(len(enemyTypes))]
		enemy := entity.NewEnemy(x, y, enemyType, g.spriteSheet)
		enemy.PatrolStart = x - 80
		enemy.PatrolEnd = x + 80
		g.enemies = append(g.enemies, enemy)
	}

	// Предметы
	g.items = make([]*entity.Item, 0)
	itemTypes := []string{
		entity.ItemCoinGold, entity.ItemCoinSilver, entity.ItemCoinBronze,
		entity.ItemGemRed, entity.ItemGemBlue, entity.ItemGemGreen,
		entity.ItemStar, entity.ItemMushroom,
	}
	numItems := 8 + levelNum*2
	for i := 0; i < numItems; i++ {
		x := 150.0 + float64(i)*180 + g.rng.Float64()*80
		y := groundY - 50 - g.rng.Float64()*120
		itemType := itemTypes[g.rng.Intn(len(itemTypes))]
		value := 10
		if itemType == entity.ItemStar {
			value = 50
		}
		item := entity.NewItem(x, y, itemType, value, g.spriteSheet)
		g.items = append(g.items, item)
	}

	// Данные (цель уровня)
	g.dataTotal = 3
	g.dataCollected = 0
	for i := 0; i < g.dataTotal; i++ {
		x := 600.0 + float64(i)*400 + g.rng.Float64()*100
		y := groundY - 100 - g.rng.Float64()*80
		item := entity.NewItem(x, y, entity.ItemData, 100, g.spriteSheet)
		g.items = append(g.items, item)
	}

	// Частицы и снаряды
	g.particles = make([]Particle, 0)
	g.projectiles = make([]*entity.Projectile, 0)

	// Камера
	g.cameraX = 0
	g.cameraY = 0
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
	if g.state == StateVictory && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.level++
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

	// Смерть игрока - респавн
	if g.player.Health.Dead {
		g.player.Health.Dead = false
		g.player.Health.Current = g.player.Health.Max
		g.startLevel()
		g.score -= 50
		if g.score < 0 {
			g.score = 0
		}
	}
}

// handleInput обрабатывает ввод
func (g *Game) handleInput() {
	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		g.player.MoveLeft()
	} else if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		g.player.MoveRight()
	} else {
		g.player.Stop()
	}

	// Прыжок
	jumpKey := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)
	if jumpKey && !g.jumpPressed {
		g.jumpPressed = true
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{255, 255, 100, 255})
	} else if !jumpKey {
		g.jumpPressed = false
	}

	// Выстрел
	shootKey := ebiten.IsKeyPressed(ebiten.KeyJ)
	if shootKey && !g.shootPressed && g.player.CanShoot() {
		g.shootPressed = true
		g.shoot()
	} else if !shootKey {
		g.shootPressed = false
	}
}

// shoot стреляет лучом
func (g *Game) shoot() {
	g.player.Shoot()

	dirX := float64(g.player.Transform.Facing)
	proj := entity.NewProjectile(
		g.player.Transform.X+g.player.Transform.Width/2,
		g.player.Transform.Y+g.player.Transform.Height/3,
		dirX*500,
		0,
		15,
		g.spriteSheet,
	)
	g.projectiles = append(g.projectiles, proj)
}

// applyPhysics применяет физику
func (g *Game) applyPhysics(dt float64) {
	g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	if g.player.Physics.VelocityY > 800 {
		g.player.Physics.VelocityY = 800
	}

	oldY := g.player.Transform.Y

	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.player.Transform.Y += g.player.Physics.VelocityY * dt

	g.player.Physics.OnGround = false

	for _, p := range g.platforms {
		platRect := p.Transform
		if entity.CheckCollision(g.player.Transform, platRect) {
			if g.player.Physics.VelocityY > 0 && oldY+g.player.Transform.Height <= p.Transform.Y+10 {
				g.player.Transform.Y = p.Transform.Y - g.player.Transform.Height
				g.player.Physics.VelocityY = 0
				g.player.Physics.OnGround = true
				g.player.ResetJump()
			}
		}
	}

	// Границы
	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
	}

	// Падение
	if g.player.Transform.Y > 800 {
		g.player.Health.TakeDamage(100)
	}
}

// updateCamera обновляет камеру
func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1

	if g.cameraX < 0 {
		g.cameraX = 0
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
			g.score += item.Value

			if item.ItemType == entity.ItemData {
				g.dataCollected++
			}

			switch item.ItemType {
			case entity.ItemMushroom:
				g.player.Health.Heal(20)
			case entity.ItemStar:
				g.player.Energy.Current = g.player.Energy.Max
			}

			g.spawnParticles(item.Transform.X+16, item.Transform.Y+16, 0, -50, 10, color.RGBA{255, 215, 0, 255})
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
				g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -50, 8, color.RGBA{255, 255, 100, 255})
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
	for _, enemy := range g.enemies {
		enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y)

		// Коллизия с игроком
		if !enemy.Health.Dead && entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 {
				g.player.Health.TakeDamage(enemy.Damage)
				g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
			}
		}
	}
}

// updateParticles обновляет частицы
func (g *Game) updateParticles(dt float64) {
	active := make([]Particle, 0)

	for i := range g.particles {
		p := &g.particles[i]
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 200 * dt
		p.Life -= dt * 0.5

		if p.Life > 0 {
			active = append(active, *p)
		}
	}

	g.particles = active
}

// spawnParticles создаёт частицы
func (g *Game) spawnParticles(x, y, vx, vy float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life: 1.0,
			Color: c,
			Size: 3 + g.rng.Float64()*4,
		})
	}
}

// checkLevelExit проверяет выход
func (g *Game) checkLevelExit() {
	// Все данные собраны и игрок справа
	if g.dataCollected >= g.dataTotal && g.player.Transform.X > 1500 {
		g.state = StateVictory
	}
}

// Particle - частица
type Particle struct {
	X, Y, VX, VY float64
	Life         float64
	Color        color.Color
	Size         float64
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
	for _, p := range g.platforms {
		p.Draw(screen, g.cameraX, g.cameraY)
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
		vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y-g.cameraY), float32(p.Size), float32(p.Size), p.Color, false)
	}

	// HUD
	if g.state == StatePlaying || g.state == StatePaused {
		g.drawHUD(screen)
	}

	// Меню/пауза/конец
	if g.state == StatePaused {
		g.drawPause(screen)
	}
	if g.state == StateGameOver {
		g.drawGameOver(screen)
	}
	if g.state == StateVictory {
		g.drawVictory(screen)
	}
}

// drawBackground отрисовывает фон
func (g *Game) drawBackground(screen *ebiten.Image) {
	// Градиент неба (тёмно-синий к фиолетовому)
	for y := 0; y < screenHeight; y++ {
		percent := float64(y) / float64(screenHeight)
		r := uint8(26 - percent*20)
		g_ := uint8(30 - percent*30)
		b := uint8(62 + percent*50)
		vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g_, b, 255}, false)
	}

	// Звёзды
	g.rng.Seed(time.Now().UnixNano())
	for i := 0; i < 50; i++ {
		starX := float32((i*37 - int(g.cameraX*0.2)) % screenWidth)
		starY := float32((i*53) % (screenHeight / 2))
		if starX < 0 {
			starX += screenWidth
		}
		vector.DrawFilledRect(screen, starX, starY, 2, 2, color.RGBA{255, 255, 255, 200}, false)
	}

	// Город на заднем плане (параллакс)
	g.drawCitySilhouette(screen)
}

// drawCitySilhouette рисует силуэт города
func (g *Game) drawCitySilhouette(screen *ebiten.Image) {
	buildingColor := color.RGBA{20, 25, 50, 255}
	windowColor := color.RGBA{0, 240, 255, 150}

	g.rng.Seed(int64(42))
	for i := 0; i < 20; i++ {
		buildingX := float32((i*80 - int(g.cameraX*0.3)) % (screenWidth + 100))
		if buildingX < -50 {
			buildingX += screenWidth + 100
		}
		buildingWidth := float32(40 + (i*17)%60)
		buildingHeight := float32(100 + (i*23)%200)
		buildingY := float32(screenHeight) - buildingHeight - 20

		vector.DrawFilledRect(screen, buildingX, buildingY, buildingWidth, buildingHeight, buildingColor, false)

		// Окна
		for wy := buildingY + 10; wy < buildingY+buildingHeight-10; wy += 20 {
			for wx := buildingX + 5; wx < buildingX+buildingWidth-10; wx += 15 {
				if g.rng.Float32() > 0.5 {
					vector.DrawFilledRect(screen, wx, wy, 8, 12, windowColor, false)
				}
			}
		}
	}
}

// drawMenu отрисовывает меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "NEON RUNNER: CYBER ESCAPE"
	ebitenutil.DebugPrintAt(screen, title, screenWidth/2-140, 150)

	subtitle := "2087. Сбеги из корпоративного ада."
	ebitenutil.DebugPrintAt(screen, subtitle, screenWidth/2-120, 200)

	instructions := []string{
		"",
		"[SPACE] Начать игру",
		"",
		"Управление:",
		"A/D - Бег",
		"W/Пробел - Прыжок (двойной!)",
		"J - Выстрел",
		"",
		"Цель: Собрать все данные (синие) и добраться до выхода!",
		"Избегай врагов или уничтожай их!",
	}

	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-150, 260+i*22)
	}
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	y := 10

	// Здоровье
	vector.DrawFilledRect(screen, 10, float32(y), 200, 20, color.RGBA{50, 50, 50, 255}, false)
	hpPercent := float32(g.player.Health.Current) / float32(g.player.Health.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*hpPercent, 20, color.RGBA{255, 100, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "HP", 220, y)

	y += 25

	// Энергия
	vector.DrawFilledRect(screen, 10, float32(y), 200, 15, color.RGBA{50, 50, 50, 255}, false)
	energyPercent := float32(g.player.Energy.Current) / float32(g.player.Energy.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*energyPercent, 15, color.RGBA{0, 240, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "NRG", 220, y)

	y += 25

	// Счёт
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), 10, y)
	y += 20

	// Данные
	ebitenutil.DebugPrintAt(screen, "DATA: "+string(rune(g.dataCollected))+"/"+string(rune(g.dataTotal)), 10, y)
	y += 20

	// Уровень
	ebitenutil.DebugPrintAt(screen, "LEVEL: "+string(rune(g.level)), screenWidth-100, 10)

	// Подсказки
	ebitenutil.DebugPrintAt(screen, "[ESC] Пауза  [J] Огонь  [W] Прыжок", 10, screenHeight-30)
}

// drawPause отрисовывает паузу
func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", screenWidth/2-50, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "[ESC] Продолжить", screenWidth/2-80, screenHeight/2)
}

// drawGameOver отрисовывает Game Over
func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{150, 50, 50, 200}, false)
	ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-80, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Рестарт", screenWidth/2-80, screenHeight/2+50)
}

// drawVictory отрисовывает победу
func (g *Game) drawVictory(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 150, 50, 200}, false)
	ebitenutil.DebugPrintAt(screen, "УРОВЕНЬ ПРОЙДЕН!", screenWidth/2-100, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Следующий уровень", screenWidth/2-120, screenHeight/2+50)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
