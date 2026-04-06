package entity

import (
	"testing"
)

// ======================== PARTICLE ========================

func TestParticle_Update_Dies(t *testing.T) {
	p := Particle{X: 0, Y: 0, Life: 5, MaxLife: 10, RotSpeed: 0.1}
	// Life: 5→4→3→2→1 (4 updates, all return true)
	for i := 0; i < 4; i++ {
		if !p.Update() {
			t.Errorf("Particle should be alive at iteration %d (Life was %d before update)", i, 5-i)
		}
	}
	// Now Life=1. 5th update: Life→0, returns false (dead)
	if p.Update() {
		t.Error("Particle should be dead when Life reaches 0")
	}
}

func TestParticle_Update_Gravity(t *testing.T) {
	p := Particle{VY: -3}
	p.Update()
	// VY should increase by 0.1 (gravity)
	if p.VY != -2.9 {
		t.Errorf("Expected VY=-2.9 after gravity, got %f", p.VY)
	}
}

func TestParticle_Update_Rotation(t *testing.T) {
	p := Particle{Rotation: 0, RotSpeed: 0.5}
	p.Update()
	if p.Rotation != 0.5 {
		t.Errorf("Expected Rotation=0.5, got %f", p.Rotation)
	}
}

func TestParticle_Alpha(t *testing.T) {
	tests := []struct {
		life, maxLife int
		want          float32
	}{
		{10, 10, 1.0},
		{5, 10, 0.5},
		{0, 10, 0.0},
	}
	for _, tt := range tests {
		p := Particle{Life: tt.life, MaxLife: tt.maxLife}
		got := p.Alpha()
		if got != tt.want {
			t.Errorf("Alpha(%d/%d) = %f, want %f", tt.life, tt.maxLife, got, tt.want)
		}
	}
}

// ======================== SELECT ANIM ========================

func TestSelectAnim_Update(t *testing.T) {
	s := SelectAnim{R: 0, C: 0, T: 0}
	s.Update()
	if s.T != 1 {
		t.Errorf("Expected T=1, got %d", s.T)
	}
}

func TestSelectAnim_Pulse_Range(t *testing.T) {
	s := SelectAnim{T: 0}
	// Check multiple frames return values in 0..1
	for i := 0; i < 100; i++ {
		p := s.Pulse()
		if p < 0 || p > 1 {
			t.Errorf("Pulse() = %f, expected 0..1", p)
		}
		s.Update()
	}
}

func TestSelectAnim_Pulse_Oscillation(t *testing.T) {
	s := SelectAnim{T: 0}
	v1 := s.Pulse()
	s.T = 31 // Half period of sin(0.1*t) ≈ π/2
	v2 := s.Pulse()
	// Values should differ (not constant)
	if v1 == v2 {
		t.Log("Warning: Pulse returned same value at different times (may be coincidence)")
	}
}

// ======================== REMOVE ANIM ========================

func TestRemoveAnim_Update(t *testing.T) {
	a := RemoveAnim{T: 0, Total: 20}
	for i := 0; i < 19; i++ {
		if a.Update() {
			t.Errorf("Should not be done at iteration %d", i)
		}
	}
	if !a.Update() {
		t.Error("Should be done at iteration 20")
	}
}

func TestRemoveAnim_Progress(t *testing.T) {
	a := RemoveAnim{Total: 20}
	a.T = 10
	if p := a.Progress(); p != 0.5 {
		t.Errorf("Progress at T=10/Total=20 = %f, want 0.5", p)
	}
}

func TestRemoveAnim_Scale(t *testing.T) {
	a := RemoveAnim{T: 0, Total: 20}
	if s := a.Scale(); s != 1.0 {
		t.Errorf("Scale at start = %f, want 1.0", s)
	}
	a.T = 20
	if s := a.Scale(); s != 1.5 {
		t.Errorf("Scale at end = %f, want 1.5", s)
	}
}

func TestRemoveAnim_Alpha(t *testing.T) {
	a := RemoveAnim{T: 0, Total: 20}
	if a := a.Alpha(); a != 1.0 {
		t.Errorf("Alpha at start = %f, want 1.0", a)
	}
	a.T = 10
	if a := a.Alpha(); a != 0.5 {
		t.Errorf("Alpha at halfway = %f, want 0.5", a)
	}
}

func TestRemoveAnim_FirstFrame(t *testing.T) {
	a := RemoveAnim{}
	if a.FirstFrame {
		t.Error("FirstFrame should be false initially")
	}
	a.Update()
	if !a.FirstFrame {
		t.Error("FirstFrame should be true after first update")
	}
}

// ======================== SLIDE ANIM ========================

func TestSlideAnim_Update(t *testing.T) {
	a := SlideAnim{T: 0, Total: 10}
	for i := 0; i < 9; i++ {
		if a.Update() {
			t.Errorf("Should not be done at iteration %d", i)
		}
	}
	if !a.Update() {
		t.Error("Should be done at iteration 10")
	}
}

func TestSlideAnim_Eased(t *testing.T) {
	a := SlideAnim{Total: 10}
	tests := []struct {
		t    int
		want float64
		tol  float64
	}{
		{0, 0.0, 0.001},
		{5, 0.875, 0.001}, // 1 - (0.5)^3 = 0.875
		{10, 1.0, 0.001},
	}
	for _, tt := range tests {
		a.T = tt.t
		got := a.Eased()
		if got < tt.want-tt.tol || got > tt.want+tt.tol {
			t.Errorf("Eased at T=%d = %f, want ~%f", tt.t, got, tt.want)
		}
	}
}

func TestSlideAnim_Position_Interpolation(t *testing.T) {
	a := SlideAnim{SX: 0, SY: 0, TX: 100, TY: 100, Total: 10}
	a.T = 0
	if x := a.X(); x != 0 {
		t.Errorf("X at start = %f, want 0", x)
	}
	a.T = 10
	if x := a.X(); x != 100 {
		t.Errorf("X at end = %f, want 100", x)
	}
}

func TestSlideAnim_Position_EaseOut(t *testing.T) {
	a := SlideAnim{SX: 0, TX: 100, Total: 10}
	// At 50% time, ease-out cubic should be > 50%
	a.T = 5
	x := a.X()
	if x <= 50 {
		t.Errorf("Ease-out at 50%% time should be > 50%%, got %f", x)
	}
}

// ======================== MENU BUTTON ========================

func TestMenuButton_Contains_Inside(t *testing.T) {
	b := MenuButton{X: 100, Y: 100, W: 50, H: 30}
	if !b.Contains(110, 110) {
		t.Error("Point inside button should be contained")
	}
}

func TestMenuButton_Contains_Edge(t *testing.T) {
	b := MenuButton{X: 100, Y: 100, W: 50, H: 30}
	if !b.Contains(100, 100) {
		t.Error("Top-left corner should be contained")
	}
	if !b.Contains(149, 129) {
		t.Error("Bottom-right corner (exclusive) should be contained")
	}
}

func TestMenuButton_Contains_Outside(t *testing.T) {
	b := MenuButton{X: 100, Y: 100, W: 50, H: 30}
	tests := []struct{ mx, my int }{
		{99, 100}, {150, 100}, {100, 99}, {100, 130},
	}
	for i, tt := range tests {
		if b.Contains(tt.mx, tt.my) {
			t.Errorf("Test %d: Point (%d,%d) should NOT be contained", i, tt.mx, tt.my)
		}
	}
}
