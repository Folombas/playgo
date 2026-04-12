package game

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"
	"os"
	"time"

	"match3_v2/internal/audio"
	"match3_v2/internal/logic"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 800
	gridSize     = 8
	tileSize     = 70
	gridOffset   = 130
	scoreY       = 60
	timerY       = 30
)

// DragState состояние перетаскивания
type DragState int

const (
	DragNone DragState = iota
	DragHovering      // Наведение на гем
	DragPicking       // Поднимаем гем
	Dragging          // Перетаскиваем
	DragSnapping      // Прилипаем к новой позиции
	DragReturning     // Возвращаем на место
	DragShaking       // Трясём при ошибке
	DragSwiping       // Свайп через одинаковые
)

// SwipeLine линия свайпа
type SwipeLine struct {
	Points     []image.Point // Ячейки на пути
	GemType    int           // Тип гема
	LinePoints []struct{ X, Y float64 } // Точки для рисования линии
	Active     bool
	Time       float64
}

// Game основной игровой объект
type Game struct {
	board         *logic.Board
	audioMgr      *audio.SoundManager
	score         int
	combo         int
	maxCombo      int
	timeLeft      int
	level         int
	targetScore   int
	gameOver      bool
	levelComplete bool
	hintTime      time.Time
	hintPair      [2]image.Point
	showHint      bool
	particles     []Particle
	trails        []Trail
	stars         []Star
	fallingGems   []FallingGem
	gemImages     map[int]*ebiten.Image
	selectorImg   *ebiten.Image
	rng           *rand.Rand

	// Drag & Drop
	dragState      DragState
	dragGem        image.Point  // Какой гем тащим
	dragStartPos   image.Point  // Начальная позиция
	dragCurrentX   float64      // Текущая X мыши
	dragCurrentY   float64      // Текущая Y мыши
	dragTargetPos  image.Point  // Целевая позиция
	dragAnimTime   float64      // Время анимации
	dragScale      float64      // Масштаб перетаскиваемого гема
	dragRotation   float64      // Вращение при перетаскивании
	dragGlowPulse  float64      // Пульсация свечения
	hoverGem       image.Point  // Гем под курсором
	hoverTime      float64      // Время наведения

	// Swipe
	swipeLine      SwipeLine
	swipeTimer     float64
	swipeMatched   bool
	screenShake    float64
}

// Particle частица для эффектов
type Particle struct {
	X, Y    float64
	VX, VY  float64
	Life    float64
	Color   color.Color
	Size    float64
	MaxLife float64
}

// Trail след при перетаскивании
type Trail struct {
	X, Y  float64
	Life  float64
	Color color.Color
	Size  float64
}

// Star звезда на фоне
type Star struct {
	X, Y    float64
	Size    float64
	Twinkle float64
	Speed   float64
}

// FallingGem анимация падения гема
type FallingGem struct {
	Row, Col int
	GemType  int
	X, Y     float64
	TargetY  float64
	Speed    float64
	Scale    float64
	Rotation float64
}

// NewGame создаёт новую игру
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	g := &Game{
		board:       logic.NewBoard(gridSize, rng),
		audioMgr:    audio.NewSoundManager(),
		score:       0,
		combo:       0,
		maxCombo:    0,
		level:       1,
		targetScore: 500,
		timeLeft:    120,
		gameOver:    false,
		levelComplete: false,
		hintTime:    time.Now(),
		showHint:    false,
		particles:   make([]Particle, 0),
		trails:      make([]Trail, 0),
		fallingGems: make([]FallingGem, 0),
		gemImages:   make(map[int]*ebiten.Image),
		rng:         rng,
		dragState:   DragNone,
		dragGem:     image.Point{-1, -1},
		hoverGem:    image.Point{-1, -1},
		dragScale:   1.0,
	}

	g.stars = g.generateStars(100)

	g.loadImages()

	// Таймер
	go func() {
		for !g.gameOver {
			time.Sleep(1 * time.Second)
			g.timeLeft--
			if g.timeLeft <= 0 {
				g.gameOver = true
			}
		}
	}()

	return g
}

// loadImages загружает спрайты
func (g *Game) loadImages() {
	gems := []string{
		"assets/gem_red.png",
		"assets/gem_blue.png",
		"assets/gem_green.png",
		"assets/gem_yellow.png",
		"assets/gem_purple.png",
	}

	for i, path := range gems {
		f, err := os.Open(path)
		if err != nil {
			fmt.Printf("Warning: could not load %s: %v\n", path, err)
			continue
		}
		img, _, err := image.Decode(f)
		f.Close()
		if err != nil {
			continue
		}
		eimg := ebiten.NewImageFromImage(img)
		g.gemImages[i] = eimg
	}

	f, err := os.Open("assets/selector.png")
	if err == nil {
		img, _, err := image.Decode(f)
		f.Close()
		if err == nil {
			g.selectorImg = ebiten.NewImageFromImage(img)
		}
	}
}

