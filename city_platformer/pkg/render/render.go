// Package render - система отрисовки PlatformerComplete
// Go365 Day 88
package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/playgo/city_platformer/pkg/entity"
	"github.com/playgo/city_platformer/pkg/level"
	"github.com/playgo/city_platformer/pkg/sprite"
)

// Renderer - система отрисовки
type Renderer struct {
	spriteSheet *sprite.SpriteSheet
	tileWidth   int
	tileHeight  int
}

// NewRenderer - создание рендерера
func NewRenderer(ss *sprite.SpriteSheet) *Renderer {
	tileW, tileH := ss.GetTileSize()
	return &Renderer{
		spriteSheet: ss,
		tileWidth:   tileW,
		tileHeight:  tileH,
	}
}

// DrawPlayer - отрисовка игрока
func (r *Renderer) DrawPlayer(screen *ebiten.Image, p entity.Player, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	// Мигание если неуязвим
	if p.Invincible > 0 && (p.Invincible%6 < 3) {
		return
	}
	
	// Определяем состояние
	state := "stand"
	if p.IsCrouching {
		state = "duck"
	} else if !p.OnGround {
		state = "jump"
	} else if p.IsMoving {
		state = "walk"
	}
	
	frame := int(p.AnimFrame)
	playerSprite := r.spriteSheet.GetPlayerFrame(state, frame)
	
	if playerSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		
		// Отражение по горизонтали
		if p.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}
		
		opts.GeoM.Translate(screenX, screenY)
		
		screen.DrawImage(playerSprite, opts)
	} else {
		r.drawPlayerFallback(screen, p, cameraX, cameraY)
	}
}

// drawPlayerFallback - резервная отрисовка
func (r *Renderer) drawPlayerFallback(screen *ebiten.Image, p entity.Player, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	c := color.RGBA{100, 200, 100, 255}
	if p.IsCrouching {
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY+20), 32, 20, c, true)
	} else {
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 32, 40, c, true)
	}
}

// DrawEnemy - отрисовка врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	frame := int(e.AnimFrame)
	enemySprite := r.spriteSheet.GetEnemyFrame(e.Type, frame)
	
	if enemySprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(enemySprite, opts)
	} else {
		r.drawEnemyFallback(screen, e, cameraX, cameraY)
	}
}

// drawEnemyFallback - резервная отрисовка врага
func (r *Renderer) drawEnemyFallback(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	var c color.RGBA
	switch e.Type {
	case "slime":
		c = color.RGBA{150, 50, 150, 255}
	case "fly":
		c = color.RGBA{200, 200, 50, 255}
	case "snail":
		c = color.RGBA{180, 100, 50, 255}
	case "fish":
		c = color.RGBA{50, 100, 200, 255}
	default:
		c = color.RGBA{150, 150, 150, 255}
	}
	
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(e.Width), float32(e.Height), c, true)
}

// DrawBoss - отрисовка босса
func (r *Renderer) DrawBoss(screen *ebiten.Image, b entity.Boss, cameraX, cameraY float64) {
	screenX := b.X - cameraX
	screenY := b.Y - cameraY
	
	// Босс - большой слизень
	bossColor := color.RGBA{180, 50, 180, 255}
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(b.Width), float32(b.Height), bossColor, true)
	
	// Глаза
	vector.DrawFilledRect(screen, float32(screenX+15), float32(screenY+15), 20, 20, color.RGBA{0, 0, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+45), float32(screenY+15), 20, 20, color.RGBA{0, 0, 0, 255}, true)
	
	// Зрачки
	vector.DrawFilledRect(screen, float32(screenX+22), float32(screenY+20), 10, 10, color.RGBA{255, 255, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+52), float32(screenY+20), 10, 10, color.RGBA{255, 255, 255, 255}, true)
}

// DrawPlatform - отрисовка платформы
func (r *Renderer) DrawPlatform(screen *ebiten.Image, p level.Platform, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	switch p.Type {
	case "grass":
		r.drawGrassPlatform(screen, screenX, screenY, p.Width, p.Height)
	case "dirt":
		r.drawDirtPlatform(screen, screenX, screenY, p.Width, p.Height)
	case "stone":
		r.drawStonePlatform(screen, screenX, screenY, p.Width, p.Height)
	case "castle":
		r.drawCastlePlatform(screen, screenX, screenY, p.Width, p.Height)
	case "box":
		r.drawBoxPlatform(screen, screenX, screenY, p.Width, p.Height)
	default:
		r.drawGrassPlatform(screen, screenX, screenY, p.Width, p.Height)
	}
}

