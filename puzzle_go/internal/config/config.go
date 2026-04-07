package config

const (
	// Размеры окна
	ScreenWidth  = 1280
	ScreenHeight = 720

	// Размеры игрового поля
	BoardCols = 8
	BoardRows = 8

	// Размеры ячеек
	CellSize = 64
	CellPadding = 4

	// Позиция игрового поля на экране
	BoardOffsetX = 400
	BoardOffsetY = 40

	// Анимации
	AnimationDuration = 0.2 // секунды
	SwapDuration      = 0.15
	FallDuration      = 0.3
	MatchDuration     = 0.25

	// Очки
	BaseMatchScore    = 100
	ComboMultiplier   = 1.5
	RainbowMatchScore = 500

	// Типы кристаллов
	CrystalTypes = 6

	// Специальные элементы
	Match4Reward = "bomb"     // 4 в ряд
	Match5Reward = "rainbow"  // 5 в ряд
	LShapeReward = "beam_h"   // L/T форма горизонтальная
	TShapeReward = "beam_v"   // L/T форма вертикальная
)

// Цвета кристаллов (для генерации)
var CrystalColors = []string{
	"red",
	"blue",
	"green",
	"yellow",
	"violet",
	"orange",
}
