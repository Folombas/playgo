package main

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/font"
)

const (
	ScreenWidth  = 1024
	ScreenHeight = 768
	StateMenu = 0; StatePlaying = 1; StateShop = 2; StatePaused = 3; StateLevelComplete = 4; StateGameOver = 5; StateBossFight = 6
	WeaponBlaster = 1; WeaponPlasma = 2; WeaponLaser = 3; WeaponMissile = 4; WeaponShotgun = 5
	EnemyDrone = 1; EnemyFighter = 2; EnemyCruiser = 3; EnemyCarrier = 4; EnemyBoss = 5
	ItemHealth = 1; ItemAmmo = 2; ItemWeapon = 3; ItemShield = 4; ItemGem = 5
)

type Achievement struct { id, name, description string; unlocked bool }
type Weapon struct { id int; name string; damage, fireRate, projectileSpeed float64; color color.RGBA; unlocked bool; level, maxLevel, upgradeCost int }
type Player struct { x, y, vx, vy, width, height, health, maxHealth, shield, maxShield float64; credits, gems, ammo, maxAmmo, level, exp, expToLevel, enemiesKilled, damageTaken, invincibleTime int; weapon *Weapon; fireCooldown float64; invincible bool; speedBoost float64 }
type Enemy struct { x, y, vx, vy, width, height, health, maxHealth, damage, speed float64; enemyType, score, animFrame int; shootTimer float64 }
type Projectile struct { x, y, vx, vy, width, height, damage float64; isPlayer bool; color color.RGBA; life int }
type Item struct { x, y float64; itemType, value, animFrame int; color color.RGBA }
type Particle struct { x, y, vx, vy, size float64; life, maxLife int; color color.RGBA }
type Boss struct { x, y, vx, width, height, health, maxHealth float64; phase, currentAttack int; attackTimer float64 }
type Game struct { player *Player; enemies []*Enemy; projectiles []*Projectile; items []*Item; particles []*Particle; boss *Boss; state, frameCount, wave, waveTimer, score, selectedWeapon int; achievements map[string]*Achievement; newAchievements []*Achievement; gameFont, smallFont font.Face }

var allAchievements = map[string]*Achievement{
	"first_kill": {id: "first_kill", name: "Первая кровь", description: "Уничтожьте первого врага"},
	"weapon_master": {id: "weapon_master", name: "Мастер оружия", description: "Купите все виды оружия"},
	"boss_slayer": {id: "boss_slayer", name: "Убийца боссов", description: "Победите босса"},
	"treasure_hunter": {id: "treasure_hunter", name: "Охотник за сокровищами", description: "Соберите 100 кристаллов"},
	"survivor": {id: "survivor", name: "Выживший", description: "Достигните 10 волны"},
	"sharpshooter": {id: "sharpshooter", name: "Снайпер", description: "Убейте 50 врагов без урона"},
	"rich": {id: "rich", name: "Богач", description: "Накопите 1000 кредитов"},
	"max_power": {id: "max_power", name: "Максимальная мощь", description: "Улучшите оружие до максимума"},
}

var weapons = map[int]*Weapon{
	WeaponBlaster: {id: WeaponBlaster, name: "Бластер", damage: 10, fireRate: 0.3, projectileSpeed: 12, color: color.RGBA{0, 255, 0, 255}, unlocked: true, level: 1, maxLevel: 5, upgradeCost: 100},
	WeaponPlasma: {id: WeaponPlasma, name: "Плазмаган", damage: 25, fireRate: 0.5, projectileSpeed: 10, color: color.RGBA{0, 200, 255, 255}, unlocked: false, level: 1, maxLevel: 5, upgradeCost: 300},
	WeaponLaser: {id: WeaponLaser, name: "Лазер", damage: 15, fireRate: 0.1, projectileSpeed: 15, color: color.RGBA{255, 0, 0, 255}, unlocked: false, level: 1, maxLevel: 5, upgradeCost: 500},
	WeaponMissile: {id: WeaponMissile, name: "Ракетница", damage: 50, fireRate: 1.0, projectileSpeed: 8, color: color.RGBA{255, 100, 0, 255}, unlocked: false, level: 1, maxLevel: 5, upgradeCost: 700},
	WeaponShotgun: {id: WeaponShotgun, name: "Дробовик", damage: 8, fireRate: 0.8, projectileSpeed: 10, color: color.RGBA{200, 0, 255, 255}, unlocked: false, level: 1, maxLevel: 5, upgradeCost: 600},
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		player: &Player{x: ScreenWidth/2, y: ScreenHeight-100, width: 40, height: 50, maxHealth: 100, health: 100, maxShield: 50, credits: 0, gems: 0, weapon: weapons[WeaponBlaster], maxAmmo: 100, ammo: 100, speedBoost: 1.0},
		state: StateMenu, achievements: allAchievements, newAchievements: make([]*Achievement, 0),
	}
}

