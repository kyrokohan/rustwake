package main

import "math/rand/v2"

// letters are the eight letters of the game's name, five pixels by seven.
var letters = map[rune][7]string{
	'R': {"####.", "#...#", "#...#", "####.", "#.#..", "#..#.", "#...#"},
	'U': {"#...#", "#...#", "#...#", "#...#", "#...#", "#...#", ".###."},
	'S': {".####", "#....", "#....", ".###.", "....#", "....#", "####."},
	'T': {"#####", "..#..", "..#..", "..#..", "..#..", "..#..", "..#.."},
	'W': {"#...#", "#...#", "#...#", "#.#.#", "#.#.#", "##.##", "#...#"},
	'A': {".###.", "#...#", "#...#", "#####", "#...#", "#...#", "#...#"},
	'K': {"#...#", "#..#.", "#.#..", "##...", "#.#..", "#..#.", "#...#"},
	'E': {"#####", "#....", "#....", "####.", "#....", "#....", "#####"},
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
