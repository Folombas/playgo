package main

// Система уровней для Bomberman
// Go365 Day 99 - Улучшение геймплея

import (
	"image/color"
	"math/rand"
)

// ======================== УРОВНИ ========================

type LevelTheme struct {
	Name        string
	BGColor     color.Color
	StoneColor  color.Color
	BrickColor  color.Color
	PlayerColor color.Color
	EnemyColors []color.Color
}

var levelThemes = []LevelTheme{
	{
		Name:        "Classic",
		BGColor:     color.RGBA{50, 50, 80, 255},
		StoneColor:  color.RGBA{120, 120, 140, 255},
		BrickColor:  color.RGBA{180, 100, 60, 255},
		PlayerColor: color.RGBA{255, 150, 200, 255},
		EnemyColors: []color.Color{
			color.RGBA{255, 120, 120, 255},
			color.RGBA{100, 100, 255, 255},
			color.RGBA{100, 255, 100, 255},
		},
	},
	{
		Name:        "Ice World",
		BGColor:     color.RGBA{30, 60, 90, 255},
		StoneColor:  color.RGBA{150, 180, 220, 255},
		BrickColor:  color.RGBA{100, 140, 180, 255},
		PlayerColor: color.RGBA{255, 200, 100, 255},
		EnemyColors: []color.Color{
			color.RGBA{200, 100, 100, 255},
			color.RGBA{150, 200, 255, 255},
			color.RGBA{100, 255, 200, 255},
		},
	},
	{
		Name:        "Lava World",
		BGColor:     color.RGBA{60, 20, 20, 255},
		StoneColor:  color.RGBA{80, 80, 80, 255},
		BrickColor:  color.RGBA{180, 60, 30, 255},
		PlayerColor: color.RGBA{100, 200, 255, 255},
		EnemyColors: []color.Color{
			color.RGBA{255, 150, 50, 255},
			color.RGBA{200, 50, 50, 255},
			color.RGBA{255, 255, 100, 255},
		},
	},
	{
		Name:        "Forest",
		BGColor:     color.RGBA{20, 40, 20, 255},
		StoneColor:  color.RGBA{100, 100, 100, 255},
		BrickColor:  color.RGBA{120, 80, 40, 255},
		PlayerColor: color.RGBA{255, 200, 150, 255},
		EnemyColors: []color.Color{
			color.RGBA{180, 80, 80, 255},
			color.RGBA{80, 150, 80, 255},
			color.RGBA{200, 200, 80, 255},
		},
	},
}

var currentLevel = 0
var levelCompleteCount = 0

func getCurrentTheme() LevelTheme {
	return levelThemes[currentLevel%len(levelThemes)]
}

func nextLevel() {
	currentLevel++
	levelCompleteCount++
}

// ======================== ГЕНЕРАЦИЯ УРОВНЕЙ ========================

type LevelConfig struct {
	Width       int
	Height      int
	EnemyCount  int
	PowerUpChance float64
	HasDoors    bool
	HasKeys     bool
}

var levelConfigs = []LevelConfig{
	{Width: 15, Height: 13, EnemyCount: 3, PowerUpChance: 0.15, HasDoors: false, HasKeys: false},
	{Width: 15, Height: 13, EnemyCount: 4, PowerUpChance: 0.18, HasDoors: true, HasKeys: true},
	{Width: 17, Height: 15, EnemyCount: 5, PowerUpChance: 0.20, HasDoors: true, HasKeys: true},
	{Width: 17, Height: 15, EnemyCount: 6, PowerUpChance: 0.22, HasDoors: true, HasKeys: true},
	{Width: 19, Height: 17, EnemyCount: 7, PowerUpChance: 0.25, HasDoors: true, HasKeys: true},
}

