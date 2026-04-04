package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"github.com/playgo/bomberman_go/internal/config"
)

// MenuScene - сцена главного меню
type MenuScene struct {
	titleFont font.Face
	hudFont   font.Face
}

// NewMenuScene создает сцену меню
func NewMenuScene() *MenuScene {
	return &MenuScene{
		titleFont: basicfont.Face7x13, // Используем встроенный шрифт с увеличением
		hudFont:   basicfont.Face7x13,
	}
}

// Update обновляет меню
func (m *MenuScene) Update() error {
	// Переход в игру по нажатию Enter
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		// TODO: Переключаемся на игровую сцену
		return nil
	}
	return nil
}

// Draw рисует меню
func (m *MenuScene) Draw(screen *ebiten.Image) {
	// Фон
	screen.Fill(color.Black)

	// Заголовок - простой встроенный шрифт
	text.Draw(screen, "BOMBERMAN GO", m.titleFont, config.ScreenWidth/2-100, config.ScreenHeight/2-50, color.White)

	// Подсказка
	green := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	
	text.Draw(screen, "Press ENTER or SPACE to start", m.hudFont, config.ScreenWidth/2-120, config.ScreenHeight/2+50, green)
	text.Draw(screen, "Arrow Keys / WASD - Move", m.hudFont, config.ScreenWidth/2-90, config.ScreenHeight/2+100, color.White)
	text.Draw(screen, "SPACE - Place Bomb", m.hudFont, config.ScreenWidth/2-70, config.ScreenHeight/2+130, color.White)

	// FPS счетчик
	text.Draw(screen, "R-Restart", m.hudFont, 10, config.ScreenHeight-20, color.White)
}

// Layout возвращает размер экрана
func (m *MenuScene) Layout(outsideWidth, outsideHeight int) (int, int) {
	return config.ScreenWidth, config.ScreenHeight
}