func (g *Game) Update() error {
	g.frameCount++
	switch g.state {
	case StateMenu: g.updateMenu()
	case StatePlaying: g.updatePlaying()
	case StateShop: g.updateShop()
	case StatePaused: g.updatePaused()
	case StateLevelComplete: g.updateLevelComplete()
	case StateGameOver: g.updateGameOver()
	case StateBossFight: g.updateBossFight()
	}
	g.updateAchievements()
	return nil
}

func (g *Game) updateMenu() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) { g.startGame() }
	if inpututil.IsKeyJustPressed(ebiten.KeyS) { g.state = StateShop }
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.state = StateMenu }
}

func (g *Game) startGame() {
	p := g.player
	p.health, p.shield, p.credits, p.gems = p.maxHealth, 0, 0, 0
	p.exp, p.level, p.expToLevel, p.enemiesKilled, p.damageTaken = 0, 1, 100, 0, 0
	g.wave, g.score, g.waveTimer = 1, 0, 0
	g.enemies, g.projectiles, g.items, g.particles = make([]*Enemy, 0), make([]*Projectile, 0), make([]*Item, 0), make([]*Particle, 0)
	g.state = StatePlaying
	playSound(SoundStart)
}

func (g *Game) updatePlaying() {
	g.waveTimer++
	if g.waveTimer > 120 && len(g.enemies) < 5+g.wave*2 { g.spawnEnemy(); g.waveTimer = 0 }
	if g.wave > 0 && g.wave%5 == 0 && len(g.enemies) == 0 && g.boss == nil { g.spawnBoss(); g.state = StateBossFight }
	g.updatePlayer(); g.updateProjectiles(); g.updateEnemies(); g.updateItems(); g.updateParticles()
	if len(g.enemies) == 0 && g.boss == nil && g.waveTimer > 60 { g.wave++; g.waveTimer = 0; g.spawnWaveBonus() }
	if g.player.health <= 0 { g.state = StateGameOver }
}

func (g *Game) updatePlayer() {
	p := g.player
	speed := 6.0 * p.speedBoost
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) { p.vx = -speed } else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) { p.vx = speed } else { p.vx *= 0.9 }
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) { p.vy = -speed } else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) { p.vy = speed } else { p.vy *= 0.9 }
	p.x += p.vx; p.y += p.vy
	if p.x < 0 { p.x = 0 }; if p.x > ScreenWidth-p.width { p.x = ScreenWidth - p.width }
	if p.y < 0 { p.y = 0 }; if p.y > ScreenHeight-p.height { p.y = ScreenHeight - p.height }
	if p.fireCooldown > 0 { p.fireCooldown-- }
	if (ebiten.IsKeyPressed(ebiten.KeySpace) || ebiten.IsKeyPressed(ebiten.KeyK)) && p.fireCooldown <= 0 && p.ammo > 0 { g.shoot(); p.fireCooldown = p.weapon.fireRate * 60; p.ammo-- }
	if ebiten.IsKeyPressed(ebiten.KeyR) && p.ammo < p.maxAmmo { p.ammo = p.maxAmmo }
	if inpututil.IsKeyJustPressed(ebiten.KeyP) { g.state = StateShop }
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.state = StatePaused }
	if p.invincibleTime > 0 { p.invincibleTime--; if p.invincibleTime <= 0 { p.invincible = false } }
}

