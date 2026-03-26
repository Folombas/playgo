// Go365 Day 86 - GO MARIO SURVIVOR v2.0.0
// Roguelite Survivor (Vampire Survivors-style)
// Авто-атака, волны врагов, прокачка, выбор апгрейдов

package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// ============================================================================
// CONSTANTS
// ============================================================================

const (
	ScreenWidth  = 1024
	ScreenHeight = 768

	// Physics
	Gravity      = 0.5
	PlayerSpeed  = 4.0

	// Combat
	BaseAttackCooldown = 40
	MaxLevel           = 99

	// Tile
	TileSize = 48
)

// ============================================================================
// COLORS
// ============================================================================

var (
	ColorUIBg        = color.RGBA{0, 0, 0, 200}
	ColorHealth      = color.RGBA{220, 50, 50, 255}
	ColorExp         = color.RGBA{100, 255, 100, 255}
	ColorGold        = color.RGBA{255, 215, 0, 255}
	ColorWeapon1     = color.RGBA{100, 200, 255, 255}
	ColorWeapon2     = color.RGBA{255, 100, 150, 255}
	ColorWeapon3     = color.RGBA{150, 255, 100, 255}
	ColorEnemyBasic  = color.RGBA{200, 80, 80, 255}
	ColorEnemyFast   = color.RGBA{255, 150, 50, 255}
	ColorEnemyTank   = color.RGBA{150, 50, 200, 255}
	ColorEnemyBoss   = color.RGBA{255, 50, 50, 255}
)

// ============================================================================
// ASSETS
// ============================================================================

type Assets struct {
	playerStand  *ebiten.Image
	playerWalk1  *ebiten.Image
	playerWalk2  *ebiten.Image
	slimeGreen   *ebiten.Image
	slimeBlue    *ebiten.Image
	coinSprite   *ebiten.Image
	gemRed       *ebiten.Image
	gameFont     font.Face
	largeFont    font.Face
}

var gameAssets *Assets

func LoadAssets() *Assets {
	assets := &Assets{}

	assets.playerStand, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_stand.png")
	assets.playerWalk1, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk1.png")
	assets.playerWalk2, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Players/128x256/Green/alienGreen_walk2.png")
	assets.slimeGreen, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeGreen.png")
	assets.slimeBlue, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Enemies/slimeBlue.png")
	assets.coinSprite, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/coinGold.png")
	assets.gemRed, _, _ = ebitenutil.NewImageFromFile("assets/PNG/Items/gemRed.png")

	assets.gameFont, _ = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 20)
	assets.largeFont, _ = loadFont("assets/fonts/SuperAdorable-MAvyp.ttf", 48)

	return assets
}

func loadFont(path string, size int) (font.Face, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ttFont, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(ttFont, &opentype.FaceOptions{
		Size: float64(size),
		DPI:  72,
	})
}

// ============================================================================
// UPGRADES
// ============================================================================

type UpgradeType int

const (
	UpgradeDamage UpgradeType = iota
	UpgradeAttackSpeed
	UpgradeMaxHealth
	UpgradeMoveSpeed
	UpgradeProjectileSize
	UpgradePierce
	UpgradeNewWeapon
)

type Upgrade struct {
	id          UpgradeType
	name        string
	description string
	icon        string
	tier        int
}

var upgradePool = []Upgrade{
	{UpgradeDamage, "Сила атаки", "+20% урона", "⚔️", 1},
	{UpgradeAttackSpeed, "Скорость атаки", "+15% скорости атаки", "⚡", 1},
	{UpgradeMaxHealth, "Макс. здоровье", "+20 HP", "❤️", 1},
	{UpgradeMoveSpeed, "Скорость бега", "+10% скорости", "👟", 1},
	{UpgradeProjectileSize, "Размер снарядов", "+25% размер", "📏", 2},
	{UpgradePierce, "Проникновение", "+1 цель", "🎯", 2},
}

// ============================================================================
// WEAPONS
// ============================================================================

type WeaponType int

const (
	WeaponMagicMissile WeaponType = iota
	WeaponFireball
	WeaponLightning
	WeaponAura
)

