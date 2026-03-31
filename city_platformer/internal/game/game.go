// Package game - основная игровая логика Sunny Adventure
// Go365 Day 91 - Доброе сказочное приключение
package game

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"sunny_adventure/internal/entity"
	"sunny_adventure/internal/level"
	"sunny_adventure/internal/render"
	"sunny_adventure/internal/sprite"
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
	state         GameState
	player        *entity.Player
	currentChunk  *level.LevelData
	enemies       []*entity.Enemy
	friends       []*entity.Friend
	clouds        []*entity.Cloud
	projectiles   []*entity.Projectile
	particles     []render.Particle
	cameraX       float64
	cameraY       float64
	score         int
	chunkNum      int // Текущий чанк
	spriteSheet   *sprite.SpriteSheet
	renderer      *render.Renderer
	rng           *rand.Rand
	levelGen      *level.LevelGenerator
	jumpPressed   bool
	shootPressed  bool
	infiniteWorld bool // Бесконечный мир
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

	g.renderer = render.NewRenderer(g.spriteSheet)
	g.levelGen = level.NewLevelGenerator(rng, tileSize, g.spriteSheet)

	return g
}

// Reset сбрасывает игру
func (g *Game) Reset() {
	g.chunkNum = 0
	g.score = 0
	g.infiniteWorld = true // Включаем бесконечный мир
	g.startChunk()
}

