# Changelog

All notable changes to the Crystal Cascade project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-04-13

### Added
- Initial release of Crystal Cascade Match-3 game
- Complete game loop with Update/Draw cycle
- 8x8 game grid with 6 gem types
- Mouse click input for tile selection and swapping
- Match detection for 3+, 4+, and 5+ combinations
- Gravity system for tile falling
- Cascade combo system with score multipliers
- Shake animation for invalid swaps
- Fade-out animation for tile removal
- Hint system (activates after 5 seconds of inactivity)
- 60-second timer with progress bar
- Score system with bonuses (50 for 4-match, 100 for 5+)
- Game Over screen with final score
- HUD with score, timer, and combo display
- Particle effects for visual feedback
- Sprite loading with fallback to colored circles
- Unit tests (12 tests covering core logic)
- Build scripts for Windows, Linux, macOS, and Web
- README with comprehensive documentation
- Web assembly (WASM) support

### Technical Details
- Engine: Ebitengine v2.9.9
- Language: Go 1.25
- Window size: 800x800
- Target FPS: 60
- Platforms: Windows, Linux, macOS, Android, Web
