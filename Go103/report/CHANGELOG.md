# Changelog - Match-3 Game

## Go103 - 13 апреля 2026

### 🎉 Initial Release

#### Core Features
- ✅ 8x8 game board with 6 tile types (colors)
- ✅ Match detection (horizontal & vertical, 3+ tiles)
- ✅ Tile swapping with validation
- ✅ Cascade system (combo chains)
- ✅ Score system (10pts/tile + bonuses)
- ✅ 60-second countdown timer
- ✅ Auto-hint after 5s inactivity
- ✅ Game over screen with final score
- ✅ New game functionality (R key / button)
- ✅ Pause system (P key)

#### Animations
- ✅ Swap animation (150ms)
- ✅ Shake animation for invalid swaps (150ms, 3 cycles)
- ✅ Match fade-out + scale down (200ms)
- ✅ Drop animation with quadratic easing (250ms)
- ✅ Hint pulsation (sin curve, 2s)

#### Audio
- ✅ Procedural sound generation (no external files)
- ✅ 4 game sounds: Swap, Match, Error, Game Over
- ✅ 16-bit PCM at 44100Hz

#### UI/UX
- ✅ Score display with high score tracking
- ✅ Timer in MM:SS format
- ✅ New Game button
- ✅ Game Over overlay
- ✅ Pause screen
- ✅ Tile selection highlighting
- ✅ Hint visualization (green pulsing border)

#### Graphics
- ✅ Programmatic fallback tiles (colored circles)
- ✅ Sprite embedding system (go:embed)
- ✅ Adaptive board sizing
- ✅ Board frame with dark background

#### Input
- ✅ Mouse click handling
- ✅ Touch input support (mobile)
- ✅ Keyboard shortcuts (R, P)
- ✅ Button hit testing

#### Cross-Platform
- ✅ Windows build (match3.exe)
- ✅ Web version (index.html + WASM)
- ✅ Android build instructions (gomobile)

#### Documentation
- ✅ README.md with full documentation
- ✅ Build instructions for all platforms
- ✅ Architecture diagram
- ✅ Testing guide
- ✅ Customization guide

### Technical Details
- **Engine**: Ebitengine v2.9.9
- **Language**: Go 1.x
- **Architecture**: Clean separation (8 files)
- **Performance**: 60 FPS game loop
- **Code Quality**: Compiles without errors

### Files Created
1. `main.go` - Entry point
2. `game.go` - Game loop
3. `board.go` - Board logic
4. `animation.go` - Animation system
5. `ui.go` - UI rendering
6. `input.go` - Input handling
7. `sounds.go` - Sound system
8. `assets.go` - Asset embedding
9. `index.html` - Web loader
10. `README.md` - Documentation
11. `.gitignore` - Git ignore rules

### Commits
1. 🎮 Init project structure
2. ✨ Add animation system
3. 🖱️ Add input handling
4. 🎨 Add UI rendering
5. 🔊 Add sound system
6. 🎲 Implement game loop
7. 🔧 Fix compilation errors
8. 📄 Add Web version + README
9. 🚀 Final cleanup & push (current)