func (g *Game) Update() error {
	if g.gameOver {
		if inpututil.IsKeyJustPressed(ebiten.KeyR) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.resetGame()
		}
		return nil
	}

	// Обработка drag & drop
	g.handleDragDrop()

	// Подсказки
	if time.Since(g.hintTime) > 5*time.Second && g.dragState == DragNone {
		g.showHint = true
		g.hintPair = g.board.FindHint()
		g.audioMgr.Play(audio.SoundHint)
	}

	// Обновление частиц и следов
	g.updateParticles()
	g.updateTrails()
	g.updateStars()
	g.updateFallingGems()

	// Тряска экрана
	if g.screenShake > 0 {
		g.screenShake -= 1.0 / 60.0
	}

	// Проверка уровня
	if g.score >= g.targetScore && !g.levelComplete {
		g.levelComplete = true
		g.audioMgr.Play(audio.SoundCombo)
		g.spawnLevelUpEffect()
	}

	// Проверка квадрата 2x2 после каждого хода
	g.checkSquareMatch()

	// Переход на следующий уровень
	if g.levelComplete && inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		g.nextLevel()
	}

	// Рестарт
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		g.resetGame()
	}

	return nil
}

func (g *Game) resetGame() {
	g.board = logic.NewBoard(gridSize, g.rng)
	g.score = 0
	g.combo = 0
	g.timeLeft = 120
	g.gameOver = false
	g.particles = nil
	g.trails = nil
	g.dragState = DragNone
	g.dragGem = image.Point{-1, -1}
	g.hoverGem = image.Point{-1, -1}
}

func (g *Game) generateStars(count int) []Star {
	stars := make([]Star, count)
	for i := range stars {
		stars[i] = Star{
			X:       g.rng.Float64() * screenWidth,
			Y:       g.rng.Float64() * screenHeight,
			Size:    1 + g.rng.Float64()*2,
			Twinkle: g.rng.Float64() * math.Pi * 2,
			Speed:   0.5 + g.rng.Float64()*2,
		}
	}
	return stars
}

func (g *Game) updateStars() {
	for i := range g.stars {
		g.stars[i].Twinkle += g.stars[i].Speed / 60.0
	}
}

func (g *Game) drawStars(screen *ebiten.Image) {
	for _, s := range g.stars {
		alpha := uint8(100 + 155*(math.Sin(s.Twinkle)*0.5+0.5))
		vector.DrawFilledCircle(screen, float32(s.X), float32(s.Y), float32(s.Size),
			color.RGBA{200, 220, 255, alpha}, false)
	}
}

func (g *Game) checkSquareMatch() {
	// Проверяем все квадраты 2x2 на поле
	for r := 0; r < gridSize-1; r++ {
		for c := 0; c < gridSize-1; c++ {
			gem := g.board.Get(r, c)
			if gem < 0 {
				continue
			}

			// Проверяем квадрат 2x2
			if g.board.Get(r, c+1) == gem &&
				g.board.Get(r+1, c) == gem &&
				g.board.Get(r+1, c+1) == gem {

				// КВАДРАТ НАЙДЕН! УДАЛЯЕМ!
				g.spawnSquareExplosion(r, c, gem)

				// Удалить все 4 гема
				g.board.Set(r, c, -1)
				g.board.Set(r, c+1, -1)
				g.board.Set(r+1, c, -1)
				g.board.Set(r+1, c+1, -1)

				// Бонусные очки!
				score := 200
				g.score += score

				// Заполнить пустоты
				g.board.FillEmpty()

				// Звук
				g.audioMgr.Play(audio.SoundCombo)
				return
			}
		}
	}
}

func (g *Game) spawnSquareExplosion(row, col, gemType int) {
	// Центр квадрата
	cx := float64(gridOffset + col*tileSize + tileSize)
	cy := float64(gridOffset + row*tileSize + tileSize)

	colors := []color.RGBA{
		{255, 80, 50, 255},
		{50, 100, 255, 255},
		{80, 255, 100, 255},
		{255, 255, 50, 255},
		{200, 80, 255, 255},
	}
	baseColor := colors[0]
	if gemType < len(colors) {
		baseColor = colors[gemType]
	}

	// 1. ОГРОМНЫЙ ВЗРЫВ - 50 частиц
	for i := 0; i < 50; i++ {
		angle := float64(i) * math.Pi * 2 / 50
		speed := 5 + g.rng.Float64()*10
		g.particles = append(g.particles, Particle{
			X:       cx,
			Y:       cy,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed - 4,
			Life:    2.0,
			MaxLife: 2.0,
			Size:    6 + g.rng.Float64()*10,
			Color:   baseColor,
		})
	}

	// 2. БЕЛЫЕ ИСКРЫ - 30 штук
	for i := 0; i < 30; i++ {
		angle := g.rng.Float64() * math.Pi * 2
		speed := 8 + g.rng.Float64()*12
		g.particles = append(g.particles, Particle{
			X:       cx,
			Y:       cy,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    1.5,
			MaxLife: 1.5,
			Size:    3 + g.rng.Float64()*4,
			Color:   color.RGBA{255, 255, 255, 255},
		})
	}

	// 3. ЭФФЕКТ УДАРНОЙ ВОЛНЫ - расширяющийся круг
	for i := 0; i < 20; i++ {
		radius := 10.0 + float64(i)*8
		g.particles = append(g.particles, Particle{
			X:       cx,
			Y:       cy,
			VX:      0,
			VY:      0,
			Life:    1.0,
			MaxLife: 1.0,
			Size:    radius,
			Color:   color.RGBA{255, 255, 255, uint8(100 - i*5)},
		})
	}

	// 4. КВАДРАТНЫЕ ЧАСТИЦЫ (визуальный эффект квадрата)
	for dx := 0; dx < 2; dx++ {
		for dy := 0; dy < 2; dy++ {
			x := float64(gridOffset + (col+dy)*tileSize + tileSize/2)
			y := float64(gridOffset + (row+dx)*tileSize + tileSize/2)

			for i := 0; i < 10; i++ {
				angle := float64(i) * math.Pi * 2 / 10
				speed := 3 + g.rng.Float64()*5
				g.particles = append(g.particles, Particle{
					X:       x,
					Y:       y,
					VX:      math.Cos(angle) * speed,
					VY:      math.Sin(angle) * speed - 2,
					Life:    1.5,
					MaxLife: 1.5,
					Size:    5 + g.rng.Float64()*5,
					Color:   baseColor,
				})
			}
		}
	}

	// Экран трясётся
	g.screenShake = 0.3
}