type Weapon struct {
	wType      WeaponType
	name       string
	damage     int
	cooldown   int
	timer      int
	level      int
	projectiles int
	pierce     int
	sizeMult   float64
}

func NewWeapon(wType WeaponType) *Weapon {
	switch wType {
	case WeaponMagicMissile:
		return &Weapon{wType: wType, name: "Magic Missile", damage: 15, cooldown: 40, timer: 0, projectiles: 1, pierce: 1, sizeMult: 1.0}
	case WeaponFireball:
		return &Weapon{wType: wType, name: "Fireball", damage: 25, cooldown: 60, timer: 0, projectiles: 1, pierce: 1, sizeMult: 1.5}
	case WeaponLightning:
		return &Weapon{wType: wType, name: "Lightning", damage: 30, cooldown: 80, timer: 0, projectiles: 1, pierce: 2, sizeMult: 1.0}
	case WeaponAura:
		return &Weapon{wType: wType, name: "Aura", damage: 10, cooldown: 20, timer: 0, projectiles: 1, pierce: 0, sizeMult: 2.0}
	}
	return &Weapon{wType: wType, name: "Unknown", damage: 10, cooldown: 40, timer: 0, projectiles: 1, pierce: 1, sizeMult: 1.0}
}

// ============================================================================
// GAME STRUCTURES
// ============================================================================

type Player struct {
	x, y         float64
	vx, vy       float64
	width        float32
	height       float32
	facing       int

	// Stats
	maxHealth    int
	health       int
	level        int
	exp          int
	maxExp       int
	gold         int

	// Combat
	damageMult   float64
	attackSpeed  float64
	moveSpeed    float64

	// Weapons
	weapons      []*Weapon

	// Status
	invincibleTimer int
}

type Enemy struct {
	x, y       float64
	vx, vy     float64
	width      float32
	height     float32
	enemyType  int
	health     int
	maxHealth  int
	damage     int
	expValue   int
	isAlive    bool
	speed      float64
}

type Projectile struct {
	x, y       float64
	vx, vy     float64
	width      float32
	height     float32
	damage     int
	owner      int
	isActive   bool
	life       int
	pierce     int
	hitCount   int
	color      color.RGBA
	wType      WeaponType
}

type Pickup struct {
	x, y       float64
	pType      int // 0: exp, 1: gold, 2: health
	value      int
	isActive   bool
	animFrame  int
}

type Particle struct {
	x, y     float64
	vx, vy   float64
	life     int
	color    color.RGBA
	size     float32
	gravity  float64
}

type DamageNumber struct {
	x, y   float64
	value  int
	isCrit bool
	life   int
	vy     float64
}

type Game struct {
	player      *Player
	enemies     []*Enemy
	projectiles []*Projectile
	pickups     []*Pickup
	particles   []*Particle
	damageNums  []*DamageNumber

	state         int
	frameCount    int
	waveNumber    int
	waveTimer     int
	enemiesToSpawn int
	spawnTimer    int

	score       int
	kills       int
	timeSurvived int

	// Level up
	isLevelUp   bool
	upgradeOptions []Upgrade
}

// ============================================================================
// INITIALIZATION
// ============================================================================

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	gameAssets = LoadAssets()

	g := &Game{
		player: &Player{
			x:           ScreenWidth / 2,
			y:           ScreenHeight / 2,
			width:       40,
			height:      56,
			facing:      1,
			maxHealth:   100,
			health:      100,
			level:       1,
			maxExp:      20,
			damageMult:  1.0,
			attackSpeed: 1.0,
			moveSpeed:   1.0,
			weapons:     []*Weapon{NewWeapon(WeaponMagicMissile)},
		},
		state:          0, // Menu
		frameCount:     0,
		waveNumber:     0,
		enemiesToSpawn: 0,
		particles:      make([]*Particle, 0),
		enemies:        make([]*Enemy, 0),
		projectiles:    make([]*Projectile, 0),
		pickups:        make([]*Pickup, 0),
		damageNums:     make([]*DamageNumber, 0),
	}

	return g
}

func (g *Game) StartGame() {
	g.state = 1 // Playing
	g.waveNumber = 1
	g.StartWave()
}

func (g *Game) StartWave() {
	g.waveTimer = 60 * 30 // 30 seconds per wave
	g.enemiesToSpawn = 5 + g.waveNumber*3
	g.spawnTimer = 0
}

