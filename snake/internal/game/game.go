// Package game содержит основную игровую логику для Simple Snake
package game

import (
	"math/rand"
	"time"
)

// Direction представляет направление движения змейки
type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

// GameState представляет текущее состояние игры
type GameState int

const (
	Menu GameState = iota
	SelectDifficulty
	Playing
	Paused
	GameOver
)

// Difficulty представляет уровень сложности
type Difficulty int

const (
	Easy Difficulty = iota
	Medium
	Hard
)

// String возвращает строковое представление сложности
func (d Difficulty) String() string {
	switch d {
	case Easy:
		return "Easy"
	case Medium:
		return "Medium"
	case Hard:
		return "Hard"
	default:
		return "Unknown"
	}
}

// EnemyCount возвращает количество врагов для уровня сложности
func (d Difficulty) EnemyCount() int {
	switch d {
	case Easy:
		return 2
	case Medium:
		return 3
	case Hard:
		return 5
	default:
		return 3
	}
}

// Point представляет позицию на сетке
type Point struct {
	X int
	Y int
}

// Enemy представляет врага (жука)
type Enemy struct {
	Pos       Point
	Direction Direction
	AnimFrame int
}

// Bomb представляет бомбу с таймером
type Bomb struct {
	Pos     Point
	Timer   int
	MaxTime int
}

// TreasureChest представляет сундук с сокровищами
type TreasureChest struct {
	Pos    Point
	Open   bool
	Arrows int
}

// Key представляет ключ
type Key struct {
	Pos Point
}

// Coin представляет монету
type Coin struct {
	Pos        Point
	Value      int
	Collected  bool
	PulsePhase float64
}

// PowerUpType представляет тип бонуса
type PowerUpType int

const (
	PowerUpSlowMotion PowerUpType = iota // Замедление времени
	PowerUpShield                         // Неуязвимость
	PowerUpShrink                         // Уменьшение змейки
	PowerUpExtraLife                      // Дополнительная жизнь
	PowerUpLightning                      // Уничтожение врагов
	PowerUpMultiplier                     // Множитель очков
)

// String возвращает строковое представление бонуса
func (p PowerUpType) String() string {
	switch p {
	case PowerUpSlowMotion:
		return "Slow Motion"
	case PowerUpShield:
		return "Shield"
	case PowerUpShrink:
		return "Shrink"
	case PowerUpExtraLife:
		return "Extra Life"
	case PowerUpLightning:
		return "Lightning"
	case PowerUpMultiplier:
		return "Multiplier"
	default:
		return "Unknown"
	}
}

// PowerUp представляет бонус
type PowerUp struct {
	Pos        Point
	Type       PowerUpType
	Active     bool
	Duration   int  // длительность в тиках для временных эффектов
	PulsePhase float64
}

// Arrow представляет стрелу
type Arrow struct {
	Pos       Point
	Direction Direction
	Active    bool
	Speed     int
}

// Config содержит конфигурацию игры
type Config struct {
	TileSize    int
	ScreenWidth int
	ScreenHeight int
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		TileSize:     20,
		ScreenWidth:  800,
		ScreenHeight: 600,
	}
}

// GridSize возвращает размер сетки
func (c *Config) GridSize() (int, int) {
	return c.ScreenWidth / c.TileSize, c.ScreenHeight / c.TileSize
}

// Game представляет основную игровую структуру
type Game struct {
	Snake      []Point
	Direction  Direction
	Food       Point
	Score      int
	GameOver   bool
	MoveTimer  int
	MoveDelay  int
	Enemies    []Enemy
	EnemyTimer int
	EnemyDelay int
	Bombs      []Bomb
	BombTimer  int
	BombDelay  int
	Chest      *TreasureChest
	Key        *Key
	Coins      []Coin
	CoinTimer  int
	CoinDelay  int
	Arrows     []Arrow
	HasKey     bool
	ArrowCount int
	State      GameState
	Difficulty Difficulty
	FoodTimer  int
	
	// Power-ups
	PowerUps       []PowerUp
	PowerUpTimer   int
	PowerUpDelay   int
	ActiveEffects  map[PowerUpType]int // тип -> оставшаяся длительность
	Lives          int                 // количество жизней
	
	config *Config
}

