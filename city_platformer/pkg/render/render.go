// Package render - система отрисовки
// Go365 Day 88 - Forest Pack
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
}

// NewRenderer - создание рендерера
func NewRenderer(ss *sprite.SpriteSheet) *Renderer {
	return &Renderer{
		spriteSheet: ss,
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
	
	// Получаем спрайт игрока
	state := "stand"
	if p.IsCrouching {
		state = "crouch"
	} else if !p.OnGround {
		state = "jump"
	} else if p.IsMoving {
		state = "walk"
	}
	
	frame := int(p.AnimFrame)
	playerSprite := r.spriteSheet.GetPlayerFrame(state, frame)
	
	if playerSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		
		// Отражение по горизонтали если смотрим влево
		if p.Facing == -1 {
			opts.GeoM.Scale(-1, 1)
			opts.GeoM.Translate(float64(p.Width), 0)
		}
		
		opts.GeoM.Translate(screenX, screenY)
		opts.GeoM.Scale(1.5, 1.5)
		
		screen.DrawImage(playerSprite, opts)
	} else {
		// Резервная отрисовка
		r.drawPlayerFallback(screen, p, cameraX, cameraY)
	}
}

// drawPlayerFallback - резервная отрисовка игрока
func (r *Renderer) drawPlayerFallback(screen *ebiten.Image, p entity.Player, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	if p.IsCrouching {
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY+12), 32, 20, color.RGBA{100, 180, 100, 255}, true)
	} else {
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 32, 32, color.RGBA{100, 180, 100, 255}, true)
	}
	
	// Глаза
	eyeX := float32(screenX + 20)
	if p.Facing == -1 {
		eyeX = float32(screenX + 8)
	}
	vector.DrawFilledRect(screen, eyeX, float32(screenY+8), 6, 6, color.RGBA{255, 255, 255, 255}, true)
	
	// Оружие
	gunX := float32(screenX + 20)
	if p.Facing == -1 {
		gunX = float32(screenX - 4)
	}
	vector.DrawFilledRect(screen, gunX, float32(screenY+18), 16, 6, color.RGBA{80, 80, 80, 255}, true)
}

// DrawEnemy - отрисовка врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	// Получаем спрайт врага
	frame := int(e.AnimFrame)
	enemySprite := r.spriteSheet.GetEnemyFrame(e.Type, frame)
	
	if enemySprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(enemySprite, opts)
	} else {
		// Резервная отрисовка
		r.drawEnemyFallback(screen, e, cameraX, cameraY)
	}
}

// drawEnemyFallback - резервная отрисовка врага
func (r *Renderer) drawEnemyFallback(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	var bodyColor color.RGBA
	
	switch e.Type {
	case "slime":
		bodyColor = color.RGBA{50, 180, 50, 255}
	case "fly":
		bodyColor = color.RGBA{200, 180, 50, 255}
	case "snail":
		bodyColor = color.RGBA{180, 100, 50, 255}
	case "fish":
		bodyColor = color.RGBA{50, 100, 200, 255}
	default:
		bodyColor = color.RGBA{150, 150, 150, 255}
	}
	
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(e.Width), float32(e.Height), bodyColor, true)
	
	// Глаза
	eyeY := float32(screenY + 8)
	vector.DrawFilledRect(screen, float32(screenX+8), eyeY, 6, 6, color.RGBA{255, 255, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX)+float32(e.Width)-14, eyeY, 6, 6, color.RGBA{255, 255, 255, 255}, true)
}

// DrawBoss - отрисовка босса
func (r *Renderer) DrawBoss(screen *ebiten.Image, b entity.Boss, cameraX, cameraY float64) {
	screenX := b.X - cameraX
	screenY := b.Y - cameraY
	
	// Тело босса (огромный лесной монстр)
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(b.Width), float32(b.Height), color.RGBA{80, 60, 40, 255}, true)
	
	// Мох/трава на боссе
	vector.DrawFilledRect(screen, float32(screenX+5), float32(screenY+5), float32(b.Width-10), 15, color.RGBA{60, 120, 60, 255}, true)
	
	// Глаза (светящиеся)
	vector.DrawFilledRect(screen, float32(screenX+15), float32(screenY+20), 18, 18, color.RGBA{255, 200, 50, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+47), float32(screenY+20), 18, 18, color.RGBA{255, 200, 50, 255}, true)
	
	// Зрачки
	vector.DrawFilledRect(screen, float32(screenX+22), float32(screenY+26), 7, 7, color.RGBA{0, 0, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+54), float32(screenY+26), 7, 7, color.RGBA{0, 0, 0, 255}, true)
	
	// Рога/ветки
	for i := 0; i < 4; i++ {
		x := float32(screenX) + float32(i*22)
		vector.DrawFilledRect(screen, x, float32(screenY-10), 10, 15, color.RGBA{100, 60, 40, 255}, true)
	}
}

