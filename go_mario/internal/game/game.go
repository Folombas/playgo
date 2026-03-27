// Package game предоставляет основную игровую логику
package game

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"

	"go_mario/internal/assets"
	"go_mario/internal/level"
	"go_mario/internal/player"
)

const (
	// Размеры экрана
	ScreenWidth  = 1024
	ScreenHeight = 768

	// Состояния игры
	StateMenu = iota
	StatePlaying
	StateGameOver
	StateWin
)

// Game представляет основную игру
type Game struct {
	state       int
	player      *player.Player
	level       *level.Level
	assets      *assets.Assets
	cameraX     float64
	frame       int
	score       int
	coins       int
	lives       int
	world       int
	levelNum    int
	font        font.Face
}

// NewGame создаёт новую игру
func NewGame() *Game {
	g := &Game{
		state:    StateMenu,
		assets:   assets.Get(),
		world:    1,
		levelNum: 1,
	}

	g.startNewGame()

	return g
}

// startNewGame начинает новую игру
func (g *Game) startNewGame() {
	g.player = player.New(100, 300)
	g.startLevel()
}

// startLevel начинает новый уровень
func (g *Game) startLevel() {
	// Создаём уровень 100x16 тайлов
	g.level = level.NewLevel(100, 16)
	g.level.Generate(int64(g.world*1000 + g.levelNum))

	// Сбрасываем позицию игрока
	g.player.Reset(100, 300)
	g.cameraX = 0
}

// Update обновляет состояние игры
func (g *Game) Update() error {
	g.frame++

	switch g.state {
	case StateMenu:
		return g.updateMenu()
	case StatePlaying:
		return g.updatePlaying()
	case StateGameOver, StateWin:
		return g.updateEndScreen()
	}

	return nil
}

// updateMenu обновляет меню
func (g *Game) updateMenu() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.state = StatePlaying
	}
	return nil
}

// updatePlaying обновляет игровой процесс
func (g *Game) updatePlaying() error {
	// Обработка ввода
	g.player.HandleInput()

	// Обновляем уровень
	g.level.Update()

	// Обновляем игрока
	g.player.Update()

	// Проверяем коллизии с тайлами
	onGround, newX, newY := g.level.CheckTileCollision(
		g.player.X, g.player.Y,
		float64(g.player.Width), float64(g.player.Height),
		g.player.VY,
	)

	g.player.X = newX
	g.player.Y = newY
	g.player.SetOnGround(onGround)

	// Ограничиваем по горизонтали
	if g.player.X < 0 {
		g.player.X = 0
	}
	if g.player.X > float64(g.level.Width*level.TileSize)-float64(g.player.Width) {
		g.player.X = float64(g.level.Width*level.TileSize) - float64(g.player.Width)
	}

	// Проверяем смерть от падения
	if g.player.Y > ScreenHeight {
		g.player.TakeDamage(3) // Мгновенная смерть
		if g.player.IsDead() {
			g.state = StateGameOver
		} else {
			g.player.Reset(100, 300)
		}
	}

	// Обновляем камеру
	g.updateCamera()

	// Проверяем коллизии с монетами
	g.checkCoinCollisions()

	// Проверяем коллизии с врагами
	g.checkEnemyCollisions()

	// Проверяем флаг победы
	g.checkFlagCollision()

	return nil
}

// updateCamera обновляет позицию камеры
func (g *Game) updateCamera() {
	targetX := g.player.X - ScreenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1

	// Ограничиваем камеру
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	maxX := float64(g.level.Width*level.TileSize) - ScreenWidth
	if g.cameraX > maxX {
		g.cameraX = maxX
	}
}

// checkCoinCollisions проверяет сбор монет
func (g *Game) checkCoinCollisions() {
	for _, coin := range g.level.Coins {
		if coin.Collected {
			continue
		}

		if checkCollision(
			g.player.X, g.player.Y,
			float64(g.player.Width), float64(g.player.Height),
			coin.X, coin.Y,
			32, 32,
		) {
			coin.Collected = true
			g.player.CollectCoin(coin.Value)
		}
	}
}

// checkEnemyCollisions проверяет коллизии с врагами
func (g *Game) checkEnemyCollisions() {
	for _, enemy := range g.level.Enemies {
		if !enemy.Alive {
			continue
		}

		if checkCollision(
			g.player.X, g.player.Y,
			float64(g.player.Width), float64(g.player.Height),
			enemy.X, enemy.Y,
			32, 32,
		) {
			// Если игрок падает сверху - убиваем врага
			if g.player.VY > 0 && g.player.Y+float64(g.player.Height) < enemy.Y+20 {
				enemy.Alive = false
				g.player.VY = -8 // Отскок
				g.score += 100
			} else {
				// Иначе игрок получает урон
				g.player.TakeDamage(1)
				if g.player.IsDead() {
					g.state = StateGameOver
				}
			}
		}
	}
}

