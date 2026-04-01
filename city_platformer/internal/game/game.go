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
	"cyber_city_runner/internal/entity"
	"cyber_city_runner/internal/sprite"
)

const (
	screenWidth  = 1280
	screenHeight = 720
	tileSize     = 48
)

// GameState - состояние игры
type GameState int

const (
	StateMenu GameState = iota
	StatePlaying
	StatePaused
	StateGameOver
	StateVictory
	StateHacking
)

// Game - основная игровая структура
type Game struct {
	state       GameState
	player      *entity.Player
	platforms   []*Platform
	enemies     []*entity.Enemy
	terminals   []*entity.Terminal
	doors       []*entity.Door
	items       []*entity.Item
	cameras     []*entity.Camera
	turrets     []*entity.Turret
	particles   []Particle
	
	cameraX     float64
	cameraY     float64
	
	score       int
	level       int
	alertLevel  int
	
	rng *rand.Rand
	spriteSheet *sprite.SpriteSheet
	
	// Ввод
	jumpPressed    bool
	dashPressed    bool
	shootPressed   bool
	hackPressed    bool
	
	// Хакерская мини-игра
	hackTarget    *entity.Terminal
	hackBarY      float64
	hackBarSpeed  float64
	hackZoneStart float64
	hackZoneEnd   float64
	
	// Стрельба
	projectiles []Projectile
	
	// Уровень
	levelWidth  int
	levelHeight int
	exitX       float64
	exitY       float64
}

// Platform - платформа
type Platform struct {
	X, Y, Width, Height float64
	Type                string
	Color               color.Color
}

// Projectile - снаряд
type Projectile struct {
	X, Y, VX, VY float64
	Width, Height float64
	Damage       int
	LifeTime     float64
	Active       bool
	FromPlayer   bool
	Color        color.Color
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
		println("Warning: sprite sheet loading error:", err.Error())
	}
	
	return g
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.level = 1
	g.score = 0
	g.alertLevel = 0
	g.startLevel()
}

// startLevel запускает уровень
func (g *Game) startLevel() {
	g.generateLevel(g.level)
}

// generateLevel генерирует уровень
func (g *Game) generateLevel(levelNum int) {
	g.levelWidth = 2000 + levelNum*500
	g.levelHeight = 600
	
	// Создаём игрока в начале уровня
	g.player = entity.NewPlayer(100, float64(g.levelHeight)-150, g.spriteSheet)
	
	// Генерация платформ
	g.platforms = make([]*Platform, 0)
	g.generateTerrain()
	
	// Враги
	g.enemies = make([]*entity.Enemy, 0)
	g.generateEnemies(levelNum)
	
	// Терминалы и двери
	g.terminals = make([]*entity.Terminal, 0)
	g.doors = make([]*entity.Door, 0)
	g.generateObstacles(levelNum)
	
	// Предметы
	g.items = make([]*entity.Item, 0)
	g.generateItems(levelNum)
	
	// Камеры и турели
	g.cameras = make([]*entity.Camera, 0)
	g.turrets = make([]*entity.Turret, 0)
	g.generateSecurity(levelNum)
	
	// Выход
	g.exitX = float64(g.levelWidth) - 100
	g.exitY = float64(g.levelHeight) - 100
	
	// Камера
	g.cameraX = 0
	g.cameraY = 0
	
	// Частицы
	g.particles = make([]Particle, 0)
	g.projectiles = make([]Projectile, 0)
}

