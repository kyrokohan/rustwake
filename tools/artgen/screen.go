package main

import (
	"math"
	"math/rand/v2"
)

// letters are the game's own letters, five pixels by seven: plain blocks, the capitals,
// the digits and the few marks the menus need. The logo is cut from them, and the game
// writes its menus with them from the strip font draws.
var letters = map[rune][7]string{
	'A': {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'B': {"####.", "#...#", "#...#", "####.", "#...#", "#...#", "####."},
	'C': {".####", "#....", "#....", "#....", "#....", "#....", ".####"},
	'D': {"####.", "#...#", "#...#", "#...#", "#...#", "#...#", "####."},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
	'F': {"#####", "#....", "#....", "####.", "#....", "#....", "#...."},
	'G': {".####", "#....", "#....", "#.###", "#...#", "#...#", ".###."},
	'H': {"#...#", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'I': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "#####"},
	'J': {"..###", "....#", "....#", "....#", "....#", "#...#", ".###."},
	'K': {"#...#", "#..#.", "#.#..", "##...", "#.#..", "#..#.", "#...#"},
	'L': {"#....", "#....", "#....", "#....", "#....", "#....", "#####"},
	'M': {"#...#", "##.##", "#.#.#", "#.#.#", "#...#", "#...#", "#...#"},
	'N': {"#...#", "##..#", "#.#.#", "#.#.#", "#..##", "#...#", "#...#"},
	'O': {".###.", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'P': {"####.", "#...#", "#...#", "####.", "#....", "#....", "#...."},
	'Q': {".###.", "#...#", "#...#", "#...#", "#.#.#", "#..#.", ".##.#"},
	'R': {"####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"},
	'S': {".####", "#....", "#....", ".###.", "....#", "....#", "####."},
	'T': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
	'U': {"#...#", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'V': {"#...#", "#...#", "#...#", "#...#", "#...#", ".#.#.", "..#.."},
	'W': {"#...#", "#...#", "#...#", "#.#.#", "#.#.#", "##.##", "#...#"},
	'X': {"#...#", "#...#", ".#.#.", "..#..", ".#.#.", "#...#", "#...#"},
	'Y': {"#...#", "#...#", ".#.#.", "..#..", "..#..", "..#..", "..#.."},
	'Z': {"#####", "....#", "...#.", "..#..", ".#...", "#....", "#####"},
	'0': {".###.", "#...#", "#..##", "#.#.#", "##..#", "#...#", ".###."},
	'1': {"..#..", ".##..", "..#..", "..#..", "..#..", "..#..", ".###."},
	'2': {".###.", "#...#", "....#", "...#.", "..#..", ".#...", "#####"},
	'3': {"####.", "....#", "....#", ".###.", "....#", "....#", "####."},
	'4': {"#...#", "#...#", "#...#", "#####", "....#", "....#", "....#"},
	'5': {"#####", "#....", "#....", "####.", "....#", "....#", "####."},
	'6': {".###.", "#....", "#....", "####.", "#...#", "#...#", ".###."},
	'7': {"#####", "....#", "...#.", "..#..", "..#..", "..#..", "..#.."},
	'8': {".###.", "#...#", "#...#", ".###.", "#...#", "#...#", ".###."},
	'9': {".###.", "#...#", "#...#", ".####", "....#", "....#", ".###."},
	'/': {"....#", "....#", "...#.", "..#..", ".#...", "#....", "#...."},
	':': {".....", "..#..", "..#..", ".....", "..#..", "..#..", "....."},
	'-': {".....", ".....", ".....", "#####", ".....", ".....", "....."},
	'.': {".....", ".....", ".....", ".....", ".....", ".##..", ".##.."},
	'!': {"..#..", "..#..", "..#..", "..#..", "..#..", ".....", "..#.."},
}

// fontFrames is how many letters the font strip holds: every character from the space to
// Z, in order, a blank for each one that has no letter.
const fontFrames = 'Z' - ' ' + 1

// font is one letter of the strip, in white, for the game to draw in any color.
func font(frame int) *pic {
	p := newPic(5, 7)
	for y, row := range letters[rune(' '+frame)] {
		for x, on := range row {
			if on == '#' {
				p.set(x, y, white)
			}
		}
	}
	return p
}

// cursor is the bolt head that marks the chosen item of a menu, and glints.
func cursor(frame int) *pic {
	p := newPic(7, 7)
	p.ball(3.5, 3.5, 3.4, 3.4, amber)
	p.fill(2, 3, 5, 4, ink) // its slot
	if frame == 1 {
		p.set(2, 1, white)
		p.set(1, 2, white)
	}
	return p
}

// splat is blood thrown on the glass: a few pools that run together, spikes out from them
// that each end in a drop, and loose drops round about. There are four, all different.
func splat(frame int) *pic {
	rng := rand.New(rand.NewPCG(91, uint64(frame)))
	p := newPic(24, 24)
	blood := tones(0xa80c14)
	for range 7 {
		a, d := rng.Float64()*2*math.Pi, 6+rng.Float64()*5
		x, y := 12+math.Cos(a)*d, 12+math.Sin(a)*d
		p.limb(12, 12, x, y, 1.7, blood)
		p.disc(x, y, 0.8+rng.Float64(), blood[2])
	}
	for range 5 {
		p.disc(12+(rng.Float64()-0.5)*6, 12+(rng.Float64()-0.5)*6, 3+rng.Float64()*2.4, blood[2])
	}
	for range 6 {
		a, d := rng.Float64()*2*math.Pi, 8+rng.Float64()*3.5
		p.disc(12+math.Cos(a)*d, 12+math.Sin(a)*d, 0.6+rng.Float64()*0.5, blood[1])
	}
	p.disc(12.5, 12.5, 2.6, blood[1]) // deep in the middle
	p.set(9, 9, blood[4])             // and wet where the light catches it
	p.set(10, 8, blood[3])
	return p
}

// logo is the game's name for the title screen: letters cut from rusted plate, three
// pixels to a pixel of the letter, lit from above, over a warning stripe.
func logo(int) *pic {
	const name, size = "RUSTWAKE", 3
	rng := rand.New(rand.NewPCG(7, 0))
	p := newPic(len(name)*6*size+1, 7*size+8)
	for i, r := range name {
		for y, row := range letters[r] {
			for x, on := range row {
				if on == '#' {
					x0, y0 := 2+(i*6+x)*size, 1+y*size
					p.fill(x0, y0, x0+size, y0+size, rust.band(0.78-0.5*float64(y)/7))
				}
			}
		}
	}
	// The edge of every letter that faces up is bright, and the edge that faces down dark.
	for y := range p.h {
		for x := range p.w {
			switch {
			case p.at(x, y).A == 0:
			case p.at(x, y-1).A == 0:
				p.set(x, y, rust[4])
			case p.at(x, y+1).A == 0:
				p.set(x, y, rust[0])
			}
		}
	}
	p.wear(rng, 220)
	p.fill(2, 7*size+3, p.w-2, 7*size+6, amber[2])
	p.stripes(2, 7*size+3, p.w-2, 7*size+6, amber[2], ink)
	p.outline(ink)
	return p
}

// iconHealth and iconAmmo mark the two numbers on the status strip.
func iconHealth(int) *pic {
	p := newPic(5, 5)
	p.fill(1, 0, 4, 5, green[2])
	p.fill(0, 1, 5, 4, green[2])
	p.fill(2, 1, 3, 4, white)
	p.fill(1, 2, 4, 3, white)
	return p
}

func iconAmmo(int) *pic {
	p := newPic(5, 5)
	for _, x := range []int{0, 2, 4} {
		p.fill(x, 0, x+1, 3, steel[3])
		p.set(x, 0, white)
	}
	p.fill(0, 3, 5, 5, amber[2])
	return p
}