func (g *Game) spawnEnemy() {
	// Spawn around player
	angle := rand.Float64() * math.Pi * 2
	dist := float64(500 + rand.Intn(200))

	ex := g.player.x + math.Cos(angle)*dist
	ey := g.player.y + math.Sin(angle)*dist

	// Clamp to screen
	ex = math.Max(50, math.Min(float64(ScreenWidth)-50, ex))
	ey = math.Max(50, math.Min(float64(ScreenHeight)-50, ey))

	enemyType := 0 // Basic
	randVal := rand.Float32()

	if g.waveNumber >= 3 && randVal < 0.1 {
		enemyType = 2 // Tank
	} else if g.waveNumber >= 2 && randVal < 0.25 {
		enemyType = 1 // Fast
	}

	enemy := &Enemy{
		x: ex,
		y: ey,
		width: 36,
		height: 36,
		enemyType: enemyType,
		isAlive: true,
	}

	switch enemyType {
	case 0: // Basic
		enemy.maxHealth = 30 + g.waveNumber*10
		enemy.health = enemy.maxHealth
		enemy.damage = 8 + g.waveNumber*2
		enemy.expValue = 10
		enemy.speed = 1.5
	case 1: // Fast
		enemy.maxHealth = 20 + g.waveNumber*5
		enemy.health = enemy.maxHealth
		enemy.damage = 6 + g.waveNumber*2
		enemy.expValue = 15
		enemy.speed = 2.5
		enemy.width = 28
		enemy.height = 28
	case 2: // Tank
		enemy.maxHealth = 100 + g.waveNumber*30
		enemy.health = enemy.maxHealth
		enemy.damage = 15 + g.waveNumber*3
		enemy.expValue = 30
		enemy.speed = 0.8
		enemy.width = 56
		enemy.height = 56
	}

	g.enemies = append(g.enemies, enemy)
}

func (g *Game) generateUpgradeOptions() {
	g.upgradeOptions = make([]Upgrade, 3)

	for i := 0; i < 3; i++ {
		idx := rand.Intn(len(upgradePool))
		g.upgradeOptions[i] = upgradePool[idx]
	}
}

func (g *Game) applyUpgrade(upgrade Upgrade) {
	p := g.player

	switch upgrade.id {
	case UpgradeDamage:
		p.damageMult += 0.2
	case UpgradeAttackSpeed:
		p.attackSpeed += 0.15
	case UpgradeMaxHealth:
		p.maxHealth += 20
		p.health += 20
	case UpgradeMoveSpeed:
		p.moveSpeed += 0.1
	case UpgradeProjectileSize:
		for _, w := range p.weapons {
			w.sizeMult += 0.25
		}
	case UpgradePierce:
		for _, w := range p.weapons {
			w.pierce += 1
		}
	case UpgradeNewWeapon:
		if len(p.weapons) < 4 {
			wType := WeaponType(rand.Intn(4))
			p.weapons = append(p.weapons, NewWeapon(wType))
		}
	}
}

// ============================================================================
// UPDATE
// ============================================================================

func (g *Game) Update() error {
	g.frameCount++

	switch g.state {
	case 0: // Menu
		g.updateMenu()
	case 1: // Playing
		g.updatePlaying()
	case 2: // LevelUp
		g.updateLevelUp()
	case 3: // GameOver
		g.updateGameOver()
	}

	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.StartGame()
	}
}

func (g *Game) updatePlaying() {
	g.timeSurvived++
	g.waveTimer--

	if g.waveTimer <= 0 {
		g.waveNumber++
		g.StartWave()
	}

	// Spawn enemies
	if g.enemiesToSpawn > 0 {
		g.spawnTimer--
		if g.spawnTimer <= 0 {
			g.spawnEnemy()
			g.enemiesToSpawn--
			g.spawnTimer = 60 - g.waveNumber*2
			if g.spawnTimer < 20 {
				g.spawnTimer = 20
			}
		}
	}

	g.updatePlayer()
	g.updateEnemies()
	g.updateProjectiles()
	g.updatePickups()
	g.updateParticles()
	g.updateDamageNumbers()
	g.checkCollisions()
	g.checkLevelUp()
}

