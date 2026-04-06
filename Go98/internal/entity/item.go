package entity

import (
	"math"

	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/helper"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// ItemType defines different item types
type ItemType int

const (
	ItemCoin ItemType = iota
	ItemGem
	ItemKey
	ItemPotion
	ItemChest // Closed chest that gives random loot
)

// Item represents a pickup item
type Item struct {
	Base
	Type      ItemType
	Value     int
	BobOffset float64
	BobTimer  float64
	Collected bool
}

func NewItem(x, y int, itemType ItemType, value int) *Item {
	return &Item{
		Base: Base{
			X:      float64(x*config.TileSize) + 4,
			Y:      float64(y*config.TileSize) + 4,
			Width:  24,
			Height: 24,
			Active: true,
		},
		Type:  itemType,
		Value: value,
	}
}

func (i *Item) Update() error {
	if !i.Active {
		return nil
	}

	// Bobbing animation
	i.BobTimer += 0.05
	i.BobOffset = math.Sin(i.BobTimer) * 3

	return nil
}

func (i *Item) Draw(screen *ebiten.Image, offsetX, offsetY float64) {
	if !i.Active || i.Collected {
		return
	}

	if i.Image == nil {
		// Fallback: colored shape
		var c color.Color
		switch i.Type {
		case ItemCoin:
			c = color.RGBA{255, 215, 0, 255}
		case ItemGem:
			c = color.RGBA{0, 200, 255, 255}
		case ItemKey:
			c = color.RGBA{255, 255, 0, 255}
		case ItemPotion:
			c = color.RGBA{255, 0, 128, 255}
		case ItemChest:
			c = color.RGBA{153, 102, 51, 255}
		}

		helper.DrawRect(screen, i.X-offsetX, i.Y-offsetY+i.BobOffset, 24, 24, c)
	} else {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(i.X-offsetX, i.Y-offsetY+i.BobOffset)
		screen.DrawImage(i.Image, op)
	}
}

// Collect marks item as collected and returns its value
func (i *Item) Collect() int {
	if i.Collected {
		return 0
	}
	i.Collected = true
	i.Active = false
	return i.Value
}
