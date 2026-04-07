package ui

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// UI управляет пользовательским интерфейсом
type UI struct {
	// Пока без шрифтов, просто рисуем прямоугольники
}

// NewUI создает новый UI
func NewUI() *UI {
	return &UI{}
}

// Draw рисует UI
func (ui *UI) Draw(screen *ebiten.Image, score int, combo int) {
	// Панель счета слева
	ui.drawScorePanel(screen, score, combo)
	
	// Заголовок
	ui.drawTitle(screen)
	
	// Подсказка внизу
	ui.drawHint(screen)
}

// drawScorePanel рисует панель счета
func (ui *UI) drawScorePanel(screen *ebiten.Image, score int, combo int) {
	// Фон панели
	panelWidth := 350
	panelHeight := 200
	panel := ebiten.NewImage(panelWidth, panelHeight)
	panel.Fill(color.RGBA{30, 30, 60, 230})
	
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(20, 20)
	screen.DrawImage(panel, op)
	
	// Рамка сверху
	border := ebiten.NewImage(panelWidth, 4)
	border.Fill(color.RGBA{100, 150, 255, 255})
	
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(20, 24)
	screen.DrawImage(border, borderOp)
	
	// Счет - просто рисуем прямоугольники вместо текста
	// Очки
	scoreWidth := float64(200)
	scoreBar := ebiten.NewImage(int(scoreWidth), 30)
	scoreBar.Fill(color.RGBA{60, 60, 120, 255})
	
	scoreOp := &ebiten.DrawImageOptions{}
	scoreOp.GeoM.Translate(40, 60)
	screen.DrawImage(scoreBar, scoreOp)
	
	// Комбо индикатор
	if combo > 1 {
		comboWidth := float64(150 * combo)
		if comboWidth > 300 {
			comboWidth = 300
		}
		comboBar := ebiten.NewImage(int(comboWidth), 20)
		comboBar.Fill(color.RGBA{255, 255, 0, 200})
		
		comboOp := &ebiten.DrawImageOptions{}
		comboOp.GeoM.Translate(40, 100)
		screen.DrawImage(comboBar, comboOp)
	}
	
	// Уровень
	level := score/1000 + 1
	levelWidth := float64(50 * level)
	if levelWidth > 300 {
		levelWidth = 300
	}
	levelBar := ebiten.NewImage(int(levelWidth), 20)
	levelBar.Fill(color.RGBA{144, 238, 144, 200})
	
	levelOp := &ebiten.DrawImageOptions{}
	levelOp.GeoM.Translate(40, 140)
	screen.DrawImage(levelBar, levelOp)
}

// drawTitle рисует заголовок
func (ui *UI) drawTitle(screen *ebiten.Image) {
	// Заголовок - красивый прямоугольник
	titleWidth := 500
	titleHeight := 60
	title := ebiten.NewImage(titleWidth, titleHeight)
	title.Fill(color.RGBA{40, 40, 80, 180})
	
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(390, 20)
	screen.DrawImage(title, opts)
	
	// Рамка заголовка
	border := ebiten.NewImage(titleWidth, 3)
	border.Fill(color.RGBA{100, 200, 255, 255})
	
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(390, 20)
	screen.DrawImage(border, borderOp)
	
	// Подзаголовок
	sub := ebiten.NewImage(400, 30)
	sub.Fill(color.RGBA{60, 60, 100, 150})
	
	subOp := &ebiten.DrawImageOptions{}
	subOp.GeoM.Translate(440, 85)
	screen.DrawImage(sub, subOp)
}

// drawHint рисует подсказку
func (ui *UI) drawHint(screen *ebiten.Image) {
	// Панель подсказок
	hintWidth := 600
	hintHeight := 40
	hint := ebiten.NewImage(hintWidth, hintHeight)
	hint.Fill(color.RGBA{30, 30, 50, 200})
	
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(340, 660)
	screen.DrawImage(hint, opts)
	
	// Маленькая рамка
	border := ebiten.NewImage(hintWidth, 2)
	border.Fill(color.RGBA{150, 150, 200, 200})
	
	borderOp := &ebiten.DrawImageOptions{}
	borderOp.GeoM.Translate(340, 660)
	screen.DrawImage(border, borderOp)
	
	// Вывод счета и комбо текстом (пока цифрами)
	scoreText := fmt.Sprintf("Score")
	_ = scoreText // Пока не используем
}