func getLevelConfig(level int) LevelConfig {
	idx := level
	if idx >= len(levelConfigs) {
		// Procedural scaling
		last := levelConfigs[len(levelConfigs)-1]
		last.EnemyCount = 7 + (level-len(levelConfigs))*2
		last.PowerUpChance = 0.25 + float64(level-len(levelConfigs))*0.02
		if last.PowerUpChance > 0.4 {
			last.PowerUpChance = 0.4
		}
		return last
	}
	return levelConfigs[idx]
}

// ======================== УМНЫЙ AI ВРАГОВ ========================

type EnemyAI struct {
	Type         int
	Path         [][2]int
	PathTimer    int
	ThinkTimer   int
	LastKnownPX  int
	LastKnownPY  int
}

func (ai *EnemyAI) Think(ex, ey, px, py int, grid [][2]int) {
	ai.ThinkTimer--
	if ai.ThinkTimer > 0 {
		return
	}
	ai.ThinkTimer = 30 // Передумываем каждые 30 кадров

	// Проверяем видимость игрока
	canSee := canSeePlayer(ex, ey, px, py, grid)
	
	if canSee {
		ai.LastKnownPX = px
		ai.LastKnownPY = py
		// Идём к игроку
		ai.Path = findPath(ex, ey, px, py, grid)
	} else if ai.LastKnownPX != -1 {
		// Идём к последней известной позиции
		ai.Path = findPath(ex, ey, ai.LastKnownPX, ai.LastKnownPY, grid)
		if len(ai.Path) == 0 {
			ai.LastKnownPX = -1
			ai.LastKnownPY = -1
		}
	} else {
		// Случайное блуждание
		if len(ai.Path) == 0 || rand.Intn(5) == 0 {
			dx := rand.Intn(3) - 1
			dy := rand.Intn(3) - 1
			tx := ex + dx
			ty := ey + dy
			if isValidTile(tx, ty, grid) {
				ai.Path = [][2]int{{tx, ty}}
			}
		}
	}
}

func canSeePlayer(ex, ey, px, py int, grid [][2]int) bool {
	// Простая проверка - если на одной линии и нет препятствий
	if ex == px {
		minY, maxY := ey, py
		if minY > maxY {
			minY, maxY = maxY, minY
		}
		for y := minY + 1; y < maxY; y++ {
			if isBlocking(ex, y, grid) {
				return false
			}
		}
		return true
	}
	if ey == py {
		minX, maxX := ex, px
		if minX > maxX {
			minX, maxX = maxX, minX
		}
		for x := minX + 1; x < maxX; x++ {
			if isBlocking(x, ey, grid) {
				return false
			}
		}
		return true
	}
	return false
}

func isBlocking(x, y int, grid [][2]int) bool {
	for _, tile := range grid {
		if tile[0] == x && tile[1] == y {
			return true
		}
	}
	return false
}

func isValidTile(x, y int, grid [][2]int) bool {
	if x < 0 || x >= GW || y < 0 || y >= GH {
		return false
	}
	return !isBlocking(x, y, grid)
}

// Простой pathfinding (BFS)
func findPath(sx, sy, tx, ty int, grid [][2]int) [][2]int {
	if sx == tx && sy == ty {
		return nil
	}

	type node struct {
		x, y int
		path [][2]int
	}

	visited := make(map[string]bool)
	queue := []node{{sx, sy, [][2]int{{sx, sy}}}}
	visited[key(sx, sy)] = true

	dirs := [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.x == tx && curr.y == ty {
			return curr.path[1:] // Возвращаем путь без начальной точки
		}

		for _, d := range dirs {
			nx, ny := curr.x+d[0], curr.y+d[1]
			k := key(nx, ny)
			
			if nx >= 0 && nx < GW && ny >= 0 && ny < GH && !visited[k] && !isBlocking(nx, ny, grid) {
				visited[k] = true
				newPath := make([][2]int, len(curr.path)+1)
				copy(newPath, curr.path)
				newPath[len(curr.path)] = [2]int{nx, ny}
				queue = append(queue, node{nx, ny, newPath})
			}
		}
	}

	return nil
}

func key(x, y int) string {
	return string(rune(x*1000 + y))
}