// generateTerrain генерирует местность
func (g *Game) generateTerrain() {
	// Пол
	groundY := float64(g.levelHeight) - 20
	
	// Создаём пол с ямами
	x := 0.0
	segmentLength := 200.0
	
	for x < float64(g.levelWidth) {
		// Ямы после первого сегмента
		if x > 300 && g.rng.Float64() < 0.2 {
			gapSize := 80.0 + g.rng.Float64()*80
			x += gapSize
			if x >= float64(g.levelWidth) {
				break
			}
		}
		
		segLen := segmentLength + g.rng.Float64()*100
		g.platforms = append(g.platforms, &Platform{
			X: x, Y: groundY, Width: segLen, Height: 20,
			Type: "ground",
			Color: color.RGBA{80, 80, 100, 255},
		})
		x += segLen
	}
	
	// Платформы
	numPlatforms := 5 + g.level*2
	for i := 0; i < numPlatforms; i++ {
		px := 300.0 + g.rng.Float64()*float64(g.levelWidth-400)
		py := float64(g.levelHeight) - 150.0 - g.rng.Float64()*250
		width := 80.0 + g.rng.Float64()*120
		
		g.platforms = append(g.platforms, &Platform{
			X: px, Y: py, Width: width, Height: 20,
			Type: "platform",
			Color: color.RGBA{60, 60, 90, 255},
		})
	}
	
	// Стены для wall run
	wallX := 400.0 + g.rng.Float64()*float64(g.levelWidth-500)
	g.platforms = append(g.platforms, &Platform{
		X: wallX, Y: float64(g.levelHeight) - 300, Width: 20, Height: 200,
		Type: "wall",
		Color: color.RGBA{70, 70, 100, 255},
	})
}

// generateEnemies генерирует врагов
func (g *Game) generateEnemies(levelNum int) {
	numEnemies := 3 + levelNum
	types := []entity.EnemyType{entity.EnemySoldier, entity.EnemyDrone, entity.EnemyRobot}
	if levelNum >= 5 {
		types = append(types, entity.EnemyElite)
	}
	
	for i := 0; i < numEnemies; i++ {
		x := 400.0 + g.rng.Float64()*float64(g.levelWidth-500)
		y := float64(g.levelHeight) - 80
		
		// Дроны выше
		enemyType := types[g.rng.Intn(len(types))]
		if enemyType == entity.EnemyDrone {
			y = float64(g.levelHeight) - 200 - g.rng.Float64()*150
		}
		
		enemy := entity.NewEnemy(x, y, enemyType, g.spriteSheet)
		enemy.PatrolStart = x - 80
		enemy.PatrolEnd = x + 80
		g.enemies = append(g.enemies, enemy)
	}
}

// generateObstacles генерирует препятствия
func (g *Game) generateObstacles(levelNum int) {
	// Двери
	numDoors := 1 + levelNum/2
	for i := 0; i < numDoors; i++ {
		x := 500.0 + float64(i)*400 + g.rng.Float64()*100
		door := entity.NewDoor(x, float64(g.levelHeight)-140, 80)
		g.doors = append(g.doors, door)
		
		// Терминал рядом
		termX := x + 70
		termY := float64(g.levelHeight) - 200
		terminal := entity.NewTerminal(termX, termY)
		terminal.LinkedDoor = door
		g.terminals = append(g.terminals, terminal)
	}
}

// generateItems генерирует предметы
func (g *Game) generateItems(levelNum int) {
	numItems := 8 + levelNum*2
	itemTypes := []string{
		entity.ItemHealth, entity.ItemEnergy, entity.ItemArmor,
		entity.ItemData, entity.ItemGrenade, entity.ItemEMP,
	}
	
	for i := 0; i < numItems; i++ {
		x := 200.0 + g.rng.Float64()*float64(g.levelWidth-300)
		y := float64(g.levelHeight) - 100 - g.rng.Float64()*200
		
		itemType := itemTypes[g.rng.Intn(len(itemTypes))]
		value := 1
		if itemType == entity.ItemData {
			value = 100
		}
		
		item := entity.NewItem(x, y, itemType, value, g.spriteSheet)
		g.items = append(g.items, item)
	}
}

// generateSecurity генерирует системы безопасности
func (g *Game) generateSecurity(levelNum int) {
	// Камеры
	numCameras := 2 + levelNum/2
	for i := 0; i < numCameras; i++ {
		x := 400.0 + float64(i)*350 + g.rng.Float64()*100
		y := float64(g.levelHeight) - 250 - g.rng.Float64()*100
		g.cameras = append(g.cameras, entity.NewCamera(x, y))
	}
	
	// Турели
	numTurrets := 1 + levelNum/3
	for i := 0; i < numTurrets; i++ {
		x := 500.0 + float64(i)*450
		y := float64(g.levelHeight) - 120
		g.turrets = append(g.turrets, entity.NewTurret(x, y))
	}
}

