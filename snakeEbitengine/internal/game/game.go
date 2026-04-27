package game

import (
	"image/color"
	"log"
	"math"
	mathrand "math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	ebitenAudio "github.com/hajimehoshi/ebiten/v2/audio"
	"golang.org/x/image/font"

	myaudio "snake/internal/audio"
	"snake/internal/entities"
	"snake/internal/settings"
	"snake/internal/assets"
	"snake/internal/types"


)

type Game struct {
	rng   *mathrand.Rand
	state types.GameState

	snake          []types.Vec
	dir            types.Vec
	nextDir        types.Vec
	ticker         float64
	speed          float64
	fruitX, fruitY int
	fruitType      int
	bombs          []types.Bomb
	ice            types.IceBlock
	iceActive      bool
	frozenTimer    float64
	score          int
	health         int
	particles      []types.Particle

	audioCtx       *ebitenAudio.Context
	sndEat         *ebitenAudio.Player
	sndBoom        *ebitenAudio.Player
	sndHeal        *ebitenAudio.Player
	sndPause       *ebitenAudio.Player
	sndMenuMove    *ebitenAudio.Player
	sndMenuSelect  *ebitenAudio.Player
	sndGhost       *ebitenAudio.Player
	sndKeyCollect  *ebitenAudio.Player
	sndKeyUse      *ebitenAudio.Player
	sndGiftOpen    *ebitenAudio.Player
	sndCoin        *ebitenAudio.Player

	shake          float64
	menuPulse      float64
	pauseCooldown  float64
	menuSelected   int
	menuButtons    []string
	buttonFlash    int
	fontFace       font.Face

    assets *assets.AssetManager


	ghostFrames      []*ebiten.Image
	ghostFrameIdx    int
	ghostAnimTimer   float64
	ghostActive      bool
	ghostX, ghostY   int
	ghostMoveTimer   float64
	ghostModeTimer   float64

	roachFrames    []*ebiten.Image
	roachFrameIdx  int
	roachActive    bool
	roachX, roachY int
	roachMoveTimer float64

	vikingFrames     []*ebiten.Image
	vikingList       []types.Viking
	vikingSpawnTimer float64

	gifts          []*types.Gift
	giftClosedImgs []*ebiten.Image
	giftOpenFrames []*ebiten.Image
	coins          []types.Coin
	coinCount      int
	coinFrames     []*ebiten.Image
	keysCollected  int
	carryingKey    bool
	keySpawnTimer  float64
	keyOnField     types.KeyOnField
	keyImg         *ebiten.Image

	settingsVolumeSlider   float64
	settingsLanguageIndex  int
	settingsDifficultyIndex int
	settingsAnimations     bool
	settingsSliderGrabbed  bool
	settingsSelected       int
}

func NewGame() *Game {
	g := &Game{
		rng:              mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:            types.STATE_MENU,
		speed:            9,
		health:           types.MaxHealth,
		menuPulse:        0,
		menuSelected:     0,
		menuButtons:      []string{"Начать игру", "Продолжить", "Новая игра", "Настройки", "Выйти из игры"},
		iceActive:        false,
		frozenTimer:      0,
		ghostActive:      false,
		ghostModeTimer:   0,
		ghostFrameIdx:    0,
		ghostAnimTimer:   0,
		roachActive:      false,
		vikingSpawnTimer: 0,
		keySpawnTimer:    5.0,
		keysCollected:    0,
		carryingKey:      false,
		coinCount:        0,
		keyOnField:       types.KeyOnField{Active: false},
		settingsVolumeSlider: 0.7,
		settingsLanguageIndex: 0,
		settingsDifficultyIndex: 1,
		settingsAnimations: true,
		settingsSelected:   0,
	}

	settings.Load()
	if settings.Current.Language == "ru" {
		g.settingsLanguageIndex = 0
	} else {
		g.settingsLanguageIndex = 1
	}
	switch settings.Current.Difficulty {
	case "easy":
		g.settingsDifficultyIndex = 0
	case "normal":
		g.settingsDifficultyIndex = 1
	case "hard":
		g.settingsDifficultyIndex = 2
	}
	g.settingsVolumeSlider = settings.Current.Volume
	g.settingsAnimations = settings.Current.BackgroundAnimation

	g.reset()
	g.initAudio()

	if err := g.loadFont(); err != nil {
		log.Printf("Шрифт не загружен: %v", err)
	}

    g.assets = assets.New()
	g.createGifts()
	return g
}

