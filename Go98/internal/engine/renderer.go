package engine

import (
	"image/color"
	"math"

	"dungeon_crawler/internal/config"
	"dungeon_crawler/internal/dungeon"
	"dungeon_crawler/internal/entity"
	"dungeon_crawler/internal/helper"

	"github.com/hajimehoshi/ebiten/v2"
)

// Camera follows the player
type Camera struct {
	X, Y float64
}

func (c *Camera) Follow(target *entity.Player) {
	// Smooth follow
	targetX := target.X + target.Width/2 - config.ScreenWidth/2
	targetY := target.Y + target.Height/2 - config.ScreenHeight/2

	c.X += (targetX - c.X) * 0.1
	c.Y += (targetY - c.Y) * 0.1

	// Clamp to dungeon bounds
	maxX := float64(config.MapWidth*config.TileSize) - config.ScreenWidth
	maxY := float64(config.MapHeight*config.TileSize) - config.ScreenHeight

	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if c.X > maxX {
		c.X = maxX
	}
	if c.Y > maxY {
		c.Y = maxY
	}
}

// Renderer handles all drawing
type Renderer struct {
	Camera Camera
}

func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderDungeon draws the dungeon tiles
func (r *Renderer) RenderDungeon(screen *ebiten.Image, d *dungeon.Dungeon) {
	tileSize := float64(config.TileSize)

	// Calculate visible tile range
	startX := int(math.Max(0, float64(int(r.Camera.X/tileSize)-1)))
	startY := int(math.Max(0, float64(int(r.Camera.Y/tileSize)-1)))
	endX := int(math.Min(float64(config.MapWidth), float64(startX+config.ViewWidth+2)))
	endY := int(math.Min(float64(config.MapHeight), float64(startY+config.ViewHeight+2)))

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			tile := d.GetTile(x, y)
			screenX := float64(x)*tileSize - r.Camera.X
			screenY := float64(y)*tileSize - r.Camera.Y

			// Skip if off screen
			if screenX < -tileSize || screenX > config.ScreenWidth ||
				screenY < -tileSize || screenY > config.ScreenHeight {
				continue
			}

			switch tile {
			case config.TileFloor:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{60, 60, 60, 255})
				// Floor detail
				helper.DrawRect(screen, screenX+1, screenY+1, tileSize-2, tileSize-2, color.RGBA{70, 70, 70, 255})
			case config.TileWall:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{100, 100, 120, 255})
				// Wall shading
				helper.DrawRect(screen, screenX, screenY, tileSize, 4, color.RGBA{80, 80, 100, 255})
				helper.DrawRect(screen, screenX, screenY, 4, tileSize, color.RGBA{90, 90, 110, 255})
			case config.TileDoor:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{139, 90, 43, 255})
				helper.DrawRect(screen, screenX+4, screenY+4, tileSize-8, tileSize-8, color.RGBA{160, 110, 60, 255})
			case config.TileStairs:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{200, 200, 100, 255})
				// Stairs symbol
				helper.DrawRect(screen, screenX+8, screenY+8, 16, 16, color.RGBA{255, 255, 200, 255})
				helper.DrawRect(screen, screenX+12, screenY+12, 8, 8, color.RGBA{255, 255, 255, 255})
			case config.TileSpikes:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{60, 60, 60, 255})
				// Spikes
				for i := 0; i < 4; i++ {
					helper.DrawRect(screen, screenX+float64(i*8)+2, screenY+16, 4, 12, color.RGBA{180, 180, 180, 255})
				}
			case config.TileWater:
				helper.DrawRect(screen, screenX, screenY, tileSize, tileSize, color.RGBA{30, 100, 180, 255})
				// Water animation
				helper.DrawRect(screen, screenX+4, screenY+8, 8, 2, color.RGBA{100, 180, 255, 255})
			}
		}
	}
}

// RenderPlayer draws the player
func (r *Renderer) RenderPlayer(screen *ebiten.Image, player *entity.Player) {
	player.Draw(screen, r.Camera.X, r.Camera.Y)
}

// RenderEnemies draws all enemies
func (r *Renderer) RenderEnemies(screen *ebiten.Image, enemies []*entity.Enemy) {
	for _, enemy := range enemies {
		enemy.Draw(screen, r.Camera.X, r.Camera.Y)
	}
}

// RenderItems draws all items
func (r *Renderer) RenderItems(screen *ebiten.Image, items []*entity.Item) {
	for _, item := range items {
		item.Draw(screen, r.Camera.X, r.Camera.Y)
	}
}

// RenderDamageNumbers draws floating damage numbers
func (r *Renderer) RenderDamageNumbers(screen *ebiten.Image, damages []*entity.DamageNumber) {
	for _, dmg := range damages {
		if !dmg.IsActive() {
			continue
		}

		helper.DrawRect(screen, dmg.X-r.Camera.X, dmg.Y-r.Camera.Y, 20, 20, dmg.Color)
	}
}
