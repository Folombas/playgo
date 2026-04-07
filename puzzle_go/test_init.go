package main

import (
	"fmt"
	"os"

	"github.com/playgo/puzzle_go/internal/audio"
	"github.com/playgo/puzzle_go/internal/render"
)

func main() {
	fmt.Println("Step 1: Starting...")
	
	fmt.Println("Step 2: Loading sprites...")
	spr := render.LoadSprites()
	fmt.Printf("Step 3: Sprites loaded. BGTile=%v, HUDBG=%v\n", spr.BGTile != nil, spr.HUDBG != nil)
	
	fmt.Println("Step 4: Initializing audio...")
	snd := audio.NewManager()
	fmt.Printf("Step 5: Audio initialized. Match=%v\n", snd.Match != nil)
	
	fmt.Println("ALL INIT STEPS PASSED!")
	fmt.Scanln()
	os.Exit(0)
}