func (g *Game) nextLevel() {
	g.level++
	g.targetScore = 500 * g.level
	g.timeLeft = 120
	g.levelComplete = false
	g.board = logic.NewBoard(gridSize, g.rng)
	g.combo = 0
	g.audioMgr.Play(audio.SoundHint)
}

func (g *Game) updateFallingGems() {
	for i := len(g.fallingGems) - 1; i >= 0; i-- {
		gem := &g.fallingGems[i]
		gem.Y += gem.Speed
		gem.Speed += 0.5
		gem.Rotation += 0.05

		if gem.Y >= gem.TargetY {
			gem.Y = gem.TargetY
			gem.Speed = -gem.Speed * 0.3 // Отскок
			gem.Scale = 1.1
			if math.Abs(gem.Speed) < 1 {
				g.fallingGems = append(g.fallingGems[:i], g.fallingGems[i+1:]...)
			}
		}

		gem.Scale += (1.0 - gem.Scale) * 0.2
	}
}

func (g *Game) spawnFallingGem(row, col, gemType int) {
	g.fallingGems = append(g.fallingGems, FallingGem{
		Row:     row,
		Col:     col,
		GemType: gemType,
		X:       float64(gridOffset + col*tileSize + tileSize/2),
		Y:       float64(gridOffset - tileSize),
		TargetY: float64(gridOffset + row*tileSize + tileSize/2),
		Speed:   2,
		Scale:   1.0,
		Rotation: 0,
	})
}

func (g *Game) drawFallingGems(screen *ebiten.Image) {
	for _, gem := range g.fallingGems {
		g.drawGem(screen,
			int(gem.X-tileSize/2+5),
			int(gem.Y-tileSize/2+5),
			gem.GemType,
			gem.Scale,
			gem.Rotation,
			1.0)
	}
}

func (g *Game) spawnLevelUpEffect() {
	// Эффект фейерверка
	for i := 0; i < 50; i++ {
		angle := float64(i) * math.Pi * 2 / 50
		speed := 5 + g.rng.Float64()*8
		g.particles = append(g.particles, Particle{
			X:       screenWidth / 2,
			Y:       screenHeight / 2,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    1.5,
			MaxLife: 1.5,
			Size:    5 + g.rng.Float64()*10,
			Color:   color.RGBA{255, 215, 0, 255},
		})
	}
}

