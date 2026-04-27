package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	mathrand "math/rand"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

const (
	screenW       = 1280
	screenH       = 800
	tileSize      = 32
	gridW         = screenW / tileSize
	gridH         = screenH / tileSize
	initialLength = 5
	maxHealth     = 100
)

type Settings struct {
	Volume              float64 `json:"volume"`
	Language            string  `json:"language"`
	Difficulty          string  `json:"difficulty"`
	BackgroundAnimation bool    `json:"background_animation"`
}

var currentSettings = Settings{
	Volume:              0.7,
	Language:            "ru",
	Difficulty:          "normal",
	BackgroundAnimation: true,
}

type Vec struct{ X, Y int }
type Bomb struct {
	X, Y  int
	Timer float64
}
type GameState int

const (
	STATE_MENU GameState = iota
	STATE_PLAYING
	STATE_PAUSED
	STATE_GAMEOVER
	STATE_SETTINGS
)

type Particle struct {
	X, Y   float64
	VX, VY float64
	Life   float64
	Color  color.RGBA
	Size   float64
	Glow   bool
}

const (
	FRUIT_APPLE = iota
	FRUIT_STRAWBERRY
	FRUIT_ORANGE
	FRUIT_BANANA
	FRUIT_PINEAPPLE
)

type IceBlock struct{ X, Y int }
type Viking struct {
	X, Y   int
	Frame  int
	Timer  float64
	Active bool
}
type Gift struct {
	X, Y   int
	Color  int
	Opened bool
	Life   float64
}
type Coin struct {
	X, Y   int
	Frame  int
	Timer  float64
	Life   float64
}
type KeyOnField struct {
	X, Y   int
	Active bool
	Life   float64
}

type Game struct {
	rng   *mathrand.Rand
	state GameState

	snake          []Vec
	dir            Vec
	nextDir        Vec
	ticker         float64
	speed          float64
	fruitX, fruitY int
	fruitType      int
	bombs          []Bomb
	ice            IceBlock
	iceActive      bool
	frozenTimer    float64
	score          int
	health         int
	particles      []Particle

	audioCtx       *audio.Context
	sndEat         *audio.Player
	sndBoom        *audio.Player
	sndHeal        *audio.Player
	sndPause       *audio.Player
	sndMenuMove    *audio.Player
	sndMenuSelect  *audio.Player
	sndGhost       *audio.Player
	sndKeyCollect  *audio.Player
	sndKeyUse      *audio.Player
	sndGiftOpen    *audio.Player
	sndCoin        *audio.Player

	shake          float64
	menuPulse      float64
	pauseCooldown  float64
	menuSelected   int
	menuButtons    []string
	buttonFlash    int
	fontFace       font.Face

	appleImg      *ebiten.Image
	strawberryImg *ebiten.Image
	orangeImg     *ebiten.Image
	bananaImg     *ebiten.Image
	pineappleImg  *ebiten.Image

	ghostFrames     []*ebiten.Image
	ghostFrameIdx   int
	ghostAnimTimer  float64
	ghostActive     bool
	ghostX, ghostY  int
	ghostMoveTimer  float64
	ghostModeTimer  float64

	roachFrames    []*ebiten.Image
	roachFrameIdx  int
	roachActive    bool
	roachX, roachY int
	roachMoveTimer float64

	vikingFrames     []*ebiten.Image
	vikingList       []Viking
	vikingSpawnTimer float64

	gifts          []*Gift
	giftClosedImgs []*ebiten.Image
	giftOpenFrames []*ebiten.Image
	coins          []Coin
	coinCount      int
	coinFrames     []*ebiten.Image
	keysCollected  int
	carryingKey    bool
	keySpawnTimer  float64
	keyOnField     KeyOnField
	keyImg         *ebiten.Image

	settingsVolumeSlider   float64
	settingsLanguageIndex  int
	settingsDifficultyIndex int
	settingsAnimations     bool
	settingsSliderGrabbed  bool
}

func loadSettings() {
	data, err := os.ReadFile("settings.json")
	if err != nil {
		currentSettings.Volume = 0.7
		currentSettings.Language = "ru"
		currentSettings.Difficulty = "normal"
		currentSettings.BackgroundAnimation = true
		saveSettings()
		return
	}
	json.Unmarshal(data, &currentSettings)
}

func saveSettings() {
	data, _ := json.MarshalIndent(currentSettings, "", "  ")
	os.WriteFile("settings.json", data, 0644)
}

func (g *Game) applySettings() {
	vol := currentSettings.Volume
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

	switch currentSettings.Difficulty {
	case "easy":
		g.speed = 6
	case "normal":
		g.speed = 9
	case "hard":
		g.speed = 12
	}
}

func makeColorTransparent(img *ebiten.Image, targetColor color.Color) *ebiten.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rgba := image.NewRGBA(bounds)
	drawImg := ebiten.NewImageFromImage(rgba)
	drawImg.Clear()
	drawImg.DrawImage(img, nil)
	pix := rgba.Pix
	stride := rgba.Stride
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			off := y*stride + x*4
			r, g, b, a := pix[off], pix[off+1], pix[off+2], pix[off+3]
			tr, tg, tb, _ := targetColor.RGBA()
			tr8 := uint8(tr >> 8)
			tg8 := uint8(tg >> 8)
			tb8 := uint8(tb >> 8)
			if absDiff(r, tr8) < 5 && absDiff(g, tg8) < 5 && absDiff(b, tb8) < 5 && a > 0 {
				pix[off+3] = 0
			}
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

func absDiff(a, b uint8) int {
	diff := int(a) - int(b)
	if diff < 0 {
		return -diff
	}
	return diff
}

func loadPNG(path string) (*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return ebiten.NewImageFromImage(img), nil
}

func loadSpriteSheet(path string, frameW, frameH, cols, rows int, removeBg bool, bgColor color.Color) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	sheetW := bounds.Dx()
	sheetH := bounds.Dy()
	var frames []*ebiten.Image
	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			x := col * frameW
			y := row * frameH
			if x+frameW > sheetW || y+frameH > sheetH {
				continue
			}
			subImg := img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(image.Rect(x, y, x+frameW, y+frameH))
			ebImg := ebiten.NewImageFromImage(subImg)
			if removeBg {
				ebImg = makeColorTransparent(ebImg, bgColor)
			}
			frames = append(frames, ebImg)
		}
	}
	return frames, nil
}

func loadSpriteSheetAuto(path string, cols, rows int, removeBg bool, bgColor color.Color) ([]*ebiten.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	frameW := bounds.Dx() / cols
	frameH := bounds.Dy() / rows
	return loadSpriteSheet(path, frameW, frameH, cols, rows, removeBg, bgColor)
}

func NewGame() *Game {
	g := &Game{
		rng:              mathrand.New(mathrand.NewSource(time.Now().UnixNano())),
		state:            STATE_MENU,
		speed:            9,
		health:           maxHealth,
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
		keyOnField:       KeyOnField{Active: false},
		settingsVolumeSlider: 0.7,
		settingsLanguageIndex: 0,
		settingsDifficultyIndex: 1,
		settingsAnimations: true,
	}
	loadSettings()
	if currentSettings.Language == "ru" {
		g.settingsLanguageIndex = 0
	} else {
		g.settingsLanguageIndex = 1
	}
	switch currentSettings.Difficulty {
	case "easy":
		g.settingsDifficultyIndex = 0
	case "normal":
		g.settingsDifficultyIndex = 1
	case "hard":
		g.settingsDifficultyIndex = 2
	}
	g.settingsVolumeSlider = currentSettings.Volume
	g.settingsAnimations = currentSettings.BackgroundAnimation

	g.reset()
	g.audioCtx = audio.NewContext(44100)
	g.sndEat = newSound(g.audioCtx, sndEat())
	g.sndBoom = newSound(g.audioCtx, sndBoom())
	g.sndHeal = newSound(g.audioCtx, sndHeal())
	g.sndPause = newSound(g.audioCtx, sndPause())
	g.sndMenuMove = newSound(g.audioCtx, sndMenuMove())
	g.sndMenuSelect = newSound(g.audioCtx, sndMenuSelect())
	g.sndGhost = newSound(g.audioCtx, sndGhost())
	g.sndKeyCollect = newSound(g.audioCtx, sndKey())
	g.sndKeyUse = newSound(g.audioCtx, sndKeyUse())
	g.sndGiftOpen = newSound(g.audioCtx, sndGiftOpen())
	g.sndCoin = newSound(g.audioCtx, sndCoin())

	g.applySettings()

	if err := g.loadFont(); err != nil {
		log.Printf("Шрифт не загружен: %v", err)
	}

	var err error
	g.appleImg, err = loadPNG("apple.png")
	if err != nil {
		log.Printf("apple.png не загружен: %v", err)
	}
	g.strawberryImg, err = loadPNG("strawberry.png")
	if err != nil {
		log.Printf("strawberry.png не загружен: %v", err)
	}
	g.orangeImg, err = loadPNG("orange.png")
	if err != nil {
		log.Printf("orange.png не загружен: %v", err)
	}
	g.bananaImg, err = loadPNG("banana.png")
	if err != nil {
		log.Printf("banana.png не загружен: %v", err)
	}
	g.pineappleImg, err = loadPNG("pineapple.png")
	if err != nil {
		log.Printf("pineapple.png не загружен: %v", err)
	}

	g.ghostFrames = make([]*ebiten.Image, 11)
	for i := 0; i <= 10; i++ {
		filename := fmt.Sprintf("skeleton-animation_%02d.png", i)
		img, err := loadPNG(filename)
		if err != nil {
			log.Printf("Не удалось загрузить %s: %v", filename, err)
			img = ebiten.NewImage(tileSize, tileSize)
			img.Fill(color.White)
		}
		g.ghostFrames[i] = img
	}

	roachFrames, err := loadSpriteSheet("roach.png", 32, 32, 4, 5, false, nil)
	if err != nil {
		log.Printf("roach.png не загружен: %v", err)
	} else {
		g.roachFrames = roachFrames
		g.roachActive = true
		g.roachX = g.rng.Intn(gridW)
		g.roachY = g.rng.Intn(gridH)
		g.roachMoveTimer = 0.5
	}

	vikingFrames, err := loadSpriteSheetAuto("2204_w053_n004_9_medicharacters_p1_9.jpg", 5, 2, true, color.White)
	if err != nil {
		log.Printf("Не удалось загрузить викингов: %v", err)
	} else {
		g.vikingFrames = vikingFrames
	}

	// Закрытые подарки (6 цветов) – правильные имена
	g.giftClosedImgs = make([]*ebiten.Image, 6)
	letters := []string{"a", "b", "c", "d", "e", "f"}
	for i := 0; i < 6; i++ {
		filename := fmt.Sprintf("gift_01%s.png", letters[i])
		img, err := loadPNG(filename)
		if err != nil {
			log.Printf("Не удалось загрузить %s: %v", filename, err)
			img = ebiten.NewImage(tileSize, tileSize)
			img.Fill(color.RGBA{150, 100, 100, 255})
		}
		g.giftClosedImgs[i] = img
	}

	// Открытые подарки (анимация 6 кадров) – правильные имена
	g.giftOpenFrames = make([]*ebiten.Image, 6)
	for i := 0; i < 6; i++ {
		filename := fmt.Sprintf("giftopen_01%s.png", letters[i])
		img, err := loadPNG(filename)
		if err != nil {
			log.Printf("Не удалось загрузить %s: %v", filename, err)
			img = ebiten.NewImage(tileSize, tileSize)
			img.Fill(color.RGBA{200, 200, 100, 255})
		}
		g.giftOpenFrames[i] = img
	}

	g.coinFrames = make([]*ebiten.Image, 4)
	for i := 0; i < 4; i++ {
		filename := fmt.Sprintf("coin_%02da.png", i+1)
		img, err := loadPNG(filename)
		if err != nil {
			log.Printf("Не удалось загрузить %s: %v", filename, err)
			img = ebiten.NewImage(tileSize, tileSize)
			img.Fill(color.RGBA{255, 215, 0, 255})
		}
		g.coinFrames[i] = img
	}

	g.keyImg, err = loadPNG("key_02d.png")
	if err != nil {
		log.Printf("key_02d.png не загружен: %v", err)
		g.keyImg = ebiten.NewImage(tileSize, tileSize)
		g.keyImg.Fill(color.RGBA{255, 215, 0, 255})
	}

	g.createGifts()
	return g
}

