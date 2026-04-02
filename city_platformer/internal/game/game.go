package game

import (
	"image/color"
	"log"

	"city_platformer/internal/entity"
	"city_platformer/internal/level"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
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
	width      int
	height     int
	state      GameState
	level      *level.Level
	player     *entity.Player
	cameraX    float64
	score      int
	lives      int
	coins      int
	levelNum   int
	fontImage  *ebiten.Image
	particles  []*entity.Particle
}

// NewGame создаёт новую игру
func NewGame() *Game {
	return &Game{
		width:    800,
		height:   600,
		state:    StateMenu,
		lives:    3,
		levelNum: 1,
	}
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.width, g.height
}

// Update обновляет логику игры
func (g *Game) Update() error {
	switch g.state {
	case StateMenu:
		g.updateMenu()
	case StatePlaying:
		g.updatePlaying()
	case StatePaused:
		g.updatePaused()
	case StateGameOver:
		g.updateGameOver()
	case StateVictory:
		g.updateVictory()
	}

	// Обновляем частицы
	g.updateParticles()

	return nil
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана
	screen.Fill(color.RGBA{135, 206, 235, 255}) // Небесно-голубой

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StatePaused:
		g.drawPaused(screen)
	case StateGameOver:
		g.drawGameOver(screen)
	case StateVictory:
		g.drawVictory(screen)
	}

	// Отрисовка частиц
	g.drawParticles(screen)
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.startLevel()
	}
}

func (g *Game) updatePlaying() {
	// Обновление игрока
	g.player.Update()

	// Обновление уровня
	g.level.Update()

	// Обновление камеры
	g.updateCamera()

	// Проверка коллизий
	g.checkCollisions()

	// Проверка победы на уровне
	if g.player.X > float64(g.level.Width-100) {
		g.levelComplete()
	}

	// Проверка падения в пропасть
	if g.player.Y > float64(g.height)+100 {
		g.playerDie()
	}

	// Пауза
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StatePaused
	}
}

func (g *Game) updatePaused() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.state = StatePlaying
	}
}

func (g *Game) updateGameOver() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.restartGame()
	}
}

func (g *Game) updateVictory() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.restartGame()
	}
}

func (g *Game) startLevel() {
	var err error
	g.level, err = level.LoadLevel(g.levelNum)
	if err != nil {
		log.Printf("Error loading level: %v", err)
		g.level = level.CreateDefaultLevel()
	}

	g.player = entity.NewPlayer(100, 300)
	g.cameraX = 0
	g.score = 0
	g.coins = 0
	g.particles = make([]*entity.Particle, 0)

	g.state = StatePlaying
}

func (g *Game) updateCamera() {
	// Камера следует за игроком
	targetX := g.player.X - float64(g.width)/2

	// Ограничение камеры
	if targetX < 0 {
		targetX = 0
	}
	if targetX > float64(g.level.Width-g.width) {
		targetX = float64(g.level.Width - g.width)
	}

	// Плавное движение камеры
	g.cameraX += (targetX - g.cameraX) * 0.1
}

func (g *Game) checkCollisions() {
	// Коллизии с платформами
	g.player.OnGround = false
	for _, platform := range g.level.Platforms {
		if g.checkPlatformCollision(platform) {
			break
		}
	}

	// Коллизии с предметами
	for i := len(g.level.Items) - 1; i >= 0; i-- {
		item := g.level.Items[i]
		if g.checkItemCollision(item) {
			g.collectItem(item)
			g.level.Items = append(g.level.Items[:i], g.level.Items[i+1:]...)
			g.spawnParticles(item.X, item.Y, 5, color.RGBA{255, 215, 0, 255})
		}
	}

	// Коллизии с врагами
	for _, enemy := range g.level.Enemies {
		if g.checkEnemyCollision(enemy) {
			// Если игрок прыгает на врага сверху
			if g.player.VY > 0 && g.player.Y+g.player.Height < enemy.Y+enemy.Height/2 {
				g.killEnemy(enemy)
				g.player.VY = -5 // Отскок
				g.score += 25
				g.spawnParticles(enemy.X, enemy.Y, 8, color.RGBA{255, 100, 100, 255})
			} else {
				g.playerHit()
			}
		}
	}
}

func (g *Game) checkPlatformCollision(platform *level.Platform) bool {
	// Простая AABB коллизия
	if g.player.X < platform.X+platform.Width &&
		g.player.X+g.player.Width > platform.X &&
		g.player.Y < platform.Y+platform.Height &&
		g.player.Y+g.player.Height > platform.Y {

		// Определяем направление коллизии
		overlapX := (g.player.Width + platform.Width) / 2 - abs((g.player.X+g.player.Width/2) - (platform.X+platform.Width/2))
		overlapY := (g.player.Height + platform.Height) / 2 - abs((g.player.Y+g.player.Height/2) - (platform.Y+platform.Height/2))

		if overlapX < overlapY {
			// Коллизия по X
			if g.player.X < platform.X {
				g.player.X = platform.X - g.player.Width
			} else {
				g.player.X = platform.X + platform.Width
			}
			g.player.VX = 0
		} else {
			// Коллизия по Y
			if g.player.VY > 0 && g.player.Y+g.player.Height <= platform.Y+overlapY {
				// Приземление на платформу
				g.player.Y = platform.Y - g.player.Height
				g.player.VY = 0
				g.player.OnGround = true
			} else if g.player.VY < 0 {
				// Удар головой
				g.player.Y = platform.Y + platform.Height
				g.player.VY = 0
			}
		}
		return true
	}
	return false
}

