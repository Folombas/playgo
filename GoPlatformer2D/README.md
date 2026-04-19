// README for Go Platformer 2D
// Mario-style platformer game

## Project Structure

- `main.go` - Main game entry point
- `sprite_loader.go` - Sprite loading system
- `ai.go` - Enemy AI system  
- `package.json` - Project configuration
- `sprites/` - Sprite assets folder

## Game Features

- Mario-style 2D platformer gameplay
- Player movement (left/right/jump)
- Platform collision
- Enemy AI with patrol and chase behaviors
- Sprite-based rendering system

## Setup

1. Place sprite files in `sprites/` folder:
   - `mario.png` - Player sprite
   - `platform.png` - Platform sprite
   - `enemy.png` - Enemy sprite
   - `sky.png` - Background sprite

2. Run the game:
   ```bash
   go run main.go
   ```

3. Or build:
   ```bash
   go build -o platformer main.go
   ```

## Controls

- Left Arrow: Move left
- Right Arrow: Move right
- Up Arrow: Jump

## Game Architecture

- ECS (Entity Component System) architecture
- Physics-based movement
- Sprite rendering system
- Modular AI system for enemies

## Dependencies

- github.com/OpenGeniusInteractive/paygo - Core game engine