// Update обновляет игру
func (g *Game) Update() error {
	// ESC - пауза/меню
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		switch g.state {
		case StatePlaying, StateHacking:
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
	
	// Хакерство
	if g.state == StateHacking {
		g.updateHacking()
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
	g.updateEnemies(dt)
	g.updateSecurity(dt)
	g.collectItems()
	g.updateProjectiles(dt)
	g.updateParticles(dt)
	g.checkWallCollisions()
	g.checkLevelExit()
	
	// Смерть игрока
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
		if g.player.WallSliding {
			g.player.WallJump()
			g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height, float64(-g.player.Physics.WallSide)*100, -150, 8, color.RGBA{100, 100, 100, 255})
		} else {
			g.player.Jump()
			if g.player.JumpCount == 1 {
				// Двойной прыжок - частицы
				g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height, 0, -100, 10, color.RGBA{0, 255, 255, 255})
			}
		}
	} else if !jumpKey {
		g.jumpPressed = false
	}
	
	// Рывок
	dashKey := ebiten.IsKeyPressed(ebiten.KeyShift)
	if dashKey && !g.dashPressed {
		g.dashPressed = true
		g.player.Dash()
		g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+g.player.Transform.Height/2, float64(-g.player.Transform.Facing)*200, 0, 15, color.RGBA{0, 255, 255, 255})
	} else if !dashKey {
		g.dashPressed = false
	}
	
	// Стрельба
	shootKey := ebiten.IsKeyPressed(ebiten.KeyJ) || ebiten.IsKeyPressed(ebiten.KeyK)
	if shootKey && !g.shootPressed {
		g.shootPressed = true
		g.shoot()
	} else if !shootKey {
		g.shootPressed = false
	}
	
	// Хакерство
	hackKey := ebiten.IsKeyPressed(ebiten.KeyE)
	if hackKey && !g.hackPressed {
		g.hackPressed = true
		g.tryStartHack()
	} else if !hackKey {
		g.hackPressed = false
	}
	
	// Граната
	if ebiten.IsKeyPressed(ebiten.KeyL) {
		g.throwGrenade()
	}
	
	// EMP
	if ebiten.IsKeyPressed(ebiten.KeyI) {
		g.useEMP()
	}
}

// shoot стреляет
func (g *Game) shoot() {
	dirX := float64(g.player.Transform.Facing)
	
	proj := Projectile{
		X: g.player.Transform.X + g.player.Transform.Width/2,
		Y: g.player.Transform.Y + g.player.Transform.Height/3,
		VX: dirX * 700,
		VY: 0,
		Width: 16, Height: 6,
		Damage: 15,
		LifeTime: 1.0,
		Active: true,
		FromPlayer: true,
		Color: color.RGBA{0, 255, 255, 255},
	}
	g.projectiles = append(g.projectiles, proj)
}

// throwGrenade бросает гранату
func (g *Game) throwGrenade() {
	if !g.player.UseGrenade() {
		return
	}
	
	dirX := float64(g.player.Transform.Facing)
	
	proj := Projectile{
		X: g.player.Transform.X + g.player.Transform.Width/2,
		Y: g.player.Transform.Y,
		VX: dirX * 400,
		VY: -300,
		Width: 12, Height: 12,
		Damage: 50,
		LifeTime: 1.5,
		Active: true,
		FromPlayer: true,
		Color: color.RGBA{255, 150, 50, 255},
	}
	g.projectiles = append(g.projectiles, proj)
}

// useEMP использует EMP-импульс
func (g *Game) useEMP() {
	if !g.player.UseEMP() {
		return
	}
	
	// Оглушить всех врагов в радиусе
	empRange := 250.0
	for _, enemy := range g.enemies {
		dx := enemy.Transform.X - g.player.Transform.X
		dy := enemy.Transform.Y - g.player.Transform.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist < empRange {
			enemy.Health.Invincible = 3.0
			enemy.State = entity.EnemyHurt
		}
	}
	
	// Отключить турели и камеры
	for _, turret := range g.turrets {
		dx := turret.Transform.X - g.player.Transform.X
		dy := turret.Transform.Y - g.player.Transform.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < empRange {
			turret.Active = false
			time.AfterFunc(5*time.Second, func() {
				if turret != nil && !turret.Hacked {
					turret.Active = true
				}
			})
		}
	}
	
	// Эффект
	g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, 0, 30, color.RGBA{200, 100, 255, 255})
}

