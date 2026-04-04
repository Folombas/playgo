package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/playgo/bomberman_go/internal/config"
)

// Game - основная структура игры
type Game struct {
	scene Scene
}

// Scene - интерфейс для всех игровых сцен
type Scene interface {
	Update() error
	Draw(screen *ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (int, int)
}

// NewGame создает новый экземпляр Game
func NewGame() *Game {
	g := &Game{
		scene: NewMenuScene(),
	}
	return g
}

// Update обновляет состояние игры
func (g *Game) Update() error {
	if err := g.scene.Update(); err != nil {
		return err
	}
	return nil
}

// Draw отрисовывает игру
func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

// Layout возвращает размер экрана
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}

// SetScene переключает сцену
func (g *Game) SetScene(s Scene) {
	g.scene = s
}