func (g *Game) handleDragDrop() {
	mx, my := ebiten.CursorPosition()

	// Обновление свайп-таймера
	if g.dragState == DragSwiping {
		g.swipeTimer += 1.0 / 60.0
	}

	// Определяем ячейку под курсором
	col := (mx - gridOffset) / tileSize
	row := (my - gridOffset) / tileSize
	hoverValid := row >= 0 && row < gridSize && col >= 0 && col < gridSize
	hoverPos := image.Point{row, col}

	switch g.dragState {
	case DragNone:
		// Начало перетаскивания ИЛИ свайпа
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && hoverValid {
			g.dragState = DragPicking
			g.dragGem = hoverPos
			g.dragStartPos = hoverPos
			g.dragCurrentX = float64(mx)
			g.dragCurrentY = float64(my)
			g.dragAnimTime = 0
			g.dragScale = 1.0
			g.dragRotation = 0
			g.showHint = false
			g.hintTime = time.Now()

			// Эффект поднятия
			g.spawnLiftEffect(hoverPos)
			g.audioMgr.Play(audio.SoundPickUp)

			// Инициализация свайпа
			g.swipeLine = SwipeLine{
				Points:     []image.Point{hoverPos},
				GemType:    g.board.Get(hoverPos.X, hoverPos.Y),
				Active:     true,
				Time:       0,
				LinePoints: []struct{ X, Y float64 }{{float64(mx), float64(my)}},
			}
			g.swipeTimer = 0
			g.swipeMatched = false
		} else if hoverValid {
			g.hoverGem = hoverPos
			g.hoverTime += 1.0 / 60.0
		} else {
			g.hoverGem = image.Point{-1, -1}
			g.hoverTime = 0
		}

	case DragPicking:
		// Поднимаем гем
		g.dragAnimTime += 1.0 / 60.0
		g.dragScale = 1.0 + math.Sin(g.dragAnimTime*math.Pi)*0.3
		g.dragRotation = math.Sin(g.dragAnimTime*math.Pi*2) * 0.1

		if g.dragAnimTime > 0.15 {
			g.dragState = Dragging
			g.dragAnimTime = 0
			g.dragScale = 1.2
		}

	case Dragging:
		// Перетаскиваем
		g.dragCurrentX = float64(mx)
		g.dragCurrentY = float64(my)
		g.dragRotation = math.Sin(float64(time.Now().UnixMilli())*0.005) * 0.15
		g.dragGlowPulse = math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7

		// Добавляем следы
		if len(g.trails) == 0 || time.Now().UnixMilli()%3 == 0 {
			gemType := g.board.Get(g.dragGem.X, g.dragGem.Y)
			g.trails = append(g.trails, Trail{
				X:     g.dragCurrentX,
				Y:     g.dragCurrentY,
				Life:  0.5,
				Color: g.getColorForGem(gemType),
				Size:  8,
			})
		}

		// Проверяем свайп - если мышь на другой ячейке того же типа
		if hoverValid {
			gemType := g.board.Get(row, col)
			if gemType == g.swipeLine.GemType && gemType >= 0 {
				// Проверяем, не добавляли ли уже эту ячейку
				alreadyAdded := false
				for _, p := range g.swipeLine.Points {
					if p.X == row && p.Y == col {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					// Проверяем соседство (ВКЛЮЧАЯ ДИАГОНАЛИ!)
					lastPt := g.swipeLine.Points[len(g.swipeLine.Points)-1]
					dr := abs(lastPt.X - row)
					dc := abs(lastPt.Y - col)
					
					// Соседство включая диагонали: max(dr, dc) == 1
					if dr <= 1 && dc <= 1 && (dr+dc) > 0 {
						g.swipeLine.Points = append(g.swipeLine.Points, image.Point{row, col})
						g.swipeLine.LinePoints = append(g.swipeLine.LinePoints, struct{ X, Y float64 }{float64(mx), float64(my)})

						// Звук свайпа
						g.audioMgr.Play(audio.SoundSwap)

						// Частицы на пути
						x := float64(gridOffset + col*tileSize + tileSize/2)
						y := float64(gridOffset + row*tileSize + tileSize/2)
						for i := 0; i < 5; i++ {
							angle := float64(i) * math.Pi * 2 / 5
							g.particles = append(g.particles, Particle{
								X:       x,
								Y:       y,
								VX:      math.Cos(angle) * 2,
								VY:      math.Sin(angle) * 2,
								Life:    0.5,
								MaxLife: 0.5,
								Size:    3 + g.rng.Float64()*3,
								Color:   g.getColorForGem(gemType),
							})
						}
					}
				}
			}
		}

		// Если перетащили на соседнюю ячейку другого типа - обычный обмен
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			if len(g.swipeLine.Points) >= 3 {
				// УДАЛИТЬ СВАЙП!
				g.swipeMatched = true
				g.processSwipeMatch()
			} else if hoverValid && g.isValidSwap(g.dragGem, hoverPos) {
				// Обычный обмен
				g.dragTargetPos = hoverPos
				g.dragState = DragSnapping
				g.dragAnimTime = 0

				g.board.Swap(g.dragGem.X, g.dragGem.Y, hoverPos.X, hoverPos.Y)
				g.audioMgr.Play(audio.SoundDrop)

				matches := g.board.FindMatches()
				if len(matches) > 0 {
					g.processMatches(matches)
					g.combo++
					if g.combo > g.maxCombo {
						g.maxCombo = g.combo
					}
					g.spawnSuccessEffect(hoverPos)

					if g.combo >= 3 {
						g.audioMgr.Play(audio.SoundCombo)
					} else {
						g.audioMgr.Play(audio.SoundMatch)
					}
				} else {
					g.board.Swap(g.dragGem.X, g.dragGem.Y, hoverPos.X, hoverPos.Y)
					g.dragState = DragShaking
					g.dragAnimTime = 0
					g.spawnFailEffect()
					g.audioMgr.Play(audio.SoundFail)
				}
			} else {
				g.dragTargetPos = g.dragStartPos
				g.dragState = DragReturning
				g.dragAnimTime = 0
				g.audioMgr.Play(audio.SoundSnap)
			}

			if !g.swipeMatched {
				g.swipeLine.Active = false
			}
		}

	case DragSwiping:
		// Анимация после свайпа
		if g.swipeTimer > 0.5 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
			g.swipeLine = SwipeLine{}
		}

	case DragSnapping:
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.2, 1.0)
		g.dragScale = 1.2 - 0.2*t

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
		}

	case DragReturning:
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.3, 1.0)
		g.dragScale = 1.2 - 0.2*t

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
		}

	case DragShaking:
		g.dragAnimTime += 1.0 / 60.0
		t := math.Min(g.dragAnimTime/0.4, 1.0)
		g.dragRotation = math.Sin(t*math.Pi*8) * 0.2 * (1 - t)
		g.dragScale = 1.0

		if t >= 1.0 {
			g.dragState = DragNone
			g.dragGem = image.Point{-1, -1}
			g.dragRotation = 0
		}
	}
}

func (g *Game) isValidSwap(from, to image.Point) bool {
	// Проверить что это соседние ячейки
	dr := abs(from.X - to.X)
	dc := abs(from.Y - to.Y)
	return dr+dc == 1
}

