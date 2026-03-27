// Go365 Day 87 - GO MARIO v2.0.0
// Простая и стабильная версия платформера Mario

package main

import (
	"image/color"
	"log"

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
	TileSize     = 40
	Gravity      = 0.5
	JumpForce    = -10.0
	Speed        = 4.0
)

// Типы тайлов
const (
	TileEmpty = iota
	TileGrass
	TileDirt
	TileBrick
	TileBox
)

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
}

type Tile struct {
	x, y int
	id   int
}

type Enemy struct {
	x, y  float64
	vx    float64
	alive bool
}

type Coin struct {
	x, y      float64
	collected bool
}

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

// Загрузка изображений
func loadImages() (map[string]*ebiten.Image, error) {
	imgs := make(map[string]*ebiten.Image)

	// Игрок
	img, _, err := ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
	if err == nil {
		imgs["player"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_walk/p1_walk.png")
	if err == nil {
		imgs["walk"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")
	if err == nil {
		imgs["jump"] = img
	}

	// Тайлы
	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
	if err == nil {
		imgs["grass"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
	if err == nil {
		imgs["dirt"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")
	if err == nil {
		imgs["brick"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")
	if err == nil {
		imgs["box"] = img
	}

	// Враги
	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
	if err == nil {
		imgs["enemy1"] = img
	}

	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")
	if err == nil {
		imgs["enemy2"] = img
	}

	// Монеты
	img, _, err = ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
	if err == nil {
		imgs["coin"] = img
	}

	return imgs, nil
}

func NewGame() *Game {
	g := &Game{
		player: &Player{
			x:      100,
			y:      300,
			width:  32,
			height: 32,
			facing: 1,
			lives:  3,
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

	// Земля
	for x := 0; x < g.levelW; x++ {
		g.tiles = append(g.tiles, &Tile{x: x, y: 12, id: TileGrass})
		g.tiles = append(g.tiles, &Tile{x: x, y: 13, id: TileDirt})
		g.tiles = append(g.tiles, &Tile{x: x, y: 14, id: TileDirt})
	}

	// Платформы
	for x := 10; x < g.levelW-10; x++ {
		if x%20 < 5 {
			for i := 0; i < 3; i++ {
				g.tiles = append(g.tiles, &Tile{x: x + i, y: 8, id: TileBrick})
			}
			g.coins = append(g.coins, &Coin{x: float64((x+1)*TileSize), y: float64(6*TileSize - 20)})
		}
	}

	// Враги
	for x := 20; x < g.levelW-10; x += 30 {
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

	if g.state == 0 { // Menu
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
			g.player.lives = 3
			g.player.coins = 0
			g.player.score = 0
			g.cameraX = 0
			g.state = 1
		}
		return nil
	}

	// Playing
	p := g.player

	// Управление
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

	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)) && p.onGround {
		p.vy = JumpForce
		p.onGround = false
	}

	// Физика
	p.vy += Gravity
	if p.vy > 8 {
		p.vy = 8
	}

	p.x += p.vx
	p.y += p.vy

	// Коллизии с тайлами
	p.onGround = false
	for _, t := range g.tiles {
		tx := float64(t.x * TileSize)
		ty := float64(t.y * TileSize)

		if p.x < tx+TileSize-4 && p.x+p.width > tx+4 &&
			p.y < ty+TileSize && p.y+p.height > ty {

			if p.vy >= 0 && p.y+p.height <= ty+20 {
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

	// Смерть от падения
	if p.y > ScreenHeight {
		p.lives--
		if p.lives <= 0 {
			g.state = 2
		} else {
			p.x = 100
			p.y = 300
			p.vx = 0
			p.vy = 0
		}
	}

	// Камера
	targetX := p.x - ScreenWidth/2
	g.cameraX += (targetX - g.cameraX) * 0.1
	if g.cameraX < 0 {
		g.cameraX = 0
	}

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

		e.x += e.vx
		if e.x < 0 || e.x > float64(g.levelW*TileSize)-32 {
			e.vx *= -1
		}

		// Коллизия с игроком
		if rectOverlap(p.x, p.y, p.width, p.height, e.x, e.y, 32, 32) {
			if p.vy > 0 && p.y+p.height < e.y+20 {
				e.alive = false
				p.vy = -6
				p.score += 200
			} else {
				p.lives--
				if p.lives <= 0 {
					g.state = 2
				} else {
					p.x = 100
					p.y = 300
					p.vx = 0
					p.vy = 0
				}
			}
		}
	}

	// Победа (флаг в конце)
	if p.x > float64((g.levelW-5)*TileSize) {
		g.state = 3
	}

	return nil
}

func rectOverlap(x1, y1, w1, h1, x2, y2, w2, h2 float64) bool {
	return x1 < x2+w2 && x1+w1 > x2 && y1 < y2+h2 && y1+h1 > y2
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Очистка
	screen.Fill(color.RGBA{100, 149, 237, 255})

	if g.state == 0 {
		g.drawMenu(screen)
		return
	}

	// Отрисовка уровня
	camX := g.cameraX

	// Тайлы
	for _, t := range g.tiles {
		tx := float64(t.x*TileSize) - camX
		ty := float64(t.y * TileSize)

		if tx < -TileSize || tx > ScreenWidth {
			continue
		}

		var img *ebiten.Image
		switch t.id {
		case TileGrass:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
		case TileDirt:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
		case TileBrick:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")
		case TileBox:
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, ty)
			screen.DrawImage(img, op)
		}
	}

	// Монеты
	for _, c := range g.coins {
		if c.collected {
			continue
		}
		cx := c.x - camX
		if cx < -32 || cx > ScreenWidth {
			continue
		}

		img, _, _ := ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cx, c.y)
			screen.DrawImage(img, op)
		}
	}

	// Враги
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		ex := e.x - camX
		if ex < -32 || ex > ScreenWidth {
			continue
		}

		var img *ebiten.Image
		if (g.frame/15)%2 == 0 {
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
		} else {
			img, _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(ex, e.y)
			if e.vx > 0 {
				op.GeoM.Scale(-1, 1)
				op.GeoM.Translate(32, 0)
			}
			screen.DrawImage(img, op)
		}
	}

	// Игрок
	p := g.player
	px := p.x - camX

	var playerImg *ebiten.Image
	if !p.onGround {
		playerImg, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")
	} else if p.vx != 0 {
		if (p.animFrame / 10) % 2 == 0 {
			playerImg, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
		} else {
			playerImg, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_walk/p1_walk.png")
		}
	} else {
		playerImg, _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
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

	// UI
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 40, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "SCORE: "+itoa(p.score), basicfont.Face7x13, 10, 25, color.White)
	text.Draw(screen, "COINS: "+itoa(p.coins), basicfont.Face7x13, 200, 25, color.White)
	text.Draw(screen, "LIVES: "+itoa(p.lives), basicfont.Face7x13, 400, 25, color.White)

	if g.state == 2 {
		g.drawGameOver(screen)
	} else if g.state == 3 {
		g.drawWin(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	text.Draw(screen, "GO MARIO", basicfont.Face7x13, ScreenWidth/2-35, 200, color.White)
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-60, 300, color.White)
	text.Draw(screen, "Arrows/WASD - Move", basicfont.Face7x13, ScreenWidth/2-50, 400, color.White)
	text.Draw(screen, "Space - Jump", basicfont.Face7x13, ScreenWidth/2-40, 420, color.White)
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, ScreenWidth/2-40, 300, color.RGBA{255, 0, 0, 255})
	text.Draw(screen, "Press ENTER to restart", basicfont.Face7x13, ScreenWidth/2-60, 400, color.White)
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "YOU WIN!", basicfont.Face7x13, ScreenWidth/2-35, 300, color.RGBA{255, 215, 0, 255})
	text.Draw(screen, "Press ENTER to menu", basicfont.Face7x13, ScreenWidth/2-60, 400, color.White)
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [12]byte
	j := len(buf)
	for i > 0 {
		j--
		buf[j] = byte(i%10 + '0')
		i /= 10
	}
	if neg {
		j--
		buf[j] = '-'
	}
	return string(buf[j:])
}

func main() {
	log.Println("Go Mario v2.0.0 - Запуск...")

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("Go Mario v2.0.0 | Go365 Day 87")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
