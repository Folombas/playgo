package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// inputMovement обрабатывает движение фигуры
func inputMovement(g *Game) {
	if g.current == nil || g.board == nil {
		return
	}

	// Влево
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		g.current.X--
		if g.board.Collides(g.current) {
			g.current.X++
		} else {
			PlaySound(SoundMove)
		}
	}

	// Вправо
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		g.current.X++
		if g.board.Collides(g.current) {
			g.current.X--
		} else {
			PlaySound(SoundMove)
		}
	}

	// Вниз (ускорение)
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		g.dropTimer += 5 // Ускорить падение
	}

	// Вниз (сдвинуть на 1)
	if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
		g.current.Y++
		if g.board.Collides(g.current) {
			g.current.Y--
		}
	}

	// Поворот (вверх)
	if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
		g.current.Rotate()
		if g.board.Collides(g.current) {
			// Попробовать сдвинуть (wall kick)
			g.current.X--
			if g.board.Collides(g.current) {
				g.current.X += 2
				if g.board.Collides(g.current) {
					g.current.X--
					g.current.RotateCCW() // Отменить поворот
				}
			}
		}
		PlaySound(SoundRotate)
	}

	// Мгновенное падение (пробел)
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		hardDrop(g)
	}
}

// hardDrop мгновенно опускает фигуру
func hardDrop(g *Game) {
	dropY := g.board.GetDropY(g.current)
	g.score += (dropY - g.current.Y) * 2 // Бонус за hard drop
	g.current.Y = dropY
	g.dropCurrent()
	PlaySound(SoundHardDrop)
}

// inputStart проверяет ввод для начала игры
func inputStart() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}

// inputPause проверяет ввод для паузы
func inputPause() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyP) ||
		inpututil.IsKeyJustPressed(ebiten.KeyEscape)
}

// inputRestart проверяет ввод для рестарта
func inputRestart() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
}