func (g *Game) updatePlayer() {
	p := g.player

	// Movement
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		p.vx = PlayerSpeed * p.moveSpeed
		p.facing = 1
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		p.vx = -PlayerSpeed * p.moveSpeed
		p.facing = -1
	} else {
		p.vx = 0
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		p.vy = PlayerSpeed * p.moveSpeed
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		p.vy = -PlayerSpeed * p.moveSpeed
	} else {
		p.vy = 0
	}

	p.x += p.vx
	p.y += p.vy

	// Clamp to screen
	p.x = math.Max(0, math.Min(float64(ScreenWidth)-float64(p.width), p.x))
	p.y = math.Max(80, math.Min(float64(ScreenHeight)-float64(p.height), p.y))

	// Auto-attack
	for _, weapon := range p.weapons {
		weapon.timer--
		if weapon.timer <= 0 {
			g.fireWeapon(weapon)
			weapon.timer = int(float64(weapon.cooldown) / p.attackSpeed)
		}
	}

	// Invincibility
	if p.invincibleTimer > 0 {
		p.invincibleTimer--
	}
}

func (g *Game) fireWeapon(weapon *Weapon) {
	p := g.player

	// Find nearest enemy
	var target *Enemy
	minDist := float64(500)

	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}
		dist := math.Hypot(e.x-p.x, e.y-p.y)
		if dist < minDist {
			minDist = dist
			target = e
		}
	}

	if target == nil {
		return
	}

	// Fire based on weapon type
	switch weapon.wType {
	case WeaponMagicMissile, WeaponFireball:
		angle := math.Atan2(target.y-p.y, target.x-p.x)
		speed := 8.0
		if weapon.wType == WeaponFireball {
			speed = 6.0
		}

		for i := 0; i < weapon.projectiles; i++ {
			spreadAngle := angle + (float64(i) - float64(weapon.projectiles-1)/2) * 0.2
			g.projectiles = append(g.projectiles, &Projectile{
				x: p.x + float64(p.width)/2,
				y: p.y + float64(p.height)/2,
				vx: math.Cos(spreadAngle) * speed,
				vy: math.Sin(spreadAngle) * speed,
				width: float32(12 * weapon.sizeMult),
				height: float32(12 * weapon.sizeMult),
				damage: int(float64(weapon.damage) * p.damageMult),
				owner: 1,
				isActive: true,
				life: 120,
				pierce: weapon.pierce,
				color: ColorWeapon1,
				wType: weapon.wType,
			})
		}

	case WeaponLightning:
		// Instant hit, pierces
		g.projectiles = append(g.projectiles, &Projectile{
			x: p.x + float64(p.width)/2,
			y: p.y + float64(p.height)/2,
			vx: 0,
			vy: 0,
			width: 4,
			height: float32(minDist),
			damage: int(float64(weapon.damage) * p.damageMult),
			owner: 1,
			isActive: true,
			life: 10,
			pierce: weapon.pierce + 1,
			color: ColorWeapon3,
			wType: weapon.wType,
		})

	case WeaponAura:
		// Damages all nearby enemies
		for _, e := range g.enemies {
			if !e.isAlive {
				continue
			}
			dist := math.Hypot(e.x-p.x, e.y-p.y)
			if dist < 150 {
				g.damageEnemy(e, weapon.damage)
				g.spawnDamageNumber(e.x, e.y, weapon.damage, false)
			}
		}
	}
}

func (g *Game) updateEnemies() {
	p := g.player

	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}

		// Move towards player
		angle := math.Atan2(p.y-e.y, p.x-e.x)
		e.vx = math.Cos(angle) * e.speed
		e.vy = math.Sin(angle) * e.speed

		// Simple separation
		for _, other := range g.enemies {
			if other == e || !other.isAlive {
				continue
			}
			dist := math.Hypot(e.x-other.x, e.y-other.y)
			if dist < 30 {
				pushAngle := math.Atan2(e.y-other.y, e.x-other.x)
				e.vx += math.Cos(pushAngle) * 0.5
				e.vy += math.Sin(pushAngle) * 0.5
			}
		}

		e.x += e.vx
		e.y += e.vy

		// Collision with player
		if g.checkCollision(p, e) && p.invincibleTimer == 0 {
			p.health -= e.damage
			p.invincibleTimer = 30
			g.spawnDamageNumber(p.x+float64(p.width)/2, p.y+float64(p.height)/2, e.damage, false)

			if p.health <= 0 {
				g.state = 3 // GameOver
			}
		}
	}
}

