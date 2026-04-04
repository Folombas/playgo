package config

// Константы игры
const (
	// Размер тайла в пикселях
	TileSize = 64

	// Размеры игрового поля (в тайлах)
	GridWidth  = 13
	GridHeight = 11

	// Размер экрана
	ScreenWidth  = GridWidth * TileSize  // 832
	ScreenHeight = GridHeight * TileSize // 704

	// Скорость игры
	TPS = 60

	// Скорость игрока (пикселей в секунду)
	PlayerSpeed = 180.0

	// Скорость врагов
	EnemySpeed = 90.0

	// Время таймера бомбы (в секундах)
	BombTimer = 2.0

	// Базовый радиус взрыва (в тайлах)
	BaseExplosionRadius = 2

	// Максимальное количество бомб
	MaxBombs = 1

	// Начальные жизни
	StartingLives = 3

	// Цвета
	ColorWhite = 0xFFFFFFFF
	ColorBlack = 0x000000FF
	ColorRed   = 0xFF0000FF
	ColorGreen = 0x00FF00FF
	ColorBlue  = 0x0000FFFF
)

// Пути к ассетам
const (
	AssetPathSprites = "assets/sprites/"
)
