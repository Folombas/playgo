// Package level предоставляет систему генерации и управления уровнями
package level

import (
	"math/rand"
)

const (
	// Размеры тайла
	TileSize = 48

	// Типы тайлов
	TileEmpty = iota
	TileGrassTop
	TileGrassMid
	TileGrassLeft
	TileGrassRight
	TileDirtTop
	TileDirtMid
	TileBrick
	TileBoxEmpty
	TileBoxItem
	TileBoxCoin
	TileBoxUsed
	TileSpike
)

// Tile представляет один тайл уровня
type Tile struct {
	X, Y int
	Type int
}

// Coin представляет монету на уровне
type Coin struct {
	X, Y      float64
	Collected bool
	Value     int // 1=бронза, 2=серебро, 3=золото
	AnimFrame int
}

// Enemy представляет врага на уровне
type Enemy struct {
	X, Y      float64
	VX        float64
	Type      int // 0=slime, 1=fly
	Alive     bool
	AnimFrame int
	Direction int // 1=вправо, -1=влево
}

// Decoration представляет декорацию
type Decoration struct {
	X, Y float64
	Type int // 0=cloud, 1=bush, 2=plant
}

// Flag представляет флаг конца уровня
type Flag struct {
	X, Y     float64
	Collected bool
	Color    int // 0=green, 1=red
}

// Level представляет игровой уровень
type Level struct {
	Tiles       []*Tile
	Coins       []*Coin
	Enemies     []*Enemy
	Decorations []*Decoration
	Flag        *Flag
	Width       int  // Ширина уровня в тайлах
	Height      int  // Высота уровня в тайлах
	HasBoss     bool // Есть ли босс в конце
}

// NewLevel создаёт новый уровень
func NewLevel(width, height int) *Level {
	return &Level{
		Width:  width,
		Height: height,
		Tiles:  make([]*Tile, 0),
		Coins:  make([]*Coin, 0),
		Enemies: make([]*Enemy, 0),
		Decorations: make([]*Decoration, 0),
	}
}

// Generate генерирует уровень процедурно
func (l *Level) Generate(seed int64) {
	rand.Seed(seed)

	// Генерируем землю
	l.generateTerrain()

	// Генерируем платформы
	l.generatePlatforms()

	// Генерируем монеты
	l.generateCoins()

	// Генерируем врагов
	l.generateEnemies()

	// Генерируем декорации
	l.generateDecorations()

	// Добавляем флаг в конце
	l.Flag = &Flag{
		X: float64((l.Width - 5) * TileSize),
		Y: float64(8 * TileSize),
		Color: rand.Intn(2),
	}
}

// generateTerrain генерирует ландшафт
func (l *Level) generateTerrain() {
	for x := 0; x < l.Width; x++ {
		// Определяем высоту земли в этой точке
		groundY := 12

		// Добавляем вариации высоты
		if x > 10 && rand.Float32() < 0.1 {
			groundY = 10 // Приподнятая платформа
		}

		// Верхний слой травы
		if x == 0 {
			l.Tiles = append(l.Tiles, &Tile{X: x, Y: groundY, Type: TileGrassLeft})
		} else if x == l.Width-1 {
			l.Tiles = append(l.Tiles, &Tile{X: x, Y: groundY, Type: TileGrassRight})
		} else {
			l.Tiles = append(l.Tiles, &Tile{X: x, Y: groundY, Type: TileGrassTop})
		}

		// Слой земли под травой
		for y := groundY + 1; y < groundY+3; y++ {
			if y == groundY+1 {
				l.Tiles = append(l.Tiles, &Tile{X: x, Y: y, Type: TileDirtTop})
			} else {
				l.Tiles = append(l.Tiles, &Tile{X: x, Y: y, Type: TileDirtMid})
			}
		}
	}
}

// generatePlatforms генерирует платформы
func (l *Level) generatePlatforms() {
	for x := 10; x < l.Width-10; x++ {
		// Случайные платформы
		if rand.Float32() < 0.15 {
			platformY := rand.Intn(4) + 6
			platformLength := rand.Intn(3) + 2

			for bx := 0; bx < platformLength; bx++ {
				if x+bx < l.Width-10 {
					// Выбираем тип блока
					tileType := TileBrick
					r := rand.Float32()
					if r < 0.2 {
						tileType = TileBoxItem
					} else if r < 0.3 {
						tileType = TileBoxCoin
					}

					l.Tiles = append(l.Tiles, &Tile{
						X: x + bx,
						Y: platformY,
						Type: tileType,
					})
				}
			}
		}
	}
}

// generateCoins генерирует монеты
func (l *Level) generateCoins() {
	for x := 10; x < l.Width-10; x++ {
		if rand.Float32() < 0.2 {
			coinY := float64(rand.Intn(6) + 4)
			value := rand.Intn(3) + 1 // 1-3

			l.Coins = append(l.Coins, &Coin{
				X: float64(x*TileSize + 15),
				Y: coinY * TileSize,
				Value: value,
			})
		}
	}
}