// tryStartHack пытается начать взлом
func (g *Game) tryStartHack() {
	// Поиск ближайшего терминала
	for _, terminal := range g.terminals {
		dx := terminal.Transform.X - g.player.Transform.X
		dy := terminal.Transform.Y - g.player.Transform.Y
		dist := math.Sqrt(dx*dx + dy*dy)
		
		if dist < 80 && !terminal.Hacked {
			g.hackTarget = terminal
			g.hackBarY = 0
			g.hackBarSpeed = 100 + g.rng.Float64()*50
			g.hackZoneStart = 0.3
			g.hackZoneEnd = 0.5
			g.state = StateHacking
			return
		}
	}
}

// updateHacking обновляет мини-игру взлома
func (g *Game) updateHacking() {
	dt := 1.0 / 60.0
	
	// Движение полоски
	g.hackBarY += g.hackBarSpeed * dt
	if g.hackBarY > 1 || g.hackBarY < 0 {
		g.hackBarSpeed = -g.hackBarSpeed
	}
	
	// Взлом (удержание E)
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		// Проверка попадания в зону
		if g.hackBarY >= g.hackZoneStart && g.hackBarY <= g.hackZoneEnd {
			g.hackTarget.Hack(dt*0.05, 0.5)
		}
		
		if g.hackTarget.Hacked {
			g.state = StatePlaying
			g.score += 50
			g.spawnParticles(g.hackTarget.Transform.X+20, g.hackTarget.Transform.Y+25, 0, -50, 20, color.RGBA{0, 255, 255, 255})
			g.hackTarget = nil
		}
	} else {
		// Прерывание
		g.hackTarget.Reset()
		g.state = StatePlaying
		g.hackTarget = nil
	}
}

// applyPhysics применяет физику
func (g *Game) applyPhysics(dt float64) {
	// Гравитация
	g.player.Physics.VelocityY += g.player.Physics.Gravity * dt
	
	// Ограничение скорости
	if g.player.Physics.VelocityY > 800 {
		g.player.Physics.VelocityY = 800
	}
	
	g.player.Transform.X += g.player.Physics.VelocityX * dt
	g.checkPlatformCollisions(true)
	
	g.player.Transform.Y += g.player.Physics.VelocityY * dt
	g.player.Physics.OnGround = false
	g.player.WallSliding = false
	g.checkPlatformCollisions(false)
	
	// Wall slide
	if !g.player.Physics.OnGround && g.player.Physics.WallSide != 0 {
		if g.player.Physics.VelocityY > 100 {
			g.player.WallSliding = true
			g.player.Physics.VelocityY *= 0.5 // Замедление падения
		}
	}
	
	// Границы уровня
	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
		g.player.Physics.VelocityX = 0
	}
	if g.player.Transform.X > float64(g.levelWidth)-g.player.Transform.Width {
		g.player.Transform.X = float64(g.levelWidth) - g.player.Transform.Width
		g.player.Physics.VelocityX = 0
	}
	
	// Падение в пропасть
	if g.player.Transform.Y > float64(g.levelHeight) {
		g.player.Health.TakeDamage(100)
	}
	
	// Трение
	if g.player.Physics.OnGround {
		g.player.Physics.VelocityX *= g.player.Physics.Friction
	}
}

// checkPlatformCollisions проверяет коллизии с платформами
func (g *Game) checkPlatformCollisions(horizontal bool) {
	playerRect := g.player.Transform
	
	for _, platform := range g.platforms {
		platRect := &entity.Transform{
			X: platform.X, Y: platform.Y,
			Width: platform.Width, Height: platform.Height,
		}
		
		if entity.CheckCollision(playerRect, platRect) {
			if horizontal {
				// Столкновение по X - стена
				if g.player.Physics.VelocityX > 0 {
					g.player.Transform.X = platform.X - g.player.Transform.Width
				} else if g.player.Physics.VelocityX < 0 {
					g.player.Transform.X = platform.X + platform.Width
				}
				g.player.Physics.VelocityX = 0
				g.player.Physics.WallSide = 1
				if g.player.Physics.VelocityX < 0 {
					g.player.Physics.WallSide = -1
				}
			} else {
				// Столкновение по Y
				if g.player.Physics.VelocityY > 0 {
					// Падение вниз
					g.player.Transform.Y = platform.Y - g.player.Transform.Height
					g.player.Physics.VelocityY = 0
					g.player.Physics.OnGround = true
					g.player.ResetJump()
				} else if g.player.Physics.VelocityY < 0 {
					// Прыжок вверх
					g.player.Transform.Y = platform.Y + platform.Height
					g.player.Physics.VelocityY = 0
				}
			}
		}
	}
}