// startChunk запускает чанок
func (g *Game) startChunk() {
	g.currentChunk = g.levelGen.GenerateChunk(g.chunkNum)

	groundLevel := float64(g.currentChunk.Height-2)*float64(g.currentChunk.TileSize)

	// Если это первый чанк или смерть - спавним в начале
	spawnX := 100.0
	if g.chunkNum > 0 && !g.player.Health.Dead {
		// Спавним у выхода из предыдущего чанка
		spawnX = float64(g.chunkNum) * float64(g.levelGen.ChunkWidth) * float64(g.levelGen.TileSize)
	}

	g.player = entity.NewPlayer(spawnX, groundLevel, g.spriteSheet)
	g.player.Physics.OnGround = true
	g.player.Physics.VelocityY = 0

	g.enemies = make([]*entity.Enemy, 0)
	g.friends = make([]*entity.Friend, 0)
	g.clouds = make([]*entity.Cloud, 0)
	g.projectiles = make([]*entity.Projectile, 0)
	g.particles = make([]render.Particle, 0)
	g.cameraX = spawnX - screenWidth/2

	// Создаём врагов из чанка
	for _, le := range g.currentChunk.Enemies {
		if le.Active {
			enemy := entity.NewEnemy(le.X, le.Y, entity.EnemyType(le.Type), g.spriteSheet)
			g.enemies = append(g.enemies, enemy)
		}
	}

	// Создаём друзей
	for _, lf := range g.currentChunk.Friends {
		friend := entity.NewFriend(lf.X, lf.Y, entity.FriendType(lf.Type), g.spriteSheet)
		g.friends = append(g.friends, friend)
	}

	// Создаём облачка
	for _, lc := range g.currentChunk.Clouds {
		cloud := entity.NewCloud(lc.X, lc.Y, lc.Num, g.spriteSheet)
		g.clouds = append(g.clouds, cloud)
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
	dt := 1.0 / 60.0

	g.handlePlayerInput()
	g.player.Update(dt)
	g.applyPhysics(dt)
	g.updateCamera()
	g.collectItems()
	g.collectFriends()
	g.collectClouds()
	g.updateProjectiles(dt)
	g.updateEnemies(dt)
	g.updateParticles(dt)
	g.checkChunkTransition() // Проверка перехода в новый чанк

	// Проверка победы (все облачка собраны + выход достигнут)
	allCloudsCollected := true
	for _, c := range g.clouds {
		if !c.Collected {
			allCloudsCollected = false
			break
		}
	}

	// В бесконечном мире переходим в следующий чанк
	if g.infiniteWorld && allCloudsCollected && g.currentChunk.CheckExitReach(g.player.Transform.X, g.player.Transform.Y, g.player.Transform.Width, g.player.Transform.Height) {
		g.chunkNum++
		g.startChunk()
		g.score += 100 // Бонус за завершение чанка
	}

	// Проверка смерти - респавн в начале чанка
	if g.player.Health.Dead {
		g.player.Health.Dead = false
		g.player.Health.Current = g.player.Health.Max
		g.startChunk()
		g.score -= 50 // Штраф за смерть
		if g.score < 0 {
			g.score = 0
		}
	}
}

// checkChunkTransition проверяет переход в соседний чанк
func (g *Game) checkChunkTransition() {
	if g.player == nil || g.currentChunk == nil {
		return
	}

	newChunkNum := int(g.player.Transform.X) / (g.levelGen.ChunkWidth * g.levelGen.TileSize)
	if newChunkNum < 0 {
		newChunkNum = 0
	}
	
	// Если перешли в новый чанк
	if newChunkNum != g.chunkNum {
		g.chunkNum = newChunkNum
		g.currentChunk = g.levelGen.GenerateChunk(g.chunkNum)
		
		// Быстрое обновление сущностей без лишних аллокаций
		g.enemies = g.enemies[:0]
		g.friends = g.friends[:0]
		g.clouds = g.clouds[:0]
		
		for _, le := range g.currentChunk.Enemies {
			if le.Active {
				g.enemies = append(g.enemies, entity.NewEnemy(le.X, le.Y, entity.EnemyType(le.Type), g.spriteSheet))
			}
		}
		for _, lf := range g.currentChunk.Friends {
			g.friends = append(g.friends, entity.NewFriend(lf.X, lf.Y, entity.FriendType(lf.Type), g.spriteSheet))
		}
		for _, lc := range g.currentChunk.Clouds {
			g.clouds = append(g.clouds, entity.NewCloud(lc.X, lc.Y, lc.Num, g.spriteSheet))
		}
	}
}

// handlePlayerInput обрабатывает ввод игрока
func (g *Game) handlePlayerInput() {
	onLadder := g.checkLadderCollision()

	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		if !onLadder {
			g.player.MoveLeft()
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		if !onLadder {
			g.player.MoveRight()
		}
	}

	// Прыжок (с поддержкой двойного)
	if (ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeySpace)) && !g.jumpPressed {
		g.jumpPressed = true
		g.player.Jump()
		g.spawnParticles(g.player.Transform.X+24, g.player.Transform.Y+g.player.Transform.Height, 0, -50, 8, color.RGBA{255, 255, 100, 255})
	} else if !ebiten.IsKeyPressed(ebiten.KeyW) && !ebiten.IsKeyPressed(ebiten.KeyUp) && !ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.jumpPressed = false
	}

	// Лазание по лестнице
	if onLadder {
		if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyUp) {
			g.player.Physics.VelocityY = -120
		} else if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
			g.player.Physics.VelocityY = 120
		}
	}

	// Выстрел лучиком света
	if ebiten.IsKeyPressed(ebiten.KeyJ) && !g.shootPressed && g.player.CanShoot() {
		g.shootPressed = true
		g.shoot()
	} else if !ebiten.IsKeyPressed(ebiten.KeyJ) {
		g.shootPressed = false
	}

	// Обнимашки (лечение)
	if ebiten.IsKeyPressed(ebiten.KeyK) {
		g.hugFriends()
	}
}

// shoot производит выстрел лучиком
func (g *Game) shoot() {
	g.player.Shoot()

	dirX := float64(g.player.Transform.Facing)
	dirY := 0.0

	projectile := entity.NewProjectile(
		g.player.Transform.X+g.player.Transform.Width/2,
		g.player.Transform.Y+g.player.Transform.Height/3,
		dirX*500,
		dirY,
		15,
		true, // Лучик добра!
		g.spriteSheet,
	)

	g.projectiles = append(g.projectiles, projectile)
}

