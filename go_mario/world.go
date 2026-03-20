package main

import (
	"math"
	"math/rand"
)

// NewWorld создаёт новый мир с процедурной генерацией
func NewWorld(seed int64) *World {
	rand.Seed(seed)
	
	width := worldWidth / blockSize
	height := worldHeight / blockSize
	
	world := &World{
		blocks: make([][]Block, width),
		width:  width,
		height: height,
		seed:   seed,
	}
	
	// Initialize blocks
	for x := range world.blocks {
		world.blocks[x] = make([]Block, height)
	}
	
	// Generate terrain using heightmap
	heightmap := generateHeightmap(width, seed)
	
	// Generate blocks
	for x := 0; x < width; x++ {
		surfaceY := heightmap[x]
		
		for y := 0; y < height; y++ {
			block := Block{
				x: x,
				y: y,
			}
			
			if y < surfaceY {
				// Air above ground
				block.typ = Air
				block.solid = false
				block.minable = false
			} else if y == surfaceY {
				// Grass on top
				block.typ = Grass
				block.solid = true
				block.minable = true
			} else if y > surfaceY && y < surfaceY+rand.Intn(3)+3 {
				// Dirt layer
				block.typ = Dirt
				block.solid = true
				block.minable = true
			} else {
				// Stone with ores
				block.typ = Stone
				block.solid = true
				block.minable = true
				
				// Generate ores
				oreChance := rand.Float64()
				if oreChance < 0.02 {
					block.typ = Diamond_Ore
				} else if oreChance < 0.05 {
					block.typ = Gold_Ore
				} else if oreChance < 0.1 {
					block.typ = Iron_Ore
				} else if oreChance < 0.2 {
					block.typ = Coal_Ore
				}
			}
			
			world.blocks[x][y] = block
		}
	}
	
	// Generate caves
	generateCaves(world)
	
	return world
}

// generateHeightmap создаёт карту высот с помощью шума
func generateHeightmap(width int, seed int64) []int {
	heightmap := make([]int, width)
	
	// Base height
	baseHeight := worldHeight/blockSize/2 - 5
	
	// Generate using multiple sine waves for pseudo-random terrain
	for x := 0; x < width; x++ {
		noise := math.Sin(float64(x)*0.1)*5 +
			math.Sin(float64(x)*0.05)*10 +
			math.Sin(float64(x)*0.02)*15
		
		heightmap[x] = baseHeight + int(noise)
	}
	
	return heightmap
}

// generateCaves создаёт пещеры
func generateCaves(world *World) {
	numCaves := rand.Intn(20) + 10
	
	for i := 0; i < numCaves; i++ {
		// Start position
		startX := rand.Intn(world.width)
		startY := rand.Intn(world.height/2) + world.height/4
		
		// Cave parameters
		length := rand.Intn(50) + 30
		direction := float64(rand.Intn(360)) * math.Pi / 180
		
		x := float64(startX)
		y := float64(startY)
		
		for j := 0; j < length; j++ {
			bx := int(x)
			by := int(y)
			
			if bx >= 0 && bx < world.width && by >= 0 && by < world.height {
				// Only carve through stone and dirt
				if world.blocks[bx][by].typ == Stone || world.blocks[bx][by].typ == Dirt {
					world.blocks[bx][by].typ = Air
					world.blocks[bx][by].solid = false
				}
			}
			
			// Move cave
			direction += (rand.Float64() - 0.5) * 0.5
			x += math.Cos(direction) * 2
			y += math.Sin(direction) * 2
		}
	}
}

// GetBlock возвращает блок по координатам
func (w *World) GetBlock(x, y int) *Block {
	if x >= 0 && x < w.width && y >= 0 && y < w.height {
		return &w.blocks[x][y]
	}
	return nil
}

// SetBlock устанавливает блок
func (w *World) SetBlock(x, y int, block Block) {
	if x >= 0 && x < w.width && y >= 0 && y < w.height {
		w.blocks[x][y] = block
	}
}

// IsSolid проверяет, твёрдый ли блок
func (w *World) IsSolid(x, y int) bool {
	block := w.GetBlock(x, y)
	return block != nil && block.solid
}

