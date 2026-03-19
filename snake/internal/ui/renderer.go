// Package ui содержит компоненты пользовательского интерфейса
package ui

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"playgo/snake/internal/game"
	"playgo/snake/internal/effects"
)

// Renderer отвечает за отрисовку игры
type Renderer struct {
	tileSize     int
	screenWidth  int
	screenHeight int
}

// NewRenderer создаёт новый рендерер
func NewRenderer(cfg *game.Config) *Renderer {
	return &Renderer{
		tileSize:     cfg.TileSize,
		screenWidth:  cfg.ScreenWidth,
		screenHeight: cfg.ScreenHeight,
	}
}

// DrawMenu отрисовывает главное меню
func (r *Renderer) DrawMenu(screen *ebiten.Image) {
	title := "SNAKE GAME"
	titleX := float32(r.screenWidth/2 - len(title)*10)
	titleY := float32(r.screenHeight/2 - 100)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	subtitle := "Go365 Go79 - Ebitengine"
	subX := float32(r.screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	startText := "Press ENTER or SPACE to Start"
	startX := float32(r.screenWidth/2 - len(startText)*6)
	ebitenutil.DebugPrintAt(screen, startText, int(startX), int(titleY+100))

	controls := []string{
		"Controls:",
		"Arrow Keys - Move",
		"SPACE - Shoot Arrow",
		"P - Pause",
		"",
		"Find the golden key and open the treasure chest!",
		"Collect arrows and shoot the bugs!",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, r.screenWidth/2-150, int(titleY)+160+i*20)
	}
}

// DrawDifficultySelection отрисовывает выбор сложности
func (r *Renderer) DrawDifficultySelection(screen *ebiten.Image, difficulty game.Difficulty) {
	title := "SNAKE GAME"
	titleX := float32(r.screenWidth/2 - len(title)*10)
	titleY := float32(r.screenHeight/2 - 150)
	ebitenutil.DebugPrintAt(screen, title, int(titleX), int(titleY))

	subtitle := "Go365 Go79 - Ebitengine"
	subX := float32(r.screenWidth/2 - len(subtitle)*5)
	ebitenutil.DebugPrintAt(screen, subtitle, int(subX), int(titleY+40))

	selectText := "Select Difficulty"
	selectX := float32(r.screenWidth/2 - len(selectText)*8)
	ebitenutil.DebugPrintAt(screen, selectText, int(selectX), int(titleY+100))

	difficulties := []struct {
		name       string
		enemyCount int
	}{
		{"Easy", 2},
		{"Medium", 3},
		{"Hard", 5},
	}

	for i, diff := range difficulties {
		y := int(titleY) + 160 + i*40
		marker := "  "
		prefix := "  "

		if game.Difficulty(i) == difficulty {
			marker = ">> "
			prefix = "<<"
			highlight := fmt.Sprintf("%s%s - %d bugs %s", marker, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, highlight, r.screenWidth/2-100, y)
		} else {
			text := fmt.Sprintf("%s%s - %d bugs %s", prefix, diff.name, diff.enemyCount, prefix)
			ebitenutil.DebugPrintAt(screen, text, r.screenWidth/2-100, y)
		}
	}

	controls := []string{
		"",
		"UP/DOWN - Change difficulty",
		"ENTER/SPACE - Start game",
	}
	for i, line := range controls {
		ebitenutil.DebugPrintAt(screen, line, r.screenWidth/2-120, int(titleY)+300+i*20)
	}
}