// checkFlagCollision проверяет достижение флага
func (g *Game) checkFlagCollision() {
	if g.level.Flag.Collected {
		return
	}

	if checkCollision(
		g.player.X, g.player.Y,
		float64(g.player.Width), float64(g.player.Height),
		g.level.Flag.X, g.level.Flag.Y,
		48, 96,
	) {
		g.level.Flag.Collected = true
		g.score += 1000
		g.levelNum++

		if g.levelNum > 3 {
			g.world++
			g.levelNum = 1
		}

		if g.world > 3 {
			g.state = StateWin
		} else {
			g.startLevel()
		}
	}
}

// updateEndScreen обновляет экран конца игры
func (g *Game) updateEndScreen() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if g.state == StateGameOver {
			g.startNewGame()
			g.state = StatePlaying
		} else {
			g.state = StateMenu
		}
	}
	return nil
}

// Draw рисует игру
func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка экрана
	screen.Fill(color.RGBA{100, 149, 237, 255}) // Cornflower blue

	switch g.state {
	case StateMenu:
		g.drawMenu(screen)
	case StatePlaying:
		g.drawPlaying(screen)
	case StateGameOver:
		g.drawPlaying(screen)
		g.drawGameOver(screen)
	case StateWin:
		g.drawPlaying(screen)
		g.drawWin(screen)
	}
}

// drawMenu рисует меню
func (g *Game) drawMenu(screen *ebiten.Image) {
	fontToUse := g.assets.GameFont
	if fontToUse == nil {
		fontToUse = g.font
	}
	if fontToUse == nil {
		return
	}
	
	// Заголовок
	title := "GO MARIO"
	titleWidth := len(title) * 7
	text.Draw(screen, title, fontToUse, (ScreenWidth-titleWidth)/2, 200, color.White)

	// Инструкция
	instruction := "Press ENTER to start"
	instWidth := len(instruction) * 7
	text.Draw(screen, instruction, fontToUse, (ScreenWidth-instWidth)/2, 400, color.White)

	// Управление
	controls := []string{
		"Controls:",
		"Arrow Keys / WASD - Move",
		"Space / W - Jump",
		"S / Down - Duck",
	}

	for i, line := range controls {
		text.Draw(screen, line, fontToUse, 100, 500+i*20, color.White)
	}
}

// drawPlaying рисует игровой процесс
func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Рисуем фон с параллаксом
	g.drawBackground(screen)

	// Рисуем декорации
	g.drawDecorations(screen, g.cameraX)

	// Рисуем тайлы
	g.drawTiles(screen, g.cameraX)

	// Рисуем монеты
	g.drawCoins(screen, g.cameraX)

	// Рисуем врагов
	g.drawEnemies(screen, g.cameraX)

	// Рисуем флаг
	g.drawFlag(screen, g.cameraX)

	// Рисуем игрока
	g.drawPlayer(screen)

	// Рисуем UI
	g.drawUI(screen)
}

// drawBackground рисует фон
func (g *Game) drawBackground(screen *ebiten.Image) {
	if g.assets.Background == nil {
		return
	}

	bgWidth := float64(g.assets.Background.Bounds().Dx())
	
	// Рисуем фон с параллаксом (медленнее камеры)
	bgX := math.Mod(-g.cameraX*0.3, bgWidth)
	if bgX < 0 {
		bgX += bgWidth
	}
	
	// Рисуем фон несколько раз чтобы заполнить экран
	for x := bgX - bgWidth; x < ScreenWidth; x += bgWidth {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x, 0)
		screen.DrawImage(g.assets.Background, op)
	}
}

// drawDecorations рисует декорации
func (g *Game) drawDecorations(screen *ebiten.Image, cameraX float64) {
	for _, dec := range g.level.Decorations {
		// Проверка видимости в камере
		if dec.X < cameraX-100 || dec.X > cameraX+ScreenWidth+100 {
			continue
		}
		
		var sprite *ebiten.Image
		switch dec.Type {
		case 0: // cloud
			idx := int(dec.X/100) % 3
			switch idx {
			case 0:
				sprite = g.assets.Cloud1
			case 1:
				sprite = g.assets.Cloud2
			case 2:
				sprite = g.assets.Cloud3
			}
		case 1: // bush
			sprite = g.assets.Bush
		case 2: // plant
			sprite = g.assets.Plant
		}

		if sprite != nil {
			decOp := &ebiten.DrawImageOptions{}
			decOp.GeoM.Translate(dec.X-cameraX, dec.Y)
			screen.DrawImage(sprite, decOp)
		}
	}
}