// NewGame создаёт новую игру
func NewGame() *Game {
	cfg := DefaultConfig()
	gridX, gridY := cfg.GridSize()
	
	g := &Game{
		Snake:      []Point{{gridX / 2, gridY / 2}, {gridX/2 - 1, gridY / 2}, {gridX/2 - 2, gridY / 2}},
		Direction:  Right,
		Score:      0,
		GameOver:   false,
		MoveDelay:  8,
		EnemyDelay: 12,
		BombDelay:  180,
		CoinDelay:  300,
		PowerUpDelay: 600, // Спавн бонуса каждые 600 тиков (~10 сек)
		Enemies:    []Enemy{},
		Bombs:      []Bomb{},
		Coins:      []Coin{},
		PowerUps:   []PowerUp{},
		Arrows:     []Arrow{},
		HasKey:     false,
		ArrowCount: 0,
		State:      Menu,
		FoodTimer:  0,
		Lives:      1,
		ActiveEffects: make(map[PowerUpType]int),
		config:     cfg,
	}
	return g
}

// Config возвращает конфигурацию игры
func (g *Game) Config() *Config {
	return g.config
}

// StartGame начинает игру
func (g *Game) StartGame() {
	g.placeFood()
	g.spawnChest()
	g.spawnKey()
	enemyCount := g.Difficulty.EnemyCount()
	for i := 0; i < enemyCount; i++ {
		g.spawnEnemy()
	}
	g.State = Playing
	g.FoodTimer = 20
}

// placeFood размещает еду в случайном месте
func (g *Game) placeFood() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		g.Food = Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		onSnake := false
		for _, segment := range g.Snake {
			if segment.X == g.Food.X && segment.Y == g.Food.Y {
				onSnake = true
				break
			}
		}
		if !onSnake {
			break
		}
	}
	g.FoodTimer = 20
}

// spawnEnemy создаёт врага
func (g *Game) spawnEnemy() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			dir := Direction(rand.Intn(4))
			g.Enemies = append(g.Enemies, Enemy{Pos: pos, Direction: dir, AnimFrame: 0})
			break
		}
	}
}

// spawnBomb создаёт бомбу
func (g *Game) spawnBomb() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 15 && dy < 10 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.Bombs = append(g.Bombs, Bomb{Pos: pos, Timer: 0, MaxTime: 180})
			break
		}
	}
}

// spawnChest создаёт сундук
func (g *Game) spawnChest() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.Chest = &TreasureChest{Pos: pos, Open: false, Arrows: 5}
			break
		}
	}
}

// spawnKey создаёт ключ
func (g *Game) spawnKey() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 20 && dy < 15 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.Key = &Key{Pos: pos}
			break
		}
	}
}

// spawnCoin создаёт монету
func (g *Game) spawnCoin() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 15 && dy < 10 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.Coins = append(g.Coins, Coin{Pos: pos, Value: 2, Collected: false, PulsePhase: 0})
			break
		}
	}
}

// spawnPowerUp создаёт бонус
func (g *Game) spawnPowerUp() {
	rand.Seed(time.Now().UnixNano())
	gridX, gridY := g.config.GridSize()
	
	// Выбираем случайный тип бонуса
	powerUpTypes := []PowerUpType{
		PowerUpSlowMotion,
		PowerUpShield,
		PowerUpShrink,
		PowerUpExtraLife,
		PowerUpLightning,
		PowerUpMultiplier,
	}
	randomType := powerUpTypes[rand.Intn(len(powerUpTypes))]
	
	for {
		pos := Point{
			X: rand.Intn(gridX),
			Y: rand.Intn(gridY),
		}
		tooClose := false
		for _, segment := range g.Snake {
			dx := segment.X - pos.X
			if dx < 0 {
				dx = -dx
			}
			dy := segment.Y - pos.Y
			if dy < 0 {
				dy = -dy
			}
			if dx < 15 && dy < 10 {
				tooClose = true
				break
			}
		}
		if !tooClose {
			g.PowerUps = append(g.PowerUps, PowerUp{
				Pos: pos,
				Type: randomType,
				Active: true,
				Duration: 0,
				PulsePhase: 0,
			})
			break
		}
	}
}

