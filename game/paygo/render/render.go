package render

import "image/color"

var clearColor color.Color

func SetClearColor(r, g, b, a float64) {
	clearColor = color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), uint8(a * 255)}
}

func Clear() {}

func Present() {}

func Config() interface{} { return nil }

func Run(game interface{}, config interface{}) { println("Game running!") }