// drawTiles рисует тайлы
func (g *Game) drawTiles(screen *ebiten.Image, cameraX float64) {
	for _, tile := range g.level.Tiles {
		tileX := float64(tile.X * level.TileSize)
		
		// Проверка видимости в камере
		if tileX < cameraX-level.TileSize || tileX > cameraX+ScreenWidth {
			continue
		}
		
		var sprite *ebiten.Image
		switch tile.Type {
		case level.TileGrassTop:
			sprite = g.assets.GrassTop
		case level.TileGrassMid:
			sprite = g.assets.GrassMid
		case level.TileDirtTop:
			sprite = g.assets.DirtTop
		case level.TileDirtMid:
			sprite = g.assets.DirtMid
		case level.TileBrick:
			sprite = g.assets.Brick
		case level.TileBoxEmpty:
			sprite = g.assets.BoxEmpty
		case level.TileBoxItem:
			sprite = g.assets.BoxItem
		case level.TileBoxCoin:
			sprite = g.assets.BoxCoin
		case level.TileBoxUsed:
			sprite = g.assets.BoxUsed
		}

		if sprite != nil {
			tileOp := &ebiten.DrawImageOptions{}
			tileOp.GeoM.Translate(
				tileX-cameraX,
				float64(tile.Y*level.TileSize),
			)
			screen.DrawImage(sprite, tileOp)
		}
	}
}

// drawCoins рисует монеты
func (g *Game) drawCoins(screen *ebiten.Image, cameraX float64) {
	for _, coin := range g.level.Coins {
		if coin.Collected {
			continue
		}

		// Проверка видимости в камере
		if coin.X < cameraX-50 || coin.X > cameraX+ScreenWidth+50 {
			continue
		}

		var sprite *ebiten.Image
		switch coin.Value {
		case 1:
			sprite = g.assets.CoinBronze
		case 2:
			sprite = g.assets.CoinSilver
		case 3:
			sprite = g.assets.CoinGold
		}

		if sprite != nil {
			coinOp := &ebiten.DrawImageOptions{}
			coinOp.GeoM.Translate(coin.X-cameraX, coin.Y)
			screen.DrawImage(sprite, coinOp)
		}
	}
}

// drawEnemies рисует врагов
func (g *Game) drawEnemies(screen *ebiten.Image, cameraX float64) {
	for _, enemy := range g.level.Enemies {
		if !enemy.Alive {
			continue
		}

		// Проверка видимости в камере
		if enemy.X < cameraX-50 || enemy.X > cameraX+ScreenWidth+50 {
			continue
		}

		var sprite *ebiten.Image
		switch enemy.Type {
		case 0: // slime
			if enemy.AnimFrame%2 == 0 {
				sprite = g.assets.SlimeWalk1
			} else {
				sprite = g.assets.SlimeWalk2
			}
		case 1: // fly
			if enemy.AnimFrame%2 == 0 {
				sprite = g.assets.FlyWalk1
			} else {
				sprite = g.assets.FlyWalk2
			}
		}

		if sprite != nil {
			enemyOp := &ebiten.DrawImageOptions{}
			enemyOp.GeoM.Translate(enemy.X-cameraX, enemy.Y)

			// Отражение по направлению
			if enemy.Direction > 0 {
				enemyOp.GeoM.Scale(-1, 1)
				enemyOp.GeoM.Translate(32, 0)
			}

			screen.DrawImage(sprite, enemyOp)
		}
	}
}

// drawFlag рисует флаг
func (g *Game) drawFlag(screen *ebiten.Image, cameraX float64) {
	if g.level.Flag == nil {
		return
	}

	var sprite *ebiten.Image
	if g.level.Flag.Color == 0 {
		sprite = g.assets.FlagGreen
	} else {
		sprite = g.assets.FlagRed
	}

	if sprite != nil {
		flagOp := &ebiten.DrawImageOptions{}
		flagOp.GeoM.Translate(g.level.Flag.X-cameraX, g.level.Flag.Y)
		screen.DrawImage(sprite, flagOp)
	}
}

