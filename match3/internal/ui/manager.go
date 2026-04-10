package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

// Цвета для UI
var (
	ColorBackground = color.RGBA{30, 30, 50, 255}
	ColorHUD        = color.RGBA{255, 255, 255, 255}
	ColorTitle      = color.RGBA{255, 215, 0, 255}
	ColorButton     = color.RGBA{100, 150, 255, 255}
	ColorScore      = color.RGBA{0, 255, 0, 255}
)

// Manager управляет отрисовкой UI элементов
type Manager struct{}

// NewManager создаёт новый менеджер UI
func NewManager() *Manager {
	return &Manager{}
}

// DrawMenu отрисовывает главное меню
func (m *Manager) DrawMenu(screen *ebiten.Image) {
	// Заголовок
	title := "MATCH-3"
	text.Draw(screen, title, basicfont.Face7x13, 280, 300, ColorTitle)

	subtitle := "Три в ряд"
	text.Draw(screen, subtitle, basicfont.Face7x13, 270, 350, ColorHUD)

	// Кнопка начала игры
	buttonText := "Нажмите ENTER для старта"
	text.Draw(screen, buttonText, basicfont.Face7x13, 180, 500, ColorButton)

	// Информация
	info := "Go365 Challenge - День 100"
	text.Draw(screen, info, basicfont.Face7x13, 200, 600, color.RGBA{150, 150, 150, 255})
}

// DrawHUD отрисовывает интерфейс во время игры
func (m *Manager) DrawHUD(screen *ebiten.Image, score int, moves int) {
	// Фон HUD
	hudBg := ebiten.NewImage(640, 80)
	hudBg.Fill(color.RGBA{40, 40, 60, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 0)
	screen.DrawImage(hudBg, op)

	// Счёт
	scoreText := fmt.Sprintf("Счёт: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 20, 30, ColorScore)

	// Ходы
	movesText := fmt.Sprintf("Ходы: %d", moves)
	text.Draw(screen, movesText, basicfont.Face7x13, 20, 50, ColorHUD)

	// Подсказка
	hint := "Клик - выбор, Shift+Клик - обмен"
	text.Draw(screen, hint, basicfont.Face7x13, 150, 70, color.RGBA{150, 150, 150, 255})

	// Кнопка выхода
	exitText := "ESC - меню"
	text.Draw(screen, exitText, basicfont.Face7x13, 520, 30, color.RGBA{255, 100, 100, 255})
}

// DrawGameOver отрисовывает экран окончания игры
func (m *Manager) DrawGameOver(screen *ebiten.Image, score int) {
	// Фон
	overlay := ebiten.NewImage(640, 960)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	screen.DrawImage(overlay, nil)

	// Game Over текст
	gameOverText := "GAME OVER"
	text.Draw(screen, gameOverText, basicfont.Face7x13, 260, 400, color.RGBA{255, 50, 50, 255})

	// Итоговый счёт
	scoreText := fmt.Sprintf("Итоговый счёт: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 240, 450, ColorScore)

	// Кнопка рестарта
	restartText := "Нажмите ENTER для рестарта"
	text.Draw(screen, restartText, basicfont.Face7x13, 180, 550, ColorButton)
}
