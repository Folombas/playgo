package main

import (
	"image/color"
	"math"
	"math/rand"
)

// ItemDatabase - база данных предметов
var ItemDatabase = map[int]Item{
	// Weapons
	1: {id: 1, name: "Wooden Sword", description: "Simple wooden sword", itemType: Weapon, rarity: Common, damage: 5, value: 10, icon: "🗡️", stackSize: 1},
	2: {id: 2, name: "Iron Sword", description: "Sharp iron sword", itemType: Weapon, rarity: Uncommon, damage: 12, value: 50, icon: "⚔️", stackSize: 1},
	3: {id: 3, name: "Golden Sword", description: "Golden enchanted sword", itemType: Weapon, rarity: Rare, damage: 20, value: 150, icon: "✨", stackSize: 1},
	4: {id: 4, name: "Diamond Blade", description: "Legendary diamond blade", itemType: Weapon, rarity: Epic, damage: 35, value: 500, icon: "💎", stackSize: 1},
	5: {id: 5, name: "Dragon Slayer", description: "Mythical dragon slayer", itemType: Weapon, rarity: Legendary, damage: 50, luck: 10, value: 2000, icon: "🐉", stackSize: 1},
	
	// Armor
	10: {id: 10, name: "Cloth Armor", description: "Simple cloth armor", itemType: Armor, rarity: Common, defense: 3, health: 10, value: 15, icon: "👕", stackSize: 1},
	11: {id: 11, name: "Iron Armor", description: "Sturdy iron armor", itemType: Armor, rarity: Uncommon, defense: 8, health: 25, value: 75, icon: "🛡️", stackSize: 1},
	12: {id: 12, name: "Knight Armor", description: "Knight's plate armor", itemType: Armor, rarity: Rare, defense: 15, health: 50, value: 200, icon: "🏰", stackSize: 1},
	13: {id: 13, name: "Dragon Scale", description: "Dragon scale armor", itemType: Armor, rarity: Epic, defense: 25, health: 100, value: 750, icon: "🐲", stackSize: 1},
	
	// Potions
	20: {id: 20, name: "Health Potion", description: "Restores 50 HP", itemType: Potion, rarity: Common, health: 50, value: 25, icon: "🧪", stackSize: 5},
	21: {id: 21, name: "Super Potion", description: "Restores 100 HP", itemType: Potion, rarity: Uncommon, health: 100, value: 50, icon: "💊", stackSize: 5},
	22: {id: 22, name: "Elixir", description: "Fully restores HP", itemType: Potion, rarity: Rare, health: 999, value: 200, icon: "✨", stackSize: 3},
	
	// Materials
	30: {id: 30, name: "Ruby", description: "Precious red gem", itemType: Material, rarity: Rare, value: 100, icon: "❤️", stackSize: 20},
	31: {id: 31, name: "Sapphire", description: "Precious blue gem", itemType: Material, rarity: Rare, value: 120, icon: "💙", stackSize: 20},
	32: {id: 32, name: "Emerald", description: "Precious green gem", itemType: Material, rarity: Rare, value: 150, icon: "💚", stackSize: 20},
	33: {id: 33, name: "Diamond", description: "Ultimate precious gem", itemType: Material, rarity: Epic, value: 300, icon: "💎", stackSize: 10},
	
	// Treasures
	40: {id: 40, name: "Gold Coin", description: "Shiny gold coin", itemType: Treasure, rarity: Common, value: 10, icon: "🪙", stackSize: 100},
	41: {id: 41, name: "Gold Bar", description: "Heavy gold bar", itemType: Treasure, rarity: Uncommon, value: 100, icon: "🧈", stackSize: 20},
	42: {id: 42, name: "Crown", description: "Royal crown", itemType: Treasure, rarity: Epic, value: 500, icon: "👑", stackSize: 1},
	43: {id: 43, name: "Ancient Artifact", description: "Mysterious ancient relic", itemType: Treasure, rarity: Legendary, value: 1000, icon: "🏺", stackSize: 1},
}

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

	// Generate biomes
	generateBiomes(world)

	// Generate chests
	generateChests(world)

	return world
}