// drawGrassPlatform - травяная платформа
func (r *Renderer) drawGrassPlatform(screen *ebiten.Image, x, y, width, height float64) {
	tileW := float64(r.tileWidth)
	tileH := float64(r.tileHeight)
	
	// Верхний ряд
	for tx := x; tx < x+width; tx += tileW {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(tx, y)
		
		// Выбираем подходящий тайл
		var tile *ebiten.Image
		if tx == x {
			tile = r.spriteSheet.GetTile("grassHalfLeft")
		} else if tx+tileW >= x+width {
			tile = r.spriteSheet.GetTile("grassHalfRight")
		} else {
			tile = r.spriteSheet.GetTile("grassHalf")
		}
		
		if tile == nil {
			tile = r.spriteSheet.GetTile("grassMid")
		}
		
		if tile != nil {
			screen.DrawImage(tile, opts)
		}
	}
	
	// Заполняем середину
	for ty := y + tileH; ty < y+height; ty += tileH {
		for tx := x; tx < x+width; tx += tileW {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(tx, ty)
			
			var tile *ebiten.Image
			if tx == x {
				tile = r.spriteSheet.GetTile("dirtLeft")
			} else if tx+tileW >= x+width {
				tile = r.spriteSheet.GetTile("dirtRight")
			} else {
				tile = r.spriteSheet.GetTile("dirtCenter")
			}
			
			if tile == nil {
				tile = r.spriteSheet.GetTile("dirtMid")
			}
			
			if tile != nil {
				screen.DrawImage(tile, opts)
			}
		}
	}
}

// drawDirtPlatform - земляная платформа
func (r *Renderer) drawDirtPlatform(screen *ebiten.Image, x, y, width, height float64) {
	r.drawTiledPlatform(screen, x, y, width, height, "dirtMid", "dirtLeft", "dirtRight", "dirtCenter")
}

// drawStonePlatform - каменная платформа
func (r *Renderer) drawStonePlatform(screen *ebiten.Image, x, y, width, height float64) {
	r.drawTiledPlatform(screen, x, y, width, height, "stoneMid", "stoneLeft", "stoneRight", "stoneCenter")
}

// drawCastlePlatform - платформа замка
func (r *Renderer) drawCastlePlatform(screen *ebiten.Image, x, y, width, height float64) {
	r.drawTiledPlatform(screen, x, y, width, height, "castleMid", "castleLeft", "castleRight", "castleCenter")
}

// drawBoxPlatform - платформа из ящиков
func (r *Renderer) drawBoxPlatform(screen *ebiten.Image, x, y, width, height float64) {
	tile := r.spriteSheet.GetTile("box")
	if tile == nil {
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{180, 100, 50, 255}, true)
		return
	}
	
	tileW := float64(tile.Bounds().Dx())
	tileH := float64(tile.Bounds().Dy())
	
	for ty := y; ty < y+height; ty += tileH {
		for tx := x; tx < x+width; tx += tileW {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(tx, ty)
			screen.DrawImage(tile, opts)
		}
	}
}

// drawTiledPlatform - общая функция отрисовки тайлами
func (r *Renderer) drawTiledPlatform(screen *ebiten.Image, x, y, width, height float64, midTile, leftTile, rightTile, centerTile string) {
	tileW := float64(r.tileWidth)
	tileH := float64(r.tileHeight)
	
	for ty := y; ty < y+height; ty += tileH {
		for tx := x; tx < x+width; tx += tileW {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(tx, ty)
			
			var tile *ebiten.Image
			if tx == x {
				tile = r.spriteSheet.GetTile(leftTile)
			} else if tx+tileW >= x+width {
				tile = r.spriteSheet.GetTile(rightTile)
			} else {
				tile = r.spriteSheet.GetTile(centerTile)
			}
			
			if tile == nil {
				tile = r.spriteSheet.GetTile(midTile)
			}
			
			if tile != nil {
				screen.DrawImage(tile, opts)
			}
		}
	}
}