func (g *Game) loadFont() error {
	face, err := entities.LoadFont("assets/fonts/font.ttf", 24)
	if err != nil {
		return err
	}
	g.fontFace = face
	return nil
}

func (g *Game) initAudio() {
	g.audioCtx = ebitenAudio.NewContext(44100)
	g.sndEat = myaudio.NewPlayer(g.audioCtx, myaudio.SndEat())
	g.sndBoom = myaudio.NewPlayer(g.audioCtx, myaudio.SndBoom())
	g.sndHeal = myaudio.NewPlayer(g.audioCtx, myaudio.SndHeal())
	g.sndPause = myaudio.NewPlayer(g.audioCtx, myaudio.SndPause())
	g.sndMenuMove = myaudio.NewPlayer(g.audioCtx, myaudio.SndMenuMove())
	g.sndMenuSelect = myaudio.NewPlayer(g.audioCtx, myaudio.SndMenuSelect())
	g.sndGhost = myaudio.NewPlayer(g.audioCtx, myaudio.SndGhost())
	g.sndKeyCollect = myaudio.NewPlayer(g.audioCtx, myaudio.SndKey())
	g.sndKeyUse = myaudio.NewPlayer(g.audioCtx, myaudio.SndKeyUse())
	g.sndGiftOpen = myaudio.NewPlayer(g.audioCtx, myaudio.SndGiftOpen())
	g.sndCoin = myaudio.NewPlayer(g.audioCtx, myaudio.SndCoin())

	g.applySettings()
}

func (g *Game) applySettings() {
	vol := settings.Current.Volume
	g.sndEat.SetVolume(vol)
	g.sndBoom.SetVolume(vol)
	g.sndHeal.SetVolume(vol)
	g.sndPause.SetVolume(vol)
	g.sndMenuMove.SetVolume(vol)
	g.sndMenuSelect.SetVolume(vol)
	g.sndGhost.SetVolume(vol)
	g.sndKeyCollect.SetVolume(vol)
	g.sndKeyUse.SetVolume(vol)
	g.sndGiftOpen.SetVolume(vol)
	g.sndCoin.SetVolume(vol)

	switch settings.Current.Difficulty {
	case "easy":
		g.speed = 6
	case "normal":
		g.speed = 9
	case "hard":
		g.speed = 12
	}
}

func (g *Game) loadImages() {
    if g.assets == nil {
        return
    }
    // Assign sprite frames and images from AssetManager
    g.giftClosedImgs = g.assets.GiftClosedImgs
    g.giftOpenFrames = g.assets.GiftOpenFrames
    g.coinFrames = g.assets.CoinFrames
    g.keyImg = g.assets.KeyImg
    // Optional: assign other animation frames if needed elsewhere
    g.ghostFrames = g.assets.GhostFrames
    g.roachFrames = g.assets.RoachFrames
    g.vikingFrames = g.assets.VikingFrames
    // Fruit images can be accessed via assets directly in rendering; no need to store them here.
}



func (g *Game) createGifts() {
	g.gifts = nil
	numGifts := 6
	for i := 0; i < numGifts; i++ {
		for tries := 0; tries < 500; tries++ {
			x := g.rng.Intn(types.GridW)
			y := g.rng.Intn(types.GridH)
			if !g.isCellOccupied(x, y) {
				g.gifts = append(g.gifts, &types.Gift{
					X:      x,
					Y:      y,
					Color:  g.rng.Intn(6),
					Opened: false,
					Life:   0,
				})
				break
			}
		}
	}
}