func (g *Game) shoot() {
	p, w := g.player, g.player.weapon
	if w.id == WeaponShotgun {
		for i := -2; i <= 2; i++ { g.projectiles = append(g.projectiles, &Projectile{x: p.x + p.width/2, y: p.y, vx: float64(i) * 2, vy: -w.projectileSpeed, width: 6, height: 15, damage: w.damage * float64(w.level), isPlayer: true, color: w.color, life: 120}) }
	} else if w.id == WeaponMissile {
		g.projectiles = append(g.projectiles, &Projectile{x: p.x + p.width/2 - 10, y: p.y, vx: 0, vy: -w.projectileSpeed, width: 12, height: 25, damage: w.damage * float64(w.level), isPlayer: true, color: w.color, life: 180})
	} else {
		g.projectiles = append(g.projectiles, &Projectile{x: p.x + p.width/2 - 3, y: p.y, vx: 0, vy: -w.projectileSpeed, width: 6, height: 15, damage: w.damage * float64(w.level), isPlayer: true, color: w.color, life: 120})
	}
	playSound(SoundShoot)
}

func (g *Game) spawnEnemy() {
	enemyType := EnemyDrone
	if g.wave > 2 && rand.Float32() < 0.3 { enemyType = EnemyFighter }
	if g.wave > 4 && rand.Float32() < 0.2 { enemyType = EnemyCruiser }
	if g.wave > 6 && rand.Float32() < 0.1 { enemyType = EnemyCarrier }
	var health, damage, speed float64; var score int
	switch enemyType {
	case EnemyDrone: health, damage, speed, score = 20, 5, 2, 10
	case EnemyFighter: health, damage, speed, score = 40, 10, 3, 25
	case EnemyCruiser: health, damage, speed, score = 80, 15, 1.5, 50
	case EnemyCarrier: health, damage, speed, score = 150, 20, 1, 100
	}
	g.enemies = append(g.enemies, &Enemy{x: float64(rand.Intn(ScreenWidth - 50)), y: -50, width: 40, height: 40, enemyType: enemyType, health: health * (1 + float64(g.wave)*0.1), maxHealth: health * (1 + float64(g.wave)*0.1), damage: damage, speed: speed, score: score})
}

func (g *Game) spawnBoss() {
	g.boss = &Boss{x: ScreenWidth/2 - 75, y: -150, vx: 2, width: 150, height: 120, health: 500 * float64(g.wave/5), maxHealth: 500 * float64(g.wave/5), phase: 1}
	playSound(SoundBoss)
}

func (g *Game) updateProjectiles() {
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		proj := g.projectiles[i]
		proj.x += proj.vx; proj.y += proj.vy; proj.life--
		if proj.life <= 0 || proj.y < -50 || proj.y > ScreenHeight+50 { g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...); continue }
		if proj.isPlayer {
			for _, enemy := range g.enemies {
				if g.checkCollision(proj, enemy) {
					enemy.health -= proj.damage
					g.spawnHitParticles(enemy.x+enemy.width/2, enemy.y+enemy.height/2, proj.color)
					g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
					if enemy.health <= 0 {
						g.player.enemiesKilled++; g.player.exp += enemy.score; g.player.credits += enemy.score; g.score += enemy.score
						g.spawnExplosionParticles(enemy.x+enemy.width/2, enemy.y+enemy.height/2)
						g.spawnItem(enemy.x, enemy.y)
						g.unlockAchievement("first_kill")
					}
					break
				}
			}
		}
	}
}

func (g *Game) updateEnemies() {
	for i := len(g.enemies) - 1; i >= 0; i-- {
		enemy := g.enemies[i]
		enemy.animFrame++; enemy.y += enemy.speed; enemy.x += math.Sin(float64(enemy.animFrame)*0.05) * enemy.speed
		enemy.shootTimer--
		if enemy.shootTimer <= 0 && enemy.y > 0 && enemy.y < ScreenHeight-200 {
			g.projectiles = append(g.projectiles, &Projectile{x: enemy.x + enemy.width/2, y: enemy.y + enemy.height, vx: 0, vy: 5, width: 8, height: 15, damage: enemy.damage, isPlayer: false, color: color.RGBA{255, 50, 50, 255}, life: 120})
			enemy.shootTimer = 120
		}
		if g.checkCollision(enemy, g.player) && !g.player.invincible { g.playerHit(enemy.damage) }
		if enemy.y > ScreenHeight || enemy.health <= 0 { g.enemies = append(g.enemies[:i], g.enemies[i+1:]...) }
	}
}