// checkWallCollisions проверяет коллизии со стенами
func (g *Game) checkWallCollisions() {
	// Для wall run
}

// updateCamera обновляет камеру
func (g *Game) updateCamera() {
	targetX := g.player.Transform.X - screenWidth/2
	targetY := g.player.Transform.Y - screenHeight/2
	
	// Плавное следование
	g.cameraX += (targetX - g.cameraX) * 0.1
	g.cameraY += (targetY - g.cameraY) * 0.1
	
	// Ограничения
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraX > float64(g.levelWidth)-screenWidth {
		g.cameraX = float64(g.levelWidth) - screenWidth
	}
	if g.cameraY < -100 {
		g.cameraY = -100
	}
	if g.cameraY > float64(g.levelHeight)-screenHeight {
		g.cameraY = float64(g.levelHeight) - screenHeight
	}
}

// updateEnemies обновляет врагов
func (g *Game) updateEnemies(dt float64) {
	playerX := g.player.Transform.X
	playerY := g.player.Transform.Y
	
	for _, enemy := range g.enemies {
		enemy.Update(dt, playerX, playerY)
		
		// Атака на игрока
		if damage, canAttack := enemy.Attack(); canAttack {
			if entity.CheckCollision(g.player.Transform, enemy.Transform) {
				g.player.Health.TakeDamage(damage)
				g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
			}
		}
	}
}

// updateSecurity обновляет системы безопасности
func (g *Game) updateSecurity(dt float64) {
	playerX := g.player.Transform.X + g.player.Transform.Width/2
	playerY := g.player.Transform.Y + g.player.Transform.Height/2
	
	// Камеры
	for _, camera := range g.cameras {
		camera.Update(dt)
		camera.Alert = camera.CanDetect(playerX, playerY)
		
		if camera.Alert {
			g.alertLevel = 1
		}
	}
	
	// Турели
	for _, turret := range g.turrets {
		turret.Update(dt, playerX, playerY)
		
		if turret.CanAttack(playerX, playerY) {
			g.player.Health.TakeDamage(turret.Attack())
			g.spawnParticles(g.player.Transform.X+16, g.player.Transform.Y+24, 0, -50, 10, color.RGBA{255, 50, 50, 255})
		}
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
			g.collectItem(item)
		}
	}
}

// collectItem обрабатывает сбор предмета
func (g *Game) collectItem(item *entity.Item) {
	switch item.ItemType {
	case entity.ItemHealth:
		g.player.Health.Heal(25)
	case entity.ItemEnergy:
		g.player.Energy.Current = g.player.Energy.Max
	case entity.ItemArmor:
		g.player.Health.AddArmor(25)
	case entity.ItemData:
		g.player.DataChips += item.Value
		g.score += item.Value
	case entity.ItemGrenade:
		g.player.Grenades++
	case entity.ItemEMP:
		g.player.EMPCharges++
	}

	// Цвет частиц по типу предмета
	particleColor := color.RGBA{255, 255, 255, 255}
	switch item.ItemType {
	case entity.ItemHealth:
		particleColor = color.RGBA{255, 50, 50, 255}
	case entity.ItemEnergy:
		particleColor = color.RGBA{50, 200, 255, 255}
	case entity.ItemArmor:
		particleColor = color.RGBA{100, 100, 150, 255}
	case entity.ItemData:
		particleColor = color.RGBA{0, 255, 0, 255}
	case entity.ItemGrenade:
		particleColor = color.RGBA{255, 150, 50, 255}
	case entity.ItemEMP:
		particleColor = color.RGBA{200, 100, 255, 255}
	}
	g.spawnParticles(item.Transform.X+12, item.Transform.Y+12, 0, -50, 10, particleColor)
}