func (g *Game) reset() {
	g.snake = nil
	cx, cy := types.GridW/2, types.GridH/2
	for i := 0; i < types.InitialLength; i++ {
		g.snake = append(g.snake, types.Vec{cx - i, cy})
	}
	g.dir = types.Vec{1, 0}
	g.nextDir = g.dir
	g.health = types.MaxHealth
	g.placeFruit()
	g.bombs = nil
	g.iceActive = false
	g.frozenTimer = 0
	g.score = 0
	g.ticker = 0
	g.shake = 0
	g.particles = nil
	g.ghostActive = false
	g.ghostModeTimer = 0
	g.ghostFrameIdx = 0
	g.ghostAnimTimer = 0
	if g.roachFrames != nil {
		g.roachActive = true
		g.roachX = g.rng.Intn(types.GridW)
		g.roachY = g.rng.Intn(types.GridH)
		g.roachMoveTimer = 0.5
		g.roachFrameIdx = 0
	}
	g.vikingList = nil
	g.vikingSpawnTimer = 3.0
	g.keysCollected = 0
	g.carryingKey = false
	g.keySpawnTimer = 5.0
	g.keyOnField.Active = false
	g.coins = nil
	g.coinCount = 0
	g.createGifts()
}

func (g *Game) placeFruit() {
	for {
		x := g.rng.Intn(types.GridW)
		y := g.rng.Intn(types.GridH)
		ok := true
		for _, s := range g.snake {
			if s.X == x && s.Y == y {
				ok = false
				break
			}
		}
		for _, b := range g.bombs {
			if b.X == x && b.Y == y {
				ok = false
				break
			}
		}
		if g.iceActive && g.ice.X == x && g.ice.Y == y {
			ok = false
		}
		if ok {
			g.fruitX, g.fruitY = x, y
			g.fruitType = g.rng.Intn(5)
			return
		}
	}
}

func (g *Game) spawnIce() {
	if !g.iceActive && g.rng.Float64() < 0.3 {
		for i := 0; i < 500; i++ {
			x := g.rng.Intn(types.GridW)
			y := g.rng.Intn(types.GridH)
			ok := true
			if x == g.fruitX && y == g.fruitY {
				ok = false
			}
			for _, s := range g.snake {
				if s.X == x && s.Y == y {
					ok = false
					break
				}
			}
			for _, b := range g.bombs {
				if b.X == x && b.Y == y {
					ok = false
					break
				}
			}
			if ok {
				g.ice = types.IceBlock{X: x, Y: y}
				g.iceActive = true
				return
			}
		}
	}
}

func (g *Game) spawnGhost() {
	if !g.ghostActive && g.rng.Float64() < 0.2 {
		for i := 0; i < 200; i++ {
			x := g.rng.Intn(types.GridW)
			y := g.rng.Intn(types.GridH)
			ok := true
			if x == g.fruitX && y == g.fruitY {
				ok = false
			}
			for _, s := range g.snake {
				if s.X == x && s.Y == y {
					ok = false
					break
				}
			}
			for _, b := range g.bombs {
				if b.X == x && b.Y == y {
					ok = false
					break
				}
			}
			if g.iceActive && g.ice.X == x && g.ice.Y == y {
				ok = false
			}
			if ok {
				g.ghostActive = true
				g.ghostX, g.ghostY = x, y
				g.ghostMoveTimer = 0.5
				g.ghostFrameIdx = 0
				g.ghostAnimTimer = 0
				return
			}
		}
	}
}

func (g *Game) spawnViking() {
	if len(g.vikingList) >= 3 {
		return
	}
	for i := 0; i < 200; i++ {
		x := g.rng.Intn(types.GridW)
		y := g.rng.Intn(types.GridH)
		if !g.isCellOccupied(x, y) {
			g.vikingList = append(g.vikingList, types.Viking{
				X:      x,
				Y:      y,
				Frame:  0,
				Timer:  0,
				Active: true,
			})
			return
		}
	}
}

