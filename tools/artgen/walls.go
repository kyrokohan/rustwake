package main

import "math/rand/v2"

// The surfaces of the three zones. A wall is 16 by 16 and a floor or a ceiling 8 by 8,
// one pixel to a face of the level's mesh. Each is drawn to sit beside copies of itself,
// and none has a hole.
var (
	teal  = tones(0x1f4c52)
	rust  = tones(0xa85a30)
	brick = tones(0x8a3e2c)
)

// plate fills a box as a flat panel: one tone, lit along the top and the left, dark along
// the bottom and the right. Sunk turns it into a panel set into the wall.
func (p *pic) plate(x0, y0, x1, y1 int, r ramp, sunk bool) {
	up, down := r[3], r[1]
	if sunk {
		up, down = r[0], r[3]
	}
	p.fill(x0, y0, x1, y1, r[2])
	p.fill(x0, y1-1, x1, y1, down)
	p.fill(x1-1, y0, x1, y1, down)
	p.fill(x0, y0, x1, y0+1, up)
	p.fill(x0, y0, x0+1, y1, up)
}

// corners puts a rivet in each corner of a 16 by 16 panel.
func (p *pic) corners(r ramp) {
	for _, at := range [][2]int{{2, 2}, {12, 2}, {2, 12}, {12, 12}} {
		p.rivet(at[0], at[1], r)
	}
}

// Zone 1, the intake: blue steel.

func wall1a(int) *pic { // a panel of plate, with a seam across it
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, slate, false)
	p.fill(1, 8, 15, 9, slate[1])
	p.fill(1, 9, 15, 10, slate[3])
	p.corners(slate)
	p.wear(rand.New(rand.NewPCG(101, 0)), 26)
	return p
}

func wall1b(int) *pic { // pipes, with collars
	p := newPic(16, 16)
	p.fill(0, 0, 16, 16, iron[0])
	p.tube(1, 0, 5, 16, steel, false)
	p.tube(6, 0, 11, 16, tones(0x4c78a0), false)
	p.tube(12, 0, 15, 16, steel, false)
	p.tube(0, 10, 6, 13, iron, true)
	p.tube(5, 3, 12, 6, iron, true)
	p.tube(11, 10, 16, 13, iron, true)
	p.wear(rand.New(rand.NewPCG(102, 0)), 20)
	return p
}

func wall1c(int) *pic { // a grille, over a warning mark
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, slate, false)
	p.plate(2, 2, 14, 10, iron, true)
	p.vent(3, 3, 13, 9, iron)
	p.fill(3, 12, 13, 14, amber[2])
	p.stripes(3, 12, 13, 14, amber[2], ink)
	p.wear(rand.New(rand.NewPCG(103, 0)), 16)
	return p
}

func floor1(int) *pic { // tread plate
	p := newPic(8, 8)
	p.plate(0, 0, 8, 8, tones(0x4a4e5a), false)
	for _, at := range [][2]int{{2, 2}, {5, 3}, {3, 5}, {6, 6}} {
		p.set(at[0], at[1], iron[3])
	}
	return p
}

func ceiling1(int) *pic { // dark panels
	p := newPic(8, 8)
	p.plate(0, 0, 8, 8, tones(0x343844), true)
	return p
}

// Zone 2, the foundry: rust and brick, with fire behind it.

func wall2a(int) *pic { // rusted plate, stained under its rivets
	rng := rand.New(rand.NewPCG(201, 0))
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, rust, false)
	for range 7 {
		x, y := rng.IntN(13)+1, rng.IntN(13)+1
		p.disc(float64(x)+1, float64(y)+1, 0.8+rng.Float64()*1.6, []rgb{rust[1], rust[3], tones(0x7a3c22)[2]}[rng.IntN(3)])
	}
	p.corners(rust)
	for _, x := range []int{2, 12} {
		p.fill(x, 4, x+1, 8, rust[0])
	}
	p.wear(rng, 30)
	return p
}

func wall2b(int) *pic { // brick, laid in courses
	p := newPic(16, 16)
	p.fill(0, 0, 16, 16, hex(0x2a1c18))
	for row := range 4 {
		for x := -8 + row%2*4; x < 16; x += 8 {
			p.plate(x, row*4, x+7, row*4+3, brick, false)
		}
	}
	p.wear(rand.New(rand.NewPCG(202, 0)), 34)
	return p
}