// DrawGame отрисовывает игровое поле
func (r *Renderer) DrawGame(screen *ebiten.Image, g *game.Game, es *effects.EffectSystem) {
	// Граница игрового поля
	vector.StrokeRect(screen, 0, 0, float32(r.screenWidth), float32(r.screenHeight), 2, color.RGBA{100, 100, 100, 255}, false)

	// Отрисовка змейки
	for i, segment := range g.Snake {
		green := color.RGBA{0, 255, 0, 255}
		if i == 0 {
			green = color.RGBA{100, 255, 100, 255}
		}
		vector.DrawFilledRect(
			screen,
			float32(segment.X*r.tileSize),
			float32(segment.Y*r.tileSize),
			float32(r.tileSize),
			float32(r.tileSize),
			green,
			false,
		)
		if i == 0 {
			r.drawSnakeEyes(screen, segment, g.Direction)
			r.drawSnakeTongue(screen, segment, g.Direction)
		}
	}

	// Отрисовка еды
	r.drawFood(screen, g.Food, g.FoodTimer)

	// Отрисовка врагов
	for _, enemy := range g.Enemies {
		r.drawEnemy(screen, enemy)
	}

	// Отрисовка бомб
	for _, bomb := range g.Bombs {
		r.drawBomb(screen, bomb)
	}

	// Отрисовка сундука
	if g.Chest != nil {
		r.drawChest(screen, *g.Chest)
	}

	// Отрисовка ключа
	if g.Key != nil {
		r.drawKey(screen, *g.Key)
	}

	// Отрисовка монет
	for _, coin := range g.Coins {
		r.drawCoin(screen, coin)
	}

	// Отрисовка стрел
	for _, arrow := range g.Arrows {
		r.drawArrow(screen, arrow)
	}

	// Отрисовка бонусов
	for _, powerUp := range g.PowerUps {
		r.drawPowerUp(screen, powerUp)
	}

	// Отрисовка счёта
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Score: %d", g.Score), 10, 10)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Arrows: %d", g.ArrowCount), 10, 25)
	if g.HasKey {
		ebitenutil.DebugPrintAt(screen, "KEY!", 10, 40)
	}
	if len(g.Coins) > 0 {
		ebitenutil.DebugPrintAt(screen, "x2 XP COINS!", 10, 55)
	}
	// Отображение жизней
	if g.Lives > 1 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Lives: %d", g.Lives), 10, 70)
	}
	// Отображение активных эффектов
	effectY := 10
	if g.Lives > 1 {
		effectY = 85
	}
	for effectType, duration := range g.ActiveEffects {
		effectName := effectType.String()[:3] // Первые 3 буквы
		seconds := duration / 60
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%s: %ds", effectName, seconds), r.screenWidth-100, effectY)
		effectY += 15
	}
}

// DrawPauseOverlay отрисовывает оверлей паузы
func (r *Renderer) DrawPauseOverlay(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0, 0, 0, 128})

	pausedText := "PAUSED"
	pausedX := r.screenWidth/2 - len(pausedText)*8
	ebitenutil.DebugPrintAt(screen, pausedText, pausedX, r.screenHeight/2-50)

	continueText := "Press P to Continue"
	contX := r.screenWidth/2 - len(continueText)*6
	ebitenutil.DebugPrintAt(screen, continueText, contX, r.screenHeight/2)
}

// DrawGameOverOverlay отрисовывает оверлей конца игры
func (r *Renderer) DrawGameOverOverlay(screen *ebiten.Image, score int, enemies int) {
	screen.Fill(color.RGBA{50, 0, 0, 180})

	gameOverText := "GAME OVER"
	gameOverX := r.screenWidth/2 - len(gameOverText)*8
	ebitenutil.DebugPrintAt(screen, gameOverText, gameOverX, r.screenHeight/2-80)

	scoreText := fmt.Sprintf("Final Score: %d", score)
	scoreX := r.screenWidth/2 - len(scoreText)*6
	ebitenutil.DebugPrintAt(screen, scoreText, scoreX, r.screenHeight/2-20)

	enemiesText := fmt.Sprintf("Enemies: %d", enemies)
	enemiesX := r.screenWidth/2 - len(enemiesText)*6
	ebitenutil.DebugPrintAt(screen, enemiesText, enemiesX, r.screenHeight/2+10)

	restartText := "Press ENTER to Restart"
	restartX := r.screenWidth/2 - len(restartText)*7
	ebitenutil.DebugPrintAt(screen, restartText, restartX, r.screenHeight/2+60)
}