func (g *Game) updateItems() {
	for i := len(g.items) - 1; i >= 0; i-- {
		item := g.items[i]
		item.animFrame++; item.y += 2
		if g.checkItemCollision(item, g.player) { g.collectItem(item); g.items = append(g.items[:i], g.items[i+1:]...); continue }
		if item.y > ScreenHeight { g.items = append(g.items[:i], g.items[i+1:]...) }
	}
}

func (g *Game) collectItem(item *Item) {
	p := g.player
	switch item.itemType {
	case ItemHealth: if p.health < p.maxHealth { p.health += 20; if p.health > p.maxHealth { p.health = p.maxHealth } }
	case ItemAmmo: if p.ammo < p.maxAmmo { p.ammo += 30; if p.ammo > p.maxAmmo { p.ammo = p.maxAmmo } }
	case ItemShield: p.shield += 25
	case ItemGem: p.gems += item.value; p.credits += item.value * 10; playSound(SoundCoin)
	}
	playSound(SoundItem)
	g.spawnCollectParticles(item.x, item.y, item.color)
}

func (g *Game) spawnItem(x, y float64) {
	randVal := rand.Float32()
	var itemType, value int; var c color.RGBA
	if randVal < 0.3 { itemType, value, c = ItemHealth, 20, color.RGBA{0, 255, 0, 255} } else if randVal < 0.5 { itemType, value, c = ItemAmmo, 30, color.RGBA{255, 165, 0, 255} } else if randVal < 0.7 { itemType, value, c = ItemShield, 25, color.RGBA{0, 100, 255, 255} } else { itemType, value, c = ItemGem, rand.Intn(5)+1, color.RGBA{200, 0, 255, 255} }
	g.items = append(g.items, &Item{x: x, y: y, itemType: itemType, value: value, color: c})
}

func (g *Game) spawnWaveBonus() { g.items = append(g.items, &Item{x: ScreenWidth / 2, y: 100, itemType: ItemGem, value: 5, color: color.RGBA{255, 215, 0, 255}}) }

func (g *Game) updateParticles() {
	for i := len(g.particles) - 1; i >= 0; i-- {
		p := g.particles[i]
		p.x += p.vx; p.y += p.vy; p.vy += 0.1; p.life--
		if p.life <= 0 { g.particles = append(g.particles[:i], g.particles[i+1:]...) }
	}
}

func (g *Game) playerHit(damage float64) {
	p := g.player
	if p.shield > 0 { shieldDmg := damage; if shieldDmg > p.shield { shieldDmg = p.shield }; p.shield -= shieldDmg; damage -= shieldDmg }
	if damage > 0 { p.health -= damage; p.invincible = true; p.invincibleTime = 60; p.damageTaken++; playSound(SoundHit); g.spawnHitParticles(p.x+p.width/2, p.y+p.height/2, color.RGBA{255, 0, 0, 255}) }
}

func (g *Game) checkCollision(a, b interface{}) bool {
	var ax, ay, aw, ah, bx, by, bw, bh float64
	switch v := a.(type) { case *Projectile: ax, ay, aw, ah = v.x, v.y, v.width, v.height; case *Enemy: ax, ay, aw, ah = v.x, v.y, v.width, v.height }
	switch v := b.(type) { case *Enemy: bx, by, bw, bh = v.x, v.y, v.width, v.height; case *Player: bx, by, bw, bh = v.x, v.y, v.width, v.height }
	return ax < bx+bw && ax+aw > bx && ay < by+bh && ay+ah > by
}

func (g *Game) checkItemCollision(item *Item, player *Player) bool { dx := (player.x + player.width/2) - item.x; dy := (player.y + player.height/2) - item.y; return math.Sqrt(dx*dx+dy*dy) < player.width/2+player.height/3 }