// generateBiomes создаёт биомы в мире
func generateBiomes(world *World) {
	world.biomes = []Biome{
		{xStart: 0, xEnd: biomeWidth/blockSize, biomeType: Forest, name: "🌲 Enchanted Forest"},
		{xStart: biomeWidth/blockSize, xEnd: 2*biomeWidth/blockSize, biomeType: Desert, name: "🏜️ Burning Desert"},
		{xStart: 2*biomeWidth/blockSize, xEnd: 3*biomeWidth/blockSize, biomeType: Mountains, name: "⛰️ Dragon Mountains"},
		{xStart: 3*biomeWidth/blockSize, xEnd: 4*biomeWidth/blockSize, biomeType: Snow, name: "❄️ Frozen Tundra"},
		{xStart: 4*biomeWidth/blockSize, xEnd: 5*biomeWidth/blockSize, biomeType: Forest, name: "🌿 Ancient Woods"},
		{xStart: 5*biomeWidth/blockSize, xEnd: world.width, biomeType: Caves, name: "💀 Dark Caves"},
	}

	// Apply biome-specific blocks
	for _, biome := range world.biomes {
		for x := biome.xStart; x < biome.xEnd && x < world.width; x++ {
			for y := 0; y < world.height; y++ {
				block := &world.blocks[x][y]
				if block.typ == Air {
					continue
				}

				switch biome.biomeType {
				case Desert:
					if block.typ == Grass {
						block.typ = Sand
					}
					// Add cactus on surface
					if block.typ == Sand && y > 0 && world.blocks[x][y-1].typ == Air {
						if rand.Float32() < 0.02 {
							world.blocks[x][y-1].typ = Cactus
						}
					}
				case Snow:
					if block.typ == Grass {
						block.typ = Snow_Block
					}
					if block.typ == Stone && rand.Float32() < 0.3 {
						block.typ = Ice
					}
				case Mountains:
					if block.typ == Dirt {
						block.typ = Stone
					}
					// More ores in mountains
					if block.typ == Stone && rand.Float32() < 0.15 {
						oreRoll := rand.Float32()
						if oreRoll < 0.05 {
							block.typ = Diamond_Ore
						} else if oreRoll < 0.1 {
							block.typ = Gold_Ore
						} else if oreRoll < 0.2 {
							block.typ = Iron_Ore
						} else {
							block.typ = Coal_Ore
						}
					}
				case Caves:
					// Ancient stone in caves
					if block.typ == Stone && y > world.height/2 {
						if rand.Float32() < 0.1 {
							block.typ = Ancient_Stone
						}
					}
				}
			}
		}
	}
}

// generateChests генерирует сундуки по миру
func generateChests(world *World) {
	world.chests = make([]Chest, 0)

	// Generate chests in different locations
	numChests := 30
	for i := 0; i < numChests; i++ {
		x := rand.Intn(world.width - 10) + 5
		y := rand.Intn(world.height/2) + 5

		// Find ground level
		groundY := y
		for groundY < world.height-1 && world.blocks[x][groundY].typ == Air {
			groundY++
		}

		if groundY < world.height-5 {
			// Determine chest type based on depth and randomness
			chestType := WoodenChest
			roll := rand.Float32()
			if groundY > world.height/2 {
				// Deeper = better chests
				if roll < 0.1 {
					chestType = AncientChest
				} else if roll < 0.3 {
					chestType = DiamondChest
				} else if roll < 0.6 {
					chestType = GoldChest
				} else {
					chestType = IronChest
				}
			} else {
				if roll < 0.7 {
					chestType = WoodenChest
				} else {
					chestType = IronChest
				}
			}

			chest := Chest{
				x:         float32(x * blockSize),
				y:         float32((groundY - 2) * blockSize),
				width:     blockSize,
				height:    blockSize,
				opened:    false,
				chestType: chestType,
				loot:      generateChestLoot(chestType),
			}
			world.chests = append(world.chests, chest)
		}
	}
}