// applyPowerUp применяет эффект бонуса
func (g *Game) applyPowerUp(powerUp PowerUp) (events []GameEvent) {
	switch powerUp.Type {
	case PowerUpSlowMotion:
		g.ActiveEffects[PowerUpSlowMotion] = 600 // 10 секунд при 60 FPS
		g.MoveDelay = g.MoveDelay * 2 // Замедление
		events = append(events, GameEvent{Type: EventPowerUpSlowMotion, Pos: powerUp.Pos})
		
	case PowerUpShield:
		g.ActiveEffects[PowerUpShield] = 480 // 8 секунд
		events = append(events, GameEvent{Type: EventPowerUpShield, Pos: powerUp.Pos})
		
	case PowerUpShrink:
		// Уменьшаем змейку на половину
		if len(g.Snake) > 3 {
			newLen := len(g.Snake) / 2
			g.Snake = g.Snake[:newLen]
		}
		events = append(events, GameEvent{Type: EventPowerUpShrink, Pos: powerUp.Pos})
		
	case PowerUpExtraLife:
		g.Lives++
		events = append(events, GameEvent{Type: EventPowerUpExtraLife, Pos: powerUp.Pos})
		
	case PowerUpLightning:
		// Уничтожаем всех врагов
		for _, enemy := range g.Enemies {
			events = append(events, GameEvent{Type: EventEnemyKill, Pos: enemy.Pos})
			g.Score++
		}
		g.Enemies = []Enemy{}
		events = append(events, GameEvent{Type: EventPowerUpLightning, Pos: powerUp.Pos})
		
	case PowerUpMultiplier:
		g.ActiveEffects[PowerUpMultiplier] = 900 // 15 секунд
		events = append(events, GameEvent{Type: EventPowerUpMultiplier, Pos: powerUp.Pos})
	}
	
	return events
}

// updatePowerUps обновляет активные эффекты
func (g *Game) updatePowerUps() {
	// Уменьшаем длительность эффектов
	for effectType, duration := range g.ActiveEffects {
		g.ActiveEffects[effectType] = duration - 1
		if g.ActiveEffects[effectType] <= 0 {
			// Эффект закончился
			delete(g.ActiveEffects, effectType)
			
			// Восстанавливаем нормальную скорость
			if effectType == PowerUpSlowMotion {
				g.MoveDelay = 8
			}
		}
	}
	
	// Обновляем пульсацию бонусов
	for i := range g.PowerUps {
		g.PowerUps[i].PulsePhase += 0.15
	}
}

// HasShield возвращает true, если активен щит
func (g *Game) HasShield() bool {
	_, ok := g.ActiveEffects[PowerUpShield]
	return ok
}

// IsSlowMotion возвращает true, если активно замедление
func (g *Game) IsSlowMotion() bool {
	_, ok := g.ActiveEffects[PowerUpSlowMotion]
	return ok
}

// GetScoreMultiplier возвращает множитель очков
func (g *Game) GetScoreMultiplier() int {
	if _, ok := g.ActiveEffects[PowerUpMultiplier]; ok {
		return 3
	}
	return 1
}

// UpdateDirection обновляет направление движения
func (g *Game) UpdateDirection(newDir Direction) {
	if newDir == Up && g.Direction != Down {
		g.Direction = Up
	} else if newDir == Down && g.Direction != Up {
		g.Direction = Down
	} else if newDir == Left && g.Direction != Right {
		g.Direction = Left
	} else if newDir == Right && g.Direction != Left {
		g.Direction = Right
	}
}