func wall2c(int) *pic { // a furnace door, with fire behind its bars
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, rust, false)
	p.plate(2, 4, 14, 14, iron, true)
	p.fill(3, 5, 13, 13, hex(0x200c06))
	for y := 6; y < 13; y += 2 {
		p.fill(3, y, 13, y+1, mix(hex(0xff8a20), hex(0xffe060), float64(y-6)/8))
	}
	p.fill(7, 5, 9, 13, iron[1]) // the bar down the middle
	p.rivet(2, 1, rust)
	p.rivet(12, 1, rust)
	return p
}

func floor2(int) *pic { // grating
	p := newPic(8, 8)
	p.fill(0, 0, 8, 8, hex(0x1a1412))
	for i := 0; i < 8; i += 4 {
		p.fill(i, 0, i+1, 8, hex(0x4c3c34))
		p.fill(0, i, 8, i+1, hex(0x5a4840))
	}
	return p
}

func ceiling2(int) *pic { // soot, and a duct
	p := newPic(8, 8)
	p.fill(0, 0, 8, 8, hex(0x1c1614))
	p.fill(0, 3, 8, 5, hex(0x2e2622))
	p.fill(0, 5, 8, 6, hex(0x120e0c))
	return p
}

// Zone 3, the core: dark teal, wired and lit from inside.

func wall3a(int) *pic { // a panel with a trace of light across it
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, teal, false)
	p.plate(3, 3, 13, 13, tones(0x183c42), true)
	p.fill(4, 8, 11, 9, hex(0x30d8c8))
	p.fill(10, 8, 11, 12, hex(0x30d8c8))
	p.set(4, 8, hex(0xc0fff8))
	p.set(10, 11, hex(0xc0fff8))
	return p
}

func wall3b(int) *pic { // ribs
	p := newPic(16, 16)
	p.fill(0, 0, 16, 16, teal[0])
	for x := 0; x < 16; x += 4 {
		p.tube(x, 0, x+3, 16, tones(0x2c6068), false)
	}
	p.fill(0, 7, 16, 8, teal[0])
	return p
}

func wall3c(int) *pic { // a screen, and the lamps under it
	rng := rand.New(rand.NewPCG(303, 0))
	p := newPic(16, 16)
	p.plate(0, 0, 16, 16, teal, false)
	p.plate(1, 2, 15, 11, iron, true)
	p.fill(2, 3, 14, 10, hex(0x062018))
	for y := 3; y < 10; y += 2 {
		p.fill(2, y, 14, y+1, hex(0x0a3222))
		for x := 3; x < 13; x++ {
			if rng.IntN(5) < 2 {
				p.set(x, y, hex(0x58ff98))
			}
		}
	}
	for i, c := range []uint32{0xff5040, 0xffc040, 0x50ff80} {
		p.set(4+i*4, 13, hex(c))
	}
	return p
}

func floor3(int) *pic { // tiles
	p := newPic(8, 8)
	for y := range 8 {
		for x := range 8 {
			p.set(x, y, []rgb{hex(0x14292d), hex(0x1a353a)}[(x/4+y/4)%2])
		}
	}
	p.set(0, 0, hex(0x2c8c88))
	p.set(4, 4, hex(0x2c8c88))
	return p
}

func ceiling3(int) *pic { // dark, with small lights
	p := newPic(8, 8)
	p.fill(0, 0, 8, 8, hex(0x0e1c20))
	p.fill(3, 3, 5, 5, hex(0x2c6068))
	return p
}

// floorExit is the pad under the way out, in any zone: lit green round its edge.
func floorExit(int) *pic {
	p := newPic(8, 8)
	p.plate(0, 0, 8, 8, tones(0x38c070), false)
	p.fill(2, 2, 6, 6, hex(0x145030))
	p.fill(3, 3, 5, 5, hex(0xa0ffb0))
	return p
}

// tiled returns a picture laid out twice across and twice down, to look at how it meets
// itself.
func tiled(p *pic) *pic {
	out := newPic(2*p.w, 2*p.h)
	for i := range 4 {
		out.paste(p, i%2*p.w, i/2*p.h)
	}
	return out
}
