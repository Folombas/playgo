package ui

import (
	"fmt"
	"image/color"
	"math"

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
type Manager struct {
	menuAnimTime float64 // Время для анимации меню
}

// NewManager создаёт новый менеджер UI
func NewManager() *Manager {
	return &Manager{
		menuAnimTime: 0,
	}
}

// UpdateMenuAnim обновляет анимацию меню
func (m *Manager) UpdateMenuAnim(deltaTime float64) {
	m.menuAnimTime += deltaTime
}

// DrawMenu отрисовывает главное меню
func (m *Manager) DrawMenu(screen *ebiten.Image) {
	// Анимация пульсации для заголовка
	pulse := 1.0 + 0.05*math.Sin(m.menuAnimTime*3)
	
	// Заголовок с эффектом градиента
	title := "MATCH-3"
	titleX := 280
	titleY := 300
	
	// Тень заголовка
	text.Draw(screen, title, basicfont.Face7x13, titleX+2, titleY+2, color.RGBA{0, 0, 0, 128})
	// Основной текст
	text.Draw(screen, title, basicfont.Face7x13, titleX, titleY, ColorTitle)
	
	// Подзаголовок с пульсацией
	subtitleAlpha := 0.7 + 0.3*math.Sin(m.menuAnimTime*2)
	subtitle := "Три в ряд"
	subY := 350
	_ = pulse // используем для альфа-эффекта
	_ = subtitleAlpha
	
	text.Draw(screen, subtitle, basicfont.Face7x13, 270, subY, color.RGBA{255, 255, 255, uint8(255 * subtitleAlpha)})
	
	// Анимированная кнопка начала игры
	buttonPulse := 1.0 + 0.1*math.Sin(m.menuAnimTime*4)
	buttonText := ">>> Нажмите ENTER для старта <<<"
	buttonY := 500
	_ = buttonPulse
	
	// Мигающий текст кнопки
	if math.Sin(m.menuAnimTime*5) > 0 {
		text.Draw(screen, buttonText, basicfont.Face7x13, 170, buttonY, color.RGBA{100, 200, 255, 255})
	} else {
		text.Draw(screen, buttonText, basicfont.Face7x13, 170, buttonY, color.RGBA{200, 255, 255, 255})
	}

	// Информация
	info := "Go365 Challenge - День 101"
	text.Draw(screen, info, basicfont.Face7x13, 200, 600, color.RGBA{150, 150, 150, 255})
	
	// Дополнительная информация
	controls := "Управление: Клик мыши / Shift+Стрелки"
	text.Draw(screen, controls, basicfont.Face7x13, 170, 650, color.RGBA{120, 120, 120, 255})
}

// DrawHUD отрисовывает интерфейс во время игры
func (m *Manager) DrawHUD(screen *ebiten.Image, score int, moves int, timeLeft int, targetScore int) {
	// Фон HUD
	hudBg := ebiten.NewImage(640, 80)
	hudBg.Fill(color.RGBA{40, 40, 60, 255})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, 0)
	screen.DrawImage(hudBg, op)

	// Счёт
	scoreText := fmt.Sprintf("Счёт: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 20, 30, ColorScore)
	
	// Прогресс-бар к целевому счёту
	if targetScore > 0 {
		progress := float64(score) / float64(targetScore)
		if progress > 1.0 {
			progress = 1.0
		}
		
		barWidth := 200
		barHeight := 12
		barX := 150
		barY := 25
		
		// Фон прогресс-бара
		bgBar := ebiten.NewImage(barWidth, barHeight)
		bgBar.Fill(color.RGBA{60, 60, 80, 255})
		bgOp := &ebiten.DrawImageOptions{}
		bgOp.GeoM.Translate(float64(barX), float64(barY))
		screen.DrawImage(bgBar, bgOp)
		
		// Заполнение прогресса
		if progress > 0 {
			fillWidth := int(float64(barWidth) * progress)
			fillBar := ebiten.NewImage(fillWidth, barHeight)
			
			// Цвет зависит от прогресса
			var fillColor color.Color
			if progress < 0.33 {
				fillColor = color.RGBA{255, 100, 100, 255} // Красный
			} else if progress < 0.66 {
				fillColor = color.RGBA{255, 255, 100, 255} // Жёлтый
			} else {
				fillColor = color.RGBA{100, 255, 100, 255} // Зелёный
			}
			
			fillBar.Fill(fillColor)
			fillOp := &ebiten.DrawImageOptions{}
			fillOp.GeoM.Translate(float64(barX), float64(barY))
			screen.DrawImage(fillBar, fillOp)
		}
		
		// Текст процента
		percentText := fmt.Sprintf("%d%%", int(progress*100))
		text.Draw(screen, percentText, basicfont.Face7x13, barX+barWidth+10, barY+10, ColorHUD)
	}

	// Ходы
	movesText := fmt.Sprintf("Ходы: %d", moves)
	text.Draw(screen, movesText, basicfont.Face7x13, 20, 50, ColorHUD)
	
	// Таймер (если есть лимит времени)
	if timeLeft > 0 {
		timerColor := ColorHUD
		// Мигание при < 10 секунд
		if timeLeft < 10 {
			// Простая пульсация через остаток от деления
			if timeLeft%2 == 0 {
				timerColor = color.RGBA{255, 50, 50, 255} // Красный
			}
		}
		
		minutes := timeLeft / 60
		seconds := timeLeft % 60
		timerText := fmt.Sprintf("⏱ %d:%02d", minutes, seconds)
		text.Draw(screen, timerText, basicfont.Face7x13, 280, 30, timerColor)
	}

	// Подсказка
	hint := "Клик - выбор, Shift+Клик - обмен"
	text.Draw(screen, hint, basicfont.Face7x13, 150, 70, color.RGBA{150, 150, 150, 255})

	// Кнопка выхода
	exitText := "ESC - меню"
	text.Draw(screen, exitText, basicfont.Face7x13, 520, 30, color.RGBA{255, 100, 100, 255})
}

// DrawGameOver отрисовывает экран окончания игры с расширенной статистикой
func (m *Manager) DrawGameOver(screen *ebiten.Image, score int, moves int, level int, combo int) {
	// Фон
	overlay := ebiten.NewImage(640, 960)
	overlay.Fill(color.RGBA{0, 0, 0, 200})
	screen.DrawImage(overlay, nil)

	// Game Over текст с тенью
	gameOverText := "GAME OVER"
	text.Draw(screen, gameOverText, basicfont.Face7x13, 262, 202, color.RGBA{0, 0, 0, 128})
	text.Draw(screen, gameOverText, basicfont.Face7x13, 260, 200, color.RGBA{255, 50, 50, 255})

	// Статистика игры
	statsY := 260
	statLineHeight := 25
	
	// Уровень
	levelText := fmt.Sprintf("Уровень: %d", level)
	text.Draw(screen, levelText, basicfont.Face7x13, 250, statsY, ColorHUD)
	
	// Итоговый счёт
	scoreText := fmt.Sprintf("Итоговый счёт: %d", score)
	text.Draw(screen, scoreText, basicfont.Face7x13, 240, statsY+statLineHeight, ColorScore)
	
	// Количество ходов
	movesText := fmt.Sprintf("Сделано ходов: %d", moves)
	text.Draw(screen, movesText, basicfont.Face7x13, 240, statsY+statLineHeight*2, ColorHUD)
	
	// Максимальное комбо
	comboText := fmt.Sprintf("Лучшее комбо: x%d", combo)
	text.Draw(screen, comboText, basicfont.Face7x13, 240, statsY+statLineHeight*3, color.RGBA{255, 215, 0, 255})
	
	// Разделитель
	dividerY := statsY + statLineHeight*4 + 10
	divider := ebiten.NewImage(300, 2)
	divider.Fill(color.RGBA{150, 150, 150, 255})
	divOp := &ebiten.DrawImageOptions{}
	divOp.GeoM.Translate(170, float64(dividerY))
	screen.DrawImage(divider, divOp)

	// Кнопка рестарта (мигающая)
	restartY := dividerY + 40
	if math.Sin(float64(statsY)*0.1) > 0 {
		restartText := ">>> Нажмите ENTER для рестарта <<<"
		text.Draw(screen, restartText, basicfont.Face7x13, 170, restartY, ColorButton)
	} else {
		restartText := "    Нажмите ENTER для рестарта    "
		text.Draw(screen, restartText, basicfont.Face7x13, 170, restartY, ColorButton)
	}
	
	// Кнопка выхода
	exitText := "ESC - вернуться в меню"
	text.Draw(screen, exitText, basicfont.Face7x13, 210, restartY+40, color.RGBA{150, 150, 150, 255})
}

// DrawPaused отрисовывает экран паузы
func (m *Manager) DrawPaused(screen *ebiten.Image) {
	// Полупрозрачный фон
	overlay := ebiten.NewImage(640, 960)
	overlay.Fill(color.RGBA{0, 0, 0, 150})
	screen.DrawImage(overlay, nil)

	// Заголовок
	pausedText := "ПАУЗА"
	text.Draw(screen, pausedText, basicfont.Face7x13, 270, 400, ColorTitle)

	// Инструкции
	resumeText := "ESC/P - продолжить"
	text.Draw(screen, resumeText, basicfont.Face7x13, 220, 450, ColorHUD)

	settingsText := "S - настройки"
	text.Draw(screen, settingsText, basicfont.Face7x13, 240, 480, ColorHUD)

	quitText := "Q - выйти в меню"
	text.Draw(screen, quitText, basicfont.Face7x13, 230, 510, color.RGBA{255, 100, 100, 255})
}

// DrawSettings отрисовывает экран настроек
func (m *Manager) DrawSettings(screen *ebiten.Image, soundManager interface{}) {
	// Фон
	overlay := ebiten.NewImage(640, 960)
	overlay.Fill(color.RGBA{20, 20, 40, 255})
	screen.DrawImage(overlay, nil)

	// Заголовок
	settingsText := "НАСТРОЙКИ"
	text.Draw(screen, settingsText, basicfont.Face7x13, 250, 200, ColorTitle)

	// Настройки звука
	volText := "Громкость: 50%"
	text.Draw(screen, volText, basicfont.Face7x13, 220, 300, ColorHUD)

	muteText := "M - вкл/выкл звук"
	text.Draw(screen, muteText, basicfont.Face7x13, 210, 350, ColorButton)

	// Назад
	backText := "ESC - назад"
	text.Draw(screen, backText, basicfont.Face7x13, 240, 500, color.RGBA{150, 150, 150, 255})
}

// DrawLevelComplete отрисовывает экран завершения уровня
func (m *Manager) DrawLevelComplete(screen *ebiten.Image, level int) {
	// Фон
	overlay := ebiten.NewImage(640, 960)
	overlay.Fill(color.RGBA{0, 50, 0, 200})
	screen.DrawImage(overlay, nil)

	// Заголовок
	levelText := fmt.Sprintf("УРОВЕНЬ %d ПРОЙДЕН!", level)
	text.Draw(screen, levelText, basicfont.Face7x13, 180, 400, ColorTitle)

	// Звёзды
	starsText := "★★★"
	text.Draw(screen, starsText, basicfont.Face7x13, 260, 450, color.RGBA{255, 215, 0, 255})

	// Продолжить
	contText := "ENTER - следующий уровень"
	text.Draw(screen, contText, basicfont.Face7x13, 170, 550, ColorButton)
}