// Update обновляет состояние игры
func (g *Game) Update() (events []GameEvent) {
	g.MoveTimer++
	if g.MoveTimer < g.MoveDelay {
		return events
	}
	g.MoveTimer = 0

	head := g.Snake[0]
	newHead := head

	switch g.Direction {
	case Up:
		newHead.Y--
	case Down:
		newHead.Y++
	case Left:
		newHead.X--
	case Right:
		newHead.X++
	}

	gridX, gridY := g.config.GridSize()

	// Проверка столкновения со стеной
	if newHead.X < 0 || newHead.X >= gridX || newHead.Y < 0 || newHead.Y >= gridY {
		// Если есть дополнительная жизнь
		if g.Lives > 1 {
			g.Lives--
			g.Snake = g.Snake[:3]
			g.Snake[0] = Point{gridX / 2, gridY / 2}
			g.Snake[1] = Point{gridX/2 - 1, gridY / 2}
			g.Snake[2] = Point{gridX/2 - 2, gridY / 2}
			g.Direction = Right
			events = append(events, GameEvent{Type: EventPowerUpExtraLife, Pos: newHead})
			return events
		}
		
		g.GameOver = true
		g.State = GameOver
		events = append(events, GameEvent{Type: EventWallCollision, Pos: newHead})
		return events
	}

	// Проверка столкновения с хвостом
	for _, segment := range g.Snake {
		if segment.X == newHead.X && segment.Y == newHead.Y {
			// Если есть дополнительная жизнь
			if g.Lives > 1 {
				g.Lives--
				g.Snake = g.Snake[:3]
				g.Snake[0] = Point{gridX / 2, gridY / 2}
				g.Snake[1] = Point{gridX/2 - 1, gridY / 2}
				g.Snake[2] = Point{gridX/2 - 2, gridY / 2}
				events = append(events, GameEvent{Type: EventPowerUpExtraLife, Pos: newHead})
				return events
			}
			
			g.GameOver = true
			g.State = GameOver
			events = append(events, GameEvent{Type: EventSelfCollision, Pos: newHead})
			return events
		}
	}

	// Добавление новой головы
	g.Snake = append([]Point{newHead}, g.Snake...)

	// Проверка поедания еды
	if newHead.X == g.Food.X && newHead.Y == g.Food.Y {
		multiplier := g.GetScoreMultiplier()
		g.Score += multiplier
		g.placeFood()
		events = append(events, GameEvent{Type: EventEatFood, Pos: g.Food})

		// Спавн нового врага каждые 2 очка
		if g.Score%2 == 0 {
			g.spawnEnemy()
		}
	} else {
		// Удаление хвоста
		g.Snake = g.Snake[:len(g.Snake)-1]
	}

	// Обновление врагов
	g.EnemyTimer++
	if g.EnemyTimer >= g.EnemyDelay {
		g.EnemyTimer = 0
		enemyEvents := g.updateEnemies()
		events = append(events, enemyEvents...)
	}

	// Спавн бомб
	g.BombTimer++
	if g.BombTimer >= g.BombDelay {
		g.BombTimer = 0
		g.spawnBomb()
	}

	// Спавн монет
	g.CoinTimer++
	if g.CoinTimer >= g.CoinDelay {
		g.CoinTimer = 0
		g.spawnCoin()
	}

	// Спавн бонусов
	g.PowerUpTimer++
	if g.PowerUpTimer >= g.PowerUpDelay {
		g.PowerUpTimer = 0
		g.spawnPowerUp()
	}

	// Обновление бомб
	bombEvents := g.updateBombs()
	events = append(events, bombEvents...)

	// Проверка сбора ключа
	if g.Key != nil && newHead.X == g.Key.Pos.X && newHead.Y == g.Key.Pos.Y {
		g.HasKey = true
		g.Key = nil
		events = append(events, GameEvent{Type: EventCollectKey, Pos: newHead})
	}

	// Проверка сбора монет
	for i := range g.Coins {
		if !g.Coins[i].Collected && newHead.X == g.Coins[i].Pos.X && newHead.Y == g.Coins[i].Pos.Y {
			g.Coins[i].Collected = true
			g.Score += g.Coins[i].Value
			events = append(events, GameEvent{Type: EventCollectCoin, Pos: g.Coins[i].Pos})
		}
	}
	// Удаление собранных монет
	for i := len(g.Coins) - 1; i >= 0; i-- {
		if g.Coins[i].Collected {
			g.Coins = append(g.Coins[:i], g.Coins[i+1:]...)
		}
	}

	// Проверка открытия сундука
	if g.Chest != nil && !g.Chest.Open && newHead.X == g.Chest.Pos.X && newHead.Y == g.Chest.Pos.Y {
		if g.HasKey {
			g.Chest.Open = true
			g.ArrowCount += g.Chest.Arrows
			g.HasKey = false
			events = append(events, GameEvent{Type: EventOpenChest, Pos: g.Chest.Pos})
		}
	}

	// Обновление стрел
	arrowEvents := g.updateArrows()
	events = append(events, arrowEvents...)

	// Проверка сбора бонусов
	for i := range g.PowerUps {
		if g.PowerUps[i].Active && newHead.X == g.PowerUps[i].Pos.X && newHead.Y == g.PowerUps[i].Pos.Y {
			g.PowerUps[i].Active = false
			powerUpEvents := g.applyPowerUp(g.PowerUps[i])
			events = append(events, powerUpEvents...)
		}
	}
	// Удаление использованных бонусов
	for i := len(g.PowerUps) - 1; i >= 0; i-- {
		if !g.PowerUps[i].Active {
			g.PowerUps = append(g.PowerUps[:i], g.PowerUps[i+1:]...)
		}
	}

	// Обновление активных эффектов
	g.updatePowerUps()

	return events
}