func (g *Game) spawnHitParticles(x, y float64, c color.RGBA) { for i := 0; i < 8; i++ { g.particles = append(g.particles, &Particle{x: x, y: y, vx: float64(rand.Intn(10)-5) * 0.8, vy: float64(rand.Intn(10)-5) * 0.8, life: 20 + rand.Intn(10), maxLife: 30, color: c, size: float64(rand.Intn(4) + 2)}) } }
func (g *Game) spawnExplosionParticles(x, y float64) { for i := 0; i < 20; i++ { angle := float64(i) * 2 * math.Pi / 20; speed := float64(rand.Intn(8)+4) * 0.7; g.particles = append(g.particles, &Particle{x: x, y: y, vx: math.Cos(angle) * speed, vy: math.Sin(angle) * speed, life: 40 + rand.Intn(20), maxLife: 60, color: color.RGBA{255, uint8(rand.Intn(100) + 100), 0, 255}, size: float64(rand.Intn(6) + 3)}) } }
func (g *Game) spawnCollectParticles(x, y float64, c color.RGBA) { for i := 0; i < 12; i++ { g.particles = append(g.particles, &Particle{x: x, y: y, vx: float64(rand.Intn(8)-4) * 0.6, vy: float64(-rand.Intn(8)-4) * 0.5, life: 30 + rand.Intn(15), maxLife: 45, color: c, size: float64(rand.Intn(5) + 2)}) } }

func (g *Game) updateAchievements() {
	if g.player.enemiesKilled >= 50 && g.player.damageTaken == 0 { g.unlockAchievement("sharpshooter") }
	if g.player.credits >= 1000 { g.unlockAchievement("rich") }
	if g.wave >= 10 { g.unlockAchievement("survivor") }
	if g.player.gems >= 100 { g.unlockAchievement("treasure_hunter") }
	if len(g.newAchievements) > 0 && g.frameCount%180 == 0 { if len(g.newAchievements) > 1 { g.newAchievements = g.newAchievements[1:] } else { g.newAchievements = make([]*Achievement, 0) } }
}

func (g *Game) unlockAchievement(id string) { if ach, ok := g.achievements[id]; ok && !ach.unlocked { ach.unlocked = true; g.newAchievements = append(g.newAchievements, ach); playSound(SoundPowerup) } }

func (g *Game) updateShop() {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyP) { g.state = StatePlaying; return }
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) { g.selectedWeapon--; if g.selectedWeapon < WeaponBlaster { g.selectedWeapon = WeaponShotgun }; playSound(SoundItem) }
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) { g.selectedWeapon++; if g.selectedWeapon > WeaponShotgun { g.selectedWeapon = WeaponBlaster }; playSound(SoundItem) }
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		w := weapons[g.selectedWeapon]
		if !w.unlocked {
			if g.player.credits >= w.upgradeCost*3 { g.player.credits -= w.upgradeCost * 3; w.unlocked = true; g.player.weapon = w; playSound(SoundPowerup) }
		} else if w.level < w.maxLevel {
			if g.player.credits >= w.upgradeCost*w.level { g.player.credits -= w.upgradeCost * w.level; w.level++; g.player.weapon = w; playSound(SoundPowerup); if w.level == w.maxLevel { g.unlockAchievement("max_power") } }
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyH) { if g.player.credits >= 50 && g.player.health < g.player.maxHealth { g.player.credits -= 50; g.player.health += 30; if g.player.health > g.player.maxHealth { g.player.health = g.player.maxHealth }; playSound(SoundItem) } }
	allUnlocked := true; for _, w := range weapons { if !w.unlocked { allUnlocked = false; break } }; if allUnlocked { g.unlockAchievement("weapon_master") }
}

func (g *Game) updateBossFight() {
	if g.boss == nil { g.state = StatePlaying; return }
	b := g.boss
	b.x += b.vx; if b.x <= 0 || b.x >= float64(ScreenWidth)-b.width { b.vx = -b.vx }
	if b.y < 50 { b.y += 2 }
	b.attackTimer--
	if b.attackTimer <= 0 { g.bossAttack(); b.attackTimer = 60 }
	for i := len(g.projectiles) - 1; i >= 0; i-- {
		proj := g.projectiles[i]
		if proj.isPlayer && proj.x < b.x+b.width && proj.x+proj.width > b.x && proj.y < b.y+b.height && proj.y+proj.height > b.y {
			b.health -= proj.damage; g.spawnHitParticles(proj.x, proj.y, proj.color); g.projectiles = append(g.projectiles[:i], g.projectiles[i+1:]...)
		}
	}
	if b.health <= 0 {
		g.boss = nil; g.state = StateLevelComplete; g.player.credits += 500; g.player.gems += 10; g.score += 1000; g.unlockAchievement("boss_slayer")
		g.spawnExplosionParticles(b.x+b.width/2, b.y+b.height/2); playSound(SoundWin)
	}
}

