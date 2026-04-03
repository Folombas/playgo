package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Action represents a game input action.
type Action int

const (
	ActionMoveLeft Action = iota
	ActionMoveRight
	ActionJump
	ActionPause
	ActionRestart
)

// Manager handles input state.
type Manager struct {
	justPressed map[Action]bool
	pressed     map[Action]bool
}

// NewManager creates a new input manager.
func NewManager() *Manager {
	return &Manager{
		justPressed: make(map[Action]bool),
		pressed:     make(map[Action]bool),
	}
}

// Update polls input state. Call once per frame.
func (m *Manager) Update() {
	// Reset justPressed
	for k := range m.justPressed {
		m.justPressed[k] = false
	}

	// Move Left
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyLeft) {
		m.pressed[ActionMoveLeft] = true
	} else {
		m.pressed[ActionMoveLeft] = false
	}

	// Move Right
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyRight) {
		m.pressed[ActionMoveRight] = true
	} else {
		m.pressed[ActionMoveRight] = false
	}

	// Jump
	if inpututil.IsKeyJustPressed(ebiten.KeyW) ||
		inpututil.IsKeyJustPressed(ebiten.KeyUp) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.justPressed[ActionJump] = true
	}
	m.pressed[ActionJump] = ebiten.IsKeyPressed(ebiten.KeyW) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeySpace)

	// Pause
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		m.justPressed[ActionPause] = true
	}

	// Restart
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		m.justPressed[ActionRestart] = true
	}
}

// IsJustPressed returns true if action was just pressed this frame.
func (m *Manager) IsJustPressed(a Action) bool {
	return m.justPressed[a]
}

// IsPressed returns true if action is currently held.
func (m *Manager) IsPressed(a Action) bool {
	return m.pressed[a]
}
