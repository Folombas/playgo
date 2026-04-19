package main

import (
	"github.com/OpenGeniusInteractive/paygo"
	"github.com/OpenGeniusInteractive/paygo/ecs"
	"github.com/OpenGeniusInteractive/paygo/input"
	"github.com/OpenGeniusInteractive/paygo/render"
	"github.com/OpenGeniusInteractive/paygo/physics"
)

// Game represents the main game state
type Game struct {
	ecs.World
	player       *ecs.Entity
	camera       *ecs.Entity
	currentLevel string
	levelData    map[string]LevelData
}

// LevelData contains level configuration
type LevelData struct {
	Name       string
	Width      float64
	Height     float64
	Background string
	Entities   []EntityData
}

// EntityData represents an entity in the level
type EntityData struct {
	Type     string
	X, Y     float64
	Width    float64
	Height   float64
	Sprite   string
	Solid    bool
	UserData interface{}
}

// NewGame creates a new game instance
func NewGame() *Game {
	game := &Game{}
	game.World = *ecs.NewWorld()
	return game
}

// Init initializes the game
func (g *Game) Init() {
	// Initialize input system
	input.Initialize()
	
	// Initialize render system
	render.SetClearColor(0.1, 0.1, 0.1, 1.0)
	
	// Load first level
	g.LoadLevel("level1")
}

// LoadLevel loads a game level
func (g *Game) LoadLevel(name string) {
	g.currentLevel = name
	
	// Load level data (in a real game, this would load from a file)
	level := g.createDefaultLevel()
	g.levelData = make(map[string]LevelData)
	g.levelData[name] = level
	
	// Clear world
	g.Clear()
	
	// Create level entities
	for _, entityData := range level.Entities {
		entity := ecs.NewEntity(&g.World)
		
		// Add sprite component
		if entityData.Sprite != "" {
			entity.AddComponent(&render.SpriteComponent{
				Sprite:  entityData.Sprite,
				Layer:   0,
				Visible: true,
			})
		}
		
		// Add transform component
		transform := &physics.TransformComponent{
			Position: physics.Vec2{X: entityData.X, Y: entityData.Y},
			Size:     physics.Vec2{X: entityData.Width, Y: entityData.Height},
		}
		entity.AddComponent(transform)
		
		// Add physics body if solid
		if entityData.Solid {
			body := physics.NewBodyStatic(physics.AABB{
				Min: physics.Vec2{X: entityData.X, Y: entityData.Y},
				Max: physics.Vec2{X: entityData.X + entityData.Width, Y: entityData.Y + entityData.Height},
			})
			entity.AddComponent(&physics.BodyComponent{
				Body: body,
			})
		}
		
		entity.SetUserData(entityData.UserData)
		g.AddEntity(entity)
	}
	
	// Create player
	g.createPlayer()
}

// createDefaultLevel creates a default test level
func (g *Game) createDefaultLevel() LevelData {
	return LevelData{
		Name:       "level1",
		Width:      1000,
		Height:     600,
		Background: "sky",
		Entities: []EntityData{
			{
				Type:   "platform",
				X:      100,
				Y:      500,
				Width:  300,
				Height: 20,
				Sprite: "sprites/platform.png",
				Solid:  true,
			},
			{
				Type:   "platform",
				X:      500,
				Y:      400,
				Width:  200,
				Height: 20,
				Sprite: "sprites/platform.png",
				Solid:  true,
			},
			{
				Type:   "enemy",
				X:      700,
				Y:      480,
				Width:  40,
				Height: 40,
				Sprite: "sprites/enemy.png",
				Solid:  true,
			},
		},
	}
}

// createPlayer creates the player character
func (g *Game) createPlayer() {
	player := ecs.NewEntity(&g.World)
	
	// Player sprite
	player.AddComponent(&render.SpriteComponent{
		Sprite:  "sprites/mario.png",
		Layer:   1,
		Visible: true,
	})
	
	// Player transform
	transform := &physics.TransformComponent{
		Position: physics.Vec2{X: 100, Y: 300},
		Size:     physics.Vec2{X: 32, Y: 48},
	}
	player.AddComponent(transform)
	
	// Player physics body
	body := physics.NewBodyDynamic(physics.AABB{
		Min: physics.Vec2{X: 100, Y: 300},
		Max: physics.Vec2{X: 132, Y: 348},
	}, 1.0)
	player.AddComponent(&physics.BodyComponent{
		Body: body,
	})
	
	// Player controller
	player.AddComponent(&PlayerController{})
	
	g.player = player
	g.AddEntity(player)
}

// PlayerController handles player input and movement
type PlayerController struct {
	physics.Velocity
	Speed     float64
	JumpForce float64
	OnGround  bool
}

// NewPlayerController creates a new player controller
func NewPlayerController() *PlayerController {
	return &PlayerController{
		Speed:     200.0,
		JumpForce: -400.0,
	}
}

// Update updates the player controller
func (pc *PlayerController) Update(entity *ecs.Entity, dt float64) {
	body := entity.GetComponent((*physics.BodyComponent)(nil)).(*physics.BodyComponent)
	
	// Input handling
	if input.IsKeyPressed("LEFT") {
		body.Velocity.X = -pc.Speed
	} else if input.IsKeyPressed("RIGHT") {
		body.Velocity.X = pc.Speed
	} else {
		body.Velocity.X = 0
	}
	
	if input.IsKeyPressed("UP") && pc.OnGround {
		body.Velocity.Y = pc.JumpForce
		pc.OnGround = false
	}
	
	// Update position
	body.Position.X += body.Velocity.X * float64(dt)
	body.Position.Y += body.Velocity.Y * float64(dt)
	
	// Simple ground check
	if body.Position.Y > 500 {
		body.Position.Y = 500
		body.Velocity.Y = 0
		pc.OnGround = true
	}
}

// Update updates the game state
func (g *Game) Update(dt float64) {
	// Update systems
	g.UpdateSystems(dt)
}

// Render renders the game
func (g *Game) Render() {
	render.Clear()
	g.RenderSystems()
	render.Present()
}

// main is the entry point
func main() {
	game := NewGame()
	
	config := &paygo.Config{
		Title:     "Go Platformer 2D",
		Width:     800,
		Height:    600,
		TargetFPS: 60,
	}
	
	paygo.Run(game, config)
}