func (g *Game) processSwipeMatch() {
	// УДАЛЕНИЕ ГЕМОВ НА ПУТИ СВАЙПА!
	points := g.swipeLine.Points
	gemType := g.swipeLine.GemType

	// Очки
	score := len(points) * 20
	if len(points) >= 5 {
		score = 150
	} else if len(points) == 4 {
		score = 80
	}
	g.score += score
	g.combo++

	// КРУТОЙ ЭФФЕКТ МОЛНИИ!
	for _, p := range points {
		x := float64(gridOffset + p.Y*tileSize + tileSize/2)
		y := float64(gridOffset + p.X*tileSize + tileSize/2)

		// Взрыв частиц
		for i := 0; i < 15; i++ {
			angle := float64(i) * math.Pi * 2 / 15
			speed := 4 + g.rng.Float64()*6
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * speed,
				VY:      math.Sin(angle) * speed - 3,
				Life:    1.5,
				MaxLife: 1.5,
				Size:    5 + g.rng.Float64()*8,
				Color:   g.getColorForGem(gemType),
			})
		}

		// Электрические искры
		for i := 0; i < 8; i++ {
			angle := g.rng.Float64() * math.Pi * 2
			speed := 6 + g.rng.Float64()*8
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * speed,
				VY:      math.Sin(angle) * speed,
				Life:    0.8,
				MaxLife: 0.8,
				Size:    2 + g.rng.Float64()*3,
				Color:   color.RGBA{255, 255, 255, 255},
			})
		}

		// Удалить гем
		g.board.Set(p.X, p.Y, -1)
	}

	// Заполнить пустоты
	g.board.FillEmpty()

	// Звук молнии!
	g.audioMgr.Play(audio.SoundMatch)
	if len(points) >= 4 {
		g.audioMgr.Play(audio.SoundCombo)
	}

	// Перейти в состояние свайпа
	g.dragState = DragSwiping
	g.swipeTimer = 0
}

func (g *Game) processMatches(matches []image.Point) {
	score := len(matches) * 10

	if len(matches) == 4 {
		score = 50
	} else if len(matches) >= 5 {
		score = 100
	}

	if g.combo > 1 {
		score = score * (1 + g.combo/2)
	}

	g.score += score

	// Частицы
	for _, m := range matches {
		x := float64(gridOffset + m.Y*tileSize + tileSize/2)
		y := float64(gridOffset + m.X*tileSize + tileSize/2)
		for i := 0; i < 12; i++ {
			angle := float64(i) * math.Pi * 2 / 12
			speed := 3 + g.rng.Float64()*4
			g.particles = append(g.particles, Particle{
				X:       x,
				Y:       y,
				VX:      math.Cos(angle) * speed,
				VY:      math.Sin(angle) * speed - 3,
				Life:    1.0,
				MaxLife: 1.0,
				Size:    4 + g.rng.Float64()*6,
				Color:   g.getColorForGem(g.board.Get(m.X, m.Y)),
			})
		}
	}

	g.board.RemoveMatches(matches)
	g.board.FillEmpty()
}

func (g *Game) spawnLiftEffect(pos image.Point) {
	x := float64(gridOffset + pos.Y*tileSize + tileSize/2)
	y := float64(gridOffset + pos.X*tileSize + tileSize/2)
	gemType := g.board.Get(pos.X, pos.Y)
	c := g.getColorForGem(gemType)

	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi * 2 / 8
		g.particles = append(g.particles, Particle{
			X:       x,
			Y:       y,
			VX:      math.Cos(angle) * 2,
			VY:      math.Sin(angle) * 2,
			Life:    0.5,
			MaxLife: 0.5,
			Size:    3 + g.rng.Float64()*3,
			Color:   c,
		})
	}
}

func (g *Game) spawnSuccessEffect(pos image.Point) {
	x := float64(gridOffset + pos.Y*tileSize + tileSize/2)
	y := float64(gridOffset + pos.X*tileSize + tileSize/2)

	// Золотой всплеск
	for i := 0; i < 20; i++ {
		angle := float64(i) * math.Pi * 2 / 20
		speed := 4 + g.rng.Float64()*5
		g.particles = append(g.particles, Particle{
			X:       x,
			Y:       y,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed - 2,
			Life:    1.2,
			MaxLife: 1.2,
			Size:    5 + g.rng.Float64()*8,
			Color:   color.RGBA{255, 215, 0, 255},
		})
	}
}

func (g *Game) spawnFailEffect() {
	// Красные частицы
	for i := 0; i < 10; i++ {
		angle := g.rng.Float64() * math.Pi * 2
		speed := 2 + g.rng.Float64()*3
		g.particles = append(g.particles, Particle{
			X:       g.dragCurrentX,
			Y:       g.dragCurrentY,
			VX:      math.Cos(angle) * speed,
			VY:      math.Sin(angle) * speed,
			Life:    0.6,
			MaxLife: 0.6,
			Size:    3 + g.rng.Float64()*4,
			Color:   color.RGBA{255, 50, 50, 255},
		})
	}
}