func (g *Game) spawnKeyOnField() {
	if g.keyOnField.Active {
		return
	}
	for i := 0; i < 200; i++ {
		x := g.rng.Intn(types.GridW)
		y := g.rng.Intn(types.GridH)
		if !g.isCellOccupied(x, y) {
			g.keyOnField.Active = true
			g.keyOnField.X = x
			g.keyOnField.Y = y
			g.keyOnField.Life = 5.0
			return
		}
	}
}

func (g *Game) collectKeyFromField() {
	if !g.keyOnField.Active {
		return
	}
	head := g.snake[0]
	if head.X == g.keyOnField.X && head.Y == g.keyOnField.Y {
		g.keysCollected++
		g.keyOnField.Active = false
		g.sndKeyCollect.Rewind()
		g.sndKeyCollect.Play()
		g.addParticles(float64(head.X*types.TileSize+types.TileSize/2), float64(head.Y*types.TileSize+types.TileSize/2), 40, color.RGBA{255, 215, 0, 255}, true)
	}
}

func (g *Game) useKey() {
	if g.keysCollected > 0 && !g.carryingKey {
		g.keysCollected--
		g.carryingKey = true
		g.sndKeyUse.Rewind()
		g.sndKeyUse.Play()
		head := g.snake[0]
		g.addParticles(float64(head.X*types.TileSize+types.TileSize/2), float64(head.Y*types.TileSize+types.TileSize/2), 40, color.RGBA{255, 200, 50, 255}, true)
	}
}

func (g *Game) openGift(gift *types.Gift) {
	if gift.Opened {
		return
	}
	if !g.carryingKey {
		return
	}
	if g.snake[0].X != gift.X || g.snake[0].Y != gift.Y {
		return
	}
	g.carryingKey = false
	gift.Opened = true
	gift.Life = 5.0
	g.sndGiftOpen.Rewind()
	g.sndGiftOpen.Play()
	g.addParticles(float64(gift.X*types.TileSize+types.TileSize/2), float64(gift.Y*types.TileSize+types.TileSize/2), 80, color.RGBA{255, 200, 100, 255}, true)

	coinCount := g.rng.Intn(5) + 3
	for i := 0; i < coinCount; i++ {
		dx := g.rng.Intn(3) - 1
		dy := g.rng.Intn(3) - 1
		nx := gift.X + dx
		ny := gift.Y + dy
		if nx < 0 {
			nx = 0
		}
		if nx >= types.GridW {
			nx = types.GridW - 1
		}
		if ny < 0 {
			ny = 0
		}
		if ny >= types.GridH {
			ny = types.GridH - 1
		}
		g.coins = append(g.coins, types.Coin{X: nx, Y: ny, Frame: 0, Timer: 0, Life: 5.0})
	}
}

func (g *Game) collectCoin(coinIdx int) {
	g.coinCount++
	g.sndCoin.Rewind()
	g.sndCoin.Play()
	coin := g.coins[coinIdx]
	g.addParticles(float64(coin.X*types.TileSize+types.TileSize/2), float64(coin.Y*types.TileSize+types.TileSize/2), 20, color.RGBA{255, 215, 0, 180}, true)
	g.coins = append(g.coins[:coinIdx], g.coins[coinIdx+1:]...)
}

func (g *Game) isCellOccupied(x, y int) bool {
	for _, s := range g.snake {
		if s.X == x && s.Y == y {
			return true
		}
	}
	for _, b := range g.bombs {
		if b.X == x && b.Y == y {
			return true
		}
	}
	if g.iceActive && g.ice.X == x && g.ice.Y == y {
		return true
	}
	if g.ghostActive && g.ghostX == x && g.ghostY == y {
		return true
	}
	for _, v := range g.vikingList {
		if v.X == x && v.Y == y {
			return true
		}
	}
	for _, gf := range g.gifts {
		if !gf.Opened && gf.X == x && gf.Y == y {
			return true
		}
	}
	for _, c := range g.coins {
		if c.X == x && c.Y == y {
			return true
		}
	}
	if g.keyOnField.Active && g.keyOnField.X == x && g.keyOnField.Y == y {
		return true
	}
	return false
}

func (g *Game) ghostModeActive() bool {
	return g.ghostModeTimer > 0
}

