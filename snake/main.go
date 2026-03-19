package main

import (
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"playgo/snake/internal/game"
	"playgo/snake/internal/effects"
	"playgo/snake/internal/ui"
)

const (
	screenWidth  = 800
	screenHeight = 600
)

// App представляет основное приложение
type App struct {
	game      *game.Game
	effects   *effects.EffectSystem
	renderer  *ui.Renderer
	background *ebiten.Image
}

// NewApp создаёт новое приложение
func NewApp() *App {
	cfg := game.DefaultConfig()
	
	app := &App{
		game:      game.NewGame(),
		effects:   effects.NewEffectSystem(),
		renderer:  ui.NewRenderer(cfg),
		background: effects.CreateGradientBackground(screenWidth, screenHeight),
	}
	return app
}

// Update обновляет состояние приложения
func (a *App) Update() error {
	// Обработка ввода в меню
	switch a.game.State {
	case game.Menu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			a.game.State = game.SelectDifficulty
		}
		return nil

	case game.SelectDifficulty:
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			if a.game.Difficulty == game.Easy {
				a.game.Difficulty = game.Hard
			} else {
				a.game.Difficulty--
			}
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			if a.game.Difficulty == game.Hard {
				a.game.Difficulty = game.Easy
			} else {
				a.game.Difficulty++
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			a.game.StartGame()
		}
		return nil

	case game.Paused:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			a.game.State = game.Playing
		}
		return nil

	case game.GameOver:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			a.game = game.NewGame()
			a.effects = effects.NewEffectSystem()
		}
		return nil

	case game.Playing:
		if inpututil.IsKeyJustPressed(ebiten.KeyP) {
			a.game.State = game.Paused
			return nil
		}
	}

	// Обработка направления движения
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		a.game.UpdateDirection(game.Up)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		a.game.UpdateDirection(game.Down)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		a.game.UpdateDirection(game.Left)
	} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		a.game.UpdateDirection(game.Right)
	}

	// Выстрел стрелой
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.game.ShootArrow()
	}

	// Обновление игры
	events := a.game.Update()

	// Обработка игровых событий и создание эффектов
	for _, event := range events {
		x := float32(event.Pos.X*a.game.Config().TileSize + a.game.Config().TileSize/2)
		y := float32(event.Pos.Y*a.game.Config().TileSize + a.game.Config().TileSize/2)
		
		switch event.Type {
		case game.EventEatFood:
			a.effects.SpawnParticles(x, y, 10, color.RGBA{255, 100, 0, 255}, 2)
		case game.EventCollectKey:
			a.effects.SpawnParticles(x, y, 15, color.RGBA{255, 215, 0, 255}, 2.5)
		case game.EventCollectCoin:
			a.effects.SpawnParticles(x, y, 15, color.RGBA{255, 215, 0, 255}, 2.5)
		case game.EventOpenChest:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{255, 215, 0, 255}, 3)
		case game.EventEnemyKill:
			a.effects.SpawnParticles(x, y, 25, color.RGBA{128, 0, 128, 255}, 4)
			a.effects.TriggerShake(5, 20)
		case game.EventEnemyCollision:
			a.effects.SpawnParticles(x, y, 30, color.RGBA{128, 0, 128, 255}, 4)
			a.effects.TriggerShake(8, 25)
		case game.EventBombExplode:
			a.effects.SpawnParticles(x, y, 40, color.RGBA{255, 100, 0, 255}, 5)
			a.effects.TriggerShake(10, 30)
		case game.EventBombCollision:
			a.effects.TriggerShake(5, 20)
		case game.EventWallCollision:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{255, 100, 100, 255}, 3)
			a.effects.TriggerShake(5, 20)
		case game.EventSelfCollision:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{255, 100, 100, 255}, 3)
			a.effects.TriggerShake(5, 20)
			
		// Power-up события
		case game.EventPowerUpSlowMotion:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{0, 191, 255, 255}, 3)
			a.effects.TriggerShake(3, 15)
		case game.EventPowerUpShield:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{65, 105, 225, 255}, 3)
		case game.EventPowerUpShrink:
			a.effects.SpawnParticles(x, y, 15, color.RGBA{34, 139, 34, 255}, 2.5)
		case game.EventPowerUpExtraLife:
			a.effects.SpawnParticles(x, y, 25, color.RGBA{255, 0, 0, 255}, 4)
			a.effects.TriggerShake(5, 20)
		case game.EventPowerUpLightning:
			a.effects.SpawnParticles(x, y, 40, color.RGBA{255, 255, 0, 255}, 5)
			a.effects.TriggerShake(8, 25)
		case game.EventPowerUpMultiplier:
			a.effects.SpawnParticles(x, y, 20, color.RGBA{50, 205, 50, 255}, 3)
		}
	}

	// Обновление эффектов
	a.effects.Update()

	return nil
}

// Draw отрисовывает приложение
func (a *App) Draw(screen *ebiten.Image) {
	// Применяем тряску экрана
	dx, dy := a.effects.ScreenShake.GetOffset()

	// Рисуем градиентный фон
	screen.DrawImage(a.background, nil)

	// Создаём временную поверхность для игры
	gameScreen := ebiten.NewImage(screenWidth, screenHeight)

	switch a.game.State {
	case game.Menu:
		a.renderer.DrawMenu(gameScreen)
		a.effects.Draw(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return

	case game.SelectDifficulty:
		a.renderer.DrawDifficultySelection(gameScreen, a.game.Difficulty)
		a.effects.Draw(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return

	case game.Paused:
		a.renderer.DrawGame(gameScreen, a.game, a.effects)
		a.effects.Draw(gameScreen)
		a.renderer.DrawPauseOverlay(gameScreen)
		screen.DrawImage(gameScreen, nil)
		return

	case game.GameOver:
		a.renderer.DrawGame(gameScreen, a.game, a.effects)
		a.effects.Draw(gameScreen)
		a.renderer.DrawGameOverOverlay(gameScreen, a.game.Score, len(a.game.Enemies))
		screen.DrawImage(gameScreen, nil)
		return

	case game.Playing:
		a.renderer.DrawGame(gameScreen, a.game, a.effects)
		a.effects.Draw(gameScreen)
	}

	// Рисуем игровую поверхность со смещением тряски
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dx), float64(dy))
	screen.DrawImage(gameScreen, op)
}

// Layout возвращает размер экрана
func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Simple Snake - Go365 Go79")

	app := NewApp()

	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}
