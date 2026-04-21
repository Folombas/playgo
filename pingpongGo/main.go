package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 800
	screenHeight = 600
	paddleWidth  = 10
	paddleHeight = 80
	ballSize     = 8
)

type Game struct {
	leftPaddleY, rightPaddleY float64
	ballX, ballY              float64
	ballVX, ballVY            float64
	leftScore, rightScore     int
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	g := &Game{
		leftPaddleY:  screenHeight/2 - paddleHeight/2,
		rightPaddleY: screenHeight/2 - paddleHeight/2,
		ballX:        screenWidth / 2,
		ballY:        screenHeight / 2,
	}
	g.resetBall()
	return g
}

func (g *Game) resetBall() {
	// Случайное направление: влево или вправо
	if rand.Float64() < 0.5 {
		g.ballVX = 4
	} else {
		g.ballVX = -4
	}
	g.ballVY = 4 * (rand.Float64()*2 - 1) // от -4 до 4
	g.ballX = screenWidth / 2
	g.ballY = screenHeight / 2
}

func (g *Game) Update() error {
	// Управление левой ракеткой (W/S)
	if ebiten.IsKeyPressed(ebiten.KeyW) && g.leftPaddleY > 0 {
		g.leftPaddleY -= 6
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) && g.leftPaddleY < screenHeight-paddleHeight {
		g.leftPaddleY += 6
	}
	// Управление правой ракеткой (стрелки вверх/вниз)
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) && g.rightPaddleY > 0 {
		g.rightPaddleY -= 6
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) && g.rightPaddleY < screenHeight-paddleHeight {
		g.rightPaddleY += 6
	}

	// Движение мяча
	g.ballX += g.ballVX
	g.ballY += g.ballVY

	// Отскок от верхней и нижней границ
	if g.ballY <= 0 || g.ballY >= screenHeight-ballSize {
		g.ballVY = -g.ballVY
	}

	// Проверка столкновения с левой ракеткой
	if g.ballX <= paddleWidth && g.ballY+ballSize >= g.leftPaddleY && g.ballY <= g.leftPaddleY+paddleHeight {
		g.ballVX = -g.ballVX
		// Корректировка позиции, чтобы не застревать
		g.ballX = paddleWidth
		// Небольшое изменение вертикальной скорости для динамики
		g.ballVY += (float64(g.leftPaddleY+paddleHeight/2) - (g.ballY + ballSize/2)) * 0.1
		g.limitVelocity()
	}

	// Проверка столкновения с правой ракеткой
	if g.ballX+ballSize >= screenWidth-paddleWidth && g.ballY+ballSize >= g.rightPaddleY && g.ballY <= g.rightPaddleY+paddleHeight {
		g.ballVX = -g.ballVX
		g.ballX = screenWidth - paddleWidth - ballSize
		g.ballVY += (float64(g.rightPaddleY+paddleHeight/2) - (g.ballY + ballSize/2)) * 0.1
		g.limitVelocity()
	}

	// Гол: мяч вышел за левый край
	if g.ballX+ballSize < 0 {
		g.rightScore++
		g.resetBall()
	}
	// Гол: мяч вышел за правый край
	if g.ballX > screenWidth {
		g.leftScore++
		g.resetBall()
	}

	// Перезапуск игры по клавише R
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.leftScore = 0
		g.rightScore = 0
		g.resetBall()
		g.leftPaddleY = screenHeight/2 - paddleHeight/2
		g.rightPaddleY = screenHeight/2 - paddleHeight/2
	}

	return nil
}

// Ограничиваем скорость, чтобы игра не становилась слишком бешеной
func (g *Game) limitVelocity() {
	maxSpeed := 8.0
	if math.Abs(g.ballVX) > maxSpeed {
		g.ballVX = math.Copysign(maxSpeed, g.ballVX)
	}
	if math.Abs(g.ballVY) > maxSpeed {
		g.ballVY = math.Copysign(maxSpeed, g.ballVY)
	}
	if math.Abs(g.ballVY) < 2 {
		g.ballVY = math.Copysign(2, g.ballVY)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Рисуем фон
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Рисуем левую ракетку
	ebitenutil.DrawRect(screen, 0, g.leftPaddleY, paddleWidth, paddleHeight, color.White)
	// Правую ракетку
	ebitenutil.DrawRect(screen, screenWidth-paddleWidth, g.rightPaddleY, paddleWidth, paddleHeight, color.White)
	// Мяч
	ebitenutil.DrawRect(screen, g.ballX, g.ballY, ballSize, ballSize, color.White)

	// Счёт
	scoreText := fmt.Sprintf("%d  |  %d", g.leftScore, g.rightScore)
	ebitenutil.DebugPrintAt(screen, scoreText, screenWidth/2-40, 20)

	// Подсказки
	ebitenutil.DebugPrintAt(screen, "W/S - left paddle", 20, screenHeight-30)
	ebitenutil.DebugPrintAt(screen, "↑/↓ - right paddle", screenWidth-200, screenHeight-30)
	ebitenutil.DebugPrintAt(screen, "R - restart", screenWidth/2-40, screenHeight-30)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowTitle("Ping Pong - Go + Ebitengine")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizable(false)
	if err := ebiten.RunGame(NewGame()); err != nil {
		panic(err)
	}
}