// Package game - основная игровая логика Food Platformer
// Go365 Day 88 - Food Edition
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
	"github.com/playgo/city_platformer/pkg/entity"
	"github.com/playgo/city_platformer/pkg/level"
	"github.com/playgo/city_platformer/pkg/render"
	"github.com/playgo/city_platformer/pkg/sprite"
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
	state       GameState
	player      *entity.Player
	platforms   []level.Platform
	foods       []level.Collectible
	enemies     []entity.Enemy
	projectiles []entity.Projectile
	particles   []render.Particle
	cameraX     float64
	cameraY     float64
	score       int
	level       int
	wave        int
	combo       int
	comboTimer  float64
	rng         *rand.Rand
	
	// Ресурсы
	spriteSheet *sprite.SpriteSheet
	levelData   *level.Level
	renderer    *render.Renderer
	
	// Босс
	boss        *entity.Boss
	bossActive  bool
}

// NewGame - создание новой игры
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	g := &Game{
		state: StateMenu,
		rng:   rng,
	}
	g.Reset()
	return g
}

// Reset - сброс игры к начальному состоянию
func (g *Game) Reset() {
	g.player = entity.NewPlayer(100, 500)
	g.level = 1
	g.wave = 1
	g.score = 0
	g.cameraX = 0
	g.cameraY = 0
	g.combo = 0
	g.comboTimer = 0
	g.bossActive = false
	g.boss = nil
	g.projectiles = make([]entity.Projectile, 0)
	g.particles = make([]render.Particle, 0)
	
	// Загрузка спрайтов
	g.spriteSheet = sprite.LoadSpriteSheet()
	g.renderer = render.NewRenderer(g.spriteSheet)
	
	// Генерация уровня
	g.generateLevel()
}

// generateLevel - генерация уровня (кухня)
func (g *Game) generateLevel() {
	g.levelData = level.GenerateLevel(g.level, g.rng)
	g.platforms = g.levelData.Platforms
	g.foods = g.levelData.Collectibles
	g.enemies = g.generateEnemies()
}