// DrawPlatform - отрисовка платформы
func (r *Renderer) DrawPlatform(screen *ebiten.Image, p level.Platform, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	switch p.Type {
	case "ground":
		r.drawGround(screen, screenX, screenY, p.Width, p.Height, cameraX, cameraY)
	case "stone":
		r.drawStonePlatform(screen, screenX, screenY, p.Width, p.Height)
	case "brick":
		r.drawBrickPlatform(screen, screenX, screenY, p.Width, p.Height)
	case "metal":
		r.drawMetalPlatform(screen, screenX, screenY, p.Width, p.Height)
	default:
		r.drawGrassPlatform(screen, screenX, screenY, p.Width, p.Height)
	}
}

// drawGround - отрисовка земли
func (r *Renderer) drawGround(screen *ebiten.Image, x, y, width, height, cameraX, cameraY float64) {
	// Верхний слой с травой
	groundTop := r.spriteSheet.Tiles["ground_top"]
	if groundTop != nil {
		tileW := float64(groundTop.Bounds().Dx())
		tileH := float64(groundTop.Bounds().Dy())
		
		for tx := x; tx < x+width; tx += tileW {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(tx, y)
			screen.DrawImage(groundTop, opts)
		}
		
		// Средний слой земли
		groundMid := r.spriteSheet.Tiles["ground_mid"]
		if groundMid != nil {
			for ty := y + tileH; ty < y+height; ty += tileH {
				for tx := x; tx < x+width; tx += tileW {
					opts := &ebiten.DrawImageOptions{}
					opts.GeoM.Translate(tx, ty)
					screen.DrawImage(groundMid, opts)
				}
			}
		}
	} else {
		// Резервная отрисовка
		vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), 8, color.RGBA{100, 180, 100, 255}, true)
		vector.DrawFilledRect(screen, float32(x), float32(y)+8, float32(width), float32(height)-8, color.RGBA{139, 90, 43, 255}, true)
	}
}

// drawGrassPlatform - отрисовка травяной платформы
func (r *Renderer) drawGrassPlatform(screen *ebiten.Image, x, y, width, height float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), 8, color.RGBA{100, 180, 100, 255}, true)
	vector.DrawFilledRect(screen, float32(x), float32(y)+8, float32(width), float32(height)-8, color.RGBA{139, 90, 43, 255}, true)
}

// drawStonePlatform - отрисовка каменной платформы
func (r *Renderer) drawStonePlatform(screen *ebiten.Image, x, y, width, height float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{150, 150, 150, 255}, true)
	
	// Детали камня
	for i := 0; i < int(width)/20; i++ {
		for j := 0; j < int(height)/20; j++ {
			vector.DrawFilledRect(screen, float32(x)+float32(i*20)+5, float32(y)+float32(j*20)+5, 10, 10, color.RGBA{130, 130, 130, 255}, true)
		}
	}
}

// drawBrickPlatform - отрисовка кирпичной платформы
func (r *Renderer) drawBrickPlatform(screen *ebiten.Image, x, y, width, height float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{180, 100, 80, 255}, true)
	
	// Кирпичи
	for i := 0; i < int(width)/30; i++ {
		for j := 0; j < int(height)/15; j++ {
			offset := float32((j % 2) * 15)
			vector.DrawFilledRect(screen, float32(x)+float32(i*30)+offset, float32(y)+float32(j*15), 28, 13, color.RGBA{150, 80, 60, 255}, true)
		}
	}
}

