package main

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed sprites/*
var spriteFS embed.FS

// TileSprites хранит спрайты для 6 типов фишек
var TileSprites []*ebiten.Image

// ButtonSprite - спрайт кнопки
var ButtonSprite *ebiten.Image

// BackgroundSprite - спрайт фона
var BackgroundSprite *ebiten.Image

// loadSprites загружает спрайты из встроенной файловой системы
// Если спрайты отсутствуют, создаёт fallback-графику программно
func loadSprites() {
	// Пытаемся загрузить спрайты из папки sprites/
	tileFiles := []string{
		"sprites/tile_0.png",
		"sprites/tile_1.png",
		"sprites/tile_2.png",
		"sprites/tile_3.png",
		"sprites/tile_4.png",
		"sprites/tile_5.png",
	}

	TileSprites = make([]*ebiten.Image, 6)
	for i, path := range tileFiles {
		img, err := loadEmbeddedImage(path)
		if err != nil {
			log.Printf("Не удалось загрузить %s, используем fallback: %v", path, err)
			TileSprites[i] = createFallbackTile(i)
		} else {
			TileSprites[i] = ebiten.NewImageFromImage(img)
		}
	}

	// Загружаем кнопку
	img, err := loadEmbeddedImage("sprites/button.png")
	if err != nil {
		log.Printf("Не удалось загрузить кнопку, используем fallback: %v", err)
		ButtonSprite = createFallbackButton()
	} else {
		ButtonSprite = ebiten.NewImageFromImage(img)
	}

	// Загружаем фон
	img, err = loadEmbeddedImage("sprites/background.png")
	if err != nil {
		log.Printf("Не удалось загрузить фон, используем fallback: %v", err)
		BackgroundSprite = createFallbackBackground()
	} else {
		BackgroundSprite = ebiten.NewImageFromImage(img)
	}

	log.Println("Спрайты загружены")
}

// loadEmbeddedImage загружает PNG изображение из embedded FS
func loadEmbeddedImage(path string) (image.Image, error) {
	data, err := spriteFS.ReadFile(path)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

// createFallbackTile создаёт программную текстуру фишки заданного цвета
func createFallbackTile(colorIndex int) *ebiten.Image {
	img := ebiten.NewImage(64, 64)

	// Цвета для 6 типов фишек
	colors := []struct{ r, g, b uint8 }{
		{255, 69, 58},   // 0: Красный
		{0, 122, 255},   // 1: Синий
		{52, 199, 89},   // 2: Зелёный
		{255, 204, 0},   // 3: Жёлтый
		{175, 82, 222},  // 4: Фиолетовый
		{255, 149, 0},   // 5: Оранжевый
	}

	c := colors[colorIndex%len(colors)]

	// Рисуем круглый спрайт
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			dx := float64(x) - 32
			dy := float64(y) - 32
			if dx*dx+dy*dy <= 30*30 {
				img.Set(x, y, &colorRGBA{c.r, c.g, c.b, 255})
			}
		}
	}

	return img
}

// colorRGBA - простая реализация color.Color для программного рисования
type colorRGBA struct{ r, g, b, a uint8 }

func (c *colorRGBA) RGBA() (uint32, uint32, uint32, uint32) {
	return uint32(c.r) * 0x101, uint32(c.g) * 0x101, uint32(c.b) * 0x101, uint32(c.a) * 0x101
}

// createFallbackButton создаёт программную кнопку
func createFallbackButton() *ebiten.Image {
	img := ebiten.NewImage(200, 50)
	for y := 0; y < 50; y++ {
		for x := 0; x < 200; x++ {
			// Синяя кнопка с закруглёнными краями
			img.Set(x, y, &colorRGBA{0, 122, 255, 255})
		}
	}
	return img
}

// createFallbackBackground создаёт программный фон
func createFallbackBackground() *ebiten.Image {
	img := ebiten.NewImage(800, 800)
	for y := 0; y < 800; y++ {
		for x := 0; x < 800; x++ {
			// Тёмно-синий градиент
			r := uint8(30 + (y * 180 / 800))
			g := uint8(30 + (y * 100 / 800))
			b := uint8(60 + (y * 120 / 800))
			img.Set(x, y, &colorRGBA{r, g, b, 255})
		}
	}
	return img
}