// drawPlayer рисует игрока
func (g *Game) drawPlayer(screen *ebiten.Image) {
	// Получаем спрайт игрока
	var sprite *ebiten.Image
	switch g.player.AnimState {
	case player.AnimIdle:
		sprite = g.assets.PlayerStand
	case player.AnimWalk:
		if g.player.AnimFrame%2 == 0 {
			sprite = g.assets.PlayerWalk1
		} else {
			sprite = g.assets.PlayerWalk2
		}
	case player.AnimJump:
		sprite = g.assets.PlayerJump
	case player.AnimDuck:
		sprite = g.assets.PlayerDuck
	case player.AnimHurt:
		sprite = g.assets.PlayerHurt
	default:
		sprite = g.assets.PlayerStand
	}

	if sprite == nil {
		return
	}

	// Мигание при неуязвимости - пропускаем отрисовку
	if g.player.Invincible && g.frame%6 < 3 {
		return
	}

	op := &ebiten.DrawImageOptions{}

	// Отражение по направлению
	if !g.player.FacingRight {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(g.player.Width), 0)
	}

	// Позиция с камерой
	op.GeoM.Translate(g.player.X-g.cameraX, g.player.Y)

	screen.DrawImage(sprite, op)
}

// drawUI рисует интерфейс
func (g *Game) drawUI(screen *ebiten.Image) {
	// Панель UI
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 50, color.RGBA{0, 0, 0, 128}, false)

	fontToUse := g.assets.GameFont
	if fontToUse == nil {
		fontToUse = g.font
	}
	if fontToUse == nil {
		return
	}

	// Счёт
	scoreText := fmt.Sprintf("SCORE: %06d", g.player.Score)
	text.Draw(screen, scoreText, fontToUse, 20, 35, color.White)

	// Монеты
	coinText := fmt.Sprintf("COINS: %d", g.player.Coins)
	text.Draw(screen, coinText, fontToUse, 300, 35, color.White)

	// Жизни
	livesText := fmt.Sprintf("LIVES: %d", g.player.Lives)
	text.Draw(screen, livesText, fontToUse, 500, 35, color.White)

	// Мир и уровень
	levelText := fmt.Sprintf("WORLD %d-%d", g.world, g.levelNum)
	text.Draw(screen, levelText, fontToUse, 700, 35, color.White)
}

// drawGameOver рисует экран проигрыша
func (g *Game) drawGameOver(screen *ebiten.Image) {
	// Полупрозрачный оверлей
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)

	fontToUse := g.assets.GameFont
	if fontToUse == nil {
		fontToUse = g.font
	}
	if fontToUse == nil {
		return
	}

	// Текст GAME OVER
	gameOver := "GAME OVER"
	goWidth := len(gameOver) * 7
	text.Draw(screen, gameOver, fontToUse, (ScreenWidth-goWidth)/2, 300, color.RGBA{255, 0, 0, 255})

	// Счёт
	scoreText := fmt.Sprintf("Final Score: %d", g.player.Score)
	sw := len(scoreText) * 7
	text.Draw(screen, scoreText, fontToUse, (ScreenWidth-sw)/2, 400, color.White)

	// Инструкция
	restart := "Press ENTER to restart"
	rw := len(restart) * 7
	text.Draw(screen, restart, fontToUse, (ScreenWidth-rw)/2, 500, color.White)
}

// drawWin рисует экран победы
func (g *Game) drawWin(screen *ebiten.Image) {
	// Полупрозрачный оверлей
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)

	fontToUse := g.assets.GameFont
	if fontToUse == nil {
		fontToUse = g.font
	}
	if fontToUse == nil {
		return
	}

	// Текст WIN
	win := "YOU WIN!"
	wWidth := len(win) * 7
	text.Draw(screen, win, fontToUse, (ScreenWidth-wWidth)/2, 300, color.RGBA{255, 215, 0, 255})

	// Счёт
	scoreText := fmt.Sprintf("Final Score: %d", g.player.Score)
	sw := len(scoreText) * 7
	text.Draw(screen, scoreText, fontToUse, (ScreenWidth-sw)/2, 400, color.White)

	// Инструкция
	menu := "Press ENTER to menu"
	mw := len(menu) * 7
	text.Draw(screen, menu, fontToUse, (ScreenWidth-mw)/2, 500, color.White)
}

// Layout возвращает размеры экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// checkCollision проверяет пересечение двух прямоугольников
func checkCollision(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

// DebugInfo возвращает отладочную информацию
func (g *Game) DebugInfo() string {
	return fmt.Sprintf("FPS: %.0f\nPos: (%.1f, %.1f)\nCam: %.1f",
		ebiten.ActualFPS,
		g.player.X, g.player.Y,
		g.cameraX,
	)
}

// DrawDebug рисует отладочную информацию
func (g *Game) DrawDebug(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, g.DebugInfo())
}