// NewInventory создаёт новый инвентарь
func NewInventory() *Inventory {
	inv := &Inventory{
		slots:    make([]InventorySlot, inventorySize),
		selected: 0,
	}
	
	// Initialize empty slots
	for i := range inv.slots {
		inv.slots[i] = InventorySlot{
			item:     Air,
			count:    0,
			maxStack: 64,
		}
	}
	
	// Give player some starting items
	inv.slots[0] = InventorySlot{item: Dirt, count: 10, maxStack: 64}
	inv.slots[1] = InventorySlot{item: Stone, count: 5, maxStack: 64}
	
	return inv
}

// AddItem добавляет предмет в инвентарь
func (inv *Inventory) AddItem(item BlockType, count int) bool {
	// Try to stack with existing items
	for i := range inv.slots {
		if inv.slots[i].item == item && inv.slots[i].count < inv.slots[i].maxStack {
			space := inv.slots[i].maxStack - inv.slots[i].count
			if space >= count {
				inv.slots[i].count += count
				return true
			} else {
				inv.slots[i].count += space
				count -= space
			}
		}
	}
	
	// Find empty slot
	for i := range inv.slots {
		if inv.slots[i].item == Air || inv.slots[i].count == 0 {
			inv.slots[i].item = item
			inv.slots[i].count = count
			return true
		}
	}
	
	return false // No space
}

// RemoveItem удаляет предмет из инвентаря
func (inv *Inventory) RemoveItem(item BlockType, count int) bool {
	// Check if we have enough
	totalCount := 0
	for _, slot := range inv.slots {
		if slot.item == item {
			totalCount += slot.count
		}
	}
	
	if totalCount < count {
		return false
	}
	
	// Remove items
	remaining := count
	for i := range inv.slots {
		if inv.slots[i].item == item {
			if inv.slots[i].count >= remaining {
				inv.slots[i].count -= remaining
				if inv.slots[i].count == 0 {
					inv.slots[i].item = Air
				}
				return true
			} else {
				remaining -= inv.slots[i].count
				inv.slots[i].count = 0
				inv.slots[i].item = Air
			}
		}
	}
	
	return true
}

// HasItem проверяет наличие предмета
func (inv *Inventory) HasItem(item BlockType, count int) bool {
	totalCount := 0
	for _, slot := range inv.slots {
		if slot.item == item {
			totalCount += slot.count
		}
	}
	return totalCount >= count
}

// NewRecipes создаёт список рецептов
func NewRecipes() []Recipe {
	return []Recipe{
		// Planks from wood
		{
			result: Plank,
			count:  4,
			ingredients: map[BlockType]int{
				Wood: 1,
			},
		},
		// Crafting table
		{
			result: Crafting_Table,
			count:  1,
			ingredients: map[BlockType]int{
				Plank: 4,
			},
		},
		// Bricks from clay (simplified: from dirt)
		{
			result: Bricks,
			count:  2,
			ingredients: map[BlockType]int{
				Dirt: 2,
				Stone: 1,
			},
		},
	}
}

// CanCraft проверяет возможность крафта
func CanCraft(recipe Recipe, inv *Inventory) bool {
	for item, count := range recipe.ingredients {
		if !inv.HasItem(item, count) {
			return false
		}
	}
	return true
}

// Craft создаёт предмет по рецепту
func Craft(recipe Recipe, inv *Inventory) bool {
	if !CanCraft(recipe, inv) {
		return false
	}
	
	// Remove ingredients
	for item, count := range recipe.ingredients {
		inv.RemoveItem(item, count)
	}
	
	// Add result
	inv.AddItem(recipe.result, recipe.count)
	return true
}

// NewCamera создаёт камеру
func NewCamera() *Camera {
	return &Camera{
		x: 0,
		y: 0,
	}
}

// NewTutorial создаёт систему обучения
func NewTutorial() *Tutorial {
	return &Tutorial{
		steps: []TutorialStep{
			{
				id:          0,
				title:       "Движение",
				description: "Используйте WASD или Стрелки для движения",
				completed:   false,
			},
			{
				id:          1,
				title:       "Прыжок",
				description: "Нажмите SPACE или W или Стрелку ВВЕРХ для прыжка",
				completed:   false,
			},
			{
				id:          2,
				title:       "Добыча блоков",
				description: "ЛКМ по блоку чтобы добыть его",
				completed:   false,
			},
			{
				id:          3,
				title:       "Размещение блоков",
				description: "ПКМ чтобы разместить выбранный блок",
				completed:   false,
			},
			{
				id:          4,
				title:       "Инвентарь",
				description: "Клавиши 1-9 для выбора слота, колёсико для прокрутки",
				completed:   false,
			},
			{
				id:          5,
				title:       "Сбор монет",
				description: "Собирайте монеты для увеличения счёта",
				completed:   false,
			},
			{
				id:          6,
				title:       "Победа над врагами",
				description: "Прыгайте на врагов сверху чтобы победить их",
				completed:   false,
			},
		},
		currentStep: 0,
		visible:     true,
		showHint:    true,
		hintTimer:   300, // 5 seconds at 60 FPS
	}
}