func (g *Game) createGifts() {
	g.gifts = nil
	numGifts := 6
	for i := 0; i < numGifts; i++ {
		for tries := 0; tries < 500; tries++ {
			x := g.rng.Intn(gridW)
			y := g.rng.Intn(gridH)
			if !g.isCellOccupied(x, y) {
				g.gifts = append(g.gifts, &Gift{
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

func (g *Game) loadFont() error {
	data, err := os.ReadFile("font.ttf")
	if err != nil {
		return err
	}
	tt, err := opentype.Parse(data)
	if err != nil {
		return err
	}
	const dpi = 72
	g.fontFace, err = opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    24,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	return err
}

func (g *Game) reset() {
	g.snake = nil
	cx, cy := gridW/2, gridH/2
	for i := 0; i < initialLength; i++ {
		g.snake = append(g.snake, Vec{cx - i, cy})
	}
	g.dir = Vec{1, 0}
	g.nextDir = g.dir
	g.health = maxHealth
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
		g.roachX = g.rng.Intn(gridW)
		g.roachY = g.rng.Intn(gridH)
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
		x := g.rng.Intn(gridW)
		y := g.rng.Intn(gridH)
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
			x := g.rng.Intn(gridW)
			y := g.rng.Intn(gridH)
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
				g.ice = IceBlock{X: x, Y: y}
				g.iceActive = true
				return
			}
		}
	}
}

func (g *Game) spawnGhost() {
	if !g.ghostActive && g.rng.Float64() < 0.2 {
		for i := 0; i < 200; i++ {
			x := g.rng.Intn(gridW)
			y := g.rng.Intn(gridH)
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
		x := g.rng.Intn(gridW)
		y := g.rng.Intn(gridH)
		if !g.isCellOccupied(x, y) {
			g.vikingList = append(g.vikingList, Viking{
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
		x := g.rng.Intn(gridW)
		y := g.rng.Intn(gridH)
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
		g.addParticles(float64(head.X*tileSize+tileSize/2), float64(head.Y*tileSize+tileSize/2), 40, color.RGBA{255, 215, 0, 255}, true)
	}
}

func (g *Game) useKey() {
	if g.keysCollected > 0 && !g.carryingKey {
		g.keysCollected--
		g.carryingKey = true
		g.sndKeyUse.Rewind()
		g.sndKeyUse.Play()
		head := g.snake[0]
		g.addParticles(float64(head.X*tileSize+tileSize/2), float64(head.Y*tileSize+tileSize/2), 40, color.RGBA{255, 200, 50, 255}, true)
	}
}

func (g *Game) openGift(gift *Gift) {
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
	g.addParticles(float64(gift.X*tileSize+tileSize/2), float64(gift.Y*tileSize+tileSize/2), 80, color.RGBA{255, 200, 100, 255}, true)

	coinCount := g.rng.Intn(5) + 3
	for i := 0; i < coinCount; i++ {
		dx := g.rng.Intn(3) - 1
		dy := g.rng.Intn(3) - 1
		nx := gift.X + dx
		ny := gift.Y + dy
		if nx < 0 {
			nx = 0
		}
		if nx >= gridW {
			nx = gridW - 1
		}
		if ny < 0 {
			ny = 0
		}
		if ny >= gridH {
			ny = gridH - 1
		}
		g.coins = append(g.coins, Coin{X: nx, Y: ny, Frame: 0, Timer: 0, Life: 5.0})
	}
}

func (g *Game) collectCoin(coinIdx int) {
	g.coinCount++
	g.sndCoin.Rewind()
	g.sndCoin.Play()
	coin := g.coins[coinIdx]
	g.addParticles(float64(coin.X*tileSize+tileSize/2), float64(coin.Y*tileSize+tileSize/2), 20, color.RGBA{255, 215, 0, 180}, true)
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

// -----------------------------------------------------------------------------
// Основной цикл Update
// -----------------------------------------------------------------------------
func (g *Game) Update() error {
	dt := 1.0 / 60.0
	g.menuPulse += 0.05
	if g.pauseCooldown > 0 {
		g.pauseCooldown -= dt
	}
	if g.buttonFlash > 0 {
		g.buttonFlash--
	}
	if g.frozenTimer > 0 {
		g.frozenTimer -= dt
		if g.frozenTimer < 0 {
			g.frozenTimer = 0
		}
	}
	if g.ghostModeTimer > 0 {
		g.ghostModeTimer -= dt
		if g.ghostModeTimer < 0 {
			g.ghostModeTimer = 0
		}
	}

	// Состояние НАСТРОЙКИ
	if g.state == STATE_SETTINGS {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = STATE_MENU
			g.pauseCooldown = 0.3
			g.sndPause.Rewind()
			g.sndPause.Play()
		}
		// Определяем выбранный параметр (простой вариант: фиксированный порядок)
		// Для простоты используем стрелки для выбора поля, но у нас нет индикатора выбора.
		// Реализуем навигацию по 4 строкам.
		selected := 0
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			selected = (selected - 1 + 4) % 4
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
			g.pauseCooldown = 0.15
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			selected = (selected + 1) % 4
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
			g.pauseCooldown = 0.15
		}
		switch selected {
		case 0: // Громкость
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				g.settingsVolumeSlider -= 0.05
				if g.settingsVolumeSlider < 0 {
					g.settingsVolumeSlider = 0
				}
				currentSettings.Volume = g.settingsVolumeSlider
				g.applySettings()
				saveSettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsVolumeSlider += 0.05
				if g.settingsVolumeSlider > 1 {
					g.settingsVolumeSlider = 1
				}
				currentSettings.Volume = g.settingsVolumeSlider
				g.applySettings()
				saveSettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 1: // Язык
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsLanguageIndex = (g.settingsLanguageIndex + 1) % 2
				if g.settingsLanguageIndex == 0 {
					currentSettings.Language = "ru"
				} else {
					currentSettings.Language = "en"
				}
				saveSettings()
				g.updateMenuButtonsLanguage()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 2: // Сложность
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				g.settingsDifficultyIndex = (g.settingsDifficultyIndex - 1 + 3) % 3
				switch g.settingsDifficultyIndex {
				case 0:
					currentSettings.Difficulty = "easy"
				case 1:
					currentSettings.Difficulty = "normal"
				case 2:
					currentSettings.Difficulty = "hard"
				}
				saveSettings()
				g.applySettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsDifficultyIndex = (g.settingsDifficultyIndex + 1) % 3
				switch g.settingsDifficultyIndex {
				case 0:
					currentSettings.Difficulty = "easy"
				case 1:
					currentSettings.Difficulty = "normal"
				case 2:
					currentSettings.Difficulty = "hard"
				}
				saveSettings()
				g.applySettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 3: // Фоновые анимации
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsAnimations = !g.settingsAnimations
				currentSettings.BackgroundAnimation = g.settingsAnimations
				saveSettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		}
		return nil
	}

	// Призрак
	if g.ghostActive {
		g.ghostAnimTimer += dt
		if g.ghostAnimTimer >= 0.1 {
			g.ghostAnimTimer = 0
			g.ghostFrameIdx = (g.ghostFrameIdx + 1) % len(g.ghostFrames)
		}
	}
	if g.ghostActive && g.state == STATE_PLAYING {
		g.ghostMoveTimer -= dt
		if g.ghostMoveTimer <= 0 {
			dx, dy := 0, 0
			switch g.rng.Intn(4) {
			case 0:
				dx = 1
			case 1:
				dx = -1
			case 2:
				dy = 1
			case 3:
				dy = -1
			}
			nx, ny := g.ghostX+dx, g.ghostY+dy
			if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH {
				g.ghostX, g.ghostY = nx, ny
			}
			g.ghostMoveTimer = 0.5
		}
	}

	// Таракан (фоновые анимации)
	if g.state == STATE_PLAYING && currentSettings.BackgroundAnimation {
		if g.roachActive {
			g.roachMoveTimer -= dt
			if g.roachMoveTimer <= 0 {
				dx, dy := 0, 0
				switch g.rng.Intn(4) {
				case 0:
					dx = 1
				case 1:
					dx = -1
				case 2:
					dy = 1
				case 3:
					dy = -1
				}
				nx, ny := g.roachX+dx, g.roachY+dy
				if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !g.isCellOccupied(nx, ny) {
					g.roachX, g.roachY = nx, ny
					if len(g.roachFrames) > 0 {
						g.roachFrameIdx = (g.roachFrameIdx + 1) % len(g.roachFrames)
					}
				}
				g.roachMoveTimer = 0.8
			}
		}
	}

	// Викинги (фоновые анимации)
	if g.state == STATE_PLAYING && currentSettings.BackgroundAnimation {
		g.vikingSpawnTimer -= dt
		if g.vikingSpawnTimer <= 0 && len(g.vikingList) < 3 {
			g.spawnViking()
			g.vikingSpawnTimer = 10.0
		}
		for i := 0; i < len(g.vikingList); i++ {
			v := &g.vikingList[i]
			v.Timer += dt
			if v.Timer >= 0.1 {
				v.Timer = 0
				v.Frame = (v.Frame + 1) % len(g.vikingFrames)
			}
			if g.rng.Float64() < 0.02 {
				dx, dy := 0, 0
				switch g.rng.Intn(4) {
				case 0:
					dx = 1
				case 1:
					dx = -1
				case 2:
					dy = 1
				case 3:
					dy = -1
				}
				nx, ny := v.X+dx, v.Y+dy
				if nx >= 0 && nx < gridW && ny >= 0 && ny < gridH && !g.isCellOccupied(nx, ny) {
					v.X, v.Y = nx, ny
				}
			}
			if v.Active && g.snake[0].X == v.X && g.snake[0].Y == v.Y {
				g.health -= 20
				g.addParticles(float64(v.X*tileSize+tileSize/2), float64(v.Y*tileSize+tileSize/2), 40, color.RGBA{200, 50, 50, 255}, true)
				g.sndHeal.Rewind()
				g.sndHeal.Play()
				g.vikingList = append(g.vikingList[:i], g.vikingList[i+1:]...)
				i--
				if g.health <= 0 {
					g.state = STATE_GAMEOVER
				}
			}
		}
	}

	// Ключ на поле
	if g.state == STATE_PLAYING {
		g.keySpawnTimer -= dt
		if g.keySpawnTimer <= 0 && !g.keyOnField.Active {
			g.spawnKeyOnField()
			g.keySpawnTimer = 15.0
		}
		if g.keyOnField.Active {
			g.keyOnField.Life -= dt
			if g.keyOnField.Life <= 0 {
				g.keyOnField.Active = false
			}
		}
		g.collectKeyFromField()
	}

	// Открытые подарки
	for i := 0; i < len(g.gifts); i++ {
		if g.gifts[i].Opened {
			g.gifts[i].Life -= dt
			if g.gifts[i].Life <= 0 {
				g.gifts = append(g.gifts[:i], g.gifts[i+1:]...)
				i--
			}
		}
	}

	// Монетки
	for i := 0; i < len(g.coins); i++ {
		g.coins[i].Life -= dt
		if g.coins[i].Life <= 0 {
			g.coins = append(g.coins[:i], g.coins[i+1:]...)
			i--
			continue
		}
		g.coins[i].Timer += dt
		if g.coins[i].Timer >= 0.1 {
			g.coins[i].Timer = 0
			g.coins[i].Frame = (g.coins[i].Frame + 1) % 4
		}
	}

	// Использование ключа по K
	if g.state == STATE_PLAYING && ebiten.IsKeyPressed(ebiten.KeyK) && g.pauseCooldown <= 0 {
		g.useKey()
		g.pauseCooldown = 0.2
	}

	// Открытие подарка при касании
	if g.state == STATE_PLAYING && g.carryingKey {
		head := g.snake[0]
		for _, gift := range g.gifts {
			if !gift.Opened && gift.X == head.X && gift.Y == head.Y {
				g.openGift(gift)
				break
			}
		}
	}

	// Сбор монеток
	if g.state == STATE_PLAYING && len(g.coins) > 0 {
		head := g.snake[0]
		for i, coin := range g.coins {
			if coin.X == head.X && coin.Y == head.Y {
				g.collectCoin(i)
				break
			}
		}
	}

	// Esc и P
	if ebiten.IsKeyPressed(ebiten.KeyEscape) && g.pauseCooldown <= 0 {
		if g.state == STATE_PLAYING || g.state == STATE_PAUSED || g.state == STATE_GAMEOVER {
			g.state = STATE_MENU
			if g.state == STATE_PAUSED {
				g.menuSelected = 1
			} else {
				g.menuSelected = 0
			}
			g.sndPause.Rewind()
			g.sndPause.Play()
		}
		g.pauseCooldown = 0.3
	}
	if ebiten.IsKeyPressed(ebiten.KeyP) && g.pauseCooldown <= 0 && (g.state == STATE_PLAYING || g.state == STATE_PAUSED) {
		if g.state == STATE_PLAYING {
			g.state = STATE_PAUSED
		} else {
			g.state = STATE_PLAYING
		}
		g.sndPause.Rewind()
		g.sndPause.Play()
		g.pauseCooldown = 0.3
	}

	// Меню
	if g.state == STATE_MENU {
		prev := g.menuSelected
		if ebiten.IsKeyPressed(ebiten.KeyUp) && g.pauseCooldown <= 0 {
			g.menuSelected--
			if g.menuSelected < 0 {
				g.menuSelected = len(g.menuButtons) - 1
			}
			g.pauseCooldown = 0.15
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && g.pauseCooldown <= 0 {
			g.menuSelected++
			if g.menuSelected >= len(g.menuButtons) {
				g.menuSelected = 0
			}
			g.pauseCooldown = 0.15
		}
		if prev != g.menuSelected {
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
		}
		if (ebiten.IsKeyPressed(ebiten.KeyEnter) || ebiten.IsKeyPressed(ebiten.KeySpace)) && g.pauseCooldown <= 0 {
			g.sndMenuSelect.Rewind()
			g.sndMenuSelect.Play()
			g.buttonFlash = 5
			switch g.menuSelected {
			case 0, 1, 2:
				g.reset()
				g.state = STATE_PLAYING
			case 3:
				g.state = STATE_SETTINGS
			case 4:
				return ebiten.Termination
			}
			g.pauseCooldown = 0.3
		}
		return nil
	}

	if g.state == STATE_PAUSED {
		return nil
	}
	if g.state == STATE_GAMEOVER {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = STATE_MENU
			g.menuSelected = 0
			g.pauseCooldown = 0.2
		}
		return nil
	}

	// Управление змейкой
	if ebiten.IsKeyPressed(ebiten.KeyUp) && g.dir.Y != 1 {
		g.nextDir = Vec{0, -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && g.dir.Y != -1 {
		g.nextDir = Vec{0, 1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.dir.X != 1 {
		g.nextDir = Vec{-1, 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.dir.X != -1 {
		g.nextDir = Vec{1, 0}
	}

	g.ticker += g.speed * dt
	if g.ticker >= 1 {
		g.ticker = 0
		g.dir = g.nextDir
		g.step()
	}

	// Бомбы
	for i := 0; i < len(g.bombs); i++ {
		g.bombs[i].Timer -= dt
		if g.bombs[i].Timer <= 0 {
			g.bombExplode(i)
			i--
		}
	}

	// Частицы
	for i := 0; i < len(g.particles); i++ {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.05
		p.Life -= 0.02
		p.Size *= 0.98
	}
	j := 0
	for _, p := range g.particles {
		if p.Life > 0 {
			g.particles[j] = p
			j++
		}
	}
	g.particles = g.particles[:j]

	g.shake *= 0.88
	return nil
}

func (g *Game) step() {
	head := g.snake[0]
	newHead := Vec{head.X + g.dir.X, head.Y + g.dir.Y}

	if newHead.X < 0 || newHead.X >= gridW || newHead.Y < 0 || newHead.Y >= gridH {
		g.triggerExplosion(head, true)
		return
	}
	if !g.ghostModeActive() {
		for _, s := range g.snake {
			if s == newHead {
				g.triggerExplosion(newHead, true)
				return
			}
		}
	}

	g.snake = append([]Vec{newHead}, g.snake...)

	ateFruit := false
	if newHead.X == g.fruitX && newHead.Y == g.fruitY {
		ateFruit = true
		switch g.fruitType {
		case FRUIT_APPLE:
			g.score += 1
			g.health = minInt(maxHealth, g.health+25)
		case FRUIT_STRAWBERRY:
			g.score += 2
			g.health =	g.health = minInt(max minIntHealth, g(maxHealth, g.health+40)
	.health+40)
		case FRUIT_	case FRUIT_ORANGEORANGE:
			g.score:
			g.score += 3 += 
			g3
			g.health = minInt(max.health = minInt(maxHealth, g.Health, g.health+health+35)
		case35)
		case FRUIT FRUIT_BANANA:
_BANANA:
			g			g.score += 1.score += 1
			g.
			g.health =health = minInt(maxHealth minInt(maxHealth, g., g.health+20health+20)
		case FR)
		case FRUIT_PUIT_PINEAPPLE:
INEAPPLE:
			g.score +=			g.score += 4 4
			g.
			g.health =health = minInt(maxHealth minInt(maxHealth, g, g.health+45)
		}
.health+45)
		}
		g		g.placeFruit.placeFruit()
	()
		g.spawnB	g.spawnBombRandomombRandom()
		g.sp()
		g.spawnIceawnIce()
		g.sp()
		g.spawnGhost()
	awnGhost()
		g.s	g.sndHeal.RndHeal.Rewindewind()
	()
		g.sndHe	g.sndHeal.Play()
al.Play()
		g		g.addParticles(float.addParticles(float64(newHead.X64(newHead.X*tile*tileSize+tileSizeSize+tileSize/2), float64(new/2), float64(newHead.YHead.Y*tileSize+t*tileSize+tileSize/2ileSize/2), ), 25, color.R25, color.RGBAGBA{50, {50, 255, 80255, 80, , 255}, true)
255}, true)
	}

	}

	if g.frozen	if g.frozenTimer > 0 && ateTimer > 0 && ateFruit {
Fruit {
			g.s	g.snake = g.snake = g.snake[:nake[:len(g.snakelen(g.snake)-1)-1]
	} else if !ate]
	} else if !ateFFruit {
ruit {
		g.snake		g.snake = g = g.snake[:len.snake[:len(g.snake)-1]
(g.snake)-1]
	}

	for	}

	for i :=  i := 0; i < len(g0; i < len(g.bombs); i.bombs); i++ {
++ {
		if		if g.b g.bombs[i].Xombs[i].X == new == newHead.XHead.X && && g.bombs g.bombs[i].[i].Y == newHeadY == newHead.Y {
			g.Y {
.health -= 			g.health -= 35
35
			g.trigger			g.triggerExplosionExplosion(newHead(newHead, g, g.health.health <=  <= 0)
0)
			g			g.bombs.bombs = append = append(g.b(g.bombs[:ombs[:i],i], g.bombs[i g.bombs[i+1+1:]...:]...)
)
					return
	return
		}
			}
	}

	if g.}

	if g.iceActiveiceActive && && newHead.X newHead.X == g == g.ice.X &&.ice.X && newHead newHead.Y == g..Y == g.ice.Yice.Y {
		g.f {
		g.frozenTimer = rozenTimer = 5.5.0
0
		g		g.iceActive = false
.iceActive =		g.addP false
		g.addParticles(floatarticles(float64(newHead.X64(newHead.X*tile*tileSize+tileSizeSize+tileSize/2), float64(new/2), floatHead.Y*tile64(newHead.Y*tileSize+tileSize/2Size+tileSize/2), ), 50,50, color.R color.RGBA{100GBA{100, , 200,200, 255 255, 255}, true)
, 255}, true)
		g		g.snd.sndHealHeal.Rewind()
.Rewind()
		g		g.sndHeal.Play.sndHeal.Play()
	}

	if()
	}

	if g. g.ghostActive &&ghostActive && newHead.X == newHead.X == g. g.ghostX &&ghostX && newHead newHead.Y ==.Y == g.ghostY {
		g.ghostModeTimer = g.ghostY {
		g.ghostModeTimer = 5.0 5.0
	
		g.	g.ghostActive = false
ghostActive = false
		g		g.sndGhost.R.sndGhost.Rewindewind()
	()
		g.s	g.sndGhostndGhost.Play.Play()
	()
		g.addParticles	g.addParticles(float64(newHead(float64(newHead.X*t.X*tileSize+tileileSize+tileSize/Size/2), float642), float64(newHead(newHead.Y*tileSize.Y*tileSize+tile+tileSize/2),Size/2), 60, color.RG 60, colorBA{200,.RGBA{200, 200 200,, 255, 255, 200 200}, true)
	}, true)
	}

	g}

	g.addP.addParticles(floatarticles(float64(newHead.X64(newHead.X*tile*tileSize+tSize+tileSizeileSize/2/2), float64(new), float64(newHead.Y*tileHead.Y*tileSize+tSize+tileSizeileSize/2/2), ), 2, color.RGBA{02, color.RGBA{0, , 180, 220180, 220, , 140}, false)
140}, false)
	if g.frozen	if g.frozenTimer >Timer > 0 {
	 0	g.add {
		g.addParticles(float64Particles(float64(newHead.X*tileSize(newHead.X*tileSize+tileSize/+tileSize/2),2), float64(newHead float64(newHead.Y*t.Y*tileSize+tileSize/ileSize+tile2), 1Size/2), 1, color, color.RG.RGBA{BA{150,150, 220 220, , 255,255, 180}, true)
	 180}, true}
}

func ()
	}
}

func (g *g *Game)Game) ghostMode ghostModeActive()Active() bool {
	return g. bool {
	return ggh.ghostModeTimer >ostModeTimer > 0 0

}

func (g *}

func (g *Game)Game) triggerExpl triggerExplosion(vosion(v Vec, Vec, fatal bool fatal bool) {
	if fatal) {
	if fatal {
	 {
		g.	g.health = 0health = 0
	
	}
	g}
	g.shake = 18
.shake = 18
	g.s	g.sndBndBoom.Room.Rewind()
	gewind()
	g.snd.sndBoomBoom.Play.Play()
	g()
	g.addParticles(float64(v.addParticles(float.X*t64(v.X*tileSize+tileSize/ileSize+tileSize/2), float64(v.Y2), float64(v.Y*tile*tileSize+tSize+tileSizeileSize/2/2), 80,), 80, color.R color.RGBAGBA{255{255, , 120, 30, 120, 30255},, 255}, true)
 true)
	g.addParticles(float64	g.addParticles(float64(v.X(v.X*tile*tileSize+tSize+tileSize/2ileSize/2), float64(v), float64(v.Y*t.Y*tileileSize+tileSize/Size+tileSize/2),2), 40 40, color, color.RG.RGBA{BA{255, 255255, 255, , 200,200, 200 200}, true)
	if}, true)
	if fatal {
 fatal {
		g		g.state =.state = STATE_G STATE_GAMEOVER
	}
AMEOVER
}

func	}
}

func (g *Game (g *Game) bombExplode(idx int) {
	b :=) bombExplode(idx int) {
	b := g.bombs[idx]
	g g.bombs[idx]
	g.bombs.bombs = append = append(g.bombs[:(g.bombs[:idx],idx], g.b g.bombs[idxombs[idx+1+1:]...)
	:]...)
	head := g.shead := g.snakenake[0]
[0]
	dx	dx := math.Abs := math.Abs(float64(float64(head.X(head.X - b - b.X))
.X))
	dy := math	dy := math.Abs.Abs(float64(float64(head.Y(head.Y - b - b.Y))
	dist.Y))
	dist := dx := dx + dy
	g + dy
	g.shake.shake =  = 1212
	g.sndBoom.R
	g.sndBoom.Rewindewind()
	g.s()
	g.sndndBBoom.Playoom.Play()
	cx := float64(b.X*tile()
	cx := float64(b.X*tileSize +Size + tileSize/2 tileSize/2)
	cy := float64)
	cy := float64(b.Y(b.Y*tileSize +*tileSize + tileSize tileSize/2)
	if/2)
	if dist <= 1 dist <= 1.5.5 {
		g. {
		g.health -=health -= 25 25
	
		g.addParticles(cx	g.addParticles(cx, cy, , cy, 120, color.RGBA{255120, color.RGBA{255, , 60,60, 30 30, , 255},255}, true)
		g true)
		g.addP.addParticles(cx,articles(c cy, 60x, cy, 60, color, color.RG.RGBA{BA{255, 200255, 200, , 50,50, 200 200}, true}, true)
		if g)
		if g.health.health <=  <= 0 {
0 {
			g			g.state = STATE_G.state = STATE_GAMEOAMEOVER
VER
				}
	} else {
	}
		g.add} else {
		g.addParticlesParticles(cx, cy(cx, 80,, cy, 80, color.R color.RGBA{255, GBA{255, 100,100, 30 30, 255},, 255}, true)
 true)
	}
	}
	g.addParticles	g.addParticles(cx(cx, cy, cy, 30,, 30, color.RGBA color.RGBA{255{255,,  200, 100, 200, 100, 200},200}, false)
 false)
}

func}

func (g (g *Game) spawn *GameBomb) spawnBombRandom()Random() {
	if g.rng {
	if g.rng.F.Float64loat64()() < 0.4 < 0.4 {
	 {
		for i	for i := 0; :=  i < 2000; i0; < 2000; i++ i++ {
		 {
			x := g.r	x := g.rng.Intng.Intn(gridn(gridW)
W)
			y := g			y.rng := g.rng.Intn.Intn(gridH)
		(gridH)
			ok	ok := true := true
		
			if x == g	if x == g.fruit.fruitX &&X && y == g.f y == g.fruitYruitY {
			 {
				ok	ok = false
		 = false
			}
	}
			for			for _, s _, s := range g.s := range g.snake {
nake {
				if				if s s.X == x && s.X == x && s.Y ==.Y == y {
 y {
										ok = false
ok = false
					break
								break	}
			
				}
			}
			for _, b :=}
			for _, range g b := range g.bombs.bombs {
			 {
				if b.X ==	if b.X == x && x && b.Y b.Y == y {
				 == y {
					ok	ok = false
					break
 = false
					break
								}
		}
			}
	}
			if g.			if g.iceActiveiceActive && g && g.ice.ice.X == x &&.X == x && g. g.ice.Yice.Y == y {
			 == y {
				ok	ok = false
			}
 = false
			}
			if ok {
			if ok {
				g				g.bombs = append.bombs(g.b = append(g.bombs, Bomb{Xombs, Bomb{X: x: x, Y, Y: y: y, Timer, Timer: 5.: 5.0})
0})
				return				return
		
				}
		}
		}
	}
}

}
	}
}

func (func (g *g *Game) addParticles(xGame) addParticles(x, y, y float64 float64, n, n int, c color int, c color.RG.RGBA,BA, glow bool glow bool) {
) {
	for i := 	for i := 0; i0; i < < n; i++ {
		a := n; i++ {
		a := g.r g.rng.Fng.Float64() * 2loat64() * 2 * math * math.Pi.Pi
		s :=
	 g.rng.F	s := g.rloat64()*ng.Float64()*4 +4 + 1 1.5
	.5
	g.particles =		g.p append(garticles = append(g.particles.particles, Particle{
			X, Particle{
			X:    :     x,
 x,
						Y:     y,
		Y:     y,
			V	VXX::    math    math.Cos(a) * s.Cos(a) * s,
		,
			VY:	VY:    math.Sin    math(a).Sin(a) * s,
		 * s,
			Life	Life: :  g.rng.F g.rng.Float64loat64()*()*1.5 + 0.41.5 + 0.4,
		,
			Color	Color: c,
		: c,
			Size	Size: :  g.r g.rng.Fng.Float64()*4 +loat64()*4 + 2 2,
		,
			Glow:	Glow:  glow  glow,
	,
		})
	}
	})
	}
}

// ---------}

// ---------- О- Отрисовка ---------трисовка ----------
func (g-
func (g *Game *Game) Draw(screen) Draw *eb(screen *ebiten.Image) {
iten.Image) {
	screen	screen.Fill(color.RGBA{12.Fill(color.RGBA{12, , 12,12, 20 20, , 255})

	ox255})

	ox, oy := 0, oy := 0.0, .00., 0.0
	if g0
.shake	if g.shake > 0. > 0.5 {
5 {
				ox =ox = (mathrand.F (mathloat64rand.Float64()*2 -()* 12 - 1) * g.sh) * g.shake
		ake
		oy = (mathoy = (mathrand.Frand.Float64loat64()*2 - 1()*2 - 1) *) * g.shake
 g.shake
	}

	}

	//	// Сетка
 Сетка
	for x := 	for x := 0; x0; x < gridW < gridW; x++ {
; x++ {
		for		for y := 0; y < grid y := 0; y < gridH;H; y++ y++ {
			c := color.RGBA{15, 15, {
			c := color.RGBA{15, 15, 25 25, 255}
, 255}
			if			if (x (x+y)%+y)%2 != 02 != 0 {
			 {
				c =	c = color.RGBA color.RGBA{18{18, , 18, 30, 18, 30, 255}
			}
			ebitenutil255}
			}
			ebitenutil.Draw.DrawRect(sRect(screen,creen, float64(x*t float64(x*tileSize)+oxileSize)+ox, float, float64(y*tileSize)+64(y*tileSize)+oy, tileSizeoy, tileSize-1-1, tile, tileSize-1,Size-1, c)
		 c)
		}
	}
	}

	// Б}

	// Бомбы
	forомбы
	for _, b := range g.b _, b := range gombs {
.bombs {
		cx :=		cx := float64(b.X*tile float64(b.X*tileSize+tileSizeSize+tileSize/2/2) + ox
) + ox
				cy := float64cy := float64(b.Y(b.Y*tile*tileSize+tileSize/2Size+tileSize/2) +) + oy oy
		base
		baseRadius :=Radius := float64 float64(tile(tileSize)/2 *Size)/2 * 1.5 1.5
	
		t := b.T	t := b.Timer
imer
		f		freq :=req := 3.0 +  3.0 + 9.0*(9.0*(1.0-m1.0-math.Minath.Min(1(1.0.0, t, t/5.0))
		pulse := /5.0))
		pulse := 1.1.0 +0 + 0.15*math 0.15.Sin*math.Sin(g.menuP(g.menuPulse*ulse*20*f20*freq)
		req)
		radius := baseRadiusradius := * pulse baseRadius * pulse
	
		r :=	r := uint8 uint8(20)
	(20)
		if t	if t <  < 2.0 {
2.0 {
			r			r = uint = uint8(80 + int(8(80 + int(175*(175*(1.0-t1.0-t/2.0)))
		}
/2.0)))
		}
				ebitenebitenutil.Dutil.DrawCircle(screenrawCircle(screen, cx, cx, cy, cy, radius, radius, color, color.RGBA{r.RGBA{r, 20,, 20, 25 25,, 255})
 255})
		ebitenutil.D		ebitenutil.DrawCircle(screen, cxrawCircle(screen, cx-2, cy-2, cy-2, radius-2-2, radius, color-2, color.RG.RGBA{0,BA{0, 0 0, 0,, 0, 100 100})
	})
		eb	ebitenutil.DrawCircle(sitenutil.DrawCircle(screen,creen, cx-radius*0 cx-radius*0.3.3, cy-radius*, cy-radius*0.0.35, radius*35, radius*0.0.25,25, color.R color.RGBAGBA{255{255, 255, 255, 255, 255, 180})
		, 180})
ebiten		ebitenutil.DrawCircleutil.DrawCircle(screen(screen, cx, cx+radius*0.+radius*02, cy.2, cy+radius+radius*0.2*0.2, radius*0, radius.2*0.2, color.RG, colorBA{255,.RGBA{255, 80 80, 80,, 80, 120 120})
		fuse})
		fuseLen :=Len := 20.0 20.0 * (t / * (t / 5 5.0)
	.0)
		fuse	fuseStartX := cx + radiusStartX := cx + radius*0.7*0
	.7
		fuseStartY := cy	fuseStartY := cy - radius*1 - radius*1.1.1
		fuseEndX := fuseStartX
		fuseEndX := fuseStartX + fuseLen* + fuseLen*0.0.7
		f7
		fuseEnduseEndY :=Y := fuseStart fuseStartY - fuseLenY - fuseLen*0*0.5
	.5
		eb	ebitenutil.Drawitenutil.DrawLine(sLine(screen, fuseStartcreen, fuseStartX, fuseStartX, fuseStartY,Y, fuseEndX, fuseEndX, fuseEnd fuseEndY, color.RY, color.RGBAGBA{80{80, 70, 50, 70, 50, , 255})
255})
		fireSize		fireSize := 3. := 3.0 +0 + 2*math 2*math.Sin.Sin(g.m(g.menuPulse*enuPulse*50)
50)
				ebitenutil.DrawCircleebitenutil.DrawCircle(screen, fuse(screen, fuseEndXEndX, fuse, fuseEndYEndY, fireSize,, fireSize, color.R color.RGBAGBA{255{255, , 100,100, 20,  20, 255})
		255})
		ebitenutil.Debitenutil.DrawCirclerawCircle(screen(screen, fuseEndX, fuseEndX, fuse, fuseEndYEndY, fireSize*0., fireSize*0.66, color.R, color.RGBA{GBA{255, 255, 255,255, 100 100, , 255})
255})
	}

	}

	// Зм	// Змейка
	forейка
	for i, i, s := range g.snake s := range g.snake {
	 {
		x :=	x := float64(s.X float64(s.X*tileSize) + ox*tileSize) + ox
	
		y :=	y := float64(s.Y*tile float64(s.Y*tileSize)Size) + oy
 + o		vary
		var base color.RG base colorBA
		if g.f.RGBA
		ifrozenTimer g.frozenTimer >  > 0 {
0 {
			base =			base = color.R color.RGBAGBA{80{80, , 180, 255180, 255, , 255}
255}
				} else if g} else if g.ghostMode.ghostModeActive() {
			Active()base = color {
			base = color.RGBA{.RGBA{20,20, 220,  220, 90, 18090,}
	 180}
		} else {
	} else {
			base			 =base = color.RGBA color.RGBA{20, {20, 220,220, 90 90, , 255}
255}
				}
		if i}
		if i > 0 {
 > 0 {
			shade			shade := uint8( := uint8(100 +100 + (i*4 (i)%100)
		*4)%100)
			if	if g.frozenTimer > g.frozenTimer > 0 0 {
			 {
				base = color.RG	base = colorBA{80, 180.RGBA{80,,  180, 255,255, shade}
 shade}
						} else if g} else.gh if g.ghostModeActive() {
				baseostModeActive() {
			 = color	base = color.RGBA{.RGBA{20,20, 220,  220, 90,90, uint8 uint8(180)}
			}(180)}
			} else {
 else {
				base =				base = color.RGBA color.RGBA{15{15, shade, , shade, 70,70, 255}
			}
 255}
					}
		}
	}
		if i	if i == 0 {
 == 0 {
			ebitenutil.D			ebitenutil.DrawRectrawRect(screen, x(screen, x-3-3, y-3, tile, y-3, tileSize+Size+6,6, tileSize tileSize+6, color.RG+6, color.RGBA{BA{0, 2000, 200, 80,,  4080, 40})
		})
			if g.frozenTimer >	if g.frozenTimer > 0 0 {
			 {
				ebitenutil.Draw	ebitenutil.DrawRect(sRect(screen, x-creen,3, y- x-3, y-3,3, tileSize tileSize+6, tile+6, tileSize+Size+6,6, color.RGBA color.RGBA{100{100, , 200,200, 255, 80})
			 255, 80})
			}
		}
			if g	if g.gh.ghostModeActive()ostModeActive() {
			 {
				eb	ebitenutil.DrawRect(sitenutil.DrawRect(screen,creen, x- x-3, y-3,3, y-3, tileSize+6, tile tileSize+6, tileSize+Size+6,6, color.RGBA{255 color.RGBA{255, , 255,255, 255,  255, 80})
80})
						}
	}
		}
			}
		ebitenebitenutil.DrawRectutil.D(screen, xrawRect(screen, x, y, y,, tile tileSize-Size-1, tileSize1, tileSize-1, base)
	-1, base)
		if i	if i == 0 {
			 == 0 {
			eyex := float64(teyex := float64(tileSizeileSize)/4 - )/4 - 2
			2
			eyeyeyey := float := float64(tileSize64(tileSize)/4 - )/42
			 - 2
ebiten			ebitenutil.Dutil.DrawRect(screenrawRect(screen, x, x+eyex, y++eyex,eyey y+eyey, , 4, 44, 4, color, color.White)
			eb.White)
			ebitenutil.DrawRect(sitenutil.DrawRect(screen,creen, x+ x+float64float64(tileSize)-eyex(tileSize)-eyex-6-6, y, y+eyey, 4+eyey, 4, 4, 4, color., color.White)
			White)
			ebitenebitenutil.Dutil.DrawRect(screen, x+eyex+rawRect(screen, x+eyex+1,1, y+ y+eyey+1eyey+1, 2,,  2, color.Black2, 2, color)
		.Black)
			ebitenutil.Draw	ebitenutil.DrawRect(screen,Rect(screen, x+ x+float64(tileSize)-float64(tileSize)-eyex-5eyex, y+ey-5, y+eyey+1,ey+1, 2,  2, 2,2, color.Black)
 color.Black)
			var			var tx, tx, ty, w, h ty, w, h float float64
64
			switch g.dir			switch g.dir {
			case Vec{1 {
			case Vec{1, 0}, 0}:
				tx, ty:
				tx, ty, w, w, h = tile, h = tileSize-Size-4, tileSize4, tileSize/2/2-2, -2, 6, 46, 4
		
			case Vec{-1, 0}	case Vec{-1, :
				tx0}:
			, ty	tx, ty, w, w, h = -, h = -2, tileSize2, tileSize/2/2-2, 6,-2, 6, 4 4
		
			case Vec{0	case Vec{0, 1}, 1}:
			:
				tx	tx, ty, ty, w, h = tile, w, h = tileSize/Size/2-2,2- tileSize-42, tileSize-4, 4, 6, 4, 6
		
			case Vec{0	case Vec{0, -, -1}1}:
				tx, ty:
				tx, ty, w, h, w = tileSize/, h = tileSize/2-2-2, -2, 2, -2, 4, 64, 6
		
			}
	}
						ebitenebitenutil.DrawRectutil.DrawRect(screen(screen, x+tx, x+tx+ox+ox, y, y+ty+oy, w+ty+oy, w, h, color, h, color.RGBA{.RGBA{255, 70, 100255, 70, , 255100, 255})
	})
		}
	}

	}
	// Фрук	}

	// Фрукты
ты
	{
	{
		cx :=		cx := float64 float64(g.fruitX(g.f*tileSize+truitX*tileSize+tileSize/2ileSize/2) + ox
		) + ox
		cy :=cy := float64(g.f float64(g.fruitYruitY*tile*tileSize+tileSizeSize+tileSize/2/2) + oy
	) + oy
		var img	var img *ebiten.Image *ebiten.Image
	
		switch g.fruitType {
		case	switch g.fruitType {
		case FRUIT FRUIT_APP_APPLE:
LE:
			img =			img = g.appleImg g.appleImg
	
		case FR	case FRUIT_STUIT_STRAWBERRYRAWBERRY:
			img = g:
			img = g.straw.strawberryberryImg
		case FRUIT_Img
		case FRORANGEUIT_:
			imgORANGE:
			img = g = g.orangeImg
.orangeImg
		case		case FRUIT FRUIT_BANANA:
			_BANANA:
			img = g.img = g.bananaImg
bananaImg
		case		case FRUIT_PINEAPPLE FRUIT_PINEAPPLE:
		:
			img = g.pine	img = g.pineappleImgappleImg
	
		}
		if	}
		if img != img != nil {
			op := nil {
			op := &ebiten.DrawImage &ebiten.DrawImageOptions{}
			w, hOptions{}
			w, h := img.Bounds().D := img.Bounds().Dx(),x(), img.Bounds(). img.Bounds().Dy()
			Dy()
			scale :=scale := 1.5
		 1.5
			op.Geo	op.GeoM.Scale(M.Scale(scale,scale, scale)
			 scale)
			op.Gop.GeoM.TranslateeoM.Translate(cx(cx-float64(w)*scale/2-float64(w)*scale, cy/2, cy-float-float64(h64(h)*scale/2)
		)*scale/2)
			op	op.GeoM..GeoM.Translate(Translate(ox, oyox, oy)
			screen.Draw)
			screen.DrawImage(img, opImage(img, op)
)
		} else {
		} else {
						ebitenebitenutil.DrawRect(screen, cxutil.DrawRect(screen, cx-float64(tileSize-float64(tileSize)/2+ox, cy-float64(t)/2+ox, cy-float64(tileSize)/2+oyileSize)/2+oy, float, float64(t64(tileSize), float64(tileSize), float64(tileSizeileSize), color.RG), color.RGBA{200BA{200, 100, 100, 50,, 50, 255 255})
		}
})
		}
	}

	}

	// Лё	// Лёд
д
	if g.ice	if g.iceActive {
		cActive {
		cx :=x := float64(g.ice.X float64(g.ice.X*tileSize+t*tileSize+tileSizeileSize/2) +/2) + ox
		 ox
cy :=		cy := float64(g. float64(g.ice.Yice.Y*tileSize+tileSize*tileSize+tileSize/2/2) + oy) + oy
		radius
		radius := float64(tileSize := float64(tileSize)/2 * )/2 * 1.2
1.2
				ebitenutil.Debitenutil.DrawCirclerawCircle(screen, cx, cy,(screen, cx, cy radius, color, radius, color.RGBA{150,.RGBA{150, 220,  220, 255,255, 255})
	 255})
		ebitenutil	ebitenutil.Draw.DrawCircle(screen,Circle(screen, cx- cx-2, cy-2, radius-2, cy-2,3, radius-3, color.RGBA{100 color.RGBA{100, 200,, 200, 240 240, 200})
		, 200})
		ebitenutil.Debitenutil.DrawCircle(screen, cx+radiusrawCircle(screen, cx+radius*0*0.3, cy-radius*.3, cy-radius*0.0.3,3, radius*0. radius*0.25,25, color.RGBA color.RGBA{255, {255, 255,255, 255 255, 200})
, 200})
	}

	//	}

	// Призрак
	if g Призрак
	if g.gh.ghostActive && lenostActive && len(g.ghost(g.ghostFramesFrames) > 0) > 0 {
		cx {
		cx := float64(g := float64(g.ghostX.ghostX*tile*tileSize+tSize+tileSize/2) + oxileSize/2) +
		 ox
		cy :=cy := float64(g. float64(g.ghostY*tileSize+tileghostY*tileSize+tileSize/2) + oSize/2) + oy
		frame :=y
		frame := g. g.ghostFrames[g.ghghostFrames[g.ghostFrameostFrameIdx]
		ifIdx]
		if frame != frame != nil {
			 nil {
			op := &ebiten.Dop := &ebiten.DrawImageOptions{}
			wrawImageOptions{}
			w, h, h := frame.Bounds().Dx(), := frame.Bounds().D frame.Boundsx(), frame.Bounds().Dy()
().Dy()
			scale :=			scale := float64 float64(tileSize) / float(tileSize) / float64(w64(w)
		)
			op.GeoM.Sc	op.GeoM.Scale(scale, scale)
			ale(scale,op.G scale)
			op.GeoM.TranslateeoM.Translate(cx(cx-float64(w-float64(w)*scale)*scale/2, cy/2, cy-float64(h)*scale-float64(h)*scale/2/2)
		)
			op.GeoM	op.GeoM.Translate(ox,.Translate(ox, oy oy)
			screen)
		.DrawImage	screen.DrawImage(frame,(frame, op)
		}
 op)
		}

	// Тара	}
	}

	// Таракан (кан (если разреесли разрешенышены фонов фоновые анимацииые анимации)
	if g.roachActive && len(g.ro)
	if g.roachActive && lenachF(g.roachFrames) > rames) > 0 &&0 && currentSettings.Background currentSettings.BackgroundAnimation {
		cAnimation {
		cx := float64x := float64(g.ro(g.roachX*tileachX*tileSize+tSize+tileSize/2) + ox
		cy := float64ileSize/2) + ox
		cy := float64(g.roachY*tile(g.roSize+tachY*tileSize+tileSize/2) + oyileSize/2) + o
	y
		frame	frame := g.roachFrames[g := g.roachFrames[g.roach.roachFrameIdx]
	FrameIdx]
		if frame !=	if frame != nil {
		 nil {
			op := &	op := &ebitenebiten.DrawImageOptions.DrawImageOptions{}
			w,{}
			w, h := h := frame.Bounds(). frame.Bounds().DxDx(), frame.Bounds(), frame.Bounds().Dy().Dy()
			scale()
			scale := float64(tileSize := float64(tileSize) / float64) / float64(w)
			op.G(w)
			op.GeoMeoM.Scale(scale, scale.Scale(scale, scale)
		)
			op.Geo	op.GeoM.M.Translate(cx-Translate(cx-float64(w)*float64(w)*scale/scale/2, cy-2, cy-float64float64(h)*scale/(h)*scale/2)
2)
			op.G			op.GeoM.TranslateeoM.Translate(ox(ox, oy)
, oy)
			screen.DrawImage			screen.DrawImage(frame,(frame, op)
		}
	}

	 op)
		}
	}

	// Ви// Викинги (екинги (если разресли разрешены фоновые анимациишены фоновые анимации)
	if len(g.vikingFrames)
	if len(g.vikingFrames) >) > 0 && current 0 && currentSettings.BackgroundAnimation {
	Settings.BackgroundAnimation {
		for _, v :=	for _, v := range g range g.vikingList {
.vikingList {
			c			cx := float64x := float64(v.X(v.X*tileSize+t*tileSize+tileSize/ileSize/2) +2) + ox
			cy := ox
			cy := float64(v.Y*tile float64(v.Y*tileSize+tileSizeSize+tileSize/2/2) + oy
			frame := g) + oy
			frame.vikingFrames := g.vikingFrames[v.Frame%len(g[v.Frame%len(g.vikingFrames.vikingFrames)]
		)]
			op := &	op := &ebiten.DrawImageOptionsebiten.D{}
			w, h :=rawImageOptions{}
			w, frame.Bounds(). h := frame.Bounds().DxDx(), frame.Bounds().Dy(), frame.Bounds().Dy()
		()
			scale	scale := float64(t := float64(tileSizeileSize) /) / float64(w)
			op.G float64(w)
			op.GeoM.Scale(scaleeoM.Scale(scale, scale)
		, scale)
			op.GeoM.	op.GeoM.Translate(cTranslate(cx-x-float64(w)*float64(w)*scale/scale/2,2, cy-float64(h)*scale/2)
			op.G cy-float64(h)*scale/2)
			op.GeoMeoM.Translate(ox.Translate(ox, o, oy)
y)
			s			screen.DrawImagecreen.D(frame,rawImage(frame, op)
 op)
		}
	}

	// Подар		}
	}

	// Пки
одарки
	for _, gift :=	for _, gift := range g.gifts {
	 range g.gifts {
		cx	cx := float := float64(g64(gift.X*tileSize+tift.X*tileileSizeSize+tileSize/2)/2) + ox
 + ox
				cycy := float64(gift.Y*t := float64(gift.Y*tileSizeileSize+tile+tileSize/Size/2)2) + o + oy
y
		var		var img * img *ebitenebiten.Image
		if gift.Image
		if gift.Opened {
		.Opened {
			img = g.giftOpenFrames	img = g.giftOpenFrames[0] //[0] // можно использовать первый можно использовать первый кад кадр илир или анима анимацию
цию
			op := &ebiten.Draw			op := &ebiten.DrawImageOptions{}
			w, h := img.Bounds().DImageOptions{}
			w, h := img.Bounds().Dx(),x(), img.B img.Bounds().ounds().Dy()
Dy()
						scalescale := float64(tile := float64(tileSize)Size) / float / float64(w64(w)
		)
			op	op.GeoM.Scale(.GeoM.Scscale,ale(scale, scale)
 scale)
						op.Gop.GeoMeoM.Translate.Translate(cx(cx-float64(w-float)*scale/264(w)*scale, cy/2, cy-float64(h-float64(h)*scale)*scale/2/2)
		)
			op.GeoM.	op.GeoTranslate(ox,M.Translate( oy)
		ox, oy)
			if gift.Life < 	if gift.Life2. < 2.0 && gift.L0 && gift.Life >ife > 0 0 {
				alpha {
				alpha := gift.Life := gift.Life /  / 2.2.0
0
				if alpha				if < 0 alpha < 0 {
				 {
					alpha	alpha =  = 0
				0
				}
			}
				op.ColorM	op.ColorM.Scale(1.Scale, 1,(1, 1, 1, alpha 1, alpha)
		)
			}
			screen.D	}
			screen.DrawImage(img,rawImage(img, op)
		 op)
		} else {
} else {
					if gift	if gift.Color >=.Color >= 0 0 && gift && gift.Color <.Color < len(g.gift len(g.giftClosedImClosedImgs)gs) {
				img = g.giftClosedImgs {
				img = g.giftClosedImgs[gift[gift.Color]
			.Color]
			} else} else {
			 {
				img	img = g = g.gift.giftClosedImClosedImgs[0]
			gs[0]
			}
		}
			op	op := & := &ebitenebiten.Draw.DrawImageOptionsImageOptions{}
		{}
			w, h := img.B	w, h := img.Bounds().ounds().DxDx(), img(), img.Bounds.Bounds().Dy().Dy()
		()
			scale	scale := float := float64(tileSize64(t) / float64ileSize) /(w)
			 float64(w)
op.G			op.GeoMeoM.Scale(scale.Scale(scale, scale, scale)
			op)
			op.GeoM..GeoM.Translate(cTranslate(cx-float64x-float64(w)*(w)*scale/2,scale/2, cy-float64(h)* cy-float64(h)*scale/2)
scale/2)
			op.G			eoMop.GeoM.Translate(ox, oy)
			s.Translate(ox, oy)
			screen.Dcreen.DrawImage(img,rawImage(img, op)
		 op)
		}
	}
	}

	// К}

	// Ключ на полелюч
	if на поле
	if g.keyOnField g.keyOnField.Active.Active && g.keyImg && g.keyImg != nil {
	 != nil	cx {
		cx := float64(g := float64(g.keyOn.keyOnField.X*tileField.X*tileSize+tSize+tileSize/ileSize/22) + ox
) + ox
		cy		cy := float64(g.key := float64(g.keyOnFieldOnField.Y*tileSize.Y*tileSize+tileSize+tileSize/2)/2) + oy
 + oy
				opop := &eb := &ebiten.DrawImageOptions{}
		w, h := giten.DrawImageOptions{}
		w, h := g.keyImg.keyImg.Bounds().Dx(), g.key.Bounds().Dx(), g.keyImg.Bounds().Dy()
		Img.Bounds().Dy()
		scale :=scale := float64(tileSize) / float float64(tileSize) / float64(w64(w)
	)
		op	op.GeoM.Scale(.GeoM.Scale(scale,scale, scale)
 scale)
				op.Gop.GeoMeoM.Translate(cx-float.Translate(cx64(w-float64(w)*scale/2)*scale/2, cy-float, cy-float64(h64(h)*scale)*scale/2)
		op.Geo/2)
		op.GeoM.M.Translate(ox, oyTranslate(ox, oy)
	)
		if g	if g.keyOn.keyOnField.LField.Lifeife < 2 < 2.0 && g.keyOn.0 && gField.L.keyOnife > 0Field.L {
		ife > 0 {
			alpha	alpha := g := g.keyOn.keyOnField.LField.Life /ife / 2 2.0.0
			if alpha < 
			if alpha0 {
 < 0 {
								alpha =alpha = 0
		 0
			}
			op.Color	}
			op.ColorM.Scale(M.Scale(1,1, 1 1, , 1,1, alpha)
		 alpha)
		}
		screen}
	.Draw	screen.DrawImage(gImage(g.keyImg.keyImg, op, op)
	)
	}

	}

	// М// Монетки
онет	for _,ки
	for _, c := range g c :=.coins range g.coins {
	 {
		if len	if len(g.co(g.coinFrames)inFrames) == 0 {
 == 0 {
			continue			continue
		}
		f
		}
		frameIdx := crameIdx := c.Frame % len.Frame % len(g.co(g.coinFrames)
		inFrames)
		img := g.coimg := g.coinFinFrames[frameIdxrames[frameIdx]
	]
		if img == nil	if img == nil {
			continue
		 {
			continue
		}
		cx}
		cx := float := float64(c.X*t64(c.X*tileSize+tileileSizeSize/+tileSize/2)2) + ox
	 + ox
		cy	cy := float64(c := float64(c.Y*tileSize.Y*tileSize+tileSize/+tileSize/2) + oy
2) + oy
				op :=op := &eb &ebiten.DrawImageiten.DrawImageOptions{}
Options{}
		w, h		w, h := img.Bounds().D := img.Bounds().Dx(), img.Bx(), img.Bounds().ounds().Dy()
		Dy()
		scale :=scale := float64 float64(tile(tileSize) / floatSize) / float64(w64(w)
		op)
		op.GeoM.Scale(.GeoM.Scale(scale,scale, scale)
 scale)
				op.Geoop.GeoMM.Translate.Translate(cx-float(cx-float64(w64(w)*scale)*scale/2, cy-float/2, cy-float64(h64(h)*scale/2)*scale/2)
		op.Geo)
		op.GeoM.M.Translate(ox,Translate(ox, oy oy)
		screen)
		screen.DrawImage(img.DrawImage(img, op)
	, op}

	)
	}

	// Частицы// Частицы
	for
	for _, p := range _, p := range g.particles {
 g.p		carticles {
		c := p.Color
 := p.Color
		if		if p.Glow {
 p.Glow {
						ebitenutil.Debitenutil.DrawRect(screenrawRect(screen, p, p.X-p.Size*.X-p.Size*1.1.5+ox,5+ox, p.Y p.Y-p.Size*1-p.Size*1.5+oy.5+oy, p.Size*3, p.Size*3, color, p.Size*3, p.Size*3, color.RG.RGBA{c.R,BA{c.R, c.G, c c.G, c.B,.B, uint8(float64 uint8(float64(c.A(c.A) * 0) * 0.4 * p.4 * p.Life.Life)})
		)})
		}
	}
		eb	ebitenutil.DrawRectitenutil(screen,.DrawRect(screen, p.X p.X-p.Size-p.Size+ox+ox, p, p.Y-p.Y-p.Size+.Size+oy,oy, p.Size*2 p.Size*2, p.Size*, p.Size*2,2, c)
	}

 c)
	}

	draw	drawText :=Text := func(str func(str string, x, y int string, x, y int, clr color, clr color.Color).Color) {
		if g {
		if g.fontFace != nil.fontFace {
			text.Draw != nil(screen {
			text.Draw(screen, str, str, g, g.fontFace, x.fontFace, x, y, cl, yr)
		, clr)
} else {
				} else {
			eb	ebitenutilitenutil.Debug.DebugPrintAtPrintAt(screen(screen, str, str, x, y)
	, x, y)
		}
	}

	}
	}

	// UI	 в// UI в зависимости от языка
 зависимости от языка
	if current	if currentSettings.Language ==Settings.Language == "ru "ru" {
		drawText("С" {
		drawText("Счёт:чёт: "+strconv.I "+strconv.Itoa(g.scoretoa), (g.score), 10, 2510, 25, color, color.White)
	.White)
	} else {
	} else	draw {
		drawText("Score:Text("Score: "+str "+strconv.Itoaconv.Itoa(g.score(g.score), ), 10,10, 25 25, color.White, color.White)
	}
	)
	}
	barXbarX := float64(s := float64(screenW - creenW - 20)
20)
	barW :=	barW := 150 150.0
	.0
	barH := barH := 14.0
14.0
	healthPct	healthPct := float64(g := float64(g.health.health) / float64(maxHealth) / float64(maxHealth)
	ebiten)
	ebitenutil.Drawutil.DrawRectRect(screen(screen, barX-bar, barX-barW, 10, barW, 10, barW,W, barH barH, color.RGBA{30, color.RGBA{30,, 30,  30, 40,40, 200})
	 200})
	ebitenutil.DebitenrawRectutil.DrawRect(screen, bar(screen, barX-barW, 10X-barW, 10, bar, barW*healthPW*healthPct,ct, barH barH, color.RG, colorBA{50,.RGBA{50, 255,  255, 80,80, 255 255})
	if})
	if currentSettings.Language == " currentSettings.Languageru" == "ru" {
	 {
		draw	drawText("ЗДText("ЗДОРОВОРОВЬЕ", intЬЕ", int(bar(barX-barW+X-barW+40), 2540), 25, color.White, color.White)
	)
	} else} else {
		draw {
		drawText("Text("HEALTH", intHEALTH", int(bar(barXX-barW+40),-barW+40), 25 25, color, color.White)
	.White)
	}

	if}

	if currentSettings.L currentSettings.Language == "anguage == "ru" {
	ru" {
		draw	drawText("КлюText("Ключи:чи: "+strconv.I "+strconv.Itoa(g.keystoa(g.keysCollectedCollected), 10,), 10, 55 55, color.RG, color.RGBA{BA{255,255, 215 215, 0,, 0, 255})
	 255})
		if g.carry	if g.carryingKeyingKey {
		 {
			drawText("	drawText("Ключ активированКлюч активирован!", 10,!", 10, 80, color.RG 80, color.RGBA{BA{255, 200255, 200, 100,, 100, 255 255})
		}
})
		}
				drawText("МdrawText("Монеты: "+онетыstrconv: "+strconv.Ito.Itoa(g.coina(g.coinCount), 10Count), 10, , 105, color.R105, color.RGBA{255GBA, {255, 215, 0215,, 255})
	} 0, 255})
	} else {
 else {
		drawText		drawText("Keys("Keys: "+: "+strstrconv.Itoa(g.keysColconv.Itoa(g.keysCollected), 10lected), 10, , 5555, color.R, color.RGBAGBA{255{255, , 215, 0215, 0, , 255})
255})
		if g.c		if g.carryingKey {
arryingKey {
						drawText("Key activated!",drawText("Key activated!", 10 10, , 80, color.R80,GBA color.RGBA{255, {255, 200, 100, 255})
200, 100, 255})
				}
	}
		drawText("	drawText("CoinsCoins: "+strconv: "+strconv.Ito.Itoa(g.coina(g.coinCount),Count),  10, 10, 105,105, color.RGBA color.RGBA{255, {255, 215,215, 0,  0, 255})
	}

255})
	}

	if current	if currentSettings.LSettings.Language == "ruanguage ==" {
 "ru" {
		drawText("ESC - меню", screenW-100		drawText("ESC - меню", screenW-100, screen, screenH-H-20,20, color. color.White)
White)
				drawTextdrawText("P("P - пауза", screen - паузаW-100,", screenW-100, screenH screenH-40-40, color, color.White)
	.White)
		draw	drawText("K -Text("K - взять ключ", screen взять ключ", screenW-W-100,100, screenH screenH-60, color-60, color.White.White)
	} else)
	} else {
	 {
		draw	drawText("Text("ESC -ESC - menu", menu", screenW screenW-100, screenH--100, screen20, color.H-20, color.White)
		White)
		drawText("PdrawText("P - pause - pause", screenW-", screen100,W-100, screenH-40 screenH, color.White-40, color)
	.White)
		drawText("	drawK -Text("K - get key", screen get key", screenW-100, screenH-60, colorW-100, screenH-60, color.White.White)
	}

	if)
	}

	if g.f g.frozenTimer > rozenTimer > 0 {
		if0 {
		if currentSettings.Language currentSettings.Language == " == "ru"ru" {
			draw {
			drawText("Text("ЗАЗАМОМОРОЗРОЗКА",КА", screenW screenW/2-60, screen/2H-30,-60, screenH- color.R30, color.RGBAGBA{100, {100, 200,200, 255 255, , 255})
		255})
} else		} else {
			draw {
			drawText("Text("FROFROZENZEN", screenW/", screenW/2-2-60,60, screenH-30 screenH-30, color.RG, color.RGBA{BA{100,100, 200 200, , 255, 255255, 255})
		}
})
		}
	if g	}
	}