// updateProjectiles обновляет снаряды
func (g *Game) updateProjectiles(dt float64) {
	active := make([]Projectile, 0)
	
	for _, p := range g.projectiles {
		if !p.Active {
			continue
		}
		
		p.X += p.VX * dt
		p.Y += p.VY * dt
		p.VY += 300 * dt // Гравитация для гранат
		p.LifeTime -= dt
		
		if p.LifeTime <= 0 {
			p.Active = false
			// Взрыв
			if p.Damage >= 50 {
				g.spawnParticles(p.X, p.Y, 0, -50, 20, color.RGBA{255, 150, 50, 255})
				// Урон врагам в радиусе
				for _, enemy := range g.enemies {
					dx := enemy.Transform.X - p.X
					dy := enemy.Transform.Y - p.Y
					dist := math.Sqrt(dx*dx + dy*dy)
					if dist < 100 {
						enemy.TakeDamage(p.Damage)
					}
				}
			}
			continue
		}
		
		// Коллизия с врагами
		if p.FromPlayer {
			for _, enemy := range g.enemies {
				enemyRect := enemy.Transform
				if entity.CheckCollision(&entity.Transform{X: p.X, Y: p.Y, Width: p.Width, Height: p.Height}, enemyRect) {
					enemy.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.X, p.Y, p.VX*0.2, 0, 8, p.Color)
					break
				}
			}
		}
		
		// Коллизия с платформами
		if p.Active {
			for _, platform := range g.platforms {
				if p.X >= platform.X && p.X <= platform.X+platform.Width &&
					p.Y >= platform.Y && p.Y <= platform.Y+platform.Height {
					p.Active = false
					g.spawnParticles(p.X, p.Y, 0, -50, 5, p.Color)
					break
				}
			}
		}
		
		if p.Active {
			active = append(active, p)
		}
	}
	
	g.projectiles = active
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
		g.particles = append(g.particles, Particle{
			X: x, Y: y,
			VX: vx + (g.rng.Float64()-0.5)*100,
			VY: vy + (g.rng.Float64()-0.5)*100,
			Life: 1.0, MaxLife: 1.0,
			Color: c,
			Size: 2 + g.rng.Float64()*3,
		})
	}
}

// checkLevelExit проверяет выход с уровня
func (g *Game) checkLevelExit() {
	exitRect := &entity.Transform{
		X: g.exitX, Y: g.exitY,
		Width: 60, Height: 80,
	}
	
	if entity.CheckCollision(g.player.Transform, exitRect) {
		// Проверка, все ли данные собраны
		if g.player.DataChips >= 100 {
			g.state = StateVictory
		}
	}
}

