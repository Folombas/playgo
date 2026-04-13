package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputProcessor обрабатывает ввод (мышь/тач/клавиатура)
type InputProcessor struct {
	boardOffsetX float64 // Смещение игрового поля по X
	boardOffsetY float64 // Смещение игрового поля по Y
	cellSize     float64 // Размер ячейки в пикселях
}

// NewInputProcessor создаёт новый обработчик ввода
func NewInputProcessor(offsetX, offsetY, cellSize float64) *InputProcessor {
	return &InputProcessor{
		boardOffsetX: offsetX,
		boardOffsetY: offsetY,
		cellSize:     cellSize,
	}
}

// Update обновляет смещения (на случай изменения размера окна)
func (ip *InputProcessor) Update(offsetX, offsetY, cellSize float64) {
	ip.boardOffsetX = offsetX
	ip.boardOffsetY = offsetY
	ip.cellSize = cellSize
}

// GetClickPosition преобразует экранные координаты в позицию на поле
func (ip *InputProcessor) GetClickPosition(screenX, screenY float64) (int, int, bool) {
	col := int((screenX - ip.boardOffsetX) / ip.cellSize)
	row := int((screenY - ip.boardOffsetY) / ip.cellSize)

	if row >= 0 && row < BoardSize && col >= 0 && col < BoardSize {
		return row, col, true
	}
	return -1, -1, false
}

// IsInsideButton проверяет находится ли точка внутри кнопки
func IsInsideButton(x, y, btnX, btnY, btnW, btnH float64) bool {
	return x >= btnX && x <= btnX+btnW && y >= btnY && y <= btnY+btnH
}

// HandleInput обрабатывает весь ввод за один кадр
// Возвращает информацию о действии игрока
type InputAction struct {
	Type          ActionType
	Row1, Col1    int // Первая позиция
	Row2, Col2    int // Вторая позиция (для swap)
	NewGame       bool // Запрос новой игры
	Pause         bool // Запрос паузы
}

type ActionType int

const (
	ActionNone ActionType = iota
	ActionSelect          // Выбор фишки
	ActionSwap            // Попытка обмена
)

// ProcessMouseInput обрабатывает ввод мыши
func (ip *InputProcessor) ProcessMouseInput() InputAction {
	// Проверяем клик мыши
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		return ip.processClick(float64(x), float64(y))
	}
	return InputAction{Type: ActionNone}
}

// ProcessTouchInput обрабатывает тач-ввод (мобильные устройства)
func (ip *InputProcessor) ProcessTouchInput() InputAction {
	// Проверяем все активные тачи
	touchIDs := ebiten.TouchIDs()
	for _, id := range touchIDs {
		if inpututil.IsTouchJustPressed(id) {
			x, y := ebiten.TouchPosition(id)
			return ip.processClick(float64(x), float64(y))
		}
	}
	return InputAction{Type: ActionNone}
}

// ProcessKeyboardInput обрабатывает клавиатуру
func ProcessKeyboardInput() InputAction {
	// R - новая игра
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		return InputAction{Type: ActionNone, NewGame: true}
	}
	// P - пауза
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		return InputAction{Type: ActionNone, Pause: true}
	}
	return InputAction{Type: ActionNone}
}

// processClick обрабатывает клик по полю
func (ip *InputProcessor) processClick(x, y float64) InputAction {
	row, col, ok := ip.GetClickPosition(x, y)
	if ok {
		return InputAction{
			Type: ActionSelect,
			Row1: row,
			Col1: col,
		}
	}
	return InputAction{Type: ActionNone}
}
