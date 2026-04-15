package main

import "image/color"

// RGBA создаёт цвет из uint8 компонентов
func RGBA(r, g, b, a uint8) color.Color {
	return color.RGBA{r, g, b, a}
}

// Lerp interpolирует два цвета
func ColorLerp(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R) + (float64(b.R)-float64(a.R))*t),
		G: uint8(float64(a.G) + (float64(b.G)-float64(a.G))*t),
		B: uint8(float64(a.B) + (float64(b.B)-float64(a.B))*t),
		A: uint8(float64(a.A) + (float64(b.A)-float64(a.A))*t),
	}
}
