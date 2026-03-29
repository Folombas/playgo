// Package render - система отрисовки
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
	
	// TODO: Использовать спрайты когда будут загружены
	// Временная отрисовка прямоугольниками
	if p.IsCrouching {
		// Приседание - более низкий прямоугольник
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY+12), 32, 20, color.RGBA{50, 150, 255, 255}, true)
	} else {
		// Стоя/бег
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), 32, 32, color.RGBA{50, 150, 255, 255}, true)
	}
	
	// Глаза (направление взгляда)
	eyeX := float32(screenX + 20)
	if p.Facing == -1 {
		eyeX = float32(screenX + 8)
	}
	vector.DrawFilledRect(screen, eyeX, float32(screenY+8), 6, 6, color.RGBA{255, 255, 255, 255}, true)
	
	// Оружие
	gunX := float32(screenX + 20)
	gunY := float32(screenY + 18)
	if p.Facing == -1 {
		gunX = float32(screenX - 4)
	}
	vector.DrawFilledRect(screen, gunX, gunY, 16, 6, color.RGBA{80, 80, 80, 255}, true)
}

// DrawEnemy - отрисовка врага
func (r *Renderer) DrawEnemy(screen *ebiten.Image, e entity.Enemy, cameraX, cameraY float64) {
	screenX := e.X - cameraX
	screenY := e.Y - cameraY
	
	var bodyColor color.RGBA
	
	// Цвет по типу врага
	switch e.Type {
	case "slime":
		bodyColor = color.RGBA{50, 200, 50, 255} // Зелёный слайм
	case "fly":
		bodyColor = color.RGBA{200, 200, 50, 255} // Жёлтая муха
	case "snail":
		bodyColor = color.RGBA{200, 100, 50, 255} // Коричневая улитка
	case "fish":
		bodyColor = color.RGBA{50, 100, 200, 255} // Синяя рыба
	default:
		bodyColor = color.RGBA{150, 150, 150, 255}
	}
	
	// Тело врага
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(e.Width), float32(e.Height), bodyColor, true)
	
	// Глаза
	eyeY := float32(screenY + 8)
	vector.DrawFilledRect(screen, float32(screenX+8), eyeY, 6, 6, color.RGBA{255, 255, 255, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX)+float32(e.Width)-14, eyeY, 6, 6, color.RGBA{255, 255, 255, 255}, true)
	
	// Зрачки
	vector.DrawFilledRect(screen, float32(screenX+10), eyeY+2, 3, 3, color.RGBA{0, 0, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX)+float32(e.Width)-12, eyeY+2, 3, 3, color.RGBA{0, 0, 0, 255}, true)
}

// DrawBoss - отрисовка босса
func (r *Renderer) DrawBoss(screen *ebiten.Image, b entity.Boss, cameraX, cameraY float64) {
	screenX := b.X - cameraX
	screenY := b.Y - cameraY
	
	// Тело босса (красный огромный враг)
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(b.Width), float32(b.Height), color.RGBA{200, 30, 30, 255}, true)
	
	// Броня/панцирь
	vector.DrawFilledRect(screen, float32(screenX+10), float32(screenY+10), float32(b.Width-20), float32(b.Height-30), color.RGBA{100, 100, 100, 255}, true)
	
	// Глаза (огромные)
	vector.DrawFilledRect(screen, float32(screenX+20), float32(screenY+15), 15, 15, color.RGBA{255, 150, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+45), float32(screenY+15), 15, 15, color.RGBA{255, 150, 0, 255}, true)
	
	// Зрачки
	vector.DrawFilledRect(screen, float32(screenX+25), float32(screenY+20), 6, 6, color.RGBA{0, 0, 0, 255}, true)
	vector.DrawFilledRect(screen, float32(screenX+50), float32(screenY+20), 6, 6, color.RGBA{0, 0, 0, 255}, true)
	
	// Шипы/украшения
	for i := 0; i < 5; i++ {
		x := float32(screenX) + float32(i*16)
		vector.DrawFilledRect(screen, x, float32(screenY-5), 8, 8, color.RGBA{150, 50, 50, 255}, true)
	}
}

// DrawPlatform - отрисовка платформы
func (r *Renderer) DrawPlatform(screen *ebiten.Image, p level.Platform, cameraX, cameraY float64) {
	screenX := p.X - cameraX
	screenY := p.Y - cameraY
	
	var topColor, bodyColor color.RGBA
	
	// Цвет по типу платформы
	switch p.Type {
	case "ground":
		topColor = color.RGBA{100, 180, 100, 255} // Трава
		bodyColor = color.RGBA{120, 80, 60, 255}  // Земля
	case "stone":
		topColor = color.RGBA{150, 150, 150, 255} // Серый камень
		bodyColor = color.RGBA{120, 120, 120, 255}
	case "brick":
		topColor = color.RGBA{180, 100, 80, 255} // Кирпич
		bodyColor = color.RGBA{150, 80, 60, 255}
	case "metal":
		topColor = color.RGBA{200, 200, 220, 255} // Металл
		bodyColor = color.RGBA{150, 150, 170, 255}
	default:
		topColor = color.RGBA{150, 150, 150, 255}
		bodyColor = color.RGBA{120, 120, 120, 255}
	}
	
	// Верхняя часть (трава/покрытие)
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(p.Width), 8, topColor, true)
	
	// Основная часть платформы
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY)+8, float32(p.Width), float32(p.Height)-8, bodyColor, true)
	
	// Детали (точки/текстура)
	for x := float32(0); x < float32(p.Width); x += 20 {
		vector.DrawFilledRect(screen, float32(screenX)+x, float32(screenY)+12, 4, 4, color.RGBA{100, 100, 100, 200}, true)
	}
}

// DrawCollectible - отрисовка предмета
func (r *Renderer) DrawCollectible(screen *ebiten.Image, c level.Collectible, cameraX, cameraY float64) {
	if c.Collected {
		return
	}
	
	screenX := c.X - cameraX
	screenY := c.Y - cameraY
	
	var gemColor color.RGBA
	
	// Цвет по типу предмета
	switch c.Type {
	case "coin":
		gemColor = color.RGBA{255, 215, 0, 255} // Золото
	case "gem_red":
		gemColor = color.RGBA{255, 50, 50, 255} // Красный
	case "gem_blue":
		gemColor = color.RGBA{50, 150, 255, 255} // Синий
	case "gem_green":
		gemColor = color.RGBA{50, 255, 100, 255} // Зелёный
	default:
		gemColor = color.RGBA{255, 215, 0, 255}
	}
	
	// Рисуем как ромб/кристалл
	centerX := float32(screenX) + float32(c.Width)/2
	centerY := float32(screenY) + float32(c.Height)/2
	
	// Внешний контур
	vector.DrawFilledRect(screen, centerX-8, centerY-10, 16, 20, gemColor, true)
	vector.DrawFilledRect(screen, centerX-10, centerY-6, 20, 12, gemColor, true)
	
	// Блик
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
		// Вражеский снаряд (красный шар)
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 8, color.RGBA{255, 50, 50, 255}, true)
		vector.DrawFilledCircle(screen, float32(screenX)+6, float32(screenY)+6, 4, color.RGBA{255, 200, 200, 255}, true)
	} else {
		// Пуля игрока (жёлтый прямоугольник)
		vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(p.Width), float32(p.Height), color.RGBA{255, 255, 100, 255}, true)
		vector.DrawFilledRect(screen, float32(screenX)+2, float32(screenY)+1, float32(p.Width-4), float32(p.Height-2), color.RGBA{255, 255, 200, 255}, true)
	}
}
