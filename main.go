package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Food types constants
const (
	FoodApple = iota
	FoodCarrot
	FoodBanana
)

var foodImages []ebiten.Image
var foodSounds []*wav.Stream

// Game represents the game state
type Game struct {
	foodImages []ebiten.Image
	sounds     []*wav.Stream
	score      int
	player     image.Rectangle
}

// NewGame initializes a new game instance
func NewGame() (*Game, error) {
	g := &Game{}

	// Load food images
	appleImg, _, err := image.DecodeConfig(getResource("apple.png"))
	if err != nil {
		return nil, err
	}
	carrotImg, _, err := image.DecodeConfig(getResource("carrot.png"))
	if err != nil {
		return nil, err
	}
	bananaImg, _, err := image.DecodeConfig(getResource("banana.png"))
	if err != nil {
		return nil, err
	}

	// Create food images with different colors
	g.foodImages = []ebiten.Image{
		createFoodImage(color.RGBA{255, 165, 0, 255}), // Apple - orange
		createFoodImage(color.RGBA{128, 128, 0, 255}), // Carrot - yellow
		createFoodImage(color.RGBA{255, 0, 0, 255}),   // Banana - red
	}

	// Load audio effects for food collection
	g.sounds, err = loadFoodSounds()
	if err != nil {
		return nil, err
	}

	// Initialize player
	g.player = image.Rect(100, 180, 100+32, 180+32)
	g.score = 0

	return g, nil
}

// loadFoodSounds loads audio effects for food collection
func loadFoodSounds() ([]*wav.Stream, error) {
	// Initialize audio context
	audioContext, err := audio.NewContext(44100)
	if err != nil {
		return nil, err
	}

	// Create simple beep sounds for different food types
	sounds := make([]*wav.Stream, len(goodImages()))

	// In a real implementation, you would load actual WAV files:
	// sound1, err := wav.DecodeWithoutAllocation(bytes.NewReader(soundData1))
	// sound2, err := wav.DecodeWithoutAllocation(bytes.NewReader(soundData2))
	// sound3, err := wav.DecodeWithoutAllocation(bytes.NewReader(soundData3))

	// For now, we'll create placeholder sounds
	// You would replace this with actual audio file loading:
	/*
		carrotSound, err := wav.DecodeWithoutAllocation(bytes.NewReader(carrotSoundData))
		if err != nil {
			return nil, err
		}
		sounds[FoodCarrot] = audioContext.NewLazySample(carrotSound)
	*/

	return sounds, nil
}

// goodImages returns the number of food images
func goodImages() int {
	return len(foodImages)
}

// createFoodImage creates a simple colored rectangle as food representation
func createFoodImage(col color.Color) *ebiten.Image {
	img := ebiten.NewImage(32, 32)
	img.Fill(col)
	return img
}

// Update updates the game state
func (g *Game) Update() error {
	// Player movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		g.player.X -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		g.player.X += 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		g.player.Y -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		g.player.Y += 5
	}

	// Food collection detection
	playerRect := g.player
	for i := 0; i < goodImages(); i++ {
		foodRect := image.Rect(100+(i*60), 100, 100+(i*60)+32, 100+32)

		if playerRect.Overlaps(foodRect) {
			// Play audio effect when food is collected
			if g.sounds[i] != nil {
				// Play sound effect through audio context
				// context.PlaySound(g.sounds[i])
			}

			// Increase score
			g.score += 10

			// Replace collected food with empty image
			g.foodImages[i] = createFoodImage(color.Transparent)
		}
	}

	return nil
}

// Draw renders the game
func (g *Game) Draw(screen *ebiten.Image) {
	// Draw food items
	for i, foodImg := range g.foodImages {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(100+(i*60)), float64(100))
		screen.DrawImage(foodImg, opts)
	}

	// Draw player
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(g.player.Min.X), float64(g.player.Min.Y))
	// Draw player as blue rectangle
	playerImg := ebiten.NewImage(32, 32)
	playerImg.Fill(color.RGBA{0, 0, 255, 255})
	screen.DrawImage(playerImg, opts)

	// Draw score
	ebitenutil.DebugPrintAt(screen, "Score: "+string(g.score), 0, 40)
}

// Layout sets the screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 640, 480
}

// getResource simulates resource loading
func getResource(name string) []byte {
	// In real implementation, this would load actual files
	return []byte{0x00} // Placeholder
}

func main() {
	// Initialize game
	game, err := NewGame()
	if err != nil {
		panic(err)
	}

	// Set up Ebiten game window
	ebiten.SetWindowTitle("Food Collection Game with Audio Effects")
	ebiten.SetWindowSize(640, 480)

	// Run the game
	if err := ebiten.RunGame(game); err != nil {
		panic(err)
	}
}