// drawMetalPlatform - отрисовка металлической платформы
func (r *Renderer) drawMetalPlatform(screen *ebiten.Image, x, y, width, height float64) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), color.RGBA{180, 180, 200, 255}, true)
	
	// Заклёпки
	for i := 0; i < int(width)/40; i++ {
		for j := 0; j < int(height)/40; j++ {
			vector.DrawFilledCircle(screen, float32(x)+float32(i*40)+20, float32(y)+float32(j*40)+20, 4, color.RGBA{120, 120, 140, 255}, true)
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
	
	// Получаем спрайт предмета
	itemSprite := r.spriteSheet.GetItem(c.Type)
	
	if itemSprite != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(itemSprite, opts)
	} else {
		// Резервная отрисовка
		r.drawCollectibleFallback(screen, c, cameraX, cameraY)
	}
}

// drawCollectibleFallback - резервная отрисовка предмета
func (r *Renderer) drawCollectibleFallback(screen *ebiten.Image, c level.Collectible, cameraX, cameraY float64) {
	screenX := c.X - cameraX
	screenY := c.Y - cameraY
	
	var gemColor color.RGBA
	
	switch c.Type {
	case "coin":
		gemColor = color.RGBA{255, 200, 50, 255}
	case "gem_red":
		gemColor = color.RGBA{255, 80, 80, 255}
	case "gem_blue":
		gemColor = color.RGBA{80, 150, 255, 255}
	case "gem_green":
		gemColor = color.RGBA{80, 255, 100, 255}
	default:
		gemColor = color.RGBA{255, 200, 50, 255}
	}
	
	centerX := float32(screenX) + float32(c.Width)/2
	centerY := float32(screenY) + float32(c.Height)/2
	
	vector.DrawFilledRect(screen, centerX-8, centerY-10, 16, 20, gemColor, true)
	vector.DrawFilledRect(screen, centerX-10, centerY-6, 20, 12, gemColor, true)
	vector.DrawFilledRect(screen, centerX-3, centerY-5, 6, 10, color.RGBA{255, 255, 255, 200}, true)
}

// DrawProjectile - отрисовка снаряда
func (r *Renderer) DrawProjectile(screen *ebiten.Image, p entity.Projectile, cameraX, cameraY float64) {
	if !p.Active {
		return
	}
	
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	if p.IsEnemy {
		// Вражеский снаряд (зелёный энергетический шар)
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 8, color.RGBA{100, 255, 100, 255}, true)
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 4, color.RGBA{200, 255, 200, 255}, true)
	} else {
		// Пуля игрока (жёлтый прямоугольник)
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(p.Width), float32(p.Height), color.RGBA{255, 220, 50, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX)+2, float32(screenY)+1, float32(p.Width-4), float32(p.Height-2), color.RGBA{255, 255, 180, 255}, true)
	}
}

// DrawBackground - отрисовка фона
func (r *Renderer) DrawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	// Слои параллакса
	r.drawParallaxLayer(screen, r.spriteSheet.BgLayerC, cameraX*0.1, 0)
	r.drawParallaxLayer(screen, r.spriteSheet.BgLayerB, cameraX*0.3, 0)
	r.drawParallaxLayer(screen, r.spriteSheet.BgLayerA, cameraX*0.5, 0)
}

// drawParallaxLayer - отрисовка слоя параллакса
func (r *Renderer) drawParallaxLayer(screen *ebiten.Image, layer *ebiten.Image, offsetX, offsetY float64) {
	if layer == nil {
		return
	}
	
	layerW := float64(layer.Bounds().Dx())
	layerH := float64(layer.Bounds().Dy())
	
	// Рисуем слой с повторением
	for x := offsetX; x < float64(screen.Bounds().Dx()); x += layerW {
		for y := offsetY; y < float64(screen.Bounds().Dy()); y += layerH {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(x, y)
			screen.DrawImage(layer, opts)
		}
	}
}

// DrawDecoration - отрисовка декораций
func (r *Renderer) DrawDecoration(screen *ebiten.Image, decType string, x, y, cameraX, cameraY float64) {
	var dec *ebiten.Image
	
	switch decType {
	case "tree":
		dec = r.spriteSheet.Tree
	case "flower":
		dec = r.spriteSheet.Flower
	case "mushroom":
		dec = r.spriteSheet.Mushroom
	case "rock":
		dec = r.spriteSheet.Rock
	case "stone":
		dec = r.spriteSheet.Stone
	}
	
	if dec != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(x-cameraX, y-cameraY)
		screen.DrawImage(dec, opts)
	}
}