// generateEnemies - генерация врагов (испорченная еда/жуки)
func (g *Game) generateEnemies() []entity.Enemy {
	enemies := make([]entity.Enemy, 0)
	enemyCount := 3 + g.level*2
	
	for i := 0; i < enemyCount; i++ {
		enemyType := "rotten"
		if g.rng.Float64() < 0.5 {
			enemyType = "bug"
		}
		x := float64(400 + i*300 + g.rng.Intn(100))
		y := float64(600 - g.rng.Intn(200))
		
		enemies = append(enemies, *entity.NewEnemy(x, y, enemyType))
	}
	
	return enemies
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
	
	// Victory - следующий уровень
	if g.state == StateVictory && ebiten.IsKeyPressed(ebiten.KeySpace) {
		g.level++
		g.generateLevel()
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
	
	// Сбор еды
	g.collectFood()
	
	// Обновление пуль
	g.updateProjectiles()
	
	// Обновление врагов
	g.updateEnemies()
	
	// Обновление частиц
	g.updateParticles()
	
	// Обновление босса
	if g.bossActive && g.boss != nil {
		g.updateBoss()
	}
	
	// Проверка победы (все враги убиты)
	if len(g.enemies) == 0 && !g.bossActive {
		if g.level % 3 == 0 && g.wave >= 3 {
			g.spawnBoss()
		} else {
			g.state = StateVictory
		}
	}
	
	// Таймер комбо
	if g.combo > 0 {
		g.comboTimer -= 1.0 / 60.0
		if g.comboTimer <= 0 {
			g.combo = 0
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
		g.spawnParticles(g.player.X+20, g.player.Y+25, 0, -2, 8, color.RGBA{200, 200, 200, 255})
	}
	
	// Приседание
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.player.Crouch()
	} else {
		g.player.Stand()
	}
	
	// Бросок ингредиентом
	if ebiten.IsKeyPressed(ebiten.KeyJ) && g.player.CanShoot() {
		g.shoot()
	}
}

// collectFood - сбор еды
func (g *Game) collectFood() {
	score, healthChange := g.levelData.CheckCollection(
		g.player.X, g.player.Y, g.player.Width, g.player.Height,
	)
	
	if score > 0 {
		g.score += score * (1 + g.combo)
		g.addCombo()
		g.spawnParticles(g.player.X+20, g.player.Y+20, 0, -2, 10, color.RGBA{255, 200, 50, 255})
	}
	
	if healthChange != 0 {
		if healthChange > 0 {
			g.player.Heal(healthChange)
		} else {
			g.player.TakeDamage(-healthChange)
		}
	}
}

// shoot - бросок ингредиентом
func (g *Game) shoot() {
	g.player.Shoot()
	
	dirX := float64(g.player.Facing)
	dirY := 0.0
	
	if g.player.IsCrouching {
		dirY = 0.3
	}
	
	projectile := entity.NewProjectile(
		g.player.X+20, g.player.Y+16,
		dirX*10, dirY*10,
		0, // тип еды
		false,
	)
	
	g.projectiles = append(g.projectiles, *projectile)
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
	for _, p := range g.platforms {
		if g.checkCollision(g.player.X+8, g.player.Y+8, p.X, p.Y, p.Width-16, p.Height-16) {
			if g.player.VY > 0 && g.player.Y+g.player.Height < p.Y+30 {
				g.player.Y = p.Y - g.player.Height
				g.player.VY = 0
				g.player.OnGround = true
			}
		}
	}
	
	// Пол
	if g.player.Y > 620 {
		g.player.Y = 620
		g.player.VY = 0
		g.player.OnGround = true
	}
	
	// Смерть от падения
	if g.player.Y > float64(screenHeight)+100 {
		g.player.TakeDamage(100)
	}
}

// updateCamera - обновление камеры
func (g *Game) updateCamera() {
	targetX := g.player.X - screenWidth/2
	targetY := g.player.Y - screenHeight/2
	
	g.cameraX += (targetX - g.cameraX) * 0.1
	g.cameraY += (targetY - g.cameraY) * 0.1
	
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	if g.cameraY < -100 {
		g.cameraY = -100
	}
}

// updateProjectiles - обновление снарядов
func (g *Game) updateProjectiles() {
	active := make([]entity.Projectile, 0)
	
	for _, p := range g.projectiles {
		p.Update()
		
		if p.Active {
			// Коллизия с врагами
			for i, enemy := range g.enemies {
				if g.checkCollision(p.X, p.Y, enemy.X, enemy.Y, p.Width, p.Height) {
					g.killEnemy(i, enemy.X, enemy.Y)
					p.Active = false
					g.addCombo()
					break
				}
			}
			
			// Коллизия с боссом
			if g.bossActive && g.boss != nil && p.Active {
				if g.checkCollision(p.X, p.Y, g.boss.X, g.boss.Y, p.Width, p.Height) {
					g.boss.TakeDamage(p.Damage)
					p.Active = false
					g.spawnParticles(p.X, p.Y, p.VX, p.VY, 5, color.RGBA{255, 100, 100, 255})
					
					if g.boss.Health <= 0 {
						g.killBoss()
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

// updateEnemies - обновление врагов
func (g *Game) updateEnemies() {
	active := make([]entity.Enemy, 0)
	
	for _, enemy := range g.enemies {
		enemy.Update()
		
		// Движение врага
		switch enemy.Type {
		case "slime":
			enemy.VX = math.Sin(float64(time.Now().UnixNano())/1e9) * 2
			enemy.VY = 0
		case "fly":
			enemy.Y += math.Sin(float64(time.Now().UnixNano())/5e8) * 0.5
		}
		
		enemy.X += enemy.VX
		enemy.Y += enemy.VY
		
		// Коллизия с игроком
		if g.checkCollision(g.player.X+8, g.player.Y+8, enemy.X+8, enemy.Y+8, 24, 24) {
			if g.player.Invincible <= 0 {
				g.playerHit()
			}
		}
		
		active = append(active, enemy)
	}
	
	g.enemies = active
}

// updateBoss - обновление босса
func (g *Game) updateBoss() {
	if g.boss == nil {
		return
	}
	
	g.boss.Update()
	
	// Движение босса
	targetX := g.player.X
	g.boss.X += (targetX - g.boss.X) * 0.02
	
	// Атака босса
	if g.rng.Float64() < 0.02 {
		g.bossAttack()
	}
	
	// Коллизия с игроком
	if g.checkCollision(g.player.X+8, g.player.Y+8, g.boss.X+20, g.boss.Y+20, 24, 40) {
		if g.player.Invincible <= 0 {
			g.playerHit()
		}
	}
}

// bossAttack - атака босса
func (g *Game) bossAttack() {
	// Босс выпускает снаряды
	for i := -2; i <= 2; i++ {
		projectile := entity.Projectile{
			X:    g.boss.X + 40,
			Y:    g.boss.Y + 30,
			VX:   float64(i) * 2,
			VY:   3,
			Width:  12,
			Height: 12,
			Life:   180,
			Active: true,
			IsEnemy: true,
		}
		g.projectiles = append(g.projectiles, projectile)
	}
}

// spawnBoss - спавн босса
func (g *Game) spawnBoss() {
	g.bossActive = true
	g.boss = entity.NewBoss(800, 400)
	g.spawnParticles(800, 400, 0, 0, 50, color.RGBA{255, 0, 0, 255})
}

// killBoss - убийство босса
func (g *Game) killBoss() {
	g.bossActive = false
	g.boss = nil
	g.score += 500
	g.spawnParticles(800, 400, 0, 0, 100, color.RGBA{255, 100, 0, 255})
	g.state = StateVictory
}

// killEnemy - убийство врага
func (g *Game) killEnemy(index int, x, y float64) {
	g.enemies = append(g.enemies[:index], g.enemies[index+1:]...)
	g.score += 25 * (1 + g.combo)
	g.spawnParticles(x+16, y+16, 0, -3, 15, color.RGBA{100, 150, 50, 255})
}

// addCombo - добавление комбо
func (g *Game) addCombo() {
	g.combo++
	g.comboTimer = 3.0
}

// playerHit - игрок получил урон
func (g *Game) playerHit() {
	g.combo = 0
	g.spawnParticles(g.player.X+20, g.player.Y+25, 0, -2, 20, color.RGBA{255, 0, 0, 255})
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
			Life: 1.0, MaxLife: 1.0,
			Color: c,
			Size: 3 + g.rng.Float64()*5,
		})
	}
}

// checkCollision - проверка коллизии AABB
func (g *Game) checkCollision(x1, y1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w2 > x2 && y1 < y2+h2 && y1+h2 > y2
}

// Draw - отрисовка игры
func (g *Game) Draw(screen *ebiten.Image) {
	// Фон (кухня)
	if g.renderer != nil {
		g.renderer.DrawBackground(screen, g.cameraX, g.cameraY)
	}
	
	// Платформы
	for _, p := range g.platforms {
		if g.renderer != nil {
			g.renderer.DrawPlatform(screen, p, g.cameraX, g.cameraY)
		}
	}
	
	// Еда
	for _, food := range g.foods {
		f := entity.NewFood(food.X, food.Y, food.TypeInt, food.Value)
		f.AnimFrame = food.AnimFrame
		if g.renderer != nil {
			g.renderer.DrawFood(screen, f, g.cameraX, g.cameraY)
		} else {
			f.Draw(screen, g.cameraX, g.cameraY)
		}
	}
	
	// Враги
	for _, enemy := range g.enemies {
		if g.renderer != nil {
			g.renderer.DrawEnemy(screen, enemy, g.cameraX, g.cameraY)
		} else {
			enemy.Draw(screen, g.cameraX, g.cameraY)
		}
	}
	
	// Босс
	if g.bossActive && g.boss != nil {
		if g.renderer != nil {
			g.renderer.DrawBoss(screen, *g.boss, g.cameraX, g.cameraY)
		}
	}
	
	// Игрок
	if g.renderer != nil {
		g.renderer.DrawPlayer(screen, *g.player, g.cameraX, g.cameraY)
	} else {
		g.player.Draw(screen, g.cameraX, g.cameraY)
	}
	
	// Снаряды
	for _, p := range g.projectiles {
		if g.renderer != nil {
			g.renderer.DrawProjectile(screen, p, g.cameraX, g.cameraY)
		} else {
			p.Draw(screen, g.cameraX, g.cameraY)
		}
	}
	
	// Частицы
	if g.renderer != nil {
		g.renderer.DrawParticles(screen, g.particles, g.cameraX, g.cameraY)
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
	case StateVictory:
		g.drawVictory(screen)
	}
}

// drawSky - отрисовка неба
func (g *Game) drawSky(screen *ebiten.Image) {
	for y := 0; y < screenHeight; y++ {
		ratio := float64(y) / float64(screenHeight)
		r := uint8(135 - ratio*35)
		gr := uint8(206 - ratio*86)
		b := uint8(235)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screenWidth), 1, color.RGBA{r, gr, b, 255}, true)
	}
}

// drawPlatformFallback - резервная отрисовка платформы
func (g *Game) drawPlatformFallback(screen *ebiten.Image, p level.Platform) {
	screenX := p.X - g.cameraX
	screenY := p.Y - g.cameraY
	
	var c color.RGBA
	switch p.Type {
	case "counter":
		c = color.RGBA{180, 140, 100, 255}
	case "floor":
		c = color.RGBA{200, 200, 200, 255}
	case "shelf":
		c = color.RGBA{160, 120, 80, 255}
	default:
		c = color.RGBA{150, 150, 150, 255}
	}
	
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(p.Width), float32(p.Height), c, true)
}

// drawEnemyFallback - резервная отрисовка врага
func (g *Game) drawEnemyFallback(screen *ebiten.Image, enemy entity.Enemy) {
	enemy.Draw(screen, g.cameraX, g.cameraY)
}

// drawParallaxBackground - параллакс фон
func (g *Game) drawParallaxBackground(screen *ebiten.Image) {
	if g.renderer != nil {
		g.renderer.DrawBackground(screen, g.cameraX, g.cameraY)
	}
}

// drawMenu - отрисовка меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	title := `
╔═══════════════════════════════════════════════╗
║     🍳 FOOD PLATFORMER 🎖️                    ║
║        LAST COOK STANDING                     ║
╠═══════════════════════════════════════════════╣
║                                               ║
║           [SPACE] - Начать игру               ║
║           [ESC] - Выход                       ║
║                                               ║
║  🎮 Управление:                               ║
║     A/D или ←/→ - Бег                         ║
║     W/↑ - Прыжок                              ║
║     S/↓ - Присесть                            ║
║     J - Бросить ингредиент                    ║
║                                               ║
║  🎯 Цель: Собери всю еду!                     ║
║  🤢 Остерегайся гнилой еды и жуков!           ║
║  🍎 Полезная еда восстанавливает здоровье!    ║
║                                               ║
╚═══════════════════════════════════════════════╝
`
	ebitenutil.DebugPrint(screen, title)
}

// drawHUD - отрисовка интерфейса
func (g *Game) drawHUD(screen *ebiten.Image) {
	if g.renderer != nil {
		g.renderer.DrawHUD(screen, g.score, 0, g.level, g.player.Health, g.player.MaxHealth)
	}
	
	// Название уровня
	levelName := "Кухня"
	if g.levelData != nil && g.levelData.Name != "" {
		levelName = g.levelData.Name
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("📍 %s", levelName), 10, 110)
	
	// Комбо
	if g.combo > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("🔥 Комбо x%d!", g.combo), 10, 130)
	}
	
	ebitenutil.DebugPrintAt(screen, "[ESC] - Пауза  [J] - Бросок", 10, screenHeight-30)
	
	// Полоска здоровья босса
	if g.bossActive && g.boss != nil && g.renderer != nil {
		g.renderer.DrawBossHealthBar(screen, g.boss.Health, g.boss.MaxHealth)
	}
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
║       🤢 ПОВАР УВОЛЕН! 🤢             ║
╠═══════════════════════════════════════╣
║     Финальный счёт: %6d                ║
║     Уровень: %2d                        ║
║     Комбо: %3d                         ║
║                                       ║
║     [SPACE] - Новая попытка           ║
║     [ESC] - Выход                     ║
╚═══════════════════════════════════════╝
`, g.score, g.level, g.combo)

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
║       🍳 УРОВЕНЬ ПРОЙДЕН! 🍳          ║
╠═══════════════════════════════════════╣
║     Счёт: %6d                          ║
║     Уровень: %2d                        ║
║                                       ║
║     [SPACE] - Следующий уровень       ║
║     [ESC] - Пауза                     ║
╚═══════════════════════════════════════╝
`, g.score, g.level)

	ebitenutil.DebugPrintAt(screen, victoryText, screenWidth/2-180, screenHeight/2-100)
}

// Layout - размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