// drawSnakeEyes отрисовывает глаза змейки
func (r *Renderer) drawSnakeEyes(screen *ebiten.Image, head game.Point, direction game.Direction) {
	x := float32(head.X * r.tileSize)
	y := float32(head.Y * r.tileSize)
	size := float32(r.tileSize)
	eyeSize := size / 6
	pupilSize := eyeSize / 2

	var leftEyeX, leftEyeY, rightEyeX, rightEyeY float32

	switch direction {
	case game.Up:
		leftEyeX = x + size/3
		leftEyeY = y + size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + size/3
	case game.Down:
		leftEyeX = x + size/3
		leftEyeY = y + 2*size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + 2*size/3
	case game.Left:
		leftEyeX = x + size/3
		leftEyeY = y + size/3
		rightEyeX = x + size/3
		rightEyeY = y + 2*size/3
	case game.Right:
		leftEyeX = x + 2*size/3
		leftEyeY = y + size/3
		rightEyeX = x + 2*size/3
		rightEyeY = y + 2*size/3
	}

	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, pupilSize, color.RGBA{0, 0, 0, 255}, false)
}

// drawSnakeTongue отрисовывает язык змейки
func (r *Renderer) drawSnakeTongue(screen *ebiten.Image, head game.Point, direction game.Direction) {
	x := float32(head.X * r.tileSize)
	y := float32(head.Y * r.tileSize)
	size := float32(r.tileSize)

	tongueColor := color.RGBA{255, 50, 50, 255}
	tongueLength := size / 2
	tongueWidth := size / 12

	var startX, startY, endX, endY float32

	switch direction {
	case game.Up:
		startX = x + size/2
		startY = y + size/4
		endX = x + size/2
		endY = y - tongueLength
	case game.Down:
		startX = x + size/2
		startY = y + 3*size/4
		endX = x + size/2
		endY = y + size + tongueLength
	case game.Left:
		startX = x + size/4
		startY = y + size/2
		endX = x - tongueLength
		endY = y + size/2
	case game.Right:
		startX = x + 3*size/4
		startY = y + size/2
		endX = x + size + tongueLength
		endY = y + size/2
	}

	vector.StrokeLine(screen, startX, startY, endX, endY, tongueWidth, tongueColor, false)

	forkLength := size / 6
	var leftForkX, leftForkY, rightForkX, rightForkY float32

	switch direction {
	case game.Up:
		leftForkX = endX - forkLength/2
		leftForkY = endY + forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY + forkLength/2
	case game.Down:
		leftForkX = endX - forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY - forkLength/2
	case game.Left:
		leftForkX = endX + forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX + forkLength/2
		rightForkY = endY + forkLength/2
	case game.Right:
		leftForkX = endX - forkLength/2
		leftForkY = endY - forkLength/2
		rightForkX = endX - forkLength/2
		rightForkY = endY + forkLength/2
	}

	vector.StrokeLine(screen, endX, endY, leftForkX, leftForkY, tongueWidth, tongueColor, false)
	vector.StrokeLine(screen, endX, endY, rightForkX, rightForkY, tongueWidth, tongueColor, false)
}

// drawFood отрисовывает еду
func (r *Renderer) drawFood(screen *ebiten.Image, food game.Point, foodTimer int) {
	x := float32(food.X * r.tileSize)
	y := float32(food.Y * r.tileSize)
	size := float32(r.tileSize)

	centerX := x + size/2
	centerY := y + size/2 + 2
	radius := size/2 - 3

	// Анимация появления (пульсация)
	pulseScale := effects.FoodPulseScale(foodTimer)

	// Основное красное тело
	vector.DrawFilledCircle(screen, centerX, centerY, radius*pulseScale, color.RGBA{255, 0, 0, 255}, false)

	// Блик на яблоке
	highlightX := centerX - radius/3
	highlightY := centerY - radius/3
	vector.DrawFilledCircle(screen, highlightX, highlightY, radius/3*pulseScale, color.RGBA{255, 100, 100, 255}, false)

	// Маленькое углубление сверху
	vector.DrawFilledCircle(screen, centerX, centerY-radius+2, 2, color.RGBA{200, 0, 0, 255}, false)

	// Коричневый черенок
	stemX := centerX
	stemY := centerY - radius
	vector.StrokeLine(screen, stemX, stemY, stemX, stemY-4, 2, color.RGBA{139, 69, 19, 255}, false)

	// Зелёный лист
	leafColor := color.RGBA{34, 139, 34, 255}
	leafBaseX := centerX + 1
	leafBaseY := stemY - 2

	// Позиции листа
	tipX := centerX + 5
	tipY := leafBaseY - 3
	botX := centerX + 2
	botY := leafBaseY + 2

	vector.DrawFilledRect(screen, leafBaseX, leafBaseY-1, 5, 2, leafColor, false)
	vector.DrawFilledCircle(screen, tipX-1, tipY, 2, leafColor, false)
	vector.DrawFilledCircle(screen, botX, botY, 2, leafColor, false)

	vector.StrokeLine(screen, leafBaseX+1, leafBaseY, tipX-1, tipY, 1, color.RGBA{100, 200, 100, 255}, false)
}

