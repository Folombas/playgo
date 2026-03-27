// Go365 Day 87 - PLATFORMER HERO v2.0.0
// Использует ВСЕ спрайты из Platformer Complete Pack!

package main

import (
	"fmt"
	"image"
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
	ScreenWidth  = 1024
	ScreenHeight = 768
	TileSize     = 48
	Gravity      = 0.6
	JumpForce    = -13.0
	Speed        = 5.0
)

// ВСЕ спрайты!
var sprites = make(map[string]*ebiten.Image)
var walkSprites = make([]*ebiten.Image, 2)

func loadAllSprites() {
	log.Println("Загрузка ВСЕХ спрайтов...")

	// ========== ИГРОК - 3 персонажа ==========
	// P1 - Зелёный
	sprites["p1_stand"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_stand.png")
	sprites["p1_jump"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_jump.png")
	sprites["p1_hurt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_hurt.png")
	sprites["p1_front"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_front.png")
	sprites["p1_duck"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p1_duck.png")

	walkSheet, _, _ := ebitenutil.NewImageFromFile("assets/sprites/mario/walk/p1_walk.png")
	if walkSheet != nil {
		bounds := walkSheet.Bounds()
		frameW := bounds.Dx() / 2
		for i := 0; i < 2; i++ {
			rect := image.Rect(i*frameW, 0, (i+1)*frameW, bounds.Dy())
			walkSprites[i] = walkSheet.SubImage(rect).(*ebiten.Image)
		}
	}

	// P2 - Красный (для 2 игрока или скинов)
	sprites["p2_stand"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p2_stand.png")
	sprites["p2_jump"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p2_jump.png")
	sprites["p2_hurt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p2_hurt.png")

	// P3 - Синий
	sprites["p3_stand"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p3_stand.png")
	sprites["p3_jump"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/mario/p3_jump.png")

	// ========== ВРАГИ - ВСЕ 6 типов ==========
	// Slime
	sprites["slime1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk1.png")
	sprites["slime2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeWalk2.png")
	sprites["slimeDead"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/slimeDead.png")

	// Fly
	sprites["fly1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly1.png")
	sprites["fly2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyFly2.png")
	sprites["flyDead"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/flyDead.png")

	// Fish
	sprites["fish1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/fishSwim1.png")
	sprites["fish2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/fishSwim2.png")

	// Snail
	sprites["snail1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/snailWalk1.png")
	sprites["snail2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/snailWalk2.png")
	sprites["snailShell"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/snailShell.png")

	// Blocker
	sprites["blocker"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/blockerBody.png")
	sprites["blockerMad"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/blockerMad.png")

	// Poker
	sprites["pokerMad"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/pokerMad.png")
	sprites["pokerSad"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/enemies/pokerSad.png")

	// ========== ТАЙЛЫ - ВСЕ ==========
	// Grass
	sprites["grass"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassMid.png")
	sprites["grassLeft"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassLeft.png")
	sprites["grassRight"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassRight.png")
	sprites["grassHill"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/grassHillLeft.png")

	// Dirt
	sprites["dirt"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirt.png")
	sprites["dirtMid"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/dirtMid.png")

	// Brick
	sprites["brick"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/brickWall.png")

	// Stone
	sprites["stone"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/stone.png")
	sprites["stoneMid"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/stoneMid.png")

	// Boxes
	sprites["boxCoin"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxCoin.png")
	sprites["boxItem"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxItem.png")
	sprites["boxEmpty"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxEmpty.png")
	sprites["boxWarning"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/boxWarning.png")

	// Spikes
	sprites["spikes"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/tiles/spikes.png")

	// ========== ПРЕДМЕТЫ - ВСЕ ==========
	// Coins - 3 типа
	sprites["coinGold"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/coinGold.png")
	sprites["coinSilver"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/coinSilver.png")
	sprites["coinBronze"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/coinBronze.png")

	// Gems - 4 типа
	sprites["gemRed"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemRed.png")
	sprites["gemBlue"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemBlue.png")
	sprites["gemGreen"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemGreen.png")
	sprites["gemYellow"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/gemYellow.png")

	// Flags - 4 цвета
	sprites["flagGreen"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/flagGreen.png")
	sprites["flagRed"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/flagRed.png")
	sprites["flagBlue"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/flagBlue.png")
	sprites["flagYellow"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/flagYellow.png")

	// Power-ups
	sprites["mushroomRed"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/mushroomRed.png")
	sprites["mushroomBrown"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/mushroomBrown.png")
	sprites["star"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/star.png")

	// Keys
	sprites["keyRed"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/keyRed.png")
	sprites["keyBlue"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/keyBlue.png")
	sprites["keyGreen"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/keyGreen.png")
	sprites["keyYellow"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/keyYellow.png")

	// Other
	sprites["bomb"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/bomb.png")
	sprites["fireball"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/items/fireball.png")

	// ========== ДЕКОРАЦИИ ==========
	// Clouds - 3 типа
	sprites["cloud1"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/cloud1.png")
	sprites["cloud2"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/cloud2.png")
	sprites["cloud3"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/cloud3.png")

	// Bush & Plants
	sprites["bush"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/bush.png")
	sprites["plant"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/plant.png")
	sprites["plantPurple"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/plantPurple.png")
	sprites["cactus"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/cactus.png")

	// Rocks
	sprites["rock"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/decorations/rock.png")

	// ========== ФОНЫ ==========
	sprites["bg"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/backgrounds/bg.png")
	sprites["bg_castle"], _, _ = ebitenutil.NewImageFromFile("assets/sprites/backgrounds/bg_castle.png")

	log.Println("✅ ВСЕ спрайты загружены!")
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
	character int // 0=p1, 1=p2, 2=p3
	coins     int
	gems      int
	score     int
	lives     int
	health    int
	maxHealth int
	hasKey    bool
}

// Тайл
type Tile struct {
	x, y int
	id   int
}

// Враг
type Enemy struct {
	x, y    float64
	vx, vy  float64
	etype   int // 0=slime, 1=fly, 2=fish, 3=snail, 4=blocker, 5=poker
	alive   bool
	frame   int
	health  int
}

// Монета
type Coin struct {
	x, y  float64
	ctype int // 0=bronze, 1=silver, 2=gold
	taken bool
}

// Gem
type Gem struct {
	x, y  float64
	gtype int // 0=red, 1=blue, 2=green, 3=yellow
	taken bool
}

// Key
type Key struct {
	x, y    float64
	ktype   int // 0=red, 1=blue, 2=green, 3=yellow
	collected bool
}

// Игра
type Game struct {
	player  *Player
	tiles   []*Tile
	enemies []*Enemy
	coins   []*Coin
	gems    []*Gem
	keys    []*Key
	cameraX float64
	levelW  int
	frame   int
	state   int // 0=menu, 1=playing, 2=gameover, 3=win
	level   int
	cloudOffset float64 // Для движущихся облаков
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	loadAllSprites()

	g := &Game{
		player: &Player{
			x:         100,
			y:         300,
			width:     40,
			height:    48,
			facing:    1,
			character: 0,
			lives:     3,
			health:    3,
			maxHealth: 3,
		},
		state:  0,
		levelW: 120,
		level:  1,
	}
	g.generateLevel()
	return g
}

func (g *Game) generateLevel() {
	g.tiles = nil
	g.enemies = nil
	g.coins = nil
	g.gems = nil
	g.keys = nil

	// Земля с травой
	for x := 0; x < g.levelW; x++ {
		// Ямы
		if x > 15 && x < g.levelW-15 && x%30 >= 25 && x%30 < 28 {
			continue
		}

		// Разная высота
		groundY := 12
		if x > 50 && x < 80 {
			groundY = 11 // Выше
		}

		g.tiles = append(g.tiles, &Tile{x: x, y: groundY, id: 1})   // grass
		g.tiles = append(g.tiles, &Tile{x: x, y: groundY + 1, id: 4}) // dirt
		g.tiles = append(g.tiles, &Tile{x: x, y: groundY + 2, id: 5}) // dirt
		g.tiles = append(g.tiles, &Tile{x: x, y: groundY + 3, id: 5}) // dirt - ещё ряд!
		g.tiles = append(g.tiles, &Tile{x: x, y: groundY + 4, id: 5}) // dirt - ещё ряд!
	}

	// Платформы из разных тайлов
	for x := 20; x < g.levelW-20; x += 25 {
		platY := 7
		platLen := 4 + rand.Intn(3)
		tileType := 2 // brick
		if rand.Float32() < 0.3 {
			tileType = 6 // stone
		}

		for i := 0; i < platLen; i++ {
			g.tiles = append(g.tiles, &Tile{x: x + i, y: platY, id: tileType})
		}

		// Блоки с предметами
		if rand.Float32() < 0.5 {
			g.tiles = append(g.tiles, &Tile{x: x + 2, y: platY - 1, id: 7}) // boxCoin
		}

		// Монеты над платформой
		coinType := rand.Intn(3)
		g.coins = append(g.coins, &Coin{
			x:     float64((x+2)*TileSize),
			y:     float64((platY-1)*TileSize - 20),
			ctype: coinType,
		})

		// Кристаллы
		if rand.Float32() < 0.3 {
			g.gems = append(g.gems, &Gem{
				x:     float64((x+3)*TileSize),
				y:     float64((platY-1)*TileSize - 20),
				gtype: rand.Intn(4),
			})
		}
	}

	// Ключи
	for x := 40; x < g.levelW-30; x += 50 {
		g.keys = append(g.keys, &Key{
			x:     float64(x*TileSize),
			y:     float64(8*TileSize),
			ktype: rand.Intn(4),
		})
	}

	// Шипы
	for x := 35; x < g.levelW-20; x += 40 {
		g.tiles = append(g.tiles, &Tile{x: x, y: 11, id: 9}) // spikes
	}

	// ВСЕ типы врагов!
	// Slime
	for x := 30; x < g.levelW-20; x += 35 {
		g.enemies = append(g.enemies, &Enemy{
			x:     float64(x * TileSize),
			y:     float64(11 * TileSize),
			vx:    -1.0,
			etype: 0, // slime
			alive: true,
		})
	}

	// Fly (летают)
	for x := 40; x < g.levelW-30; x += 45 {
		g.enemies = append(g.enemies, &Enemy{
			x:     float64(x * TileSize),
			y:     float64(6 * TileSize),
			vx:    -1.5,
			vy:    0,
			etype: 1, // fly
			alive: true,
		})
	}

	// Fish (в воздухе как декор)
	for x := 60; x < g.levelW-40; x += 50 {
		g.enemies = append(g.enemies, &Enemy{
			x:     float64(x * TileSize),
			y:     float64(5 * TileSize),
			vx:    -0.8,
			etype: 2, // fish
			alive: true,
		})
	}

	// Snail (медленные)
	for x := 50; x < g.levelW-30; x += 55 {
		g.enemies = append(g.enemies, &Enemy{
			x:     float64(x * TileSize),
			y:     float64(11 * TileSize),
			vx:    -0.5,
			etype: 3, // snail
			alive: true,
		})
	}

	// Blocker (большие)
	if g.level > 1 {
		g.enemies = append(g.enemies, &Enemy{
			x:      float64(70 * TileSize),
			y:      float64(10 * TileSize),
			vx:     -0.3,
			etype:  4, // blocker
			alive:  true,
			health: 3,
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
			g.player.health = g.player.maxHealth
			g.player.coins = 0
			g.player.gems = 0
			g.player.score = 0
			g.player.hasKey = false
			g.cameraX = 0
			g.level++
			g.generateLevel()
			g.state = 1
		}
		return nil
	}

	g.updatePlayer()
	g.updateCamera()
	g.updateClouds()
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

	p.vy += Gravity
	if p.vy > 12 {
		p.vy = 12
	}

	p.x += p.vx
	p.y += p.vy

	// Коллизии
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

	if p.x < 0 {
		p.x = 0
	}
	if p.x > float64(g.levelW*TileSize)-p.width {
		p.x = float64(g.levelW*TileSize) - p.width
	}

	if p.y > ScreenHeight {
		p.health = 0
	}

	if p.health <= 0 {
		p.lives--
		if p.lives > 0 {
			p.health = p.maxHealth
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

func (g *Game) updateClouds() {
	// Облака двигаются туда-сюда
	g.cloudOffset += 0.3
	if g.cloudOffset > 200 {
		g.cloudOffset = -200
	}
}

func (g *Game) updateEnemies() {
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		e.x += e.vx
		e.y += e.vy
		e.frame++

		// Разворот
		if e.frame%120 == 0 && e.etype != 1 {
			e.vx *= -1
		}

		// Fish плавает
		if e.etype == 2 {
			e.y += float64(e.frame%60-30) * 0.1
		}
	}
}

func (g *Game) checkCollisions() {
	p := g.player

	// Монеты
	for _, c := range g.coins {
		if c.taken {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, c.x-16, c.y-16, 32, 32) {
			c.taken = true
			p.coins++
			p.score += (c.ctype + 1) * 100
		}
	}

	// Кристаллы
	for _, gm := range g.gems {
		if gm.taken {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, gm.x-16, gm.y-16, 32, 32) {
			gm.taken = true
			p.gems++
			p.score += (gm.gtype + 1) * 200
		}
	}

	// Ключи
	for _, k := range g.keys {
		if k.collected {
			continue
		}
		if rectOverlap(p.x, p.y, p.width, p.height, k.x-16, k.y-16, 32, 32) {
			k.collected = true
			p.hasKey = true
			p.score += 500
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
				p.vy = -7
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
	// Небо
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

	// Фон
	if sprites["bg"] != nil {
		bgX := -camX * 0.2
		for x := bgX; x < ScreenWidth; x += 800 {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(x, 0)
			screen.DrawImage(sprites["bg"], op)
		}
	}

	// Облака (все 3 типа) - ДВИЖУЩИЕСЯ!
	for i := 0; i < 15; i++ {
		baseX := float64(i * 80)
		// Движение туда-сюда
		cloudX := baseX + g.cloudOffset + float64(i%5)*20
		// Зацикливание
		for cloudX < -150 {
			cloudX += ScreenWidth + 300
		}
		for cloudX > ScreenWidth+150 {
			cloudX -= ScreenWidth + 300
		}
		
		y := float32(40 + i%4*25)
		cloudName := fmt.Sprintf("cloud%d", (i%3)+1)
		if sprites[cloudName] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cloudX, float64(y))
			screen.DrawImage(sprites[cloudName], op)
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
			img = sprites["brick"]
		case 3:
			img = sprites["stone"]
		case 4:
			img = sprites["dirt"]
		case 5:
			img = sprites["dirtMid"]
		case 6:
			img = sprites["stoneMid"]
		case 7:
			img = sprites["boxCoin"]
		case 8:
			img = sprites["boxItem"]
		case 9:
			img = sprites["spikes"]
		}

		if img != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, ty)
			screen.DrawImage(img, op)
		}
	}

	// Декорации - кусты и растения
	for x := 10; x < g.levelW; x += 25 {
		tx := float64(x*TileSize) - camX
		if tx > -50 && tx < ScreenWidth+50 {
			if sprites["bush"] != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(tx, float64(11*TileSize-32))
				screen.DrawImage(sprites["bush"], op)
			}
		}
	}

	// Кактусы и растения
	for x := 20; x < g.levelW; x += 40 {
		tx := float64(x*TileSize) - camX
		if sprites["cactus"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, float64(11*TileSize-40))
			screen.DrawImage(sprites["cactus"], op)
		}
	}

	// Камни
	for x := 15; x < g.levelW; x += 50 {
		tx := float64(x*TileSize) - camX
		if sprites["rock"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(tx, float64(11*TileSize-20))
			screen.DrawImage(sprites["rock"], op)
		}
	}

	// Монеты (все 3 типа)
	for _, c := range g.coins {
		if c.taken {
			continue
		}
		cx := c.x - camX
		if cx < -32 || cx > ScreenWidth+32 {
			continue
		}
		coinName := fmt.Sprintf("coin%d", []string{"Bronze", "Silver", "Gold"}[c.ctype])
		if sprites[coinName] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(cx, c.y)
			screen.DrawImage(sprites[coinName], op)
		}
	}

	// Кристаллы (все 4 типа)
	for _, gm := range g.gems {
		if gm.taken {
			continue
		}
		gx := gm.x - camX
		if gx < -32 || gx > ScreenWidth+32 {
			continue
		}
		// Используем gemRed для всех
		if sprites["gemRed"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(gx, gm.y)
			screen.DrawImage(sprites["gemRed"], op)
		}
	}

	// Ключи
	for _, k := range g.keys {
		if k.collected {
			continue
		}
		kx := k.x - camX
		if kx < -32 || kx > ScreenWidth+32 {
			continue
		}
		if sprites["keyRed"] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(kx, k.y)
			screen.DrawImage(sprites["keyRed"], op)
		}
	}

	// ВСЕ враги!
	for _, e := range g.enemies {
		if !e.alive {
			continue
		}
		ex := e.x - camX
		if ex < -50 || ex > ScreenWidth+50 {
			continue
		}

		var img *ebiten.Image
		switch e.etype {
		case 0: // slime
			if (e.frame/10)%2 == 0 {
				img = sprites["slime1"]
			} else {
				img = sprites["slime2"]
			}
		case 1: // fly
			if (e.frame/10)%2 == 0 {
				img = sprites["fly1"]
			} else {
				img = sprites["fly2"]
			}
		case 2: // fish
			if (e.frame/15)%2 == 0 {
				img = sprites["fish1"]
			} else {
				img = sprites["fish2"]
			}
		case 3: // snail
			if (e.frame/20)%2 == 0 {
				img = sprites["snail1"]
			} else {
				img = sprites["snail2"]
			}
		case 4: // blocker
			img = sprites["blocker"]
		case 5: // poker
			img = sprites["pokerMad"]
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
	prefix := "p1"
	if p.character == 1 {
		prefix = "p2"
	} else if p.character == 2 {
		prefix = "p3"
	}

	if !p.onGround {
		playerImg = sprites[prefix+"_jump"]
	} else if p.vx != 0 {
		playerImg = walkSprites[(p.animFrame/8)%2]
	} else {
		playerImg = sprites[prefix+"_stand"]
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

	// Флаг
	flagX := float64((g.levelW-8)*TileSize) - camX
	if flagX > -50 && flagX < ScreenWidth+50 {
		vector.DrawFilledRect(screen, float32(flagX), 4*TileSize, 5, 8*TileSize, color.RGBA{139, 90, 43, 255}, false)
		flagName := "flagGreen"
		if g.level%2 == 0 {
			flagName = "flagRed"
		} else if g.level%3 == 0 {
			flagName = "flagBlue"
		}
		if sprites[flagName] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(flagX+5, 4*TileSize)
			screen.DrawImage(sprites[flagName], op)
		}
	}

	g.drawUI(screen)
}

func (g *Game) drawUI(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 45, color.RGBA{0, 0, 0, 200}, false)

	p := g.player
	text.Draw(screen, fmt.Sprintf("SCORE: %06d", p.score), basicfont.Face7x13, 15, 30, color.White)
	text.Draw(screen, fmt.Sprintf("COINS: %d", p.coins), basicfont.Face7x13, 250, 30, color.RGBA{255, 215, 0, 255})
	text.Draw(screen, fmt.Sprintf("GEMS: %d", p.gems), basicfont.Face7x13, 400, 30, color.RGBA{0, 255, 255, 255})

	hearts := ""
	for i := 0; i < p.health; i++ {
		hearts += "❤"
	}
	text.Draw(screen, "LIVES: "+hearts, basicfont.Face7x13, 550, 30, color.RGBA{255, 100, 100, 255})

	if p.hasKey {
		text.Draw(screen, "KEY: 🗝️", basicfont.Face7x13, 750, 30, color.RGBA{255, 255, 0, 255})
	}

	text.Draw(screen, fmt.Sprintf("LEVEL %d", g.level), basicfont.Face7x13, 900, 30, color.White)
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	text.Draw(screen, "PLATFORMER HERO", basicfont.Face7x13, ScreenWidth/2-70, 150, color.White)
	text.Draw(screen, "v2.0.0 - ALL SPRITES EDITION", basicfont.Face7x13, ScreenWidth/2-95, 190, color.RGBA{200, 200, 200, 255})
	text.Draw(screen, "Press ENTER to start", basicfont.Face7x13, ScreenWidth/2-65, 300, color.White)

	controls := []string{
		"CONTROLS:",
		"Arrow Keys / WASD - Move",
		"Space / Up - Jump",
		"Collect: Coins, Gems, Keys",
		"Jump on enemies!",
		"Reach the flag!",
	}
	for i, line := range controls {
		text.Draw(screen, line, basicfont.Face7x13, ScreenWidth/2-80, 380+i*28, color.RGBA{200, 200, 255, 255})
	}

	text.Draw(screen, "Uses ALL sprites from Platformer Complete Pack!", basicfont.Face7x13, ScreenWidth/2-140, 560, color.RGBA{255, 215, 0, 255})
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "GAME OVER", basicfont.Face7x13, ScreenWidth/2-40, 300, color.RGBA{255, 50, 50, 255})
	text.Draw(screen, fmt.Sprintf("Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-40, 400, color.White)
	text.Draw(screen, "Press ENTER to retry", basicfont.Face7x13, ScreenWidth/2-55, 500, color.White)
}

func (g *Game) drawWin(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, false)
	text.Draw(screen, "LEVEL COMPLETE!", basicfont.Face7x13, ScreenWidth/2-60, 300, color.RGBA{0, 255, 0, 255})
	text.Draw(screen, fmt.Sprintf("Score: %06d", g.player.score), basicfont.Face7x13, ScreenWidth/2-40, 400, color.White)
	text.Draw(screen, "Press ENTER for next level", basicfont.Face7x13, ScreenWidth/2-75, 500, color.White)
}

func (g *Game) Layout(w, h int) (int, int) {
	return ScreenWidth, ScreenHeight
}

func main() {
	log.Println("🎮 PLATFORMER HERO v2.0.0")
	log.Println("🎨 ALL SPRITES EDITION")

	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("PLATFORMER HERO v2.0.0 | ALL SPRITES | Go365 Day 87")

	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
