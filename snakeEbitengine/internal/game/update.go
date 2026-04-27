package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"snake/internal/settings"
	"snake/internal/types"
)

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

	if g.state == types.STATE_SETTINGS {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.state = types.STATE_MENU
			g.pauseCooldown = 0.3
			g.sndPause.Rewind()
			g.sndPause.Play()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			g.settingsSelected = (g.settingsSelected - 1 + 4) % 4
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
			g.pauseCooldown = 0.15
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) {
			g.settingsSelected = (g.settingsSelected + 1) % 4
			g.sndMenuMove.Rewind()
			g.sndMenuMove.Play()
			g.pauseCooldown = 0.15
		}
		switch g.settingsSelected {
		case 0:
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				g.settingsVolumeSlider -= 0.05
				if g.settingsVolumeSlider < 0 {
					g.settingsVolumeSlider = 0
				}
				settings.Current.Volume = g.settingsVolumeSlider
				g.applySettings()
				settings.Save()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsVolumeSlider += 0.05
				if g.settingsVolumeSlider > 1 {
					g.settingsVolumeSlider = 1
				}
				settings.Current.Volume = g.settingsVolumeSlider
				g.applySettings()
				settings.Save()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 1:
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsLanguageIndex = (g.settingsLanguageIndex + 1) % 2
				if g.settingsLanguageIndex == 0 {
					settings.Current.Language = "ru"
				} else {
					settings.Current.Language = "en"
				}
				settings.Save()
				g.updateMenuButtonsLanguage()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 2:
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
				g.settingsDifficultyIndex = (g.settingsDifficultyIndex - 1 + 3) % 3
				switch g.settingsDifficultyIndex {
				case 0:
					settings.Current.Difficulty = "easy"
				case 1:
					settings.Current.Difficulty = "normal"
				case 2:
					settings.Current.Difficulty = "hard"
				}
				settings.Save()
				g.applySettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsDifficultyIndex = (g.settingsDifficultyIndex + 1) % 3
				switch g.settingsDifficultyIndex {
				case 0:
					settings.Current.Difficulty = "easy"
				case 1:
					settings.Current.Difficulty = "normal"
				case 2:
					settings.Current.Difficulty = "hard"
				}
				settings.Save()
				g.applySettings()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		case 3:
			if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyRight) {
				g.settingsAnimations = !g.settingsAnimations
				settings.Current.BackgroundAnimation = g.settingsAnimations
				settings.Save()
				g.sndMenuMove.Rewind()
				g.sndMenuMove.Play()
				g.pauseCooldown = 0.15
			}
		}
		return nil
	}

	if g.ghostActive {
		g.ghostAnimTimer += dt
		if g.ghostAnimTimer >= 0.1 {
			g.ghostAnimTimer = 0
			if len(g.ghostFrames) > 0 {
				g.ghostFrameIdx = (g.ghostFrameIdx + 1) % len(g.ghostFrames)
			}
		}
	}
	if g.ghostActive && g.state == types.STATE_PLAYING {
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
			if nx >= 0 && nx < types.GridW && ny >= 0 && ny < types.GridH {
				g.ghostX, g.ghostY = nx, ny
			}
			g.ghostMoveTimer = 0.5
		}
	}

	if g.state == types.STATE_PLAYING && settings.Current.BackgroundAnimation {
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
				if nx >= 0 && nx < types.GridW && ny >= 0 && ny < types.GridH && !g.isCellOccupied(nx, ny) {
					g.roachX, g.roachY = nx, ny
					if len(g.roachFrames) > 0 {
						g.roachFrameIdx = (g.roachFrameIdx + 1) % len(g.roachFrames)
					}
				}
				g.roachMoveTimer = 0.8
			}
		}
	}

	if g.state == types.STATE_PLAYING && settings.Current.BackgroundAnimation {
		g.vikingSpawnTimer -= dt
		if g.vikingSpawnTimer <= 0 && len(g.vikingList) < 3 {
			g.spawnViking()
			g.vikingSpawnTimer = 10.0
		}
		for i := 0; i < len(g.vikingList); i++ {
			v := &g.vikingList[i]
			v.Timer += dt
			if v.Timer >= 0.1 && len(g.vikingFrames) > 0 {
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
				if nx >= 0 && nx < types.GridW && ny >= 0 && ny < types.GridH && !g.isCellOccupied(nx, ny) {
					v.X, v.Y = nx, ny
				}
			}
			if v.Active && g.snake[0].X == v.X && g.snake[0].Y == v.Y {
				g.health -= 20
				g.addParticles(float64(v.X*types.TileSize+types.TileSize/2), float64(v.Y*types.TileSize+types.TileSize/2), 40, color.RGBA{200, 50, 50, 255}, true)
				g.sndHeal.Rewind()
				g.sndHeal.Play()
				g.vikingList = append(g.vikingList[:i], g.vikingList[i+1:]...)
				i--
				if g.health <= 0 {
					g.state = types.STATE_GAMEOVER
				}
			}
		}
	}

	if g.state == types.STATE_PLAYING {
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

	for i := 0; i < len(g.gifts); i++ {
		if g.gifts[i].Opened {
			g.gifts[i].Life -= dt
			if g.gifts[i].Life <= 0 {
				g.gifts = append(g.gifts[:i], g.gifts[i+1:]...)
				i--
			}
		}
	}

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

	if g.state == types.STATE_PLAYING && ebiten.IsKeyPressed(ebiten.KeyK) && g.pauseCooldown <= 0 {
		g.useKey()
		g.pauseCooldown = 0.2
	}

	if g.state == types.STATE_PLAYING && g.carryingKey {
		head := g.snake[0]
		for _, gift := range g.gifts {
			if !gift.Opened && gift.X == head.X && gift.Y == head.Y {
				g.openGift(gift)
				break
			}
		}
	}

	if g.state == types.STATE_PLAYING && len(g.coins) > 0 {
		head := g.snake[0]
		for i, coin := range g.coins {
			if coin.X == head.X && coin.Y == head.Y {
				g.collectCoin(i)
				break
			}
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) && g.pauseCooldown <= 0 {
		if g.state == types.STATE_PLAYING || g.state == types.STATE_PAUSED || g.state == types.STATE_GAMEOVER {
			g.state = types.STATE_MENU
			if g.state == types.STATE_PAUSED {
				g.menuSelected = 1
			} else {
				g.menuSelected = 0
			}
			g.sndPause.Rewind()
			g.sndPause.Play()
		}
		g.pauseCooldown = 0.3
	}
	if ebiten.IsKeyPressed(ebiten.KeyP) && g.pauseCooldown <= 0 && (g.state == types.STATE_PLAYING || g.state == types.STATE_PAUSED) {
		if g.state == types.STATE_PLAYING {
			g.state = types.STATE_PAUSED
		} else {
			g.state = types.STATE_PLAYING
		}
		g.sndPause.Rewind()
		g.sndPause.Play()
		g.pauseCooldown = 0.3
	}

	if g.state == types.STATE_MENU {
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
				g.state = types.STATE_PLAYING
			case 3:
				g.state = types.STATE_SETTINGS
			case 4:
				return ebiten.Termination
			}
			g.pauseCooldown = 0.3
		}
		return nil
	}

	if g.state == types.STATE_PAUSED {
		return nil
	}
	if g.state == types.STATE_GAMEOVER {
		if inputPressed() && g.pauseCooldown <= 0 {
			g.state = types.STATE_MENU
			g.menuSelected = 0
			g.pauseCooldown = 0.2
		}
		return nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) && g.dir.Y != 1 {
		g.nextDir = types.Vec{0, -1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) && g.dir.Y != -1 {
		g.nextDir = types.Vec{0, 1}
	}
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.dir.X != 1 {
		g.nextDir = types.Vec{-1, 0}
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.dir.X != -1 {
		g.nextDir = types.Vec{1, 0}
	}

	g.ticker += g.speed * dt
	if g.ticker >= 1 {
		g.ticker = 0
		g.dir = g.nextDir
		g.step()
	}

	for i := 0; i < len(g.bombs); i++ {
		g.bombs[i].Timer -= dt
		if g.bombs[i].Timer <= 0 {
			g.bombExplode(i)
			i--
		}
	}

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