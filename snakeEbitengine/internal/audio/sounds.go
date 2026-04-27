package audio

func SndEat() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.1, 600, 0.5, "sine", 400)
	t2 := SynthWave(sr, 0.1, 1200, 0.3, "sine", 200)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndBoom() []byte {
	sr := 44100
	low := SynthWave(sr, 0.3, 80, 0.9, "sine", -30)
	n := SynthWave(sr, 0.3, 0, 0.7, "noise", 0)
	mid := SynthWave(sr, 0.15, 300, 0.4, "square", -200)
	return MixToWAV(sr, [][]int16{low, n, mid})
}

func SndHeal() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.15, 400, 0.4, "sine", 300)
	t2 := SynthWave(sr, 0.2, 800, 0.3, "sine", 100)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndPause() []byte {
	sr := 44100
	t := SynthWave(sr, 0.08, 220, 0.4, "square", -50)
	return MixToWAV(sr, [][]int16{t})
}

func SndMenuMove() []byte {
	sr := 44100
	t := SynthWave(sr, 0.05, 800, 0.3, "sine", 100)
	return MixToWAV(sr, [][]int16{t})
}

func SndMenuSelect() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.1, 400, 0.5, "sine", 200)
	t2 := SynthWave(sr, 0.1, 700, 0.5, "sine", -100)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndGhost() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.2, 500, 0.4, "sine", -200)
	t2 := SynthWave(sr, 0.2, 800, 0.3, "sine", 100)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndKey() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.15, 880, 0.5, "sine", -400)
	t2 := SynthWave(sr, 0.15, 440, 0.4, "sine", -200)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndKeyUse() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.1, 600, 0.5, "sine", -200)
	t2 := SynthWave(sr, 0.1, 800, 0.4, "sine", 100)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndGiftOpen() []byte {
	sr := 44100
	t1 := SynthWave(sr, 0.2, 300, 0.6, "sine", 400)
	t2 := SynthWave(sr, 0.2, 600, 0.5, "sine", -300)
	return MixToWAV(sr, [][]int16{t1, t2})
}

func SndCoin() []byte {
	sr := 44100
	t := SynthWave(sr, 0.1, 1000, 0.4, "sine", -600)
	return MixToWAV(sr, [][]int16{t})
}