func (g *Game) triggerExplosion(v types.Vec, fatal bool) {
	if fatal {
		g.health = 0
	}
	g.shake = 18
	g.sndBoom.Rewind()
	g.sndBoom.Play()
	g.addParticles(float64(v.X*types.TileSize+types.TileSize/2), float64(v.Y*types.TileSize+types.TileSize/2), 80, color.RGBA{255, 120, 30, 255}, true)
	g.addParticles(float64(v.X*types.TileSize+types.TileSize/2), float64(v.Y*types.TileSize+types.TileSize/2), 40, color.RGBA{255, 255, 200, 200}, true)
	if fatal {
		g.state = types.STATE_GAMEOVER
	}
}

func (g *Game) bombExplode(idx int) {
	b := g.bombs[idx]
	g.bombs = append(g.bombs[:idx], g.bombs[idx+1:]...)
	head := g.snake[0]
	dx := math.Abs(float64(head.X - b.X))
	dy := math.Abs(float64(head.Y - b.Y))
	dist := dx + dy
	g.shake = 12
	g.sndBoom.Rewind()
	g.sndBoom.Play()
	cx := float64(b.X*types.TileSize + types.TileSize/2)
	cy := float64(b.Y*types.TileSize + types.TileSize/2)
	if dist <= 1.5 {
		g.health -= 25
		g.addParticles(cx, cy, 120, color.RGBA{255, 60, 30, 255}, true)
		g.addParticles(cx, cy, 60, color.RGBA{255, 200, 50, 200}, true)
		if g.health <= 0 {
			g.state = types.STATE_GAMEOVER
		}
	} else {
		g.addParticles(cx, cy, 80, color.RGBA{255, 100, 30, 255}, true)
	}
	g.addParticles(cx, cy, 30, color.RGBA{255, 200, 100, 200}, false)
}

func (g *Game) spawnBombRandom() {
	if g.rng.Float64() < 0.4 {
		for i := 0; i < 2000; i++ {
			x := g.rng.Intn(types.GridW)
			y := g.rng.Intn(types.GridH)
			ok := true

			if x == g.fruitX && y == g.fruitY {
				ok = false
			}
			for _, s := range g.snake {
				if s.X == x && s.Y == y {
					ok = false
					break
				}
			}
			for _, b := range g.bombs {
				if b.X == x && b.Y == y {
					ok = false
					break
				}
			}
			if g.iceActive && g.ice.X == x && g.ice.Y == y {
				ok = false
			}
			if ok {
				g.bombs = append(g.bombs, types.Bomb{X: x, Y: y, Timer: 5.0})
				return
			}
		}
	}
}

func (g *Game) addParticles(x, y float64, n int, c color.RGBA, glow bool) {
	for i := 0; i < n; i++ {
		a := g.rng.Float64() * 2 * math.Pi
		s := g.rng.Float64()*4 + 1.5
		g.particles = append(g.particles, types.Particle{
			X:    x,
			Y:    y,
			VX:   math.Cos(a) * s,
			VY:   math.Sin(a) * s,
			Life: g.rng.Float64()*1.5 + 0.4,
			Color: c,
			Size:  g.rng.Float64()*4 + 2,
			Glow:  glow,
		})
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func inputPressed() bool {
	return ebiten.IsKeyPressed(ebiten.KeyEnter) ||
		ebiten.IsKeyPressed(ebiten.KeySpace) ||
		ebiten.IsKeyPressed(ebiten.KeyUp) ||
		ebiten.IsKeyPressed(ebiten.KeyDown) ||
		ebiten.IsKeyPressed(ebiten.KeyLeft) ||
		ebiten.IsKeyPressed(ebiten.KeyRight)
}

func (g *Game) updateMenuButtonsLanguage() {
	if settings.Current.Language == "ru" {
		g.menuButtons = []string{"Начать игру", "Продолжить", "Новая игра", "Настройки", "Выйти из игры"}
	} else {
		g.menuButtons = []string{"Start game", "Continue", "New game", "Settings", "Exit game"}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return types.ScreenW, types.ScreenH
}