// generateChestLoot генерирует содержимое сундука
func generateChestLoot(chestType ChestType) []Item {
	loot := make([]Item, 0)

	// Number of items based on chest type
	numItems := 1
	switch chestType {
	case WoodenChest:
		numItems = 1 + rand.Intn(2)
	case IronChest:
		numItems = 2 + rand.Intn(2)
	case GoldChest:
		numItems = 3 + rand.Intn(2)
	case DiamondChest:
		numItems = 4 + rand.Intn(2)
	case AncientChest:
		numItems = 5 + rand.Intn(2)
	}

	// Loot tables by chest type
	for i := 0; i < numItems; i++ {
		itemRoll := rand.Float32()
		itemId := 0

		switch chestType {
		case WoodenChest:
			if itemRoll < 0.3 {
				itemId = 1 // Wooden Sword
			} else if itemRoll < 0.6 {
				itemId = 20 // Health Potion
			} else if itemRoll < 0.8 {
				itemId = 40 // Gold Coin
			} else {
				itemId = 10 // Cloth Armor
			}
		case IronChest:
			if itemRoll < 0.25 {
				itemId = 2 // Iron Sword
			} else if itemRoll < 0.5 {
				itemId = 21 // Super Potion
			} else if itemRoll < 0.75 {
				itemId = 41 // Gold Bar
			} else {
				itemId = 11 // Iron Armor
			}
		case GoldChest:
			if itemRoll < 0.2 {
				itemId = 3 // Golden Sword
			} else if itemRoll < 0.4 {
				itemId = 30 // Ruby
			} else if itemRoll < 0.6 {
				itemId = 31 // Sapphire
			} else if itemRoll < 0.8 {
				itemId = 22 // Elixir
			} else {
				itemId = 12 // Knight Armor
			}
		case DiamondChest:
			if itemRoll < 0.15 {
				itemId = 4 // Diamond Blade
			} else if itemRoll < 0.35 {
				itemId = 33 // Diamond
			} else if itemRoll < 0.55 {
				itemId = 32 // Emerald
			} else if itemRoll < 0.75 {
				itemId = 42 // Crown
			} else {
				itemId = 13 // Dragon Scale
			}
		case AncientChest:
			if itemRoll < 0.1 {
				itemId = 5 // Dragon Slayer
			} else if itemRoll < 0.3 {
				itemId = 43 // Ancient Artifact
			} else if itemRoll < 0.5 {
				itemId = 33 // Diamond
			} else if itemRoll < 0.7 {
				itemId = 22 // Elixir
			} else {
				itemId = 42 // Crown
			}
		}

		if itemId != 0 {
			item := ItemDatabase[itemId]
			loot = append(loot, item)
		}
	}

	return loot
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

// NewAchievementAlbum создаёт альбом достижений
func NewAchievementAlbum() *AchievementAlbum {
	return &AchievementAlbum{
		achievements: []Achievement{
			{
				id: 0, achType: FirstSteps, title: "Первые шаги",
				description: "Сделайте первый шаг в игре",
				medalTier: Bronze, completed: false, progress: 0, maxProgress: 1,
				icon: "🥉",
			},
			{
				id: 1, achType: BlockMiner, title: "Шахтёр-новичок",
				description: "Добудьте 10 блоков",
				medalTier: Bronze, completed: false, progress: 0, maxProgress: 10,
				icon: "⛏️",
			},
			{
				id: 2, achType: BlockMiner, title: "Опытный шахтёр",
				description: "Добудьте 50 блоков",
				medalTier: Silver, completed: false, progress: 0, maxProgress: 50,
				icon: "🥈",
			},
			{
				id: 3, achType: BlockMiner, title: "Мастер шахты",
				description: "Добудьте 100 блоков",
				medalTier: Gold, completed: false, progress: 0, maxProgress: 100,
				icon: "🥇",
			},
			{
				id: 4, achType: CoinCollector, title: "Коллекционер",
				description: "Соберите 10 монет",
				medalTier: Bronze, completed: false, progress: 0, maxProgress: 10,
				icon: "🪙",
			},
			{
				id: 5, achType: CoinCollector, title: "Богач",
				description: "Соберите 50 монет",
				medalTier: Silver, completed: false, progress: 0, maxProgress: 50,
				icon: "💰",
			},
			{
				id: 6, achType: CoinCollector, title: "Магнат",
				description: "Соберите 100 монет",
				medalTier: Gold, completed: false, progress: 0, maxProgress: 100,
				icon: "👑",
			},
			{
				id: 7, achType: EnemySlayer, title: "Охотник",
				description: "Победите 5 врагов",
				medalTier: Bronze, completed: false, progress: 0, maxProgress: 5,
				icon: "⚔️",
			},
			{
				id: 8, achType: EnemySlayer, title: "Воин",
				description: "Победите 20 врагов",
				medalTier: Silver, completed: false, progress: 0, maxProgress: 20,
				icon: "🗡️",
			},
			{
				id: 9, achType: EnemySlayer, title: "Легенда",
				description: "Победите 50 врагов",
				medalTier: Gold, completed: false, progress: 0, maxProgress: 50,
				icon: "🏆",
			},
			{
				id: 10, achType: DiamondFinder, title: "Ценитель",
				description: "Найдите алмазную руду",
				medalTier: Platinum, completed: false, progress: 0, maxProgress: 1,
				icon: "💎",
			},
			{
				id: 11, achType: WorldExplorer, title: "Путешественник",
				description: "Пройдите 1000 блоков",
				medalTier: Silver, completed: false, progress: 0, maxProgress: 1000,
				icon: "🗺️",
			},
			{
				id: 12, achType: Builder, title: "Строитель",
				description: "Разместите 50 блоков",
				medalTier: Silver, completed: false, progress: 0, maxProgress: 50,
				icon: "🏗️",
			},
			{
				id: 13, achType: Survivor, title: "Выживший",
				description: "Достигните 10 жизней",
				medalTier: Gold, completed: false, progress: 0, maxProgress: 10,
				icon: "❤️",
			},
			{
				id: 14, achType: Champion, title: "Чемпион",
				description: "Завершите все квесты",
				medalTier: Diamond, completed: false, progress: 0, maxProgress: 1,
				icon: "👑",
			},
		},
		totalUnlocked: 0,
		showAlbum:     false,
	}
}

// UpdateAchievement обновляет прогресс достижения
func (a *AchievementAlbum) UpdateAchievement(achType AchievementType, progress int) {
	for i := range a.achievements {
		if a.achievements[i].achType == achType && !a.achievements[i].completed {
			a.achievements[i].progress = progress
			if progress >= a.achievements[i].maxProgress {
				a.achievements[i].completed = true
				a.totalUnlocked++
				// Add to pending notifications
				a.pendingNotifications = append(a.pendingNotifications, a.achievements[i])
			}
		}
	}
}

// GetNextNotification возвращает следующее уведомление и удаляет его из очереди
func (a *AchievementAlbum) GetNextNotification() *Achievement {
	if len(a.pendingNotifications) > 0 {
		ach := a.pendingNotifications[0]
		a.pendingNotifications = a.pendingNotifications[1:]
		return &ach
	}
	return nil
}

// GetMedalColor возвращает цвет медали
func GetMedalColor(tier MedalTier) color.RGBA {
	switch tier {
	case Bronze:
		return color.RGBA{205, 127, 50, 255}
	case Silver:
		return color.RGBA{192, 192, 192, 255}
	case Gold:
		return color.RGBA{255, 215, 0, 255}
	case Platinum:
		return color.RGBA{224, 224, 224, 255}
	case Diamond:
		return color.RGBA{185, 252, 255, 255}
	default:
		return color.RGBA{128, 128, 128, 255}
	}
}

// GetTierName возвращает название уровня медали
func GetTierName(tier MedalTier) string {
	switch tier {
	case Bronze:
		return "Бронза"
	case Silver:
		return "Серебро"
	case Gold:
		return "Золото"
	case Platinum:
		return "Платина"
	case Diamond:
		return "Алмаз"
	default:
		return "Обычная"
	}
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
