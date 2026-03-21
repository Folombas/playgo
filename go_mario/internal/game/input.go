package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState - состояние ввода
type InputState struct {
	Left      bool
	Right     bool
	Up        bool
	Down      bool
	Jump      bool
	Action    bool
	Menu      bool
	Pause     bool
	Save      bool
	Load      bool
	Mine      bool
	Place     bool
	Inventory int
}

// GetInput получает текущее состояние ввода
func GetInput() *InputState {
	input := &InputState{}

	// Movement
	input.Left = ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	input.Right = ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)
	input.Up = ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW)
	input.Down = ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS)

	// Jump
	input.Jump = ebiten.IsKeyPressed(ebiten.KeySpace) || input.Up

	// Action
	input.Action = ebiten.IsKeyPressed(ebiten.KeyE) || ebiten.IsKeyPressed(ebiten.KeyEnter)

	// Menu/Pause
	input.Menu = inpututil.IsKeyJustPressed(ebiten.KeyEscape)
	input.Pause = inpututil.IsKeyJustPressed(ebiten.KeyP)

	// Save/Load
	input.Save = inpututil.IsKeyJustPressed(ebiten.KeyF5)
	input.Load = inpututil.IsKeyJustPressed(ebiten.KeyF9)

	// Mining/Placing
	input.Mine = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	input.Place = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)

	// Inventory slots
	for i := 0; i < inventorySize; i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(i + 49)) {
			input.Inventory = i
			break
		}
	}

	// Mouse wheel for inventory
	_, wheelY := ebiten.Wheel()
	if wheelY > 0 {
		input.Inventory--
		if input.Inventory < 0 {
			input.Inventory = inventorySize - 1
		}
	} else if wheelY < 0 {
		input.Inventory++
		if input.Inventory >= inventorySize {
			input.Inventory = 0
		}
	}

	return input
}

// IsMoving проверяет, движется ли игрок
func (i *InputState) IsMoving() bool {
	return i.Left || i.Right || i.Up || i.Down
}

// JustPressedAction проверяет, нажата ли кнопка действия только что
func (i *InputState) JustPressedAction() bool {
	return i.Action || i.Menu
}