func (g *Game) updateProjectiles() {
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		proj := g.projectiles[i]

		if proj.wType == WeaponLightning {
			proj.life--
			if proj.life <= 0 {
				proj.isActive = false
			}
			continue
		}

		proj.x += proj.vx
		proj.y += proj.vy
		proj.life--

		if proj.life <= 0 || proj.x < -100 || proj.x > float64(ScreenWidth)+100 ||
			proj.y < -100 || proj.y > float64(ScreenHeight)+100 {
			proj.isActive = false
		}
	}
}

func (g *Game) updatePickups() {
	p := g.player

	for _, pickup := range g.pickups {
		if !pickup.isActive {
			continue
		}

		pickup.animFrame++

		// Magnet effect
		dist := math.Hypot(pickup.x-p.x, pickup.y-p.y)
		if dist < 100 {
			pickup.x += (p.x - pickup.x) * 0.1
			pickup.y += (p.y - pickup.y) * 0.1
		}

		// Collect
		if dist < 30 {
			pickup.isActive = false

			switch pickup.pType {
			case 0: // Exp
				p.exp += pickup.value
			case 1: // Gold
				p.gold += pickup.value
			case 2: // Health
				p.health = int(math.Min(float64(p.health)+float64(pickup.value), float64(p.maxHealth)))
			}
		}
	}
}

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx
		p.y += p.vy
		p.vy += p.gravity
		p.life--
		if p.life <= 0 {
			g.particles = append(g.particles[:i], g.particles[i+1:]...)
		}
	}
}

func (g *Game) updateDamageNumbers() {
	for i := len(g.damageNums) - 1; i >= 0; i-- {
		dn := g.damageNums[i]
		dn.y += dn.vy
		dn.vy -= 0.2
		dn.life--
		if dn.life <= 0 {
			g.damageNums = append(g.damageNums[:i], g.damageNums[i+1:]...)
		}
	}
}

func (g *Game) checkCollisions() {
	// Projectiles vs Enemies
	for _, proj := range g.projectiles {
		if !proj.isActive || proj.wType == WeaponAura {
			continue
		}

		for _, e := range g.enemies {
			if !e.isAlive || proj.hitCount >= proj.pierce+1 {
				continue
			}

			if proj.x < e.x+float64(e.width) &&
				proj.x+float64(proj.width) > e.x &&
				proj.y < e.y+float64(e.height) &&
				proj.y+float64(proj.height) > e.y {

				g.damageEnemy(e, proj.damage)
				g.spawnDamageNumber(e.x, e.y, proj.damage, false)

				if proj.wType != WeaponLightning {
					proj.hitCount++
				}
			}
		}
	}
}

func (g *Game) damageEnemy(e *Enemy, damage int) {
	e.health -= damage

	if e.health <= 0 {
		e.isAlive = false
		g.kills++
		g.score += e.expValue

		// Drop pickup
		g.pickups = append(g.pickups, &Pickup{
			x: e.x,
			y: e.y,
			pType: 0, // Exp
			value: e.expValue,
			isActive: true,
		})

		// Death particles
		g.spawnDeathParticles(e.x, e.y)
	}
}

func (g *Game) checkLevelUp() {
	p := g.player

	if p.exp >= p.maxExp && p.level < MaxLevel {
		p.level++
		p.exp -= p.maxExp
		p.maxExp = int(float64(p.maxExp) * 1.3)
		p.health = p.maxHealth

		g.isLevelUp = true
		g.state = 2 // LevelUp
		g.generateUpgradeOptions()
	}
}

func (g *Game) checkCollision(a interface{}, b interface{}) bool {
	switch va := a.(type) {
	case *Player:
		if vb, ok := b.(*Enemy); ok {
			return va.x < vb.x+float64(vb.width) &&
				va.x+float64(va.width) > vb.x &&
				va.y < vb.y+float64(vb.height) &&
				va.y+float64(va.height) > vb.y
		}
	}
	return false
}

