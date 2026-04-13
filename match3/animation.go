package main

import (
	"math"
	"time"
)

// AnimationType defines the type of animation.
type AnimationType int

const (
	AnimSwap      AnimationType = iota // Swapping two tiles
	AnimShake                            // Invalid swap shake
	AnimRemove                           // Tile disappearing
	AnimFall                             // Tile falling into place
	AnimPulse                            // Hint pulsing
)

// Animation represents a single animation in progress.
type Animation struct {
	Type      AnimationType
	Tiles     []*Tile
	Targets   [][2]int // target positions for swap
	Progress  float64  // 0.0 to 1.0
	Duration  time.Duration
	StartTime time.Time
	Done      bool

	// Swap-specific
	SwapA *Tile
	SwapB *Tile

	// Shake-specific
	ShakeOffset float64

	// Fall-specific
	FallFromRow int
	FallToRow   int
}

// NewAnimation creates a new animation with the given parameters.
func NewAnimation(typ AnimationType, tiles []*Tile, duration time.Duration) *Animation {
	return &Animation{
		Type:      typ,
		Tiles:     tiles,
		Duration:  duration,
		StartTime: time.Now(),
		Progress:  0,
		Done:      false,
	}
}

// NewSwapAnimation creates a swap animation between two tiles.
func NewSwapAnimation(a, b *Tile, duration time.Duration) *Animation {
	anim := &Animation{
		Type:      AnimSwap,
		Tiles:     []*Tile{a, b},
		Duration:  duration,
		StartTime: time.Now(),
		SwapA:     a,
		SwapB:     b,
	}
	return anim
}

// NewShakeAnimation creates a shake animation for invalid swap.
func NewShakeAnimation(tiles []*Tile, duration time.Duration) *Animation {
	return &Animation{
		Type:      AnimShake,
		Tiles:     tiles,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

// NewRemoveAnimation creates a remove (fade+shrink) animation.
func NewRemoveAnimation(tiles []*Tile, duration time.Duration) *Animation {
	return &Animation{
		Type:      AnimRemove,
		Tiles:     tiles,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

// NewFallAnimation creates a fall animation for tiles.
func NewFallAnimation(tiles []*Tile, duration time.Duration) *Animation {
	return &Animation{
		Type:      AnimFall,
		Tiles:     tiles,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

// NewPulseAnimation creates a pulsing hint animation.
func NewPulseAnimation(tiles []*Tile, duration time.Duration) *Animation {
	return &Animation{
		Type:      AnimPulse,
		Tiles:     tiles,
		Duration:  duration,
		StartTime: time.Now(),
	}
}

// Update advances the animation based on elapsed time. Returns true if finished.
func (a *Animation) Update() bool {
	elapsed := time.Since(a.StartTime)
	a.Progress = math.Min(float64(elapsed)/float64(a.Duration), 1.0)

	if a.Progress >= 1.0 {
		a.Done = true
	}

	switch a.Type {
	case AnimSwap:
		a.updateSwap()
	case AnimShake:
		a.updateShake()
	case AnimRemove:
		a.updateRemove()
	case AnimFall:
		a.updateFall()
	case AnimPulse:
		a.updatePulse()
	}

	return a.Done
}

// updateSwap interpolates positions between two tiles.
func (a *Animation) updateSwap() {
	t := easeInOutCubic(a.Progress)

	// SwapA moves from its original to SwapB's position
	a.SwapA.X = a.SwapA.X + (a.SwapB.X-a.SwapA.X)*t // simplified — actual handled in game
	a.SwapB.X = a.SwapB.X + (a.SwapA.X-a.SwapB.X)*t
}

// updateShake oscillates tiles left and right.
func (a *Animation) updateShake() {
	// 3 cycles at 50ms each = 150ms total, ±4 pixels
	cycle := a.Progress * 3.0
	offset := math.Sin(cycle*math.Pi*2) * 4.0
	a.ShakeOffset = offset
}

// updateRemove shrinks and fades tiles.
func (a *Animation) updateRemove() {
	t := a.Progress
	for _, tile := range a.Tiles {
		tile.Scale = 1.0 - t
		tile.Alpha = 1.0 - t
	}
}

// updateFall moves tiles from their start position to target.
func (a *Animation) updateFall() {
	// Fall animation handled in game.Update by setting Y based on progress
}

// updatePulse oscillates tile scale for hint.
func (a *Animation) updatePulse() {
	t := a.Progress
	pulse := 1.0 + math.Sin(t*math.Pi*4)*0.1
	for _, tile := range a.Tiles {
		tile.Scale = pulse
	}
}

// easeInOutCubic provides smooth easing.
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// AnimationManager manages a queue of animations.
type AnimationManager struct {
	Current   *Animation
	IsPlaying bool
}

// Start begins an animation.
func (m *AnimationManager) Start(anim *Animation) {
	m.Current = anim
	m.IsPlaying = true
}

// Update advances the current animation. Returns true when all animations are done.
func (m *AnimationManager) Update() bool {
	if !m.IsPlaying || m.Current == nil {
		m.IsPlaying = false
		return true
	}

	done := m.Current.Update()
	if done {
		m.IsPlaying = false
		m.Current = nil
	}
	return !m.IsPlaying
}

// IsAnimating returns true if any animation is in progress.
func (m *AnimationManager) IsAnimating() bool {
	return m.IsPlaying
}
