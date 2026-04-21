package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

const (
	width  = 60
	height = 15
	ground = height - 2
)

type Object struct {
	x, y int
	typ  byte // 'c' - кактус, 'm' - монетка
}

var (
	playerX     = 5
	playerY     = ground
	playerJump  = false
	jumpVel     = 0
	jumpGravity = 1

	objects     []Object
	score       = 0
	gameOver    = false
	speed       = 1
	speedTicker *time.Ticker
)

func main() {
	err := termbox.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.SetCursor(0, 0)

	rand.Seed(time.Now().UnixNano())
	go handleInput()
	go generateObjects()
	speedTicker = time.NewTicker(10 * time.Second)
	go increaseSpeed()

	gameLoop()
}

func handleInput() {
	for {
		ev := termbox.PollEvent()
		if ev.Type != termbox.EventKey {
			continue
		}
		// Обработка специальных клавиш
		switch ev.Key {
		case termbox.KeyArrowUp:
			if !playerJump && !gameOver {
				playerJump = true
				jumpVel = -4
			}
			continue
		case termbox.KeyCtrlC:
			termbox.Close()
			fmt.Printf("Game over! Score: %d\n", score)
			return
		}
		// Обработка обычных символов (пробел, q)
		if ev.Ch != 0 {
			switch ev.Ch {
			case ' ', 'w', 'W':
				if !playerJump && !gameOver {
					playerJump = true
					jumpVel = -4
				}
			case 'q', 'Q':
				termbox.Close()
				fmt.Printf("Game over! Score: %d\n", score)
				return
			}
		}
	}
}

func generateObjects() {
	for !gameOver {
		delay := 800 - (speed * 30)
		if delay < 300 {
			delay = 300
		}
		time.Sleep(time.Duration(delay) * time.Millisecond)
		if gameOver {
			break
		}
		typ := byte('c')
		if rand.Intn(100) < 30 {
			typ = 'm'
		}
		objects = append(objects, Object{x: width - 1, y: ground, typ: typ})
	}
}

func increaseSpeed() {
	for !gameOver {
		<-speedTicker.C
		if speed < 8 {
			speed++
		}
	}
}

func gameLoop() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for !gameOver {
		<-ticker.C
		update()
		draw()
	}
	drawGameOver()
	time.Sleep(3 * time.Second)
}

func update() {
	if playerJump {
		playerY += jumpVel
		jumpVel += jumpGravity
		if playerY >= ground {
			playerY = ground
			playerJump = false
			jumpVel = 0
		}
	} else {
		playerY = ground
	}

	newObjects := []Object{}
	for _, obj := range objects {
		obj.x -= speed
		if obj.x > 0 {
			newObjects = append(newObjects, obj)
		}
	}
	objects = newObjects

	for i := 0; i < len(objects); i++ {
		obj := objects[i]
		if obj.x == playerX && obj.y == playerY {
			if obj.typ == 'c' {
				gameOver = true
				return
			} else if obj.typ == 'm' {
				score++
				objects = append(objects[:i], objects[i+1:]...)
				i--
			}
		}
	}
}

func draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	for x := 0; x < width; x++ {
		termbox.SetCell(x, ground+1, '_', termbox.ColorWhite, termbox.ColorDefault)
	}

	ch := 'D'
	if playerJump {
		ch = '^'
	}
	termbox.SetCell(playerX, playerY, ch, termbox.ColorYellow, termbox.ColorDefault)

	for _, obj := range objects {
		var ch rune
		var color termbox.Attribute
		if obj.typ == 'c' {
			ch = '#'
			color = termbox.ColorGreen
		} else {
			ch = '$'
			color = termbox.ColorYellow
		}
		termbox.SetCell(obj.x, obj.y, ch, color, termbox.ColorDefault)
	}

	scoreText := fmt.Sprintf("Score: %d  Speed: %d", score, speed)
	for i, c := range scoreText {
		termbox.SetCell(i, 0, c, termbox.ColorCyan, termbox.ColorDefault)
	}
	termbox.Flush()
}

func drawGameOver() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	msg := fmt.Sprintf("GAME OVER! Score: %d", score)
	msgX := width/2 - len(msg)/2
	for i, c := range msg {
		termbox.SetCell(msgX+i, height/2, c, termbox.ColorRed, termbox.ColorDefault)
	}
	restartMsg := "Press Q to exit"
	restartX := width/2 - len(restartMsg)/2
	for i, c := range restartMsg {
		termbox.SetCell(restartX+i, height/2+2, c, termbox.ColorWhite, termbox.ColorDefault)
	}
	termbox.Flush()
}