func (g *Game) updateLevelUp() {
	// Select upgrade with keys 1, 2, 3
	for i := 0; i < len(g.upgradeOptions); i++ {
		if inpututil.IsKeyJustPressed(ebiten.Key(i + 1 + 48)) { // 1, 2, 3
			g.applyUpgrade(g.upgradeOptions[i])
			g.isLevelUp = false
			g.state = 1 // Playing
			g.upgradeOptions = nil
			break
		}
	}
}

func (g *Game) updateGameOver() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		*g = *NewGame()
	}
}

// ============================================================================
// PARTICLES
// ============================================================================

func (g *Game) spawnDeathParticles(x, y float64) {
	for i := 0; i < 15; i++ {
		g.particles = append(g.particles, &Particle{
			x: x + 20,
			y: y + 20,
			vx: (rand.Float64() - 0.5) * 8,
			vy: (rand.Float64() - 0.5) * 8,
			life: 30 + rand.Intn(20),
			color: ColorEnemyBasic,
			size: 3 + rand.Float32()*3,
			gravity: 0.3,
		})
	}
}

func (g *Game) spawnDamageNumber(x, y float64, value int, isCrit bool) {
	g.damageNums = append(g.damageNums, &DamageNumber{
		x: x,
		y: y,
		value: value,
		isCrit: isCrit,
		life: 40,
		vy: 2,
	})
}

// ============================================================================
// DRAW
// ============================================================================

func (g *Game) Draw(screen *ebiten.Image) {
	switch g.state {
	case 0:
		g.drawMenu(screen)
	case 1, 2:
		g.drawPlaying(screen)
	case 3:
		g.drawGameOver(screen)
	}
}

func (g *Game) drawMenu(screen *ebiten.Image) {
	// Background gradient
	for y := 0; y < ScreenHeight; y++ {
		r := uint8(20 + float64(y)/ScreenHeight*30)
		g := uint8(30 + float64(y)/ScreenHeight*40)
		b := uint8(60 + float64(y)/ScreenHeight*60)
		vector.DrawFilledRect(screen, 0, float32(y), ScreenWidth, 1, color.RGBA{r, g, b, 255}, true)
	}

	// Title
	title := "🍄 GO MARIO SURVIVOR 🍄"
	if gameAssets.largeFont != nil {
		bounds := text.BoundString(gameAssets.largeFont, title)
		text.Draw(screen, title, gameAssets.largeFont, ScreenWidth/2-bounds.Dx()/2, 180, ColorGold)
	}

	// Subtitle
	if gameAssets.gameFont != nil {
		text.Draw(screen, "Roguelite Survivor", gameAssets.gameFont, ScreenWidth/2-80, 250, color.White)
	}

	// Instructions
	instructions := []string{
		"⬅️ ➡️ ⬆️ ⬇️ / WASD - Движение",
		"⚔️ Авто-атака ближайшего врага",
		"📦 Собирай опыт для прокачки",
		"🎯 Выбери апгрейд цифрами 1/2/3",
		"",
		"Нажми ENTER для старта",
	}

	y := 350
	for _, line := range instructions {
		if gameAssets.gameFont != nil {
			bounds := text.BoundString(gameAssets.gameFont, line)
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-bounds.Dx()/2, y, color.White)
		}
		y += 32
	}

	// Features
	features := []string{
		"🎮 Волны врагов",
		"⚔️ Разные оружия",
		"📈 Система прокачки",
		"💀 Бесконечный геймплей",
	}

	y = 550
	for _, line := range features {
		if gameAssets.gameFont != nil {
			bounds := text.BoundString(gameAssets.gameFont, line)
			text.Draw(screen, line, gameAssets.gameFont, ScreenWidth/2-bounds.Dx()/2, y, ColorGold)
		}
		y += 28
	}
}

