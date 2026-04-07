package main

import (
	"fmt"
	"os"
	"time"

	"github.com/playgo/puzzle_go/internal/audio"
	"github.com/playgo/puzzle_go/internal/game"
	"github.com/playgo/puzzle_go/internal/render"
)

func main() {
	fmt.Println("1. Init subsystems...")
	spr := render.LoadSprites()
	snd := audio.NewManager()
	
	fmt.Println("2. Create game...")
	g := game.NewGame(spr, snd)
	fmt.Println("3. Game created")
	
	// Try calling Update manually (without Ebiten)
	fmt.Println("4. Testing Update calls...")
	for i := 0; i < 10; i++ {
		err := g.Update()
		if err != nil {
			fmt.Printf("Update[%d] returned error: %v\n", i, err)
		} else {
			fmt.Printf("Update[%d] OK\n", i)
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	fmt.Println("5. All updates passed!")
	fmt.Println("Press Enter to exit...")
	fmt.Scanln()
	os.Exit(0)
}
