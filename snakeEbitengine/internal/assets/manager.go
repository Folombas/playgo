package assets

import (
    "fmt"
    "image/color"
    "log"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

    "snake/internal/constants"
    "snake/internal/entities"
    "snake/internal/types"
)

type AssetManager struct {
    // Images
    Apple, Strawberry, Orange, Banana, Pineapple *ebiten.Image
    GhostFrames      []*ebiten.Image
    RoachFrames      []*ebiten.Image
    VikingFrames     []*ebiten.Image
    GiftClosedImgs   []*ebiten.Image
    GiftOpenFrames   []*ebiten.Image
    CoinFrames       []*ebiten.Image
    KeyImg           *ebiten.Image
}

// New creates a new AssetManager and loads all assets. Errors are logged and placeholder images are used when loading fails.
func New() *AssetManager {
    am := &AssetManager{}
    am.loadImages()
    return am
}

func (am *AssetManager) loadImages() {
	// Helper that loads an image (PNG or JPEG) and returns a placeholder on error.
	loadImage := func(path string) *ebiten.Image {
		img, _, err := ebitenutil.NewImageFromFile(path)
		if err != nil {
			log.Printf("failed to load %s: %v, using placeholder", path, err)
			img = ebiten.NewImage(types.TileSize, types.TileSize)
			img.Fill(color.White)
		}
		return img
	}


    // Fruits
    am.Apple = loadImage("assets/fruits/apple.png")
    am.Strawberry = loadImage("assets/fruits/strawberry.png")
    am.Orange = loadImage("assets/fruits/orange.png")
    am.Banana = loadImage("assets/fruits/banana.png")
    am.Pineapple = loadImage("assets/fruits/pineapple.png")

    // Ghost animation frames (11 frames)
    am.GhostFrames = make([]*ebiten.Image, constants.GhostFrames)
    for i := 0; i < constants.GhostFrames; i++ {
        filename := fmt.Sprintf("assets/ghost/skeleton-animation_%02d.png", i)
        am.GhostFrames[i] = loadImage(filename)
    }

    // Roach sprite sheet (auto‑slice)
    if frames, err := entities.LoadSpriteSheet("assets/roach/roach.png", constants.RoachFramesX, constants.RoachFramesY, constants.RoachCols, constants.RoachRows, false, nil); err != nil {
        log.Printf("roach sprite sheet load error: %v", err)
    } else {
        am.RoachFrames = frames
    }

    // Viking sprite sheet (auto‑slice)
    if frames, err := entities.LoadSpriteSheetAuto("assets/vikings/2204_w053_n004_9_medicharacters_p1_9.jpg", constants.VikingCols, constants.VikingRows, true, color.White); err != nil {
        log.Printf("viking sprite sheet load error: %v", err)
    } else {
        am.VikingFrames = frames
    }

    // Gifts (closed and open)
    letters := []string{"a", "b", "c", "d", "e", "f"}
    am.GiftClosedImgs = make([]*ebiten.Image, constants.GiftCount)
    for i := 0; i < constants.GiftCount; i++ {
        fname := fmt.Sprintf("assets/gifts/gift_01%s.png", letters[i])
        am.GiftClosedImgs[i] = loadImage(fname)
    }
    am.GiftOpenFrames = make([]*ebiten.Image, constants.GiftCount)
    for i := 0; i < constants.GiftCount; i++ {
        fname := fmt.Sprintf("assets/gifts/giftopen_01%s.png", letters[i])
        am.GiftOpenFrames[i] = loadImage(fname)
    }

    // Coins
    am.CoinFrames = make([]*ebiten.Image, constants.CoinFrames)
    for i := 0; i < constants.CoinFrames; i++ {
        fname := fmt.Sprintf("assets/coins/coin_%02da.png", i+1)
        am.CoinFrames[i] = loadImage(fname)
    }

    // Key image
    am.KeyImg = loadImage("assets/keys/key_02d.png")
}

// Helper imports not present in this snippet – they are added via go.mod imports.
