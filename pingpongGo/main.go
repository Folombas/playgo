package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	width  = 60
	height = 20
)

var (
	// позиции мяча
	ballX, ballY   = width/2, height/2
	ballVx, ballVy = 1, 1

	// позиции ракеток (Y координата верхнего края)
	leftPaddleY   = height/2 - 2
	rightPaddleY  = height/2 - 2
	paddleHeight  = 4
	leftScore     = 0
	rightScore    = 0
	gameRunning   = true
)

func main() {
	// инициализация termbox
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()

	// отключаем курсор
	termbox.SetCursor(0, 0)

	// запускаем обработку клавиш в горутине
	go handleInput()

	// игровой цикл
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for gameRunning {
		<-ticker.C
		update()
		draw()
	}
}

// обработка клавиатуры
func handleInput() {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyArrowUp:
				if rightPaddleY > 0 {
					rightPaddleY--
				}
			case termbox.KeyArrowDown:
				if rightPaddleY < height-paddleHeight {
					rightPaddleY++
				}
			case termbox.KeyChar:
				switch ev.Ch {
				case 'w', 'W':
					if leftPaddleY > 0 {
						leftPaddleY--
					}
				case 's', 'S':
					if leftPaddleY < height-paddleHeight {
						leftPaddleY++
					}
				case 'q', 'Q':
					gameRunning = false
					return
				}
			}
		}
	}
}

// обновление физики
func update() {
	// движение мяча
	ballX += ballVx
	ballY += ballVy

	// отскок от верхней и нижней стен
	if ballY <= 0 {
		ballY = 0
		ballVy = -ballVy
	}
	if ballY >= height-1 {
		ballY = height - 1
		ballVy = -ballVy
	}

	// столкновение с левой ракеткой
	if ballX <= 2 && ballY >= leftPaddleY && ballY < leftPaddleY+paddleHeight {
		ballX = 3
		ballVx = -ballVx
	}

	// столкновение с правой ракеткой
	if ballX >= width-3 && ballY >= rightPaddleY && ballY < rightPaddleY+paddleHeight {
		ballX = width - 4
		ballVx = -ballVx
	}

	// гол (мяч ушёл за левый край)
	if ballX <= 0 {
		rightScore++
		resetBall()
	}
	// гол (мяч ушёл за правый край)
	if ballX >= width-1 {
		leftScore++
		resetBall()
	}
}

// сброс мяча в центр после гола
func resetBall() {
	ballX, ballY = width/2, height/2
	ballVx = -ballVx // меняем направление, чтобы не было патовой ситуации
	ballVy = 1
	if ballVy == 0 {
		ballVy = 1
	}
	time.Sleep(500 * time.Millisecond)
}

// отрисовка поля, ракеток, мяча и счёта
func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// левая ракетка
	for i := 0; i < paddleHeight; i++ {
		termbox.SetCell(1, leftPaddleY+i, '█', termbox.ColorWhite, termbox.ColorDefault)
	}

	// правая ракетка
	for i := 0; i < paddleHeight; i++ {
		termbox.SetCell(width-2, rightPaddleY+i, '█', termbox.ColorWhite, termbox.ColorDefault)
	}

	// мяч
	termbox.SetCell(ballX, ballY, '●', termbox.ColorYellow, termbox.ColorDefault)

	// счёт
	scoreText := fmt.Sprintf("%d : %d", leftScore, rightScore)
	textX := width/2 - len(scoreText)/2
	for i, ch := range scoreText {
		termbox.SetCell(textX+i, 0, ch, termbox.ColorGreen, termbox.ColorDefault)
	}

	// рамка (простая)
	for x := 0; x < width; x++ {
		termbox.SetCell(x, 0, '─', termbox.ColorBlue, termbox.ColorDefault)
		termbox.SetCell(x, height-1, '─', termbox.ColorBlue, termbox.ColorDefault)
	}
	for y := 0; y < height; y++ {
		termbox.SetCell(0, y, '│', termbox.ColorBlue, termbox.ColorDefault)
		termbox.SetCell(width-1, y, '│', termbox.ColorBlue, termbox.ColorDefault)
	}

	termbox.Flush()
}