// updateEnemies обновляет врагов
func (g *Game) updateEnemies() (events []GameEvent) {
	for i := range g.Enemies {
		enemy := &g.Enemies[i]
		enemy.AnimFrame++

		newPos := enemy.Pos
		switch enemy.Direction {
		case Up:
			newPos.Y--
		case Down:
			newPos.Y++
		case Left:
			newPos.X--
		case Right:
			newPos.X++
		}

		gridX, gridY := g.config.GridSize()
		
		// Проверка границ
		if newPos.X < 0 || newPos.X >= gridX || newPos.Y < 0 || newPos.Y >= gridY {
			enemy.Direction = Direction(rand.Intn(4))
			continue
		}

		enemy.Pos = newPos

		// Случайное изменение направления
		if rand.Intn(10) < 2 {
			enemy.Direction = Direction(rand.Intn(4))
		}

		// Проверка столкновения со змейкой
		for _, segment := range g.Snake {
			if segment.X == enemy.Pos.X && segment.Y == enemy.Pos.Y {
				// Если есть щит, не умираем
				if g.HasShield() {
					// Уничтожаем врага
					events = append(events, GameEvent{Type: EventEnemyKill, Pos: enemy.Pos})
					// Удаляем врага из списка (будет обработано в updateEnemies)
					g.Enemies = append(g.Enemies[:i], g.Enemies[i+1:]...)
					g.Score++
					return events
				}
				
				// Если есть дополнительная жизнь
				if g.Lives > 1 {
					g.Lives--
					// Отбрасываем змейку
					g.Snake = g.Snake[:3]
					g.Snake[0] = Point{gridX / 2, gridY / 2}
					g.Snake[1] = Point{gridX/2 - 1, gridY / 2}
					g.Snake[2] = Point{gridX/2 - 2, gridY / 2}
					events = append(events, GameEvent{Type: EventPowerUpExtraLife, Pos: enemy.Pos})
					return events
				}
				
				g.GameOver = true
				g.State = GameOver
				events = append(events, GameEvent{Type: EventEnemyCollision, Pos: enemy.Pos})
				return events
			}
		}
	}
	return events
}