.gh	if g.ghostModeostModeActive()Active() {
	 {
		if currentSettings.Language ==	if currentSettings.Language == "ru "ru" {
" {
						drawText("ПdrawText("ПРИЗРИЗРАЧРАЧНЫЙНЫЙ РЕ РЕЖИЖИМ",М", screen screenW/2W/2-100, screen-100, screenH-H-60,60, color.R color.RGBAGBA{200, {200, 200,200, 255 255, , 255})
		} else255})
		} else {
			draw {
			drawText("Text("GHOSTGHOST MODE MODE", screenW/", screenW/2-100,2- screenH-60100, screenH-60,, color color.RG.RGBA{200, 200, 255, 255BA{200, 200, 255, 255})
	})
		}
	}
	}

	}

	switch g	switch g.state {
.state {
	case STATE	case STATE_MEN_MENU:
U:
				ebitenebitenutil.DrawRectutil.DrawRect(screen, (screen, 0, 00, 0, screen, screenW, screenH, colorW, screenH, color.RGBA{.RGBA{0,0, 0,  0, 0, 2550, 255})
		drawText("})
		drawText("S NS N A K E   A K E   R E V I R E V I V E V E D", D", screenW/2 screenW-180, /2-180, 150,150, color.R color.RGBAGBA{255, {255, 200,200, 100 100, 255})
		, 255})
		startY := 280
startY := 280
				step :=step := 50 50
	
		for i, btn	for i, btn := range := range g.menuButtons {
			y := g.menuButtons {
			y := startY startY + i + i*step*step
			if i
			if i == g == g.menuSelected {
				b.menuSelected {
				bg :=g := color.R color.RGBA{100GBA{100, , 100, 150, 100, 150, 255}
				if255}
				if g.buttonFlash > g.button 0Flash > 0 {
				 {
					bg = color	bg = color.RG.RGBA{200,BA{ 200200, 200, 255,, 255, 255 255}
			}
				}
					}
				ebitenebitenutil.DrawRectutil.DrawRect(screen, screen(screen, screenW/W/2-2-150,150, float64 float64(y)-15, 300(y)-15, 300, , 35,35, bg)
 bg)
								drawTextdrawText(btn, screen(btn, screenW/W/2-len(2-len(btn)*btn)*3,3, y, y, color.R color.RGGBA{255, BA{255255, 0, 255, 0, , 255})
255})
			} else			} else {
			 {
				drawText(	drawbtn,Text(btn, screenW/2 screenW/2-len-len(btn(btn)*3)*3, y, color.White, y, color)
		.White)
			}
	}
				}
	}
		if current	if currentSettings.LSettings.Language ==anguage == "ru" {
 "ru			drawText" {
			drawText("С("Стретрелки влки вверх/верх/внивниз,з, Enter - Enter - выбор", выбор", screenW/2-220 screenW/2-220, screen, screenH-70,H- color.RGBA70, color.RGBA{200{200, , 200,200, 200 200, , 255})
255})
				} else {
			draw} else {
		Text("	drawText("Arrow keys up/dArrow keys up/down,own, Enter - Enter - select", select", screenW screenW/2-240/2, screen-240, screenH-70,H- color.RGBA70, color.RGBA{200, 200,{200,  200200, 200, , 255})
		255})
		}
	case STATE_P}
	case STATE_PAUSAUSED:
ED:
		ebiten		ebitenutil.DrawRectutil.D(screen, rawRect(screen0,, 0, 0, screenW, 0, screenW, screenH screenH, color, color.RG.RGBA{BA{0,0, 0 0, , 0,0, 200})
	 200	if currentSettings.Language ==})
		if currentSettings.L "ru" {
anguage == "ru" {
						drawTextdrawText("П("ПАУАУЗАЗА", screenW/2-", screenW/2-40, screenH/240, screenH/2, color, color.RG.RGBA{BA{255,255, 255, 150, 255,  255150,})
			draw 255})
			drawText("Text("НажНажмитемите P для P для продолжения продолжения", screenW/", screenW/2-150,2-150, screenH screenH/2/2+40+40, color.White, color)
		.White)
	} else {
	}			drawText else {
			drawText("PA("PAUSED", screenUSEDW/2-", screen40,W/2-40, screenH screenH/2, color/2, color.RG.RGBA{BA{255,255, 255 255, , 150, 255})
		150, 255	draw})
			drawText("Text("Press PPress P to resume", screenW/ to resume", screen2-W/2-150, screenH150, screenH/2/2+40+40, color, color.White.White)
		}
)
		}
	case STATE	case STATE_GAM_GAMEOVEREOVER:
	:
		ebitenutil	ebitenutil.Draw.DrawRect(sRect(screen, 100, creen, 100, 80,80, screenW-200, screen screenW-200, screenH-H-180, color.R180, color.RGBA{40GBA{40, 0, 0, 0, 0, 255})
, 		if255})
		if currentSettings.Language currentSettings == "ru.Language == "ru"" {
		 {
			draw	drawText("Text("ИГИГРА ОКОНЧЕНА",РА ОКОНЧЕНА", screenW/2 screenW/2-80-80, screen, screenH/2-H/2-40,40, color.R color.RGBA{255, GBA{255, 100, 100100, 100, , 255})
			255})
			drawTextdrawText("С("Счёт:чёт: "+str "+strconv.Iconv.Itoatoa(g.score(g.score), screen), screenW/2-W/2-60,60, screenH/2 screenH/2, color, color.White.White)
		)
			draw	drawText("НажText("Нажмите любуюмите любую клави клавишу для меншу для менюю", screenW/2-", screenW/2-150,150, screenH screenH/2/2+40, color.White+40, color.White)
	)
		}	} else {
			 else {
drawText			drawText("GAME OVER("G", screenAME OVER", screenW/W/2-2-80, screenH/2-4080, screenH/2-40, color, color.RG.RGBA{BA{255,255, 100, 100, 100,  255100, 255})
			draw})
			drawText("Text("Score:Score: "+str "+strconv.Itoa(g.scoreconv.Itoa(g.score),), screen screenW/W/2-2-60,60, screenH screenH/2/2, color, color.White)
		.White)
			drawText("Press any	drawText("Press any key for menu", screenW key for menu", screenW/2/2-150-150, screen, screenH/2+H/40, color.2+40, color.White)