func (g *Game) getColorForGem(gem int) color.Color {
	colors := []color.Color{
		color.RGBA{255, 80, 80, 255},
		color.RGBA{80, 80, 255, 255},
		color.RGBA{80, 255, 80, 255},
		color.RGBA{255, 255, 80, 255},
		color.RGBA{180, 80, 255, 255},
	}
	if gem >= 0 && gem < len(colors) {
		return colors[gem]
	}
	return color.White
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := &g.particles[i]
		p.X += p.VX
		p.Y += p.VY
		p.VY += 0.25
		p.Life -= 1.0 / 60.0

		if p.Life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updateTrails() {
	for i := len(g.trails) - 1; i >= 0; i-- {
		g.trails[i].Life -= 1.0 / 60.0
		if g.trails[i].Life <= 0 {
			g.trails = append(g.trails[:i], g.trails[i+1:]...)
		}
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Фон
	screen.Fill(color.RGBA{10, 5, 30, 255})

	// Звёзды
	g.drawStars(screen)

	// Заголовок
	g.drawHeader(screen)

	// Игровое поле
	g.drawBoard(screen)

	// Следы перетаскивания
	g.drawTrails(screen)

	// Падающие гемы
	g.drawFallingGems(screen)

	// Частицы
	g.drawParticles(screen)

	// Перетаскиваемый гем
	if g.dragState == DragPicking || g.dragState == Dragging ||
		g.dragState == DragSnapping || g.dragState == DragReturning ||
		g.dragState == DragShaking {
		g.drawDraggedGem(screen)
	}

	// Линия свайпа
	if g.dragState == Dragging || (g.dragState == DragSwiping && g.swipeTimer < 0.5) {
		g.drawSwipeLine(screen)
	}

	// Game Over
	if g.gameOver {
		g.drawGameOver(screen)
	}

	// Тряска экрана - рисуем поверх всё с вибрацией
	if g.screenShake > 0 {
		shakeX := math.Sin(float64(time.Now().UnixMilli())*0.05) * 10 * g.screenShake
		shakeY := math.Cos(float64(time.Now().UnixMilli())*0.05) * 10 * g.screenShake
		
		// Белая вспышка
		alpha := uint8(g.screenShake * 100)
		vector.DrawFilledRect(screen, 0, 0, screenWidth, screenHeight,
			color.RGBA{255, 255, 255, alpha}, false)
		
		// Сдвиг курсора для эффекта
		_ = shakeX
		_ = shakeY
	}
}

func (g *Game) drawHeader(screen *ebiten.Image) {
	drawRoundedRect(screen, 20, 10, 760, 100, 15, color.RGBA{40, 30, 70, 200})

	// Уровень
	g.drawCenteredText(screen, fmt.Sprintf("LEVEL %d", g.level), screenWidth/2, 15, 14, color.RGBA{100, 200, 255, 255})

	// Прогресс-бар к следующему уровню
	progress := float64(g.score) / float64(g.targetScore)
	if progress > 1 {
		progress = 1
	}
	barX := 250.0
	barW := 300.0
	drawRoundedRect(screen, barX, 32, barW, 8, 4, color.RGBA{30, 20, 50, 255})
	drawRoundedRect(screen, barX, 32, barW*progress, 8, 4, color.RGBA{100, 200, 255, 255})

	// Счёт
	g.drawCenteredText(screen, fmt.Sprintf("%d", g.score), screenWidth/2, scoreY+15, 32, color.RGBA{255, 215, 0, 255})

	mins := g.timeLeft / 60
	secs := g.timeLeft % 60
	timeColor := color.RGBA{100, 255, 100, 255}
	if g.timeLeft < 30 {
		timeColor = color.RGBA{255, 100, 100, 255}
	}
	g.drawCenteredText(screen, "TIME", 120, timerY, 14, color.RGBA{150, 150, 200, 255})
	g.drawCenteredText(screen, fmt.Sprintf("%d:%02d", mins, secs), 120, timerY+20, 24, timeColor)

	if g.combo > 1 {
		g.drawCenteredText(screen, fmt.Sprintf("COMBO x%d!", g.combo), 680, timerY+10, 20, color.RGBA{255, 165, 0, 255})
	}

	g.drawCenteredText(screen, "R - Restart | Drag gems", 680, timerY+35, 12, color.RGBA{150, 150, 200, 255})
}

func (g *Game) drawBoard(screen *ebiten.Image) {
	// Фон доски
	drawRoundedRect(screen, float64(gridOffset-10), float64(gridOffset-10),
		float64(gridSize*tileSize+20), float64(gridSize*tileSize+20), 10,
		color.RGBA{30, 20, 60, 255})

	// Ячейки
	for r := 0; r < gridSize; r++ {
		for c := 0; c < gridSize; c++ {
			x := gridOffset + c*tileSize
			y := gridOffset + r*tileSize

			// Фон ячейки
			cellColor := color.RGBA{50, 40, 90, 255}
			if (r+c)%2 == 0 {
				cellColor = color.RGBA{60, 50, 100, 255}
			}

			// Подсветка при наведении
			if g.hoverGem.X == r && g.hoverGem.Y == c && g.dragState == DragNone {
				pulse := math.Sin(float64(time.Now().UnixMilli())*0.006)*0.15 + 0.15
				r2 := uint8(float64(cellColor.R) + 50*pulse)
				g2 := uint8(float64(cellColor.G) + 50*pulse)
				b2 := uint8(float64(cellColor.B) + 80*pulse)
				cellColor = color.RGBA{r2, g2, b2, 255}
			}

			// Подсветка целевой ячейки при перетаскивании
			if g.dragState == Dragging && g.isValidSwap(g.dragGem, image.Point{r, c}) {
				pulse := math.Sin(float64(time.Now().UnixMilli())*0.01)*0.3 + 0.5
				cellColor = color.RGBA{
					uint8(100 + 100*pulse),
					uint8(50 + 100*pulse),
					uint8(150 + 100*pulse),
					255,
				}
			}
			
			// Подсветка при свайпе (одинаковые гемы включая диагонали)
			if g.dragState == Dragging && len(g.swipeLine.Points) > 0 {
				gemAtCell := g.board.Get(r, c)
				if gemAtCell == g.swipeLine.GemType && gemAtCell >= 0 {
					lastPt := g.swipeLine.Points[len(g.swipeLine.Points)-1]
					dr := abs(lastPt.X - r)
					dc := abs(lastPt.Y - c)
					if dr <= 1 && dc <= 1 && (dr+dc) > 0 {
						// Можно свапнуть по диагонали!
						pulse := math.Sin(float64(time.Now().UnixMilli())*0.015)*0.4 + 0.6
						cellColor = color.RGBA{
							uint8(50 + 200*pulse),
							uint8(200 + 55*pulse),
							uint8(50),
							255,
						}
					}
				}
			}

			drawRoundedRect(screen, float64(x+2), float64(y+2), tileSize-4, tileSize-4, 8, cellColor)

			// Гем (не рисуем перетаскиваемый)
			gem := g.board.Get(r, c)
			if gem >= 0 && !(g.dragGem.X == r && g.dragGem.Y == c &&
				(g.dragState == DragPicking || g.dragState == Dragging ||
					g.dragState == DragSnapping || g.dragState == DragShaking)) {
				g.drawGem(screen, x, y, gem, 1.0, 0, 1.0)
			}

			// Подсказка
			if g.showHint {
				if (r == g.hintPair[0].X && c == g.hintPair[0].Y) ||
					(r == g.hintPair[1].X && c == g.hintPair[1].Y) {
					pulse := math.Sin(float64(time.Now().UnixMilli())*0.008)*0.3 + 0.7
					alpha := uint8(180 * pulse)
					op := &ebiten.DrawImageOptions{}
					op.ColorScale.ScaleAlpha(float32(alpha) / 255.0)
					op.GeoM.Translate(float64(x), float64(y))
					if g.selectorImg != nil {
						screen.DrawImage(g.selectorImg, op)
					}
				}
			}
		}
	}
}

func (g *Game) drawGem(screen *ebiten.Image, x, y int, gem int, scale, rotation, alpha float64) {
	if img, ok := g.gemImages[gem]; ok {
		op := &ebiten.DrawImageOptions{}
		baseScale := float64(tileSize-10) / float64(img.Bounds().Dx()) * scale
		op.GeoM.Scale(baseScale, baseScale)

		// Вращение вокруг центра
		if rotation != 0 {
			cx := float64(img.Bounds().Dx()) / 2
			cy := float64(img.Bounds().Dy()) / 2
			op.GeoM.Translate(-cx, -cy)
			op.GeoM.Rotate(rotation)
			op.GeoM.Translate(cx, cy)
		}

		op.ColorScale.ScaleAlpha(float32(alpha))
		op.GeoM.Translate(float64(x+5), float64(y+5))
		screen.DrawImage(img, op)
	} else {
		colors := []color.RGBA{
			{255, 80, 80, uint8(255 * alpha)},
			{80, 80, 255, uint8(255 * alpha)},
			{80, 255, 80, uint8(255 * alpha)},
			{255, 255, 80, uint8(255 * alpha)},
			{180, 80, 255, uint8(255 * alpha)},
		}
		c := colors[0]
		if gem < len(colors) {
			c = colors[gem]
		}
		drawRoundedRect(screen, float64(x+5), float64(y+5), tileSize-10, tileSize-10, 10, c)
	}
}

func (g *Game) drawDraggedGem(screen *ebiten.Image) {
	gemType := g.board.Get(g.dragGem.X, g.dragGem.Y)
	if gemType < 0 {
		return
	}

	// Позиция: прилипание или мышь
	var drawX, drawY float64
	if g.dragState == DragSnapping || g.dragState == DragReturning {
		drawX = float64(gridOffset + g.dragTargetPos.Y*tileSize + tileSize/2)
		drawY = float64(gridOffset + g.dragTargetPos.X*tileSize + tileSize/2)
	} else if g.dragState == DragShaking {
		shake := math.Sin(g.dragAnimTime*math.Pi*8) * 8 * (1 - g.dragAnimTime/0.4)
		drawX = g.dragCurrentX + shake
		drawY = g.dragCurrentY
	} else {
		drawX = g.dragCurrentX
		drawY = g.dragCurrentY
	}

	// ОГНЕННОЕ КОЛЬЦО вместо белого круга!
	if g.dragState == Dragging {
		g.drawFireRing(screen, drawX, drawY, gemType)
	}

	// Сам гем
	g.drawGem(screen, int(drawX-tileSize/2+5), int(drawY-tileSize/2+5), gemType,
		g.dragScale, g.dragRotation, 0.9)
}

// drawFireRing рисует огненное кольцо вокруг перетаскиваемого гема
func (g *Game) drawFireRing(screen *ebiten.Image, x, y float64, gemType int) {
	time := float64(time.Now().UnixMilli()) * 0.003

	// Цвета огня в зависимости от типа гема
	colors := []color.RGBA{
		{255, 80, 50, 200},   // красный - огненное кольцо
		{50, 100, 255, 200},  // синий - ледяное кольцо
		{80, 255, 100, 200},  // зелёный - токсичное кольцо
		{255, 255, 50, 200},  // жёлтый - электрическое кольцо
		{200, 80, 255, 200},  // фиолетовый - магическое кольцо
	}

	baseColor := colors[0]
	if gemType < len(colors) {
		baseColor = colors[gemType]
	}

	// Вращающиеся сегменты кольца
	for i := 0; i < 12; i++ {
		angle := float64(i)*math.Pi*2/12 + time
		nextAngle := float64(i+1)*math.Pi*2/12 + time

		radius := 45.0 + math.Sin(time*2+float64(i))*5

		// Внешнее кольцо
		x1 := x + math.Cos(angle)*radius
		y1 := y + math.Sin(angle)*radius
		x2 := x + math.Cos(nextAngle)*radius
		y2 := y + math.Sin(nextAngle)*radius

		alpha := uint8(150 + math.Sin(time+float64(i)*0.5)*100)
		vector.StrokeLine(screen,
			float32(x1), float32(y1),
			float32(x2), float32(y2),
			4, color.RGBA{baseColor.R, baseColor.G, baseColor.B, alpha}, false)

		// Огненные частицы
		if i%3 == 0 {
			particleRadius := radius + 10 + math.Sin(time*3+float64(i))*8
			px := x + math.Cos(angle+time)*particleRadius
			py := y + math.Sin(angle+time)*particleRadius

			size := 3 + math.Sin(time*4+float64(i))*2
			vector.DrawFilledCircle(screen, float32(px), float32(py), float32(size),
				color.RGBA{baseColor.R, baseColor.G, baseColor.B, uint8(180 + math.Sin(time*2)*75)}, false)
		}
	}

	// Внутреннее вращающееся кольцо
	for i := 0; i < 8; i++ {
		angle := float64(i)*math.Pi*2/8 - time*1.5
		radius := 35.0 + math.Cos(time*1.5+float64(i))*3

		x1 := x + math.Cos(angle)*radius
		y1 := y + math.Sin(angle)*radius

		size := 2 + math.Sin(time*3+float64(i))*1.5
		vector.DrawFilledCircle(screen, float32(x1), float32(y1), float32(size),
			color.RGBA{255, 255, 255, uint8(100 + math.Sin(time+float64(i))*50)}, false)
	}

	// Пульсирующее свечение
	glowRadius := 55.0 + math.Sin(time*2)*10
	alpha := uint8(40 + math.Sin(time*3)*20)
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(glowRadius),
		color.RGBA{baseColor.R, baseColor.G, baseColor.B, alpha}, false)
}

func (g *Game) drawTrails(screen *ebiten.Image) {
	for _, t := range g.trails {
		alpha := uint8(255 * (t.Life / 0.5))
		if c, ok := t.Color.(color.RGBA); ok {
			vector.DrawFilledCircle(screen, float32(t.X), float32(t.Y), float32(t.Size*t.Life/0.5),
				color.RGBA{c.R, c.G, c.B, alpha}, false)
		}
	}
}

func (g *Game) drawParticles(screen *ebiten.Image) {
	for _, p := range g.particles {
		alpha := uint8(255 * (p.Life / p.MaxLife))
		if c, ok := p.Color.(color.RGBA); ok {
			vector.DrawFilledCircle(screen, float32(p.X), float32(p.Y), float32(p.Size),
				color.RGBA{c.R, c.G, c.B, alpha}, false)
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 180})
	drawRoundedRect(screen, 150, 250, 500, 300, 20, color.RGBA{40, 30, 70, 255})

	g.drawCenteredText(screen, "TIME'S UP!", screenWidth/2, 300, 48, color.RGBA{255, 100, 100, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Final Score: %d", g.score), screenWidth/2, 380, 32, color.RGBA{255, 215, 0, 255})
	g.drawCenteredText(screen, fmt.Sprintf("Level Reached: %d", g.level), screenWidth/2, 430, 24, color.RGBA{100, 200, 255, 255})
	g.drawCenteredText(screen, "Press R or ENTER to restart", screenWidth/2, 500, 20, color.RGBA{200, 200, 255, 255})
}

func (g *Game) drawCenteredText(screen *ebiten.Image, str string, x, y int, size int, c color.Color) {
	ebitenutil.DebugPrintAt(screen, str, x-len(str)*size/4, y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func drawRoundedRect(screen *ebiten.Image, x, y, w, h float64, cr float64, c color.Color) {
	vector.DrawFilledRect(screen, float32(x+cr), float32(y), float32(w-cr*2), float32(h), c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+cr), float32(w), float32(h-cr*2), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+cr), float32(y+h-cr), float32(cr), c, false)
	vector.DrawFilledCircle(screen, float32(x+w-cr), float32(y+h-cr), float32(cr), c, false)
}

func (g *Game) drawSwipeLine(screen *ebiten.Image) {
	if len(g.swipeLine.LinePoints) < 2 {
		return
	}

	alpha := uint8(255)
	if g.dragState == DragSwiping {
		alpha = uint8(255 * (1 - g.swipeTimer/0.5))
	}

	// Свечение линии
	for i := 1; i < len(g.swipeLine.LinePoints); i++ {
		p1 := g.swipeLine.LinePoints[i-1]
		p2 := g.swipeLine.LinePoints[i]

		// Толстая светящаяся линия
		vector.StrokeLine(screen,
			float32(p1.X), float32(p1.Y),
			float32(p2.X), float32(p2.Y),
			12, color.RGBA{255, 255, 255, alpha/3}, false)

		// Средняя линия
		vector.StrokeLine(screen,
			float32(p1.X), float32(p1.Y),
			float32(p2.X), float32(p2.Y),
			6, color.RGBA{100, 200, 255, alpha/2}, false)

		// Тонкая яркая линия
		vector.StrokeLine(screen,
			float32(p1.X), float32(p1.Y),
			float32(p2.X), float32(p2.Y),
			2, color.RGBA{255, 255, 255, alpha}, false)
	}

	// Электрические разряды (молнии)
	if g.dragState == Dragging && len(g.swipeLine.Points) >= 3 {
		for _, p := range g.swipeLine.LinePoints {
			for i := 0; i < 3; i++ {
				angle := float64(i) * math.Pi * 2 / 3 + float64(time.Now().UnixMilli())*0.01
				ex := p.X + math.Cos(angle)*15
				ey := p.Y + math.Sin(angle)*15

				vector.StrokeLine(screen,
					float32(p.X), float32(p.Y),
					float32(ex), float32(ey),
					1, color.RGBA{200, 230, 255, alpha}, false)
			}
		}
	}
}
