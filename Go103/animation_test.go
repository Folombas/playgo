package main

import (
	"testing"
	"time"
)

func TestNewAnimation(t *testing.T) {
	anim := NewAnimation(AnimSwap, nil, 100*time.Millisecond)
	if anim.Type != AnimSwap {
		t.Errorf("Expected AnimSwap, got %v", anim.Type)
	}
	if anim.Done {
		t.Error("New animation should not be done")
	}
}

func TestNewSwapAnimation(t *testing.T) {
	t1 := &Tile{Row: 0, Col: 0}
	t2 := &Tile{Row: 0, Col: 1}
	anim := NewSwapAnimation(t1, t2, 150*time.Millisecond)

	if anim.Type != AnimSwap {
		t.Errorf("Expected AnimSwap, got %v", anim.Type)
	}
	if len(anim.Tiles) != 2 {
		t.Errorf("Expected 2 tiles, got %d", len(anim.Tiles))
	}
}

func TestNewShakeAnimation(t *testing.T) {
	tiles := []*Tile{{}, {}}
	anim := NewShakeAnimation(tiles, 150*time.Millisecond)

	if anim.Type != AnimShake {
		t.Errorf("Expected AnimShake, got %v", anim.Type)
	}
}

func TestNewRemoveAnimation(t *testing.T) {
	tiles := []*Tile{{}, {}, {}}
	anim := NewRemoveAnimation(tiles, 150*time.Millisecond)

	if anim.Type != AnimRemove {
		t.Errorf("Expected AnimRemove, got %v", anim.Type)
	}
}

func TestAnimationManager(t *testing.T) {
	mgr := AnimationManager{}
	if mgr.IsAnimating() {
		t.Error("New manager should not be animating")
	}

	anim := NewAnimation(AnimPulse, nil, 50*time.Millisecond)
	mgr.Start(anim)

	if !mgr.IsAnimating() {
		t.Error("Manager should be animating after Start")
	}

	// Wait for animation to complete
	time.Sleep(60 * time.Millisecond)
	mgr.Update()

	if mgr.IsAnimating() {
		t.Error("Animation should be done after duration")
	}
}

func TestEaseInOutCubic(t *testing.T) {
	tests := []struct {
		input    float64
		expected float64
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
	}
	for _, tc := range tests {
		result := easeInOutCubic(tc.input)
		if result < tc.expected-0.01 || result > tc.expected+0.01 {
			t.Errorf("easeInOutCubic(%f) = %f, expected ~%f",
				tc.input, result, tc.expected)
		}
	}
}
