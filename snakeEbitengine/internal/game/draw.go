package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"

	"snake/internal/render"
	"snake/internal/types"
)

func (g *Game) Draw(screen *ebiten.Image) {
	drawText := func(str string, x, y int, clr color.Color) {
		if g.fontFace != nil {
			text.Draw(screen, str, g.fontFace, x, y, clr)
		} else {
			ebitenutil.DebugPrintAt(screen, str, x, y)
		}
	}
	render.Draw(screen, g, drawText)
}

func (g *Game) GetState() types.GameState { return g.state }
func (g *Game) GetSnake() []types.Vec { return g.snake }
func (g *Game) GetDir() types.Vec { return g.dir }
func (g *Game) GetFruitX() int { return g.fruitX }
func (g *Game) GetFruitY() int { return g.fruitY }
func (g *Game) GetFruitType() int { return g.fruitType }
func (g *Game) GetBombs() []types.Bomb { return g.bombs }
func (g *Game) GetIce() types.IceBlock { return g.ice }
func (g *Game) GetIceActive() bool { return g.iceActive }
func (g *Game) GetFrozenTimer() float64 { return g.frozenTimer }
func (g *Game) GetScore() int { return g.score }
func (g *Game) GetHealth() int { return g.health }
func (g *Game) GetParticles() []types.Particle { return g.particles }
func (g *Game) GetShake() float64 { return g.shake }
func (g *Game) GetMenuPulse() float64 { return g.menuPulse }
func (g *Game) GetMenuSelected() int { return g.menuSelected }
func (g *Game) GetMenuButtons() []string { return g.menuButtons }
func (g *Game) GetButtonFlash() int { return g.buttonFlash }
func (g *Game) GetAppleImg() *ebiten.Image { return g.assets.Apple }
func (g *Game) GetStrawberryImg() *ebiten.Image { return g.assets.Strawberry }
func (g *Game) GetOrangeImg() *ebiten.Image { return g.assets.Orange }
func (g *Game) GetBananaImg() *ebiten.Image { return g.assets.Banana }
func (g *Game) GetPineappleImg() *ebiten.Image { return g.assets.Pineapple }
func (g *Game) GetGhostFrames() []*ebiten.Image { return g.assets.GhostFrames }
func (g *Game) GetGhostFrameIdx() int { return g.ghostFrameIdx }
func (g *Game) GetGhostActive() bool { return g.ghostActive }
func (g *Game) GetGhostX() int { return g.ghostX }
func (g *Game) GetGhostY() int { return g.ghostY }
func (g *Game) GetGhostModeTimer() float64 { return g.ghostModeTimer }
func (g *Game) GetRoachFrames() []*ebiten.Image { return g.assets.RoachFrames }
func (g *Game) GetRoachFrameIdx() int { return g.roachFrameIdx }
func (g *Game) GetRoachActive() bool { return g.roachActive }
func (g *Game) GetRoachX() int { return g.roachX }
func (g *Game) GetRoachY() int { return g.roachY }
func (g *Game) GetVikingFrames() []*ebiten.Image { return g.vikingFrames }
func (g *Game) GetVikingList() []types.Viking { return g.vikingList }
func (g *Game) GetGifts() []*types.Gift { return g.gifts }
func (g *Game) GetGiftClosedImgs() []*ebiten.Image { return g.giftClosedImgs }
func (g *Game) GetGiftOpenFrames() []*ebiten.Image { return g.giftOpenFrames }
func (g *Game) GetCoins() []types.Coin { return g.coins }
func (g *Game) GetCoinFrames() []*ebiten.Image { return g.coinFrames }
func (g *Game) GetCoinCount() int { return g.coinCount }
func (g *Game) GetKeysCollected() int { return g.keysCollected }
func (g *Game) GetCarryingKey() bool { return g.carryingKey }
func (g *Game) GetKeyOnField() types.KeyOnField { return g.keyOnField }
func (g *Game) GetKeyImg() *ebiten.Image { return g.keyImg }
func (g *Game) GetSettingsVolumeSlider() float64 { return g.settingsVolumeSlider }
func (g *Game) GetSettingsLanguageIndex() int { return g.settingsLanguageIndex }
func (g *Game) GetSettingsDifficultyIndex() int { return g.settingsDifficultyIndex }
func (g *Game) GetSettingsAnimations() bool { return g.settingsAnimations }
func (g *Game) GetSettingsSelected() int { return g.settingsSelected }