func (g *Game) drawPlaying(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{30, 40, 50, 255})

	// Draw pickups
	for _, p := range g.pickups {
		if !p.isActive {
			continue
		}
		c := ColorExp
		if p.pType == 1 {
			c = ColorGold
		} else if p.pType == 2 {
			c = ColorHealth
		}
		vector.DrawFilledCircle(screen, float32(p.x)+15, float32(p.y)+15, 10, c, true)
	}

	// Draw player
	p := g.player
	if p.invincibleTimer == 0 || g.frameCount%4 < 2 {
		var playerImg *ebiten.Image
		if g.frameCount%12 < 6 {
			playerImg = gameAssets.playerWalk1
		} else {
			playerImg = gameAssets.playerWalk2
		}

		if playerImg != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(p.x, p.y)
			op.GeoM.Scale(0.5, 0.5)
			screen.DrawImage(playerImg, op)
		} else {
			vector.DrawFilledRect(screen, float32(p.x), float32(p.y), float32(p.width), float32(p.height), color.RGBA{0, 255, 100, 255}, true)
		}
	}

	// Draw enemies
	for _, e := range g.enemies {
		if !e.isAlive {
			continue
		}
		c := ColorEnemyBasic
		if e.enemyType == 1 {
			c = ColorEnemyFast
		} else if e.enemyType == 2 {
			c = ColorEnemyTank
		}
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y), e.width, e.height, c, true)

		// Health bar
		barWidth := e.width
		healthPercent := float32(e.health) / float32(e.maxHealth)
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y)-6, barWidth, 4, color.RGBA{80, 0, 0, 255}, true)
		vector.DrawFilledRect(screen, float32(e.x), float32(e.y)-6, barWidth*healthPercent, 4, ColorHealth, true)
	}

	// Draw projectiles
	for _, proj := range g.projectiles {
		if !proj.isActive {
			continue
		}
		if proj.wType == WeaponLightning {
			// Draw lightning line
			vector.StrokeLine(screen, float32(proj.x), float32(proj.y),
				float32(proj.x), float32(proj.y)+proj.height, 3, proj.color, true)
		} else {
			vector.DrawFilledCircle(screen, float32(proj.x), float32(proj.y), float32(proj.width)/2, proj.color, true)
		}
	}

	// Draw particles
	for _, part := range g.particles {
		alpha := uint8(255 * part.life / 50)
		c := color.RGBA{part.color.R, part.color.G, part.color.B, alpha}
		vector.DrawFilledCircle(screen, float32(part.x), float32(part.y), part.size, c, true)
	}

	// Draw damage numbers
	for _, dn := range g.damageNums {
		if gameAssets.gameFont != nil {
			txt := fmt.Sprintf("%d", dn.value)
			if dn.isCrit {
				txt = fmt.Sprintf("💥%d!", dn.value)
			}
			if dn.isCrit {
				text.Draw(screen, txt, gameAssets.gameFont, int(dn.x), int(dn.y), color.RGBA{255, 215, 0, 255})
			} else {
				text.Draw(screen, txt, gameAssets.gameFont, int(dn.x), int(dn.y), color.White)
			}
		}
	}

	// Draw UI
	g.drawUI(screen)

	// Draw level up overlay
	if g.state == 2 {
		g.drawLevelUpOverlay(screen)
	}
}