func (g *Game) bossAttack() {
	b := g.boss
	switch b.currentAttack {
	case 0: for i := -3; i <= 3; i++ { g.projectiles = append(g.projectiles, &Projectile{x: b.x + b.width/2, y: b.y + b.height, vx: float64(i) * 1.5, vy: 4, width: 10, height: 20, damage: 15, isPlayer: false, color: color.RGBA{255, 0, 100, 255}, life: 180}) }
	case 1: g.projectiles = append(g.projectiles, &Projectile{x: b.x + b.width/2 - 20, y: b.y + b.height, vx: 0, vy: 8, width: 40, height: 60, damage: 30, isPlayer: false, color: color.RGBA{255, 0, 0, 255}, life: 60})
	case 2: dx := g.player.x - b.x; dy := g.player.y - b.y; dist := math.Sqrt(dx*dx + dy*dy); g.projectiles = append(g.projectiles, &Projectile{x: b.x + b.width/2, y: b.y + b.height, vx: dx / dist * 5, vy: dy / dist * 5, width: 15, height: 15, damage: 20, isPlayer: false, color: color.RGBA{150, 0, 255, 255}, life: 180})
	}
	b.currentAttack = (b.currentAttack + 1) % 3
}

func (g *Game) updatePaused() { if inpututil.IsKeyJustPressed(ebiten.KeyEscape) { g.state = StatePlaying } }
func (g *Game) updateLevelComplete() { if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) { g.state = StatePlaying } }
func (g *Game) updateGameOver() { if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) { g.state = StateMenu } }
func minF(a, b float64) float64 { if a < b { return a }; return b }
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) { return ScreenWidth, ScreenHeight }

type SoundType int
const ( SoundShoot SoundType = iota; SoundHit; SoundExplosion; SoundCoin; SoundPowerup; SoundItem; SoundStart; SoundWin; SoundBoss )
var audioCtx *audio.Context
func initAudio() { audioCtx = audio.NewContext(44100) }
func generateSound(frequency, duration float64, waveType int) []byte {
	sampleRate := 44100; numSamples := int(float64(sampleRate) * duration); samples := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ { t := float64(i) / float64(sampleRate); envelope := 1.0 - float64(i)/float64(numSamples); var value float64
		switch waveType { case 0: value = math.Sin(2 * math.Pi * frequency * t); case 1: if math.Sin(2*math.Pi*frequency*t) >= 0 { value = 1.0 } else { value = -1.0 }; case 2: value = float64(rand.Intn(2000)-1000) / 1000 }
		value *= envelope * 0.3; sample := int16(value * 32767); samples[i*2] = byte(sample); samples[i*2+1] = byte(sample >> 8) }
	return samples
}
func playSound(sound SoundType) {
	if audioCtx == nil { return }
	var samples []byte
	switch sound { case SoundShoot: samples = generateSound(800, 0.1, 1); case SoundHit: samples = generateSound(200, 0.15, 2); case SoundExplosion: samples = generateSound(100, 0.3, 2); case SoundCoin: samples = generateSound(1200, 0.1, 0); case SoundPowerup: samples = generateSound(600, 0.2, 0); case SoundItem: samples = generateSound(1000, 0.08, 0); case SoundStart: samples = generateSound(500, 0.3, 0); case SoundWin: samples = generateSound(800, 0.4, 0); case SoundBoss: samples = generateSound(150, 0.5, 1) }
	if len(samples) > 0 { player := audioCtx.NewPlayerFromBytes(samples); player.Play() }
}

func main() { initAudio(); ebiten.SetWindowSize(ScreenWidth, ScreenHeight); ebiten.SetWindowTitle("🚀 Space Warrior"); ebiten.SetVsyncEnabled(true); game := NewGame(); if err := ebiten.RunGame(game); err != nil { log.Fatal(err) } }