// DrawCollectible - отрисовка предмета
func (r *Renderer) DrawCollectible(screen *ebiten.Image, c level.Collectible, cameraX, cameraY float64) {
	if c.Collected {
		return
	}
	
	screenX := c.X - cameraX
	screenY := c.Y - cameraY
	
	var itemSprite *ebiten.Image
	
	switch c.Type {
	case "coin", "coin_gold":
		itemSprite = r.spriteSheet.GetCoin("gold")
	case "coin_silver":
		itemSprite = r.spriteSheet.GetCoin("silver")
	case "coin_bronze":
		itemSprite = r.spriteSheet.GetCoin("bronze")
	case "gem_red":
		itemSprite = r.spriteSheet.GetGem("red")
	case "gem_blue":
		itemSprite = r.spriteSheet.GetGem("blue")
	case "gem_green":
		itemSprite = r.spriteSheet.GetGem("green")
	case "gem_yellow":
		itemSprite = r.spriteSheet.GetGem("yellow")
	case "star":
		itemSprite = r.spriteSheet.Star
	default:
		itemSprite = r.spriteSheet.GetCoin("gold")
	}
	
	if itemSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(itemSprite, opts)
	}
}

// DrawProjectile - отрисовка снаряда
func (r *Renderer) DrawProjectile(screen *ebiten.Image, p entity.Projectile, cameraX, cameraY float64) {
	if !p.Active {
		return
	}
	
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	if p.IsEnemy {
		// Вражеский снаряд - фиолетовый шар
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 8, color.RGBA{180, 50, 180, 255}, true)
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 4, color.RGBA{255, 150, 255, 255}, true)
	} else {
		// Пуля игрока - оранжевый шар
		vector.DrawFilledCircle(screen, float32(screenX)+4, float32(screenY)+4, 6, color.RGBA{255, 150, 50, 255}, true)
		vector.DrawFilledCircle(screen, float32(screenX)+4, float32(screenY)+4, 3, color.RGBA{255, 255, 200, 255}, true)
	}
}

// DrawBackground - отрисовка фона
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	// Небо - градиент
	for y := 0; y < screen.Bounds().Dy(); y++ {
		ratio := float64(y) / float64(screen.Bounds().Dy())
		r := uint8(135 - ratio*35)
		g := uint8(206 - ratio*86)
		b := uint8(235)
		vector.DrawFilledRect(screen, 0, float32(y), float32(screen.Bounds().Dx()), 1, color.RGBA{r, g, b, 255}, true)
	}
	
	// Облака (параллакс)
	r.drawClouds(screen, cameraX*0.2, 50)
	r.drawClouds(screen, cameraX*0.4, 150)
}

// drawClouds - отрисовка облаков
func (r *Renderer) drawClouds(screen *ebiten.Image, offsetX float64, baseY float64) {
	clouds := []string{"cloud1", "cloud2", "cloud3"}
	
	for i := 0; i < 8; i++ {
		cloudType := clouds[i%len(clouds)]
		cloud := r.spriteSheet.GetDecoration(cloudType)
		
		if cloud != nil {
			x := float64(i*200) - offsetX
			y := baseY + float64((i%3)*30)
			
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(x, y)
			screen.DrawImage(cloud, opts)
		}
	}
}

// DrawDecoration - отрисовка декораций
func (r *Renderer) DrawDecoration(screen *ebiten.Image, decType string, x, y, cameraX, cameraY float64) {
	dec := r.spriteSheet.GetDecoration(decType)
	
	if dec != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(x-cameraX, y-cameraY)
		screen.DrawImage(dec, opts)
	}
}

// DrawHUD - отрисовка интерфейса
func (r *Renderer) DrawHUD(screen *ebiten.Image, score, lives, level, wave, combo int) {
	// Фон HUD
	vector.DrawFilledRect(screen, 10, 10, 250, 100, color.RGBA{0, 0, 0, 128}, true)
	vector.DrawFilledRect(screen, 10, 10, 250, 100, color.RGBA{255, 255, 255, 64}, false)
}

// DrawBossHealthBar - полоска здоровья босса
func (r *Renderer) DrawBossHealthBar(screen *ebiten.Image, health, maxHealth int) {
	barWidth := 400
	barHeight := 20
	x := float32(screen.Bounds().Dx())/2 - float32(barWidth)/2
	y := float32(50)
	
	// Фон
	vector.DrawFilledRect(screen, x, y, float32(barWidth), float32(barHeight), color.RGBA{50, 50, 50, 255}, true)
	
	// Здоровье
	healthPercent := float32(health) / float32(maxHealth)
	vector.DrawFilledRect(screen, x+2, y+2, float32(barWidth-4)*healthPercent, float32(barHeight-4), color.RGBA{255, 50, 50, 255}, true)
	
	// Рамка
	vector.DrawFilledRect(screen, x, y, float32(barWidth), float32(barHeight), color.RGBA{255, 255, 255, 255}, false)
}