// Particle - частица
type Particle struct {
	X, Y, VX, VY float64
	Life, MaxLife float64
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
		vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y-g.cameraY), float32(p.Width), float32(p.Height), p.Color, false)
	}
	
	// Предметы
	for _, item := range g.items {
		item.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Терминалы
	for _, terminal := range g.terminals {
		terminal.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Двери
	for _, door := range g.doors {
		door.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Камеры
	for _, camera := range g.cameras {
		camera.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Турели
	for _, turret := range g.turrets {
		turret.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Враги
	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Снаряды
	for _, p := range g.projectiles {
		if p.Active {
			vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y-g.cameraY), float32(p.Width), float32(p.Height), p.Color, false)
		}
	}
	
	// Игрок
	if g.player != nil {
		g.player.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Частицы
	for _, p := range g.particles {
		vector.DrawFilledRect(screen, float32(p.X-g.cameraX), float32(p.Y-g.cameraY), float32(p.Size), float32(p.Size), p.Color, false)
	}
	
	// Выход
	g.drawExit(screen)
	
	// HUD
	if g.state == StatePlaying || g.state == StateHacking {
		g.drawHUD(screen)
	}
	
	// Хакерская мини-игра
	if g.state == StateHacking {
		g.drawHacking(screen)
	}
	
	// Пауза
	if g.state == StatePaused {
		g.drawPause(screen)
	}
	
	// Game Over
	if g.state == StateGameOver {
		g.drawGameOver(screen)
	}
	
	// Victory
	if g.state == StateVictory {
		g.drawVictory(screen)
	}
}

// drawBackground отрисовывает фон
func (g *Game) drawBackground(screen *ebiten.Image) {
	// Градиент неба (тёмный киберпанк)
	for y := 0; y < screenHeight; y++ {
		percent := float64(y) / float64(screenHeight)
		r := uint8(20 + percent*30)
		g := uint8(20 + percent*20)
		b := uint8(40 + percent*40)
		vector.DrawFilledRect(screen, 0, float32(y), screenWidth, 1, color.RGBA{r, g, b, 255}, false)
	}
	
	// Силуэт города на фоне
	for i := 0; i < 20; i++ {
		buildingX := float32(i*80 - int(g.cameraX*0.3)%80)
		buildingHeight := float32(100 + (i*17)%150)
		vector.DrawFilledRect(screen, buildingX, float32(screenHeight)-buildingHeight, 70, buildingHeight, color.RGBA{30, 30, 50, 255}, false)
		
		// Окна
		for wy := float32(20); wy < buildingHeight-20; wy += 30 {
			for wx := float32(10); wx < 60; wx += 15 {
				if (i+int(wy))%3 == 0 {
					vector.DrawFilledRect(screen, buildingX+wx, float32(screenHeight)-buildingHeight+wy, 8, 15, color.RGBA{200, 200, 100, 200}, false)
				}
			}
		}
	}
}

// drawMenu отрисовывает меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := "CYBER CITY RUNNER"
	ebitenutil.DebugPrintAt(screen, title, screenWidth/2-150, 150)
	
	subtitle := "Go365 Day 92 - Киберпанк-платформер"
	ebitenutil.DebugPrintAt(screen, subtitle, screenWidth/2-120, 200)
	
	instructions := []string{
		"",
		"[SPACE] Начать игру",
		"",
		"Управление:",
		"A/D - Бег",
		"W/Пробел - Прыжок (двойной в воздухе)",
		"Shift - Рывок",
		"J/K - Стрельба",
		"E - Хакерство",
		"L - Граната",
		"I - EMP",
		"",
		"Цель: Собрать 100 данных и добраться до выхода",
	}
	
	for i, line := range instructions {
		ebitenutil.DebugPrintAt(screen, line, screenWidth/2-150, 280+i*20)
	}
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	y := 10
	
	// Здоровье
	vector.DrawFilledRect(screen, 10, float32(y), 200, 20, color.RGBA{50, 50, 50, 255}, false)
	hpPercent := float32(g.player.Health.Current) / float32(g.player.Health.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*hpPercent, 20, color.RGBA{0, 255, 100, 255}, false)
	ebitenutil.DebugPrintAt(screen, "HP", 220, y)
	
	y += 25
	
	// Энергия
	vector.DrawFilledRect(screen, 10, float32(y), 200, 15, color.RGBA{50, 50, 50, 255}, false)
	energyPercent := float32(g.player.Energy.Current) / float32(g.player.Energy.Max)
	vector.DrawFilledRect(screen, 10, float32(y), 200*energyPercent, 15, color.RGBA{0, 200, 255, 255}, false)
	ebitenutil.DebugPrintAt(screen, "NRG", 220, y)
	
	y += 25
	
	// Броня
	vector.DrawFilledRect(screen, 10, float32(y), 200, 15, color.RGBA{50, 50, 50, 255}, false)
	armorPercent := float32(g.player.Health.Armor) / 100.0
	vector.DrawFilledRect(screen, 10, float32(y), 200*armorPercent, 15, color.RGBA{100, 100, 150, 255}, false)
	ebitenutil.DebugPrintAt(screen, "ARM", 220, y)
	
	y += 30
	
	// Данные
	ebitenutil.DebugPrintAt(screen, "DATA: "+string(rune(g.player.DataChips))+" / 100", 10, y)
	y += 20
	
	// Счёт
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), 10, y)
	y += 20
	
	// Гранаты
	ebitenutil.DebugPrintAt(screen, "GRENADES: "+string(rune(g.player.Grenades)), 10, y)
	y += 20
	
	// EMP
	ebitenutil.DebugPrintAt(screen, "EMP: "+string(rune(g.player.EMPCharges)), 10, y)
	
	// Уровень
	ebitenutil.DebugPrintAt(screen, "LEVEL: "+string(rune(g.level)), screenWidth-100, 10)
	
	// Подсказки
	ebitenutil.DebugPrintAt(screen, "[ESC] Пауза  [Shift] Рывок  [E] Хак  [L] Граната  [I] EMP", 10, screenHeight-30)
}