// drawEnemy отрисовывает врага (жука)
func (r *Renderer) drawEnemy(screen *ebiten.Image, enemy game.Enemy) {
	x := float32(enemy.Pos.X * r.tileSize)
	y := float32(enemy.Pos.Y * r.tileSize)
	size := float32(r.tileSize) * 1.5

	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{128, 0, 128, 255}, false)

	headX := x + size/2
	headY := y + size/2
	vector.DrawFilledCircle(screen, headX, headY, size/3, color.RGBA{100, 0, 100, 255}, false)

	// Анимированные ноги
	legOffset := float32((enemy.AnimFrame % 20) / 10.0 * 3)
	if enemy.AnimFrame%40 < 20 {
		legOffset = -legOffset
	}

	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+size/3, legY, x-size/4, legY+legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	for i := 0; i < 3; i++ {
		legY := y + size/4 + float32(i)*size/4
		vector.StrokeLine(screen, x+2*size/3, legY, x+5*size/4, legY-legOffset+float32(i)*2, 2, color.RGBA{100, 0, 100, 255}, false)
	}

	// Анимированные усики
	antennaAngle := float32((enemy.AnimFrame % 30) / 30.0 * 1.0)
	if enemy.AnimFrame%60 < 30 {
		antennaAngle = -antennaAngle
	}

	vector.StrokeLine(screen, headX-size/6, headY-size/3, headX-size/2-antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)
	vector.StrokeLine(screen, headX+size/6, headY-size/3, headX+size/2+antennaAngle*size, headY-size/2-antennaAngle*size, 1, color.RGBA{150, 50, 50, 255}, false)

	// Рот
	mouthX := headX
	mouthY := headY + size/8
	vector.DrawFilledCircle(screen, mouthX, mouthY, size/8, color.RGBA{50, 0, 0, 255}, false)

	// Красные светящиеся глаза
	eyeSize := size / 5
	leftEyeX := headX - size/6
	leftEyeY := headY - size/8
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX, leftEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, leftEyeX-eyeSize/4, leftEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)

	rightEyeX := headX + size/6
	rightEyeY := headY - size/8
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize+2, color.RGBA{255, 50, 0, 100}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize, color.RGBA{255, 100, 0, 180}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize-2, color.RGBA{255, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX, rightEyeY, eyeSize/3, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, rightEyeX+eyeSize/4, rightEyeY-eyeSize/4, eyeSize/5, color.RGBA{255, 255, 255, 255}, false)
}

// drawBomb отрисовывает бомбу
func (r *Renderer) drawBomb(screen *ebiten.Image, bomb game.Bomb) {
	x := float32(bomb.Pos.X * r.tileSize)
	y := float32(bomb.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	vector.DrawFilledCircle(screen, x+size/2, y+size/2, size/2-2, color.RGBA{0, 0, 0, 255}, false)
	vector.DrawFilledCircle(screen, x+size/3, y+size/3, size/6, color.RGBA{50, 50, 50, 255}, false)

	fuseX := x + size/2
	fuseY := y + size/4
	vector.StrokeLine(screen, fuseX, fuseY, fuseX, fuseY-size/3, 2, color.RGBA{139, 69, 19, 255}, false)

	// Мигание искры
	blinkPhase := float32(bomb.Timer%5) * 0.4
	sparkSize := size/6 + blinkPhase*size/8

	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize, color.RGBA{255, 200, 0, 200}, false)
	vector.DrawFilledCircle(screen, fuseX, fuseY-size/3, sparkSize/2, color.RGBA{255, 255, 255, 255}, false)

	// Частицы искр
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 3; i++ {
		particleX := fuseX + float32(rand.Intn(8)-4)
		particleY := fuseY - size/3 + float32(rand.Intn(8)-4)
		vector.DrawFilledCircle(screen, particleX, particleY, 1, color.RGBA{255, 100, 0, 255}, false)
	}
}