White)
				}
	case}
	case STATE_SET STATE_SETTINGS:
	TINGS	eb:
		ebitenutil.Drawitenutil.DrawRect(sRect(screencreen,, 0 0, 0, screenW, 0,, screen screenW, screenH, color.RH, color.RGBAGBA{0{0, , 0,0, 0 0, , 220})
		if currentSettings.Language220})
		if currentSettings == "ru".Language == "ru" {
		 {
			draw	drawText("Text("НАСТНАСТРОЙКИРОЙ", screenW/КИ", screenW/2-2-100,100, 100 100, color, color.RG.RGBA{BA{255,255, 255,  255, 150,150, 255 255})
	})
		}	} else {
 else {
						drawTextdrawText("SET("SETTINGSTINGS", screenW/", screenW/2-2-80,80, 100 100, color, color.RGBA{255,.RGBA{ 255255, 255, 150,, 150, 255 255})
	})
		}
	}
		y := 		y := 180
		stepY180
		 := stepY := 60
60
				slidersliderX :=X := 500 500
		sl
		sliderW := 300
iderW := 		// Громкость300
		// Г
		if currentромкость
		if currentSettings.LSettings.Language == "ruanguage == "ru" {
			drawText("Г" {
			drawTextромкость("Громкость:", 300,:",  y,300, y, color. color.White)
White)
		} else		} else {
		 {
			drawText("Volume:",	drawText("Volume:", 300 300, y, color, y, color.White.White)
	)
		}
			}
		ebitenutil.DrawRectebitenutil.DrawRect(screen(screen, float, float6464(slider(sliderX),X), float64 float64(y)-(y)-5,5, float64 float64(sliderW), (sliderW), 10,10, color.R color.RGBAGBA{100{100, 100, 100, 255, 100, 100, 255})
		handleX := sliderX + int(g})
		handleX := sliderX + int(g.settings.settingsVolumeSliderVolumeSlider*float*float64(s64(sliderW))
		ebitenutil.DrawRectliderW))
		ebitenutil.DrawRect(screen(screen, float, float64(64(handleX-8), float64(yhandleX-8), float64(y)-15)-15, , 16,16, 24 24, color, color.RG.RGBA{255,BA{255, 200, 100 200, 100, 255})
		drawText(fmt.Sprintf("%d%%, 255})
		drawText(fmt.Sprintf("%d%%", int(g.settingsVolumeSlider*100)), slider", int(g.settingsVolumeSliderX+sliderW+20, y+*100)), sliderX+sliderW+20,5, y+5, color. color.White)
White)
		y		y += step += stepY
		Y
		// Язык// Я
		ifзык
	 currentSettings.L	if currentSettings.Language == "ru" {
anguage == "ru" {
						drawText("ЯdrawText("Язык:", зык:", 300, y,300, y, color. color.White)
White)
			if			if g.s g.settingsLanguageettingsLanguageIndex == 0Index == 0 {
				draw {
				drawText("РусText("Русский",ский", 500 500, y, color, y, color.RGBA{.RGBA{255,255, 255 255, 0, 255})
		, 0, 255})
			} else {
	}				 else {
				drawText("English", 500,drawText("English",  y,500, y, color.RGBA color.RGBA{200{200, , 200,200, 200, 255})
 200, 255})
						}
	}
		}	} else {
			drawText else {
			drawText("Language:", ("Language:", 300,300, y, y, color. color.White)
			ifWhite)
			if g.settingsLanguageIndex == 0 {
				draw g.settingsLanguageIndex == 0 {
				drawText("Text("Russian",Russian", 500 500, y, y, color, color.RG.RGBA{BA{200,200, 200,  200, 200, 255})
		200, 255})
			}	} else {
 else {
								drawTextdrawText("English", 500,("English",  y, color.R500, y,GBA color.RGBA{255, 255,{255, 255, 0 0, , 255})
			255})
			}
	}
		}
		y += step	}
		y += stepY
Y
				// Сложность// Сложность
	
		if currentSettings.L	if currentSettings.Language ==anguage == "ru "ru" {
" {
						drawText("СdrawText("Сложность:", ложность:", 300,300, y, color. y, color.White)
White)
			switch g.settingsDifficulty			switch g.settingsDifficultyIndex {
			caseIndex {
			case 0 0:
			:
				draw	drawText("Text("Лёгкая",Лёгкая 500,", 500, y, color.R y, color.RGBAGBA{100{100, 200,, 200, 100 100, 255})
, 255})
			case 1			case 1:
			:
				drawText("	drawText("СредСредняя", 500няя", 500, y, color.RG, y, color.RGBA{255,BA{255, 255 255, 0,, 0, 255})
		 255})
			case 	case 2:
2:
								drawTextdrawText("Сложная("Сложная", ", 500,500, y, color.RGBA{255, 100, 100 y, color.RGBA{255, 100, 100, , 255})
255})
						}
	}
		}	} else {
			 else {
			drawText("DifficultydrawText("Difficulty:", 300,:", 300, y, y, color. color.White)
White)
			switch			switch g.s g.settingsDifficultyettingsDifficultyIndex {
Index {
			case 0:
						case 0	drawText(":
				drawText("Easy",Easy", 500 500, y, y, color, color.RGBA{.RGBA{100, 200100, 200, , 100,100, 255 255})
			case 1:
				})
			case 1:
drawText("Normal				drawText("Normal", ", 500,500, y, color.RGBA{255 y, color.RGBA{255, 255,, 255, 0 0, 255})
, 255})
			case			case 2 2:
			:
				draw	drawText("Text("Hard",Hard",  500, y, color500, y, color.RG.RGBA{BA{255,255, 100,  100, 100, 255100, 255})
			}
})
			}
				}
	}
		y +=	y += stepY stepY
		//
	 Фонов	// Фоновые анимацииые анимации
	
		if current	if currentSettings.LSettings.Language ==anguage == "ru" {
 "ru" {
			drawText			("ФdrawText("Фоновые анимаоновые анимации:",ции:", 300 300, y, y, color.White, color.White)
			if g)
			if g.settings.settingsAnimationsAnimations {
				draw {
				drawText("ВкText("Вкл",л", 600 600, y, color, y, color.RG.RGBA{100, 255BA{100,, 100, 255, 100, 255 255})
		})
			}	} else {
				 else {
				drawTextdrawText("("Выкл", Выкл", 600,600, y, color.RGBA y, color.RGBA{255{255, 100, 100, , 100, 100, 255})
255})
						}
	}
		} else {
				} else {
			drawText("BGdrawText animations:",("BG animations:", 300, y 300, color, y, color.White.White)
		)
			if g	if g.settings.settingsAnimations {
			Animations {
				drawText("	drawText("ON",ON", 600 600, y, y, color.RGBA{, color.RGBA{100, 255, 100, 255100, 255, 100, 255})
		})
			} else {
	} else {
								drawText("OFFdrawText("OFF", 600, y,", 600, y, color.RGBA color.RGBA{255{255, 100,, 100, 100 100, , 255})
			255})
			}
		}
}
		}
				drawText("ESCdrawText("ESC - назад", screen - назад", screenW/2-80,W/2-80, screenH screenH-50, color-50, color.RG.RGBA{BA{200, 200, 200, 200, 200, 255})
	}
}