// generateEnemies генерирует врагов
func (l *Level) generateEnemies() {
	for x := 15; x < l.Width-10; x++ {
		if rand.Float32() < 0.08 {
			enemyType := rand.Intn(2) // 0=slime, 1=fly
			enemyY := float64(11 * TileSize)

			// Летающие враги выше
			if enemyType == 1 {
				enemyY = float64(rand.Intn(4) + 6) * TileSize
			}

			l.Enemies = append(l.Enemies, &Enemy{
				X: float64(x * TileSize),
				Y: enemyY,
				Type: enemyType,
				Alive: true,
				Direction: -1, // Начинаем движение влево
				VX: -1.5,
			})
		}
	}
}

// generateDecorations генерирует декорации
func (l *Level) generateDecorations() {
	for x := 0; x < l.Width; x++ {
		// Кусты на земле
		if rand.Float32() < 0.1 {
			l.Decorations = append(l.Decorations, &Decoration{
				X: float64(x*TileSize),
				Y: float64(11*TileSize - 32),
				Type: 1, // bush
			})
		}

		// Облака на небе
		if rand.Float32() < 0.05 {
			l.Decorations = append(l.Decorations, &Decoration{
				X: float64(x*TileSize),
				Y: float64(rand.Intn(100) + 50),
				Type: 0, // cloud
			})
		}

		// Растения
		if rand.Float32() < 0.03 {
			l.Decorations = append(l.Decorations, &Decoration{
				X: float64(x*TileSize),
				Y: float64(11*TileSize - 48),
				Type: 2, // plant
			})
		}
	}
}

// GetTileAt возвращает тайл в указанных координатах
func (l *Level) GetTileAt(x, y int) *Tile {
	for _, tile := range l.Tiles {
		if tile.X == x && tile.Y == y {
			return tile
		}
	}
	return nil
}

// IsSolid проверяет, является ли тайл твёрдым
func (l *Level) IsSolid(x, y int) bool {
	tile := l.GetTileAt(x, y)
	if tile == nil {
		return false
	}

	// Все тайлы кроме пустых - твёрдые
	return tile.Type != TileEmpty
}

// Update обновляет состояние уровня
func (l *Level) Update() {
	// Обновляем анимацию монет
	for _, coin := range l.Coins {
		if !coin.Collected {
			coin.AnimFrame = (coin.AnimFrame + 1) % 4
		}
	}

	// Обновляем врагов
	for _, enemy := range l.Enemies {
		if !enemy.Alive {
			continue
		}

		enemy.X += enemy.VX
		enemy.AnimFrame = (enemy.AnimFrame + 1) % 4

		// Простой AI - разворот по таймеру
		if rand.Float32() < 0.02 {
			enemy.Direction *= -1
			enemy.VX = float64(enemy.Direction) * 1.5
		}
	}
}

// CheckTileCollision проверяет коллизию игрока с тайлами
// Возвращает true если игрок на земле
func (l *Level) CheckTileCollision(playerX, playerY, playerWidth, playerHeight float64, playerVY float64) (onGround bool, newX, newY float64) {
	onGround = false
	newX, newY = playerX, playerY

	// Проверяем тайлы вокруг игрока
	startX := int(playerX / TileSize)
	endX := int((playerX + playerWidth) / TileSize)
	startY := int(playerY / TileSize)
	endY := int((playerY + playerHeight) / TileSize)

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			tile := l.GetTileAt(x, y)
			if tile == nil || tile.Type == TileEmpty {
				continue
			}

			tileX := float64(x * TileSize)
			tileY := float64(y * TileSize)

			// Проверка пересечения
			if playerX < tileX+TileSize &&
				playerX+playerWidth > tileX &&
				playerY < tileY+TileSize &&
				playerY+playerHeight > tileY {

				// Определяем сторону коллизии
				overlapX := (playerX + playerWidth/2) - (tileX + TileSize/2)
				overlapY := (playerY + playerHeight/2) - (tileY + TileSize/2)

				halfWidth := playerWidth/2 + TileSize/2
				halfHeight := playerHeight/2 + TileSize/2

				if abs(overlapX) < halfWidth && abs(overlapY) < halfHeight {
					// Вычисляем глубину пересечения
					dx := halfWidth - abs(overlapX)
					dy := halfHeight - abs(overlapY)

					if dx < dy {
						// Коллизия по горизонтали
						if overlapX < 0 {
							newX = tileX - playerWidth
						} else {
							newX = tileX + TileSize
						}
					} else {
						// Коллизия по вертикали
						if overlapY < 0 {
							newY = tileY - playerHeight
							onGround = true
						} else {
							newY = tileY + TileSize
						}
					}
				}
			}
		}
	}

	return
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