func (g *Game) drawUI(screen *ebiten.Image) {
	p := g.player

	// UI Background
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, 70, ColorUIBg, true)

	// Health bar
	vector.DrawFilledRect(screen, 20, 10, 200, 16, color.RGBA{80, 0, 0, 255}, true)
	healthPercent := float32(p.health) / float32(p.maxHealth)
	vector.DrawFilledRect(screen, 20, 10, 200*healthPercent, 16, ColorHealth, true)

	// Exp bar
	vector.DrawFilledRect(screen, 20, 32, 250, 10, color.RGBA{0, 50, 0, 255}, true)
	expPercent := float32(p.exp) / float32(p.maxExp)
	vector.DrawFilledRect(screen, 20, 32, 250*expPercent, 10, ColorExp, true)

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("HP %d/%d", p.health, p.maxHealth), gameAssets.gameFont, 230, 24, color.White)
		text.Draw(screen, fmt.Sprintf("LVL %d", p.level), gameAssets.gameFont, 350, 24, ColorGold)
		text.Draw(screen, fmt.Sprintf("EXP %d/%d", p.exp, p.maxExp), gameAssets.gameFont, 420, 24, ColorExp)

		// Stats
		minutes := g.timeSurvived / 3600
		seconds := (g.timeSurvived % 3600) / 60
		text.Draw(screen, fmt.Sprintf("⏱️ %d:%02d", minutes, seconds), gameAssets.gameFont, 550, 24, color.White)
		text.Draw(screen, fmt.Sprintf("💀 %d", g.kills), gameAssets.gameFont, 680, 24, ColorEnemyBasic)
		text.Draw(screen, fmt.Sprintf("📊 %d", g.score), gameAssets.gameFont, 780, 24, ColorGold)

		// Wave
		text.Draw(screen, fmt.Sprintf("🌊 WAVE %d", g.waveNumber), gameAssets.gameFont, 900, 24, ColorGold)

		// Weapons
		x := 20
		y := 50
		for _, w := range p.weapons {
			weaponColor := ColorWeapon1
			if w.wType == WeaponFireball {
				weaponColor = ColorWeapon2
			} else if w.wType == WeaponLightning {
				weaponColor = ColorWeapon3
			}
			vector.DrawFilledRect(screen, float32(x), float32(y), 40, 6, color.RGBA{50, 50, 50, 255}, true)
			cdPercent := float32(w.timer) / float32(w.cooldown)
			vector.DrawFilledRect(screen, float32(x), float32(y), 40*(1-cdPercent), 6, weaponColor, true)
			x += 45
		}
	}
}

func (g *Game) drawLevelUpOverlay(screen *ebiten.Image) {
	// Overlay
	vector.DrawFilledRect(screen, 0, 0, ScreenWidth, ScreenHeight, color.RGBA{0, 0, 0, 180}, true)

	if gameAssets.largeFont != nil {
		text.Draw(screen, "🎉 LEVEL UP!", gameAssets.largeFont, ScreenWidth/2-120, 250, ColorGold)
	}

	if gameAssets.gameFont != nil {
		text.Draw(screen, fmt.Sprintf("Level %d reached! Choose upgrade:", g.player.level), gameAssets.gameFont, ScreenWidth/2-150, 320, color.White)

		for i, upgrade := range g.upgradeOptions {
			y := 380 + i*80
			vector.DrawFilledRect(screen, float32(ScreenWidth/2-200), float32(y-30), 400, 60, ColorUIBg, true)
			vector.DrawFilledRect(screen, float32(ScreenWidth/2-200), float32(y-30), 400, 60, color.RGBA{255, 215, 0, 50}, true)

			text.Draw(screen, fmt.Sprintf("[%d] %s %s", i+1, upgrade.icon, upgrade.name), gameAssets.gameFont, ScreenWidth/2-180, y, ColorGold)
			text.Draw(screen, upgrade.description, gameAssets.gameFont, ScreenWidth/2-180, y+25, color.White)
		}
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	screen.Fill(color.RGBA{50, 20, 20, 255})

	if gameAssets.largeFont != nil {
		text.Draw(screen, "💀 GAME OVER", gameAssets.largeFont, ScreenWidth/2-150, ScreenHeight/2-80, ColorHealth)

		minutes := g.timeSurvived / 3600
		seconds := (g.timeSurvived % 3600) / 60

		text.Draw(screen, fmt.Sprintf("Time Survived: %d:%02d", minutes, seconds), gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2, color.White)
		text.Draw(screen, fmt.Sprintf("Wave Reached: %d", g.waveNumber), gameAssets.gameFont, ScreenWidth/2-100, ScreenHeight/2+35, color.White)
		text.Draw(screen, fmt.Sprintf("Enemies Killed: %d", g.kills), gameAssets.gameFont, ScreenWidth/2-100, ScreenHeight/2+70, color.White)
		text.Draw(screen, fmt.Sprintf("Final Score: %d", g.score), gameAssets.gameFont, ScreenWidth/2-100, ScreenHeight/2+105, ColorGold)
		text.Draw(screen, "Press ENTER to restart", gameAssets.gameFont, ScreenWidth/2-120, ScreenHeight/2+160, color.White)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth, ScreenHeight
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowTitle("GO MARIO SURVIVOR - Go365 Day 86 | Roguelite")
	ebiten.SetVsyncEnabled(true)

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