func (g *Game) checkItemCollision(item *level.Item) bool {
	return g.player.X < item.X+item.Width &&
		g.player.X+g.player.Width > item.X &&
		g.player.Y < item.Y+item.Height &&
		g.player.Y+g.player.Height > item.Y
}

func (g *Game) checkEnemyCollision(enemy *entity.Enemy) bool {
	return g.player.X < enemy.X+enemy.Width &&
		g.player.X+g.player.Width > enemy.X &&
		g.player.Y < enemy.Y+enemy.Height &&
		g.player.Y+g.player.Height > enemy.Y
}

func (g *Game) collectItem(item *level.Item) {
	switch item.Type {
	case "coin":
		g.coins++
		g.score += 10
		if g.coins >= 100 {
			g.coins = 0
			g.lives++
		}
	case "star":
		g.score += 50
	case "heart":
		g.lives++
	}
}

func (g *Game) killEnemy(enemy *entity.Enemy) {
	enemy.Alive = false
}

func (g *Game) playerHit() {
	g.lives--
	g.spawnParticles(g.player.X, g.player.Y, 10, color.RGBA{255, 0, 0, 255})

	if g.lives <= 0 {
		g.state = StateGameOver
	} else {
		// Респаун игрока
		g.player.X = 100
		g.player.Y = 300
		g.player.VX = 0
		g.player.VY = 0
		g.cameraX = 0
	}
}

func (g *Game) playerDie() {
	g.lives--
	g.spawnParticles(g.player.X, g.player.Y, 15, color.RGBA{255, 50, 50, 255})

	if g.lives <= 0 {
		g.state = StateGameOver
	} else {
		g.player.X = 100
		g.player.Y = 300
		g.player.VX = 0
		g.player.VY = 0
		g.cameraX = 0
	}
}

func (g *Game) levelComplete() {
	g.levelNum++
	if g.levelNum > 3 {
		g.state = StateVictory
	} else {
		g.startLevel()
	}
}

func (g *Game) restartGame() {
	g.levelNum = 1
	g.lives = 3
	g.score = 0
	g.coins = 0
	g.startLevel()
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.Update()
		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) spawnParticles(x, y float64, count int, c color.Color) {
	for i := 0; i < count; i++ {
		g.particles = append(g.particles, entity.NewParticle(x, y, c))
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Заголовок
	ebitenutil.DebugPrintAt(screen, "PIXEL QUEST", g.width/2-60, 150)
	ebitenutil.DebugPrintAt(screen, "Go365 Challenge - Day 93", g.width/2-80, 200)

	// Инструкция
	ebitenutil.DebugPrintAt(screen, "Press ENTER to Start", g.width/2-70, 300)
	ebitenutil.DebugPrintAt(screen, "Controls:", g.width/2-40, 380)
	ebitenutil.DebugPrintAt(screen, "A/D - Move", g.width/2-40, 410)
	ebitenutil.DebugPrintAt(screen, "W/Space - Jump", g.width/2-40, 430)
	ebitenutil.DebugPrintAt(screen, "ESC - Pause", g.width/2-40, 450)
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Отрисовка уровня
	g.level.Draw(screen, g.cameraX)

	// Отрисовка игрока
	g.player.Draw(screen, g.cameraX)

	// HUD
	g.drawHUD(screen)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Фон HUD
	vector.DrawFilledRect(screen, 0, 0, float32(g.width), 40, color.RGBA{0, 0, 0, 128}, false)

	// Текст HUD
	ebitenutil.DebugPrintAt(screen, "SCORE: "+itoa(g.score), 10, 10)
	ebitenutil.DebugPrintAt(screen, "COINS: "+itoa(g.coins), 150, 10)
	ebitenutil.DebugPrintAt(screen, "LIVES: "+itoa(g.lives), 300, 10)
	ebitenutil.DebugPrintAt(screen, "LEVEL: "+itoa(g.levelNum), 450, 10)
}

func (g *Game) drawPaused(screen *ebiten.Image) {
	g.drawPlaying(screen)

	// Полупрозрачный оверлей
	vector.DrawFilledRect(screen, 0, 0, float32(g.width), float32(g.height), color.RGBA{0, 0, 0, 128}, false)

	// Текст паузы
	ebitenutil.DebugPrintAt(screen, "PAUSED", g.width/2-40, g.height/2-20)
	ebitenutil.DebugPrintAt(screen, "Press ESC to Resume", g.width/2-70, g.height/2+20)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 0, 0, 255})
	ebitenutil.DebugPrintAt(screen, "GAME OVER", g.width/2-50, g.height/2-20)
	ebitenutil.DebugPrintAt(screen, "Press ENTER to Restart", g.width/2-80, g.height/2+20)
}

func (g *Game) drawVictory(screen *ebiten.Image) {
	screen.Fill(color.RGBA{255, 215, 0, 255})
	ebitenutil.DebugPrintAt(screen, "VICTORY!", g.width/2-50, g.height/2-20)
	ebitenutil.DebugPrintAt(screen, "Final Score: "+itoa(g.score), g.width/2-60, g.height/2+10)
	ebitenutil.DebugPrintAt(screen, "Press ENTER to Play Again", g.width/2-80, g.height/2+50)
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		p.Draw(screen, g.cameraX)
	}
}

// Вспомогательные функции
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