// hugFriends обнимает друзей (лечение)
func (g *Game) hugFriends() {
	for _, friend := range g.friends {
		if !friend.Collected && entity.CheckCollision(g.player.Transform, friend.Transform) {
			friend.Collected = true
			g.player.Health.Heal(20)
			g.score += 10
			g.spawnParticles(friend.Transform.X+16, friend.Transform.Y+16, 0, -50, 10, color.RGBA{255, 100, 200, 255})
			break
		}
	}
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

	for _, p := range g.currentChunk.Platforms {
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
				if g.player.State != entity.PlayerJumping {
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

	// Границы уровня
	if g.player.Transform.X < 0 {
		g.player.Transform.X = 0
	}
	maxX := float64(g.currentChunk.Width*g.currentChunk.TileSize) - g.player.Transform.Width
	if g.player.Transform.X > maxX {
		g.player.Transform.X = maxX
	}

	// Падение в пропасть
	if g.player.Transform.Y > float64(g.currentChunk.Height)*float64(tileSize) {
		g.player.Health.TakeDamage(100)
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
	maxCameraX := float64(g.currentChunk.Width*g.currentChunk.TileSize) - screenWidth
	if g.cameraX > maxCameraX {
		g.cameraX = maxCameraX
	}
	if g.cameraY < -100 {
		g.cameraY = -100
	}
	maxCameraY := float64(g.currentChunk.Height*g.currentChunk.TileSize) - screenHeight
	if g.cameraY > maxCameraY {
		g.cameraY = maxCameraY
	}
}

// checkLadderCollision проверяет, находится ли игрок на лестнице
func (g *Game) checkLadderCollision() bool {
	playerCenter := g.player.Transform.X + g.player.Transform.Width/2
	playerBottom := g.player.Transform.Y + g.player.Transform.Height

	for _, platform := range g.currentChunk.Platforms {
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
	for _, item := range g.currentChunk.Items {
		if item.Collected {
			continue
		}
		// Создаём временный Transform для проверки коллизии
		itemTransform := &entity.Transform{
			X:      item.X,
			Y:      item.Y,
			Width:  32,
			Height: 32,
		}
		if entity.CheckCollision(g.player.Transform, itemTransform) {
			g.collectItem(item)
		}
	}
}

// collectItem обрабатывает сбор одного предмета
func (g *Game) collectItem(item *level.LevelItem) {
	item.Collected = true

	switch item.Type {
	case "coinGold", "coinSilver", "coinBronze":
		g.score += item.Value
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 10, color.RGBA{255, 215, 0, 255})
	case "gemRed", "gemBlue", "gemGreen", "gemYellow":
		g.score += item.Value
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 15, color.RGBA{100, 200, 255, 255})
	case "star":
		g.score += item.Value
		g.player.Light.Current = g.player.Light.Max
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 20, color.RGBA{255, 255, 255, 255})
	case "mushroomRed", "mushroomBrown":
		g.player.Grow()
		g.spawnParticles(item.X+16, item.Y+16, 0, -100, 10, color.RGBA{255, 100, 100, 255})
	}
}

// collectFriends собирает друзей
func (g *Game) collectFriends() {
	for _, friend := range g.friends {
		if !friend.Collected && entity.CheckCollision(g.player.Transform, friend.Transform) {
			friend.Collected = true
			g.player.FriendCount++
			g.score += 25
			g.spawnParticles(friend.Transform.X+16, friend.Transform.Y+16, 0, -100, 15, color.RGBA{255, 150, 200, 255})
		}
	}
}

// collectClouds собирает облачка
func (g *Game) collectClouds() {
	for _, cloud := range g.clouds {
		if !cloud.Collected && entity.CheckCollision(g.player.Transform, cloud.Transform) {
			cloud.Collected = true
			g.score += 50
			g.spawnParticles(cloud.Transform.X+24, cloud.Transform.Y+16, 0, -50, 20, color.RGBA{200, 200, 255, 255})
		}
	}
}

// updateProjectiles обновляет снаряды
func (g *Game) updateProjectiles(dt float64) {
	active := make([]*entity.Projectile, 0)

	for _, p := range g.projectiles {
		p.Update(dt)

		if p.Active {
			// Коллизия с врагами (лучики добра превращают врагов в друзей!)
			for _, enemy := range g.enemies {
				if !enemy.Converted && entity.CheckCollision(p.Transform, enemy.Transform) {
					enemy.Health.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.Transform.X, p.Transform.Y, p.VelocityX*0.2, 0, 8, color.RGBA{255, 255, 100, 255})

					if enemy.Health.Dead && !enemy.Converted {
						enemy.Convert() // Превратить в друга!
						g.score += 30
					}
					break
				}
			}

			// Коллизия с платформами
			if p.Active {
				for _, platform := range g.currentChunk.Platforms {
					if platform.Solid && entity.CheckCollision(p.Transform, &platform.Transform) {
						p.Active = false
						g.spawnParticles(p.Transform.X, p.Transform.Y, 0, -50, 3, color.RGBA{255, 255, 200, 255})
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
	for _, enemy := range g.enemies {
		enemy.Update(dt, g.player.Transform.X, g.player.Transform.Y)
		enemy.AI(dt, g.player.Transform.X, g.player.Transform.Y)

		// Коллизия с игроком
		if !enemy.Converted && entity.CheckCollision(g.player.Transform, enemy.Transform) {
			if g.player.Health.Invincible <= 0 {
				g.player.Health.TakeDamage(enemy.Damage)
				g.spawnParticles(g.player.Transform.X+24, g.player.Transform.Y+24, 0, -100, 10, color.RGBA{255, 50, 50, 255})
			}
		}
	}
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
	g.renderer.DrawBackground(screen, g.cameraX, g.cameraY, g.chunkNum)

	if g.currentChunk != nil {
		// Платформы
		for _, p := range g.currentChunk.Platforms {
			g.renderer.DrawPlatform(screen, p, g.cameraX, g.cameraY)
		}

		// Декорации
		for _, decor := range g.currentChunk.Decors {
			g.renderer.DrawDecor(screen, decor, g.cameraX, g.cameraY)
		}

		// Предметы
		for _, item := range g.currentChunk.Items {
			if !item.Collected {
				entityItem := entity.NewItem(item.X, item.Y, item.Type, item.Value, g.spriteSheet)
				g.renderer.DrawItem(screen, entityItem, g.cameraX, g.cameraY)
			}
		}

		// Друзья
		for _, friend := range g.friends {
			if !friend.Collected {
				g.renderer.DrawFriend(screen, friend, g.cameraX, g.cameraY)
			}
		}

		// Облачка
		for _, cloud := range g.clouds {
			if !cloud.Collected {
				g.renderer.DrawCloud(screen, cloud, g.cameraX, g.cameraY)
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

		// Выход (если есть)
		if g.currentChunk.ExitX > 0 {
			g.renderer.DrawExit(screen, g.currentChunk.ExitX, g.currentChunk.ExitY, g.cameraX, g.cameraY)
		}
	}

	g.renderer.DrawParticles(screen, g.particles, g.cameraX, g.cameraY)

	switch g.state {
	case StateMenu:
		g.renderer.DrawMenu(screen)
	case StatePlaying:
		g.drawHUD(screen)
	case StatePaused:
		g.drawHUD(screen)
		g.renderer.DrawPause(screen)
	case StateGameOver:
		g.renderer.DrawGameOver(screen, g.score, g.chunkNum)
	case StateVictory:
		g.renderer.DrawVictory(screen, g.score, g.player.FriendCount)
	}
}

// drawHUD отрисовывает интерфейс
func (g *Game) drawHUD(screen *ebiten.Image) {
	g.renderer.DrawHUD(
		screen,
		g.player.Health.Current,
		g.player.Health.Max,
		g.player.Light.Current,
		g.player.Light.Max,
		g.score,
		g.chunkNum,
		g.currentChunk.Name,
		g.player.FriendCount,
		len(g.clouds),
	)

	ebitenutil.DebugPrintAt(screen, "[ESC] Пауза  [J] Морковка  [K] Обнимашки  🌍 Бесконечный мир!", 10, screenHeight-30)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
