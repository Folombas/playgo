package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputHandler обрабатывает ввод от мыши и тачскрина
type InputHandler struct {
	boardOffsetX float64
	boardOffsetY float64
	windowWidth  int
	windowHeight int
}

// NewInputHandler создаёт обработчик ввода
func NewInputHandler(boardOffsetX, boardOffsetY float64, windowWidth, windowHeight int) *InputHandler {
	return &InputHandler{
		boardOffsetX: boardOffsetX,
		boardOffsetY: boardOffsetY,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,
	}
}

// GetClickPosition возвращает позицию клика/касания в координатах сетки
func (ih *InputHandler) GetClickPosition() *Position {
	// Проверяем мышь
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		return ih.screenToGrid(float64(mx), float64(my))
	}

	// Проверяем тач (для мобильных устройств)
	if touches := ebiten.TouchIDs(); len(touches) > 0 {
		// Берём первый тач
		touchID := touches[0]
		tx, ty := ebiten.TouchPosition(touchID)
		if tx != 0 || ty != 0 {
			return ih.screenToGrid(float64(tx), float64(ty))
		}
	}

	return nil
}

// screenToGrid преобразует экранные координаты в координаты сетки
func (ih *InputHandler) screenToGrid(screenX, screenY float64) *Position {
	// Вычитаем смещение доски
	x := screenX - ih.boardOffsetX
	y := screenY - ih.boardOffsetY

	// Проверяем, находится ли клик в пределах доски
	if x < 0 || x >= BoardPixel || y < 0 || y >= BoardPixel {
		return nil
	}

	// Преобразуем в координаты сетки
	col := int(x / CellSize)
	row := int(y / CellSize)

	// Проверяем границы
	if row < 0 || row >= BoardSize || col < 0 || col >= BoardSize {
		return nil
	}

	return &Position{Row: row, Col: col}
}

// IsResetPressed проверяет, нажата ли клавиша сброса (R)
func (ih *InputHandler) IsResetPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyR)
}

// IsPausePressed проверяет, нажата ли клавиша паузы (P)
func (ih *InputHandler) IsPausePressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP)
}

// IsClickOnNewGame проверяет, был ли клик на кнопке "Новая игра"
func (ih *InputHandler) IsClickOnNewGame(buttonX, buttonY, buttonW, buttonH float64) bool {
	// Проверяем мышь
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		x, y := float64(mx), float64(my)
		return x >= buttonX && x <= buttonX+buttonW && y >= buttonY && y <= buttonY+buttonH
	}

	// Проверяем тач
	if touches := ebiten.TouchIDs(); len(touches) > 0 {
		// Аналогично используем курсор как fallback
		mx, my := ebiten.CursorPosition()
		x, y := float64(mx), float64(my)
		return x >= buttonX && x <= buttonX+buttonW && y >= buttonY && y <= buttonY+buttonH
	}

	return false
}

// UpdateBoardOffset обновляет смещение доски (при изменении размера окна)
func (ih *InputHandler) UpdateBoardOffset(boardOffsetX, boardOffsetY float64) {
	ih.boardOffsetX = boardOffsetX
	ih.boardOffsetY = boardOffsetY
}
