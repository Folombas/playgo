package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Manager handles input polling.
type Manager struct{}

// NewManager creates a new input manager.
func NewManager() *Manager {
	return &Manager{}
}

// JustJump returns true if jump key was just pressed.
func (m *Manager) JustJump() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsKeyJustPressed(ebiten.KeyUp)
}

// JustEnter returns true if Enter was just pressed.
func (m *Manager) JustEnter() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter)
}

// JustEscape returns true if Escape was just pressed.
func (m *Manager) JustEscape() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}
