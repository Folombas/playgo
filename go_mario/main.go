// Go365 Day 87 - HERO ADVENTURE v1.0.0
// Простой и стабильный платформер

package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 600
	TileSize     = 48
	Gravity      = 0.6
	JumpForce    = -12.0
	Speed        = 4.0
)

// Кэш спрайтов
var sprites = make(map[string]*ebiten.Image)

func loadSprites() {
	// Игрок
	sprites["player"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
	sprites["player_walk1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
	sprites["player_walk2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_walk/p1_walk.png")
	sprites["player_jump"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")

	// Тайлы
	sprites["grass"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
	sprites["dirt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
	sprites["brick"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")
	sprites["box"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")

	// Враги
	sprites["enemy1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
	sprites["enemy2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")

	// Предметы
	sprites["coin"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
	sprites["flag"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/flagGreen.png")
	sprites["cloud"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/cloud1.png")
	sprites["bush"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/bush.png")

	// Фон
	sprites["bg"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/backgrounds/bg.png")
}

// Игрок
type Player struct {
	x, y      float64
	vx, vy    float64
	width     float64
	height    float64
	onGround  bool
	facing    int
	animFrame int
	coins     int
	score     int
	lives     int
	health    int
}

// Тайл
type Tile struct {
	x, y int
	id   int // 1=grass, 2=dirt, 3=brick, 4=box
}

// Враг
type Enemy struct {
	x, y  float64
	vx    float64
	alive bool
	frame int
}

// Монета
type Coin struct {
	x, y      float64
	collected bool
	frame     int
}

// Игра
type Game struct {
	player  *Player
	tiles   []*Tile
	enemies []*Enemy
	coins   []*Coin
	cameraX float64
	levelW  int
	frame   int
	state   int // 0=menu, 1=playing, 2=gameover, 3=win
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	loadSprites()

	g := &Game{
		player: &Player{
			x:      100,
			y:      300,
			width:  32,
			height: 48,
			facing: 1,
			lives:  3,
			health: 3,
		},
		state:  0,
		levelW: 100,
	}
	g.generateLevel()
	return g
}

func (g *Game) generateLevel() {
	g.tiles = nil
	g.enemies = nil
	g.coins = nil

	// Земля с травой
	for x := 0; x < g.levelW; x++ {
		// Ямы
		if x > 10 && x < g.levelW-10 && x%25 >= 20 && x%25 < 23 {
			continue
		}
		g.tiles = append(g.tiles, &Tile{x: x, y: 12, id: 1}) // grass
		g.tiles = append(g.tiles, &Tile{x: x, y: 13, id: 2}) // dirt
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: 2}) // dirt
	}

	// Платформы
	for x := 15; x < g.levelW-15; x += 20 {
		platY := 7
		for i := 0; i < 4; i++ {
			id := 3 // brick
			if i == 2 {
				id = 4 // box
			}
			g.tiles = append(g.tiles, &Tile{x: x + i, y: platY, id: id})
		}
		// Монета над платформой
		g.coins = append(g.coins, &Coin{x: float64((x+2)*TileSize), y: float64((platY-1)*TileSize - 20)})
	}

	// Враги
	for x := 30; x < g.levelW-20; x += 35 {
		g.enemies = append(g.enemies, &Enemy{
			x:     float64(x * TileSize),
			y:     float64(11 * TileSize),
			vx:    -1.0,
			alive: true,
		})
	}
}

func (g *Game) Update() error {
	g.frame++

	if g.state == 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.state = 1
		}
		return nil
	}

	if g.state == 2 || g.state == 3 {
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.player.x = 100
			g.player.y = 300
			g.player.vx = 0
			g.player.vy = 0
			g.player.health = 3
			g.player.coins = 0
			g.player.score = 0
			g.cameraX = 0
			g.generateLevel()
			g.state = 1
		}
		return nil
	}

	g.updatePlayer()
	g.updateCamera()
	g.updateEnemies()
	g.checkCollisions()

	// Победа
	if g.player.x > float64((g.levelW-8)*TileSize) {
		g.state = 3
	}

	return nil
}

func (g *Game) updatePlayer() {
	p := g.player

	// Движение
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = Speed
		p.facing = 1
		p.animFrame++
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -Speed
		p.facing = -1
		p.animFrame++
	} else {
		p.vx = 0
	}

	// Прыжок
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
	}

	// Физика
	p.vy += Gravity
	if p.vy > 10 {
		p.vy = 10
	}

	p.x += p.vx
	p.y += p.vy

	// Коллизии с тайлами
	p.onGround = false
	for _, t := range g.tiles {
		tx := float64(t.x * TileSize)
		ty := float64(t.y * TileSize)

		if p.x < tx+TileSize-5 && p.x+p.width > tx+5 &&
			p.y < ty+TileSize && p.y+p.height > ty {

			if p.vy >= 0 && p.y+p.height <= ty+25 {
				p.y = ty - p.height
				p.vy = 0
				p.onGround = true
			}
		}
	}

	// Границы
	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(g.levelW*TileSize)-p.width {
		p.x = float64(g.levelW*TileSize) - p.width
	}

	// Смерть
	if p.y > ScreenHeight {
		p.health = 0
	}

	if p.health <= 0 {
		p.lives--
		if p.lives > 0 {
			p.health = 3
			p.x = 100
			p.y = 300
		} else {
			g.state = 2
		}
	}
}

func (g *Game) updateCamera() {
	targetX := g.player.x - ScreenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}
	maxX := float64(g.levelW*TileSize) - ScreenWidth
	if g.cameraX > maxX {
		g.cameraX = maxX
	}
}

func (g *Game) updateEnemies() {
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		e.x += e.vx
		e.frame++
		if e.frame%120 == 0 {
			e.vx *= -1
		}
	}
}

func (g *Game) checkCollisions() {
	p := g.player

	// Монеты
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, c.x-16, c.y-16, 32, 32) {
			c.collected = true
			p.coins++
			p.score += 100
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, e.x, e.y, 40, 40) {
			if p.vy > 0 && p.y+p.height < e.y+20 {
				// Прыгнул сверху
				e.alive = false
				p.vy = -6
				p.score += 200
			} else {
				// Получил урон
				p.health--
				p.vy = -5
				if p.x < e.x {
					p.vx = -8
				} else {
					p.vx = 8
				}
			}
		}
	}
}

func rectOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка - голубое небо
	screen.Fill(color.RGBA{100, 149, 237, 255})

	if g.state == 0 {
		g.drawMenu(screen)
		return
	}

	g.drawGame(screen)

	if g.state == 2 {
		g.drawGameOver(screen)
	} else if g.state == 3 {
		g.drawWin(screen)
	}
}

func (g *Game) drawGame(screen *ebiten.Image) {
	camX := g.cameraX

	// Фон (параллакс)
	if sprites["bg"] != nil {
		bgX := -camX * 0.3
		for x := bgX; x < ScreenWidth; x += 800 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, 0)
			screen.DrawImage(sprites["bg"], op)
		}
	}

	// Облака
	for i := 0; i < 10; i++ {
		x := float32((i*100 - int(camX*0.2)) % (ScreenWidth + 100))
		if x < -100 {
			x += ScreenWidth + 100
		}
		y := float32(50 + i%3*30)
		if sprites["cloud"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			screen.DrawImage(sprites["cloud"], op)
		}
	}

	// Тайлы
	for _, t := range g.tiles {
		tx := float64(t.x*TileSize) - camX
		ty := float64(t.y * TileSize)

		if tx < -TileSize || tx > ScreenWidth {
			continue
		}

		var img *ebiten.Image
		switch t.id {
		case 1:
			img = sprites["grass"]
		case 2:
			img = sprites["dirt"]
		case 3:
			img = sprites["brick"]
		case 4:
			img = sprites["box"]
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, ty)
			screen.DrawImage(img, op)
		}
	}

	// Кусты
	for x := 5; x < g.levelW; x += 30 {
		tx := float64(x*TileSize) - camX
		if tx > -50 && tx < ScreenWidth+50 {
			if sprites["bush"] != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(tx, float64(11*TileSize-32))
				screen.DrawImage(sprites["bush"], op)
			}
		}
	}

	// Монеты
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		cx := c.x - camX
		if cx < -32 || cx > ScreenWidth+32 {
			continue
		}
		if sprites["coin"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cx, c.y)
			screen.DrawImage(sprites["coin"], op)
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		ex := e.x - camX
		if ex < -50 || ex > ScreenWidth+50 {
			continue
		}

		var img *ebiten.Image
		if (e.frame / 10) % 2 == 0 {
			img = sprites["enemy1"]
		} else {
			img = sprites["enemy2"]
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(ex, e.y)
			if e.vx > 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(40, 0)
			}
			screen.DrawImage(img, op)
		}
	}

	// Игрок
	p := g.player
	px := p.x - camX

	var playerImg *ebiten.Image
	if !p.onGround {
		playerImg = sprites["player_jump"]
	} else if p.vx != 0 {
		if (p.animFrame / 10) % 2 == 0 {
			playerImg = sprites["player_walk1"]
		} else {
			playerImg = sprites["player_walk2"]
		}
	} else {
		playerImg = sprites["player"]
	}

	if playerImg != nil {
		op := &ebiten.DrawImageOptions{}
		if p.facing < 0 {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(p.width, 0)
		}
		op.GeoM.Translate(px, p.y)
		screen.DrawImage(playerImg, op)
	}

	// Флаг в конце
	flagX := float64((g.levelW-8)*TileSize) - camX
	if flagX > -50 && flagX < ScreenWidth+50 {
		vector.DrawFilledRect(screen, float32(flagX), float32(4*TileSize), 5, 8*TileSize, color.RGBA{139, 90, 43, 255}, false)
		if sprites["flag"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(flagX+5, float64(4*TileSize))
			screen.DrawImage(sprites["flag"], op)
		}
	}

	// UI
	g.drawUI(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 180}, false)

	p := g.player
	text.Draw(screen, fmt.Sprintf("SCORE: %06d", p.score), basicfont.Face7x13, 10, 25, color.White)
	text.Draw(screen, fmt.Sprintf("COINS: %d", p.coins), basicfont.Face7x13, 200, 25, color.RGBA{255, 215, 0, 255})

	hearts := ""
	for i := 0; i < p.health; i++ {
		hearts += "❤"
	}
	text.Draw(screen, "LIVES: "+hearts, basicfont.Face7x13, 400, 25, color.RGBA{255, 100, 100, 255})
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	text.Draw(screen, "HERO ADVENTURE", basicfont.Face7x13, ScreenWidth/2-60, 200, color.White)
	text.Draw(screen, "v1.0.0", basicfont.Face7x13, ScreenWidth/2-20, 240, color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-65, 350, color.White)

	controls := []string{
		"Controls:",
		"Arrow Keys / WASD - Move",
		"Space / Up - Jump",
		"Jump on enemies to defeat them!",
		"Reach the flag to win!",
	}
	for i, line := range controls {
		text.Draw(screen, line, basicfont.Face7x13, ScreenWidth/2-80, 420+i*25, color.RGBA{200, 200, 200, 255})
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, ScreenWidth/2-40, 300, color.RGBA{255, 50, 50, 255})
	text.Draw(screen, fmt.Sprintf("Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-40, 400, color.White)
	text.Draw(screen, "Press ENTER to retry", basicfont.Face7x13, ScreenWidth/2-55, 500, color.White)
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "YOU WIN!", basicfont.Face7x13, ScreenWidth/2-35, 300, color.RGBA{255, 215, 0, 255})
	text.Draw(screen, fmt.Sprintf("Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-40, 400, color.White)
	text.Draw(screen, "Press ENTER to play again", basicfont.Face7x13, ScreenWidth/2-70, 500, color.White)
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	log.Println("🎮 HERO ADVENTURE v1.0.0")
	log.Println("Загрузка спрайтов...")

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("HERO ADVENTURE v1.0.0 | Go365 Day 87")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