// NewQuests создаёт список квестов
func NewQuests() []Quest {
	return []Quest{
		{
			id:          0,
			title:       "Первые шаги",
			description: "Добудьте 5 блоков",
			objective:   "0/5 блоков",
			completed:   false,
			reward:      50,
		},
		{
			id:          1,
			title:       "Коллекционер",
			description: "Соберите 10 монет",
			objective:   "0/10 монет",
			completed:   false,
			reward:      100,
		},
		{
			id:          2,
			title:       "Охотник на врагов",
			description: "Победите 3 врагов",
			objective:   "0/3 врагов",
			completed:   false,
			reward:      150,
		},
		{
			id:          3,
			title:       "Шахтёр",
			description: "Найдите и добудьте алмазную руду",
			objective:   "Алмаз не найден",
			completed:   false,
			reward:      500,
		},
	}
}

// NewCheckpoints создаёт контрольные точки
func NewCheckpoints() []Checkpoint {
	return []Checkpoint{
		{x: 200, y: 500, activated: false},
		{x: 800, y: 500, activated: false},
		{x: 1500, y: 500, activated: false},
		{x: 2200, y: 500, activated: false},
		{x: 3000, y: 500, activated: false},
	}
}

// NewHealthPacks создаёт аптечки
func NewHealthPacks() []HealthPack {
	packs := make([]HealthPack, 10)
	for i := range packs {
		packs[i] = HealthPack{
			x:          float32(400 + i*350),
			y:          400,
			vy:         0,
			healAmount: 1,
			collected:  false,
		}
	}
	return packs
}

// UpdateTutorial обновляет состояние туториала
func (t *Tutorial) Update(g *Game) {
	if !t.visible || t.currentStep >= len(t.steps) {
		return
	}

	// Check if current step is completed
	if t.currentStep < len(t.steps) {
		step := &t.steps[t.currentStep]
		
		// Check completion based on step ID
		switch step.id {
		case 0: // Movement
			if g.player.animFrame > 10 {
				step.completed = true
				t.currentStep++
			}
		case 1: // Jump
			// Will be checked in Update
		case 2: // Mining
			// Will be checked in handleBlockInteraction
		case 3: // Placing
			// Will be checked in handleBlockInteraction
		case 4: // Inventory
			// Auto-complete after mining
			if t.steps[2].completed {
				step.completed = true
				t.currentStep++
			}
		case 5: // Coins
			if g.player.coins > 0 {
				step.completed = true
				t.currentStep++
			}
		case 6: // Enemies
			// Will be checked in updateEnemies
		}
	}

	// Update hint timer
	t.hintTimer--
	if t.hintTimer <= 0 {
		t.hintTimer = 300
		t.showHint = !t.showHint
	}
}

// CompleteStep завершает шаг туториала
func (t *Tutorial) CompleteStep(id int) {
	if id >= 0 && id < len(t.steps) {
		t.steps[id].completed = true
		if id == t.currentStep && t.currentStep < len(t.steps)-1 {
			t.currentStep++
		}
	}
}

// GetCurrentHint возвращает текущую подсказку
func (t *Tutorial) GetCurrentHint() string {
	if t.currentStep >= len(t.steps) {
		return ""
	}
	step := t.steps[t.currentStep]
	if step.completed {
		return ""
	}
	return step.title + ": " + step.description
}

// Update обновляет позицию камеры
func (c *Camera) Update(playerX, playerY float64) {
	// Center camera on player
	targetX := playerX - screenWidth/2
	targetY := playerY - screenHeight/2
	
	// Clamp to world bounds
	if targetX < 0 {
		targetX = 0
	}
	if targetX > float64(worldWidth-screenWidth) {
		targetX = float64(worldWidth - screenWidth)
	}
	if targetY < 0 {
		targetY = 0
	}
	if targetY > float64(worldHeight-screenHeight) {
		targetY = float64(worldHeight - screenHeight)
	}
	
	// Smooth camera movement
	c.x += (targetX - c.x) * 0.1
	c.y += (targetY - c.y) * 0.1
}
