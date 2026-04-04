package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/playgo/bomberman_go/internal/config"
	"github.com/playgo/bomberman_go/internal/entity"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// GameplayScene - основная игровая сцена
type GameplayScene struct {
	grid    *entity.Grid
	player  *entity.Player
	enemies []*entity.Enemy
	bombs   []*entity.Bomb
	explos  []*entity.Explosion

	// Загруженные спрайты
	sprites map[string]*ebiten.Image
	
	// Шрифт для HUD
	hudFont font.Face
}

// NewGameplayScene создает игровую сцену
func NewGameplayScene() *GameplayScene {
	g := &GameplayScene{
		sprites: make(map[string]*ebiten.Image),
		hudFont: basicfont.Face7x13,
	}
	g.loadAssets()
	g.initGame()
	
	return g
}

// loadAssets загружает все спрайты
func (g *GameplayScene) loadAssets() {
	// Загружаем спрайты из файлов
	spriteFiles := map[string]string{
		"player_stand":  "player_stand.png",
		"player_walk1":  "player_walk1.png",
		"player_walk2":  "player_walk2.png",
		"bomb":          "bomb.png",
		"brick":         "brick.png",
		"stone":         "stone.png",
		"grass":         "grass.png",
		"enemy1":        "enemy1.png",
		"enemy2":        "enemy2.png",
		"explosion":     "explosion.png",
		"heart":         "heart.png",
	}

	for name, file := range spriteFiles {
		img, _, err := ebitenutil.NewImageFromFile(config.AssetPathSprites + file)
		if err != nil {
			fmt.Printf("Warning: could not load sprite %s: %v\n", file, err)
			continue
		}
		g.sprites[name] = img
	}
}

// initGame инициализирует игровое состояние
func (g *GameplayScene) initGame() {
	// Создаем игровое поле
	g.grid = entity.NewGrid()

	// Создаем игрока
	g.player = entity.NewPlayer(1, 1)

	// Создаем врагов
	g.enemies = entity.SpawnEnemies(3, g.grid)

	// Инициализируем массивы
	g.bombs = make([]*entity.Bomb, 0)
	g.explos = make([]*entity.Explosion, 0)
}

// Update обновляет игровое состояние
func (g *GameplayScene) Update() error {
	// Обновляем игрока
	g.player.Update(g.grid, g.bombs)

	// Обновляем врагов
	for _, enemy := range g.enemies {
		enemy.Update(g.grid)
	}

	// Обновляем бомбы
	for i := len(g.bombs) - 1; i >= 0; i-- {
		if g.bombs[i].Update() {
			// Бомба взорвалась - создаем взрыв
			explos := entity.NewExplosion(g.bombs[i].X, g.bombs[i].Y, g.player.ExplosionRadius, g.grid)
			g.explos = append(g.explos, explos)
			g.bombs = append(g.bombs[:i], g.bombs[i+1:]...)
		}
	}

	// Обновляем взрывы
	for i := len(g.explos) - 1; i >= 0; i-- {
		if g.explos[i].Update() {
			g.explos = append(g.explos[:i], g.explos[i+1:]...)
		}
	}

	// Проверяем коллизии врагов с взрывами
	for _, enemy := range g.enemies {
		for _, explos := range g.explos {
			enemyGridX := int(enemy.X / float64(config.TileSize))
			enemyGridY := int(enemy.Y / float64(config.TileSize))
			if explos.Contains(enemyGridX, enemyGridY) {
				enemy.Kill()
			}
		}
	}

	// Проверяем коллизии игрока с врагами
	for _, enemy := range g.enemies {
		if !enemy.IsDead() && g.player.CollidesWith(enemy) {
			g.player.TakeDamage()
		}
	}

	// Проверяем коллизии игрока с взрывами
	for _, explos := range g.explos {
		if g.player.IsInExplosion(explos) {
			g.player.TakeDamage()
		}
	}

	// Проверяем ввод для установки бомбы
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.player.TryPlaceBomb(&g.bombs)
	}

	// Перезапуск по R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.initGame()
	}

	return nil
}

// Draw отрисовывает игровую сцену
func (g *GameplayScene) Draw(screen *ebiten.Image) {
	// Очищаем экран
	screen.Fill(color.RGBA{R: 50, G: 150, B: 50, A: 255})

	// Рисуем сетку
	g.grid.Draw(screen, g.sprites)

	// Рисуем бомбы
	for _, bomb := range g.bombs {
		bomb.Draw(screen, g.sprites["bomb"])
	}

	// Рисуем взрывы
	for _, explos := range g.explos {
		explos.Draw(screen, g.sprites["explosion"])
	}

	// Рисуем врагов
	for _, enemy := range g.enemies {
		if !enemy.IsDead() {
			enemy.Draw(screen, g.sprites["enemy1"], g.sprites["enemy2"])
		}
	}

	// Рисуем игрока
	g.player.Draw(screen, g.sprites["player_stand"], g.sprites["player_walk1"], g.sprites["player_walk2"])

	// Рисуем HUD
	g.drawHUD(screen)
}

// drawHUD рисует интерфейс
func (g *GameplayScene) drawHUD(screen *ebiten.Image) {
	// Фон HUD
	hudBg := ebiten.NewImage(config.ScreenWidth, 40)
	hudBg.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 180})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 0)
	screen.DrawImage(hudBg, op)

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}

	// Жизни
	livesText := fmt.Sprintf("Lives: %d", g.player.Lives)
	text.Draw(screen, livesText, g.hudFont, 20, 28, white)

	// Бомбы
	bombsText := fmt.Sprintf("Bombs: %d/%d", g.player.ActiveBombs, g.player.MaxBombs)
	text.Draw(screen, bombsText, g.hudFont, 150, 28, white)

	// Радиус взрыва
	radiusText := fmt.Sprintf("Fire: %d", g.player.ExplosionRadius)
	text.Draw(screen, radiusText, g.hudFont, 300, 28, white)

	// Враги alive
	aliveCount := 0
	for _, e := range g.enemies {
		if !e.IsDead() {
			aliveCount++
		}
	}
	enemiesText := fmt.Sprintf("Enemies: %d", aliveCount)
	text.Draw(screen, enemiesText, g.hudFont, 450, 28, white)

	// Управление
	text.Draw(screen, "R-Restart SPACE-Bomb WASD-Move", g.hudFont, 550, 28, green)
}

// Layout возвращает размер экрана
func (g *GameplayScene) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