// drawChest отрисовывает сундук
func (r *Renderer) drawChest(screen *ebiten.Image, chest game.TreasureChest) {
	x := float32(chest.Pos.X * r.tileSize)
	y := float32(chest.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	chestColor := color.RGBA{139, 69, 19, 255}
	if chest.Open {
		chestColor = color.RGBA{100, 50, 10, 255}
	}
	vector.DrawFilledRect(screen, x+2, y+4, size-4, size-6, chestColor, false)

	lidColor := color.RGBA{255, 215, 0, 255}
	if chest.Open {
		vector.StrokeLine(screen, x+2, y+4, x+size-2, y+4, 2, lidColor, false)
	} else {
		vector.DrawFilledRect(screen, x+2, y+2, size-4, size/3, lidColor, false)
	}

	if !chest.Open {
		vector.DrawFilledCircle(screen, x+size/2, y+size/2, 3, color.RGBA{255, 215, 0, 255}, false)
	}
}

// drawKey отрисовывает ключ
func (r *Renderer) drawKey(screen *ebiten.Image, key game.Key) {
	x := float32(key.Pos.X * r.tileSize)
	y := float32(key.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	keyColor := color.RGBA{255, 215, 0, 255}

	headSize := size / 3
	vector.DrawFilledCircle(screen, x+size/2, y+size/3, headSize, keyColor, false)

	shaftWidth := size / 8
	shaftHeight := size / 2
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2, shaftWidth, shaftHeight, keyColor, false)

	toothSize := size / 6
	vector.DrawFilledRect(screen, x+size/2-shaftWidth/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize, keyColor, false)
	vector.DrawFilledRect(screen, x+size/2, y+size/2+shaftHeight-toothSize, shaftWidth, toothSize/2, keyColor, false)
}

// drawCoin отрисовывает монету с пульсацией
func (r *Renderer) drawCoin(screen *ebiten.Image, coin game.Coin) {
	x := float32(coin.Pos.X * r.tileSize)
	y := float32(coin.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	coinRadius := size/2 - 4
	centerX := x + size/2
	centerY := y + size/2

	// Пульсация монеты
	pulseScale := effects.PulseScale(coin.PulsePhase, 0.1)

	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius*pulseScale, color.RGBA{255, 215, 0, 255}, false)

	innerRadius := coinRadius - 3
	vector.DrawFilledCircle(screen, centerX, centerY, innerRadius*pulseScale, color.RGBA{255, 235, 100, 255}, false)

	dotRadius := innerRadius / 2
	vector.DrawFilledCircle(screen, centerX, centerY, dotRadius*pulseScale, color.RGBA{255, 200, 0, 255}, false)

	sparkleOffset := coinRadius - 1
	vector.DrawFilledCircle(screen, centerX, centerY-sparkleOffset*pulseScale, 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX, centerY+sparkleOffset*pulseScale, 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX-sparkleOffset*pulseScale, centerY, 1, color.RGBA{255, 255, 255, 255}, false)
	vector.DrawFilledCircle(screen, centerX+sparkleOffset*pulseScale, centerY, 1, color.RGBA{255, 255, 255, 255}, false)

	// Анимированное свечение
	glowPhase := (time.Now().UnixMilli() / 200) % 2
	glowIntensity := uint8(100 + glowPhase*50)
	vector.DrawFilledCircle(screen, centerX, centerY, coinRadius*pulseScale+2, color.RGBA{255, 215, 0, glowIntensity}, false)
}

// drawArrow отрисовывает стрелу
func (r *Renderer) drawArrow(screen *ebiten.Image, arrow game.Arrow) {
	x := float32(arrow.Pos.X * r.tileSize)
	y := float32(arrow.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	arrowColor := color.RGBA{192, 192, 192, 255}

	shaftLength := size / 2
	shaftWidth := float32(2)

	var startX, startY, endX, endY float32

	switch arrow.Direction {
	case game.Up:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2
		endY = y + size/2 - shaftLength
	case game.Down:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2
		endY = y + size/2 + shaftLength
	case game.Left:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2 - shaftLength
		endY = y + size/2
	case game.Right:
		startX = x + size/2
		startY = y + size/2
		endX = x + size/2 + shaftLength
		endY = y + size/2
	}

	vector.StrokeLine(screen, startX, startY, endX, endY, shaftWidth, arrowColor, false)

	headSize := size / 6
	var headX1, headY1, headX2, headY2, headX3, headY3 float32

	switch arrow.Direction {
	case game.Up:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize/2
		headY2 = endY + headSize
		headX3 = endX + headSize/2
		headY3 = endY + headSize
	case game.Down:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize/2
		headY2 = endY - headSize
		headX3 = endX + headSize/2
		headY3 = endY - headSize
	case game.Left:
		headX1 = endX
		headY1 = endY
		headX2 = endX + headSize
		headY2 = endY - headSize/2
		headX3 = endX + headSize
		headY3 = endY + headSize/2
	case game.Right:
		headX1 = endX
		headY1 = endY
		headX2 = endX - headSize
		headY2 = endY - headSize/2
		headX3 = endX - headSize
		headY3 = endY + headSize/2
	}

	vector.StrokeLine(screen, headX1, headY1, headX2, headY2, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX2, headY2, headX3, headY3, shaftWidth, arrowColor, false)
	vector.StrokeLine(screen, headX3, headY3, headX1, headY1, shaftWidth, arrowColor, false)
}

// drawPowerUp отрисовывает бонус
func (r *Renderer) drawPowerUp(screen *ebiten.Image, powerUp game.PowerUp) {
	x := float32(powerUp.Pos.X * r.tileSize)
	y := float32(powerUp.Pos.Y * r.tileSize)
	size := float32(r.tileSize)

	// Пульсация бонуса
	pulseScale := effects.PulseScale(powerUp.PulsePhase, 0.15)

	centerX := x + size/2
	centerY := y + size/2
	radius := size/2 - 3

	// Цвет в зависимости от типа бонуса
	var powerUpColor color.RGBA
	switch powerUp.Type {
	case game.PowerUpSlowMotion:
		powerUpColor = color.RGBA{0, 191, 255, 255} // Голубой
	case game.PowerUpShield:
		powerUpColor = color.RGBA{65, 105, 225, 255} // Синий
	case game.PowerUpShrink:
		powerUpColor = color.RGBA{34, 139, 34, 255} // Зелёный
	case game.PowerUpExtraLife:
		powerUpColor = color.RGBA{255, 0, 0, 255} // Красный
	case game.PowerUpLightning:
		powerUpColor = color.RGBA{255, 255, 0, 255} // Жёлтый
	case game.PowerUpMultiplier:
		powerUpColor = color.RGBA{50, 205, 50, 255} // Светло-зелёный
	}

	// Основная форма бонуса (ромб)
	vector.DrawFilledCircle(screen, centerX, centerY, radius*pulseScale, powerUpColor, false)

	// Блик
	highlightX := centerX - radius/3
	highlightY := centerY - radius/3
	vector.DrawFilledCircle(screen, highlightX, highlightY, radius/4*pulseScale, color.RGBA{255, 255, 255, 200}, false)

	// Символ бонуса (точка в центре)
	vector.DrawFilledCircle(screen, centerX, centerY, 2, color.RGBA{255, 255, 255, 255}, false)

	// Свечение
	glowPhase := (time.Now().UnixMilli() / 200) % 2
	glowIntensity := uint8(100 + glowPhase*80)
	vector.DrawFilledCircle(screen, centerX, centerY, radius*pulseScale+3, color.RGBA{powerUpColor.R, powerUpColor.G, powerUpColor.B, glowIntensity}, false)
}