200, 255})
	}
}

func (func (g *Game)g *Game) updateMenu updateMenuButtonsLanguage() {
ButtonsLanguage() {
	if currentSettings.L	if currentSettings.Language ==anguage == "ru" {
 "ru" {
		g		g.menuButtons =.menuButtons = []string []string{"Начать игру", "П{"Начать игру", "Продолжитьродол", "жить", "Новая игра",Новая "На игра", "Настройки", "стройки", "ВыйВыйти из игры"}
ти из игры"}
	}	} else {
		g else {
		g.menu.menuButtons =Buttons = []string []string{"Start game",{"Start game", "Continue", " "Continue", "New game", "Settings",New game", "Settings", "Exit "Exit game"}
	}
 game"}
	}
}

func (g}

func (g *Game) Layout(outside *Game) Layout(outsideWidth, outsideHeightWidth, int) outsideHeight int) (int, int (int, int) {
) {
	return screen	return screenW, screenHW, screenH
}

func input
}

func inputPressed() bool {
Pressed()	return e bool {
	return ebitenbiten.IsKey.IsKeyPressed(Pressed(ebiten.KeyEnterebiten.KeyEnter) ||
		ebiten.IsKey) ||
		ebiten.IsKeyPressed(Pressed(ebitenebiten.KeySpace.KeySpace) ||
		ebiten) ||
		ebiten.IsKey.IsKeyPressedPressed(ebiten.KeyUp(ebiten) ||
.KeyUp) ||
				ebitenebiten.IsKey.IsKeyPressed(Pressed(ebiten.KeyDown)ebiten.KeyDown ||
		) ||
ebiten.IsKey		Pressed(ebiten.IsKeyPressed(ebitenebiten.KeyLeft.KeyLeft) ||
) ||
				ebitenebiten.IsKeyPressed(.IsKeyPressed(ebitenebiten.KeyRight.KeyRight)
}

)
}

func minfunc minInt(aInt(a, b int, b int) int {
	if a) int {
	if a < b < b {
	 {
		return a	return a

		}
	return}
	return b
 b
}

//}

// ----- Аудио ----- А (синтезудио (син) -----тез) -----
func
func newSound newSound(ctx(ctx *audio.Context *audio.Context, data, data []byte []byte) *audio.Player {
	d,) *audio.Player {
	d, err := err := wav.Decode(ctx wav.Decode(ctx, bytes.NewReader, bytes.NewReader(data))
	if err(data))
	if err != nil != nil {
		log.Printf {
		log.Printf("w("wav decode err:av decode err: %v", err %v)
	", err)
		return nil
		return nil
	}
	p}
	p, err := audio, err := audio.NewPlayer(ctx, d)
	if err.NewPlayer(ctx, d)
	if err != nil != nil {
	 {
		log.Printf("audio	log.Printf("audio player err player err: %v",: %v", err)
		return nil
	}
 err)
		return nil
	return p
}

	}
	return p
}

func synthWave(sfunc synthWave(sr intr int, dur, freq, dur, freq, amp, amp float64, wave float64, wave string, freq string, freqSSweepweep float64 float64) []int16) []int16 {
	n {
	n := int(float64 := int(float64(sr(sr) *) * dur)
	out := make([] dur)
	outint16, n)
	for := make([]int16, n)
	for i := 0 i := 0; i < n; i++ {
; i < n; i++ {
		t		t := float := float64(i) /64(i) / float64 float64(sr)
	(sr)
		f :=	f := freq + freq + freqS freqSweep*t
		varweep s float*t
		var s float64
64
		switch		switch wave {
		case wave {
		case "s "sine":
ine":
			s = math			s = math.Sin(2.Sin(2 * math.Pi * math * f.Pi * f * t)
	 * t	case ")
		case "square":
square":
			if			if math.Sin(2* math.Sin(2*math.Pi*fmath.Pi*f*t)*t) >=  >= 0 {
0 {
				s = 				s = 1
1
			} else			} else {
				s = {
				s = -1
		 -1
			}
	}
		case		case "noise":
 "noise":
			s = mathrand.N			s = mathrand.NormFloatormFloat64()
64()
		default		default:
			s = math.S:
			s = math.Sin(2 *in(2 * math.Pi * f * t)
 math.Pi * f * t)
				}
}
		att		att, dec, dec, sus, sus, rel, rel := 0. := 0.005,005, 0 0.02.02, , 0.0.6, dur*6, dur*0.3
		0.3
		env :=env := 1 1.0.0
		if t < att
		if t {
			env < att {
		 = t / att	env
	 = t / att
		}	} else if else if t t < att+ < att+dec {
dec {
			env =			env = 1 1 - (t-att)/dec - (t-att)/dec*(1*(1-sus-sus)
	)
		}	} else if else if t > dur-rel {
		 t > dur-rel {
			env = sus	env = sus * (dur - * ( t)dur - t) / rel / rel
	
		}	} else {
		 else {
				env =env = sus
		 sus
}
		val		}
		val := s := s * amp * env * amp * env
	
		if val > 	if val > 1 {
1 {
						val = 1val = 1
		} else if
		} else if val val < -1 < -1 {
		 {
			val = -1
	val = -		}
	1
	out[i		}
	] = int16	out[i] = int16(val *(val * 327 32767)
67)
	}
	return out	}
	return out
}

func mixToW
}

func mixToWAV(sAV(sr int, tracks [][]int16) []r int, tracks [][]int16) []byte {
	maxLenbyte {
	maxLen :=  := 0
	for _,0
	for _, t := t := range tracks range tracks {
		if len(t {
		if len(t) > max) > maxLen {
Len {
			max			maxLen = len(t)
	Len = len(t)
		}
	}
	}
	}
	mix	mix := make([]int := make([]int32, maxLen32,)
	for maxLen)
	for _, t := range tracks {
 _, t := range tracks {
		for i :=		for i := 0 0; i < len; i < len(t);(t); i++ i++ {
		 {
			mix[i]	m += intix[i]32(t[i])
 += int32(t[i])
		}
	}
	var peak int		}
	}
	var peak int32
	for _,32
	for _, v v := range mix := {
		if v range mix {
	 < 	if v < 0 {
0 {
			v			v = -v
 = -v
				}
		if v}
		if v > peak > peak {
		 {
			peak = v	peak
		}
	}
 = v
		}
	scale :=	}
	scale :=  1.0
1.0
	if peak > 32767	if peak >  {
	32767 {
		scale	scale =  = 3276732767.0.0 / float / float64(64(peak)
peak)
	}
	buf	}
	buf := & := &bytes.Bbytes.Buffer{}
uffer{}
	dataSize	dataSize := maxLen * := maxLen * 2 2
	buf.Write
	buf.WriteString("RIFFString("RIFF")
	writeLE")
	Uint32(bufwriteLEUint32(buf, uint32(, uint32(36+36+dataSizedataSize))
	buf.WriteString))
	buf.Write("String("WAVEfmt ")
	writeWAVEfmt ")
LEU	writeLEUint32(buf,int32(buf, 16 16)
	writeLE)
	writeLEUintUint16(buf, 16(buf, 1)
	write1)
LEU	writeLEUint16(bufint16(buf,, 1 1)
	writeLE)
	writeLEUintUint32(buf32(buf, uint, uint32(sr))
	write32(sr))
	writeLEULEUint32(buf,int32(buf, uint32 uint32(sr*2(sr*2))
	))
	writeLEwriteLEUint16(bufUint16(buf, , 2)
2)
	write	writeLEUint16LEUint16(buf,(buf, 16 16)
	b)
	buf.Writeuf.WriteString("String("data")
	writeLEUdata")
	writeLEUint32int32(buf,(buf, uint32 uint32(dataSize(dataSize))
	for))
	for i := i := 0 0; i; i < maxLen; < maxLen; i++ {
		v := i++ int16 {
		v := int16(float64(float64(mix(mix[i]) * scale)
	[i]) * scale)
			__ = binary.Write(buf = binary.Write(buf, binary, binary.Little.LittleEndianEndian, v)
	}
	return buf.B, v)
	}
	returnytes()
 buf.Bytes()
}

func}

func writeLE writeLEUintUint16(w16(w io.W io.Writer,riter, v uint v uint16)16) { _ { _ = binary.Write(w = binary.Write(w, binary, binary.Little.LittleEndian, v) }
Endian, vfunc write) }
func writeLEUint32LEUint32(w io.Writer(w io.Writer, v, v uint32 uint32) { _ =) { _ = binary.Write(w, binary.LittleEnd binary.Write(w, binary.Lian,ittleEndian, v) v) }

func }

func snd sndEatEat() []byte {
() []byte {
	sr := 	sr44100
	t := 441001 :=
	t1 := synthWave synthWave(sr(sr, 0., 0.1, 6001, 600, 0., 5, "s0.5,ine", "sine", 400)
	t 4002 := synthWave(sr)
	t2 := synthWave(sr, , 0.1,0.1, 1200, 1200, 0 0.3.3, ", "sinesine", 200)
", 200)
	return mixToWAV(s	return mixToWAV(sr,r, [][] [][]int16int16{t1, t{t1, t2})
2})
}
func sndBoom}
func sndBoom() []byte {
() []byte {
	sr	sr := 441 := 4410000
	low := synthWave
	low := synthWave(sr, (sr, 0.3,0. 803, 80, 0., 0.9,9, "sine", "sine", -30 -30)
	n := synth)
	n := synthWave(sr,Wave(s 0r, 0.3, .30,, 0, 0.7 0.7, ", "noisenoise", 0)
	mid", 0)
 := synth	midWave(sr, 0 := synthWave(sr, 0.15.15, 300, 0.4, 300, 0.4, "square", -200, "square",)
	return -200)
	return mixTo mixToWAVWAV(sr,(sr, [][]int [][]int16{low,16{low, n, mid})
 n, mid})
}
func sndHe}
func sndHeal() []al() []byte {
byte {
	sr	sr :=  := 4410044100
	t
	t1 := synthWave(sr, 1 := synthWave(sr0.15,, 0.15, 400 400, , 0.0.4,4, "s "sine",ine", 300 300)
	t)
	t2 := synthWave(sr2 := synthWave(sr, , 0.0.2,2, 800 800, 0., 0.3, "s3, "sine", 100)
	return mixToine", 100)
	return mixToWAV(srWAV(sr, [][]int16{t, [][]int16{t1,1, t2 t2})
}
})
}
func sfunc sndPause() []bytendPause() {
	s []byte {
	sr :=r := 441 44100
00
	t :=	t := synthWave synthWave(sr(sr, 0., 0.08,08, 220,  220, 0.0.4,4, "square "square", -", -50)
50)
	return mix	return mixToWAV(sToWAV(sr, [][]int16{t})
r, [][]int16{t})
}
func}
func snd sndMenuMove() []MenuMove() []byte {
	srbyte {
	sr := 44100
	t := synthWave(sr, 0 := 44100
	t := synthWave(sr, 0.05.05, , 800,800, 0 0.3.3, ", "sinesine", ", 100)
100)
	return mix	return mixToWAV(sr,ToWAV(sr, [][]int16 [][]int16{t})
}
func snd{t})
}
func sndMenuSelectMenuSelect() []byte {
() []byte {
	sr	sr := 44100 := 44100
	t
	t1 :=1 := synthWave synthWave(sr(sr, 0., 0.1, 4001, 400, , 0.5, "s0.5, "sine", 200ine", 200)
	t)
	t2 :=2 := synthWave synthWave(sr, (sr, 0.0.1,1, 700, 0.5, 700, 0.5, "s "sine",ine", -100 -100)
	return mixTo)
	returnWAV mixToWAV(sr(sr, [][]int, [][]int16{t16{t1, t2})
}
1, t2})
}
func sfunc sndGhost() []ndGhost() []byte {
byte {
	sr := 	sr := 4410044100
	t
	t1 :=1 := synthWave(sr synthWave, (sr, 0.2, 5000.2, 500, 0., 0.4, "sine",4, -200 "sine", -200)
	t2 :=)
	t2 := synthWave(sr synthWave(sr, , 0.0.2,2, 800 800, , 0.0.3,3, "sine", "sine", 100 100)
	return)
	return mixTo mixToWAVWAV(sr(sr, [][]int,16{t1, [][]int16{t1, t2})
}
 t2})
}
func sfunc sndKeyndKey() []byte {
	sr() []byte {
 := 	sr := 44100
	t44100
	t1 :=1 := synthWave synthWave(sr(sr, , 0.0.15,15, 880,  8800.5,, 0. "s5, "sine",ine", -400 -400)
	t)
	t2 :=2 := synthWave synthWave(sr(sr, 0., 15, 4400.15,, 0.4, 440, 0.4, "s "sine", -200)
	returnine", - mixToWAV200)
	return mixTo(srWAV(sr,, [][]int [][]int16{t16{t1,1, t2})
}
 t2func sndKey})
}
func sUse()ndKeyUse() []byte {
	s []byte {
	sr :=r := 441 44100
00
	t1	t1 := synth := synthWave(sWave(sr,r, 0.1 0.1, 600,, 600, 0.5 0.5, ", "sinesine", -200)
	t2", -200)
 := synth	t2 := synthWave(sr,Wave(s 0r, 0..1, 800,1, 800, 0 0.4.4, ", "sinesine", ", 100)
	return mix100)
ToW	return mixToWAV(sAV(sr, [][]r, [][]int16int16{t1{t1, t2})
}
func snd, t2})
}
func sndGiftGiftOpen() []byteOpen() []byte {
	s {
	sr :=r := 44100
 44100
	t1	t1 := synth := synthWave(sWave(sr, 0r, 0.2.2, , 300,300, 0 0.6.6, ", "sine", sine400)
	t2", 400)
 := synth	t2 := synthWave(sr,Wave(sr, 0 0.2, 600,.2, 600, 0.5 0, "sine.5, "sine", -", -300)
300)
	return mixToW	return mixToWAV(sr,AV(s [][]int16r, [][]{t1int16{t1, t, t2})
2})
}
func}
func sndCoin() snd []byteCoin() []byte {
	sr := {
	s 44100
r := 441	t :=00
	t := synthWave(sr synthWave(sr, , 0.1,0.1, 100 1000,0,  0.4, "0.4sine, "sine", -600", -600)
	return mix)
	return mixToWAV(sr,ToWAV(sr, [][] [][]intint1616{t})
{t})
}

func}

func main() main() {
	 {
	ebiten.SetWindowebiten.SetWindowSize(sSize(screenWcreenW, screenH)
, screenH)
	eb	ebiten.Setiten.SetWindowTitleWindowTitle("З("Змейкамейка:: Возр Возрождение")
	ebождение")
	ebiten.SetFullscreeniten.Set(true)
Fullscreen(true)
	if err := e	if err := ebitenbiten.RunGame(New.RunGame(NewGame()); errGame()); != nil {
 err != nil {
		log.Fatal		log(err)
.Fatal(err)
	}
	}
}