// updateBombs обновляет бомбы
func (g *Game) updateBombs() (events []GameEvent) {
	for i := len(g.Bombs) - 1; i >= 0; i-- {
		bomb := &g.Bombs[i]
		bomb.Timer++

		// Проверка столкновения со змейкой
		for _, segment := range g.Snake {
			if segment.X == bomb.Pos.X && segment.Y == bomb.Pos.Y {
				g.GameOver = true
				g.State = GameOver
				events = append(events, GameEvent{Type: EventBombCollision, Pos: bomb.Pos})
				return events
			}
		}

		// Взрыв бомбы
		if bomb.Timer >= bomb.MaxTime {
			events = append(events, GameEvent{Type: EventBombExplode, Pos: bomb.Pos})
			
			// Проверка урона змейке
			for _, segment := range g.Snake {
				dx := segment.X - bomb.Pos.X
				if dx < 0 {
					dx = -dx
				}
				dy := segment.Y - bomb.Pos.Y
				if dy < 0 {
					dy = -dy
				}
				if dx < 3 && dy < 3 {
					g.GameOver = true
					g.State = GameOver
					return events
				}
			}
			g.Bombs = append(g.Bombs[:i], g.Bombs[i+1:]...)
		}
	}
	return events
}

// updateArrows обновляет стрелы
func (g *Game) updateArrows() (events []GameEvent) {
	for i := len(g.Arrows) - 1; i >= 0; i-- {
		arrow := &g.Arrows[i]
		if !arrow.Active {
			g.Arrows = append(g.Arrows[:i], g.Arrows[i+1:]...)
			continue
		}

		arrow.Speed++
		if arrow.Speed < 3 {
			continue
		}
		arrow.Speed = 0

		// Движение стрелы
		switch arrow.Direction {
		case Up:
			arrow.Pos.Y--
		case Down:
			arrow.Pos.Y++
		case Left:
			arrow.Pos.X--
		case Right:
			arrow.Pos.X++
		}

		gridX, gridY := g.config.GridSize()
		
		// Проверка границ
		if arrow.Pos.X < 0 || arrow.Pos.X >= gridX || arrow.Pos.Y < 0 || arrow.Pos.Y >= gridY {
			arrow.Active = false
			continue
		}

		// Проверка попадания во врага
		for j := len(g.Enemies) - 1; j >= 0; j-- {
			enemy := &g.Enemies[j]
			if arrow.Pos.X == enemy.Pos.X && arrow.Pos.Y == enemy.Pos.Y {
				g.Enemies = append(g.Enemies[:j], g.Enemies[j+1:]...)
				arrow.Active = false
				g.Score++
				events = append(events, GameEvent{Type: EventEnemyKill, Pos: enemy.Pos})
				break
			}
		}
	}
	return events
}

// ShootArrow выпускает стрелу
func (g *Game) ShootArrow() {
	if g.ArrowCount <= 0 {
		return
	}
	
	head := g.Snake[0]
	arrow := Arrow{
		Pos:       head,
		Direction: g.Direction,
		Active:    true,
		Speed:     0,
	}
	g.Arrows = append(g.Arrows, arrow)
	g.ArrowCount--
}

// GameEventType представляет тип игрового события
type GameEventType int

const (
	EventNone GameEventType = iota
	EventEatFood
	EventCollectKey
	EventCollectCoin
	EventOpenChest
	EventEnemyKill
	EventEnemyCollision
	EventBombExplode
	EventBombCollision
	EventWallCollision
	EventSelfCollision
	// Power-up события
	EventPowerUpSlowMotion
	EventPowerUpShield
	EventPowerUpShrink
	EventPowerUpExtraLife
	EventPowerUpLightning
	EventPowerUpMultiplier
)

// GameEvent представляет игровое событие
type GameEvent struct {
	Type GameEventType
	Pos  Point
}
