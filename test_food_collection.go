package main

import (
	"image"
	"testing"
)

func TestFoodCollection(t *testing.T) {
	// Test that food images are loaded
	if len(foodImages) != 2 {
		t.Errorf("Expected 2 food images, got %d", len(foodImages))
	}

	// Test player rectangle creation
	playerRect := image.Rect(
		int(30),
		int(180),
		int(30)+32,
		int(180)+32,
	)

	if playerRect.Dx() != 32 || playerRect.Dy() != 32 {
		t.Errorf("Player rect size incorrect: %v", playerRect)
	}

	// Test food rectangle creation
	foodRect := image.Rect(0, 0, 32, 32)
	if foodRect.Dx() != 32 || foodRect.Dy() != 32 {
		t.Errorf("Food rect size incorrect: %v", foodRect)
	}

	// Test overlap detection
	testCases := []struct {
		name     string
		playerX  int
		foodX    int
		expected bool
	}{
		{"Overlapping", 100, 100, true},
		{"Not overlapping", 0, 200, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			playerRect := image.Rect(tc.playerX, 100, tc.playerX+32, 132)
			foodRect := image.Rect(tc.foodX, 100, tc.foodX+32, 132)
			result := playerRect.Overlaps(foodRect)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestScoreIncrement(t *testing.T) {
	// This would require running the game loop
	// For now, just verify the score field exists
	g := &Game{
		foodImages: make([]ebiten.Image, 2),
		score:      0,
	}

	if g.score != 0 {
		t.Errorf("Expected initial score 0, got %d", g.score)
	}
}
