package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	width  = 60
	height = 20
)

var (
	ballX, ballY   = width/2, height/2
	ballVx, ballVy = 1, 1
	leftPaddleY    = height/2 - 2
	rightPaddleY   = height/2 - 2
	paddleHeight   = 4
	leftScore      = 0
	rightScore     = 0
	gameRunning    = true
)

func main() {
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()
	termbox.SetCursor(0, 0)

	go handleInput()

	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for gameRunning {
		<-ticker.C
		update()
		draw()
	}
}

func handleInput() {
	for {
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}
		// Обработка специальных клавиш (стрелки, Ctrl+C и т.д.)
		switch ev.Key {
		case termbox.KeyArrowUp:
			if rightPaddleY > 0 {
				rightPaddleY--
			}
			continue
		case termbox.KeyArrowDown:
			if rightPaddleY < height-paddleHeight {
				rightPaddleY++
			}
			continue
		case termbox.KeyCtrlC, termbox.KeyCtrlD:
			gameRunning = false
			return
		}
		// Обработка обычных символов (буквы, цифры)
		if ev.Ch != 0 {
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

func update() {
	ballX += ballVx
	ballY += ballVy

	if ballY <= 0 {
		ballY = 0
		ballVy = -ballVy
	}
	if ballY >= height-1 {
		ballY = height - 1
		ballVy = -ballVy
	}

	// Левая ракетка
	if ballX <= 2 && ballY >= leftPaddleY && ballY < leftPaddleY+paddleHeight {
		ballX = 3
		ballVx = -ballVx
	}
	// Правая ракетка
	if ballX >= width-3 && ballY >= rightPaddleY && ballY < rightPaddleY+paddleHeight {
		ballX = width - 4
		ballVx = -ballVx
	}

	if ballX <= 0 {
		rightScore++
		resetBall()
	}
	if ballX >= width-1 {
		leftScore++
		resetBall()
	}
}

func resetBall() {
	ballX, ballY = width/2, height/2
	ballVx = -ballVx
	ballVy = 1
	time.Sleep(500 * time.Millisecond)
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for i := 0; i < paddleHeight; i++ {
		termbox.SetCell(1, leftPaddleY+i, '█', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(width-2, rightPaddleY+i, '█', termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.SetCell(ballX, ballY, '●', termbox.ColorYellow, termbox.ColorDefault)

	scoreText := fmt.Sprintf("%d : %d", leftScore, rightScore)
	for i, ch := range scoreText {
		termbox.SetCell(width/2-len(scoreText)/2+i, 0, ch, termbox.ColorGreen, termbox.ColorDefault)
	}

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