// drawHacking отрисовывает мини-игру взлома
func (g *Game) drawHacking(screen *ebiten.Image) {
	// Затемнение фона
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
	
	// Окно взлома
	hackX := screenWidth/2 - 150
	hackY := screenHeight/2 - 100
	vector.DrawFilledRect(screen, float32(hackX), float32(hackY), 300, 200, color.RGBA{20, 20, 40, 255}, false)
	vector.StrokeRect(screen, float32(hackX), float32(hackY), 300, 200, 2, color.RGBA{0, 255, 255, 255}, false)
	
	ebitenutil.DebugPrintAt(screen, "HACKING...", hackX+100, hackY+10)
	
	// Зона взлома
	zoneY := hackY + 60
	zoneHeight := int((g.hackZoneEnd - g.hackZoneStart) * 100)
	vector.DrawFilledRect(screen, float32(hackX+20), float32(zoneY)+float32(g.hackZoneStart*100), 260, float32(zoneHeight), color.RGBA{0, 255, 0, 100}, false)
	
	// Движущаяся полоска
	barY := float32(zoneY) + float32(g.hackBarY*100)
	vector.DrawFilledRect(screen, float32(hackX+20), barY, 260, 3, color.RGBA{255, 255, 0, 255}, false)
	
	// Прогресс
	progressY := hackY + 170
	progressWidth := g.hackTarget.HackProgress * 260
	vector.DrawFilledRect(screen, float32(hackX+20), float32(progressY), float32(progressWidth), 15, color.RGBA{0, 255, 255, 255}, false)
	
	ebitenutil.DebugPrintAt(screen, "Удерживай E в зелёной зоне!", hackX+50, hackY+130)
}

// drawPause отрисовывает паузу
func (g *Game) drawPause(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 0, 0, 150}, false)
	ebitenutil.DebugPrintAt(screen, "ПАУЗА", screenWidth/2-50, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "[ESC] Продолжить", screenWidth/2-80, screenHeight/2)
}

// drawGameOver отрисовывает Game Over
func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{50, 0, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "GAME OVER", screenWidth/2-80, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "SCORE: "+string(rune(g.score)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Рестарт", screenWidth/2-80, screenHeight/2+50)
}

// drawVictory отрисовывает победу
func (g *Game) drawVictory(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight, color.RGBA{0, 50, 0, 200}, false)
	ebitenutil.DebugPrintAt(screen, "УРОВЕНЬ ПРОЙДЕН!", screenWidth/2-100, screenHeight/2-50)
	ebitenutil.DebugPrintAt(screen, "DATA: "+string(rune(g.player.DataChips)), screenWidth/2-60, screenHeight/2)
	ebitenutil.DebugPrintAt(screen, "[SPACE] Следующий уровень", screenWidth/2-120, screenHeight/2+50)
}

// drawExit отрисовывает выход
func (g *Game) drawExit(screen *ebiten.Image) {
	x := g.exitX - g.cameraX
	y := g.exitY - g.cameraY
	
	// Портал выхода
	vector.DrawFilledRect(screen, float32(x), float32(y), 60, 80, color.RGBA{0, 255, 0, 100}, false)
	vector.StrokeRect(screen, float32(x), float32(y), 60, 80, 3, color.RGBA{0, 255, 0, 255}, false)
	
	// Мигание
	if int(time.Now().UnixMilli()/200)%2 == 0 {
		vector.DrawFilledRect(screen, float32(x+10), float32(y+10), 40, 60, color.RGBA{0, 255, 0, 50}, false)
	}
	
	ebitenutil.DebugPrintAt(screen, "EXIT", int(x)+15, int(y)-20)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
