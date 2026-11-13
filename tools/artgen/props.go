package main

import "math/rand/v2"

var (
	green = tones(0x48c868)
	lime  = hex(0xa0ffb0)
	olive = tones(0x8a7c52)
	drum  = tones(0x2f7f86)
	pink  = tones(0xf070f0)
)

// weapon is the rivet driver as the player holds it, seen from behind: a barrel wound with
// a heating coil on a body that widens toward the eye. Frame 0 is at rest; 1 to 3 are a
// shot, with the flash at the muzzle, the coil white hot and the kick back.
func weapon(frame int) *pic {
	rng := rand.New(rand.NewPCG(31, uint64(frame)))
	g := newPic(48, 36)

	// The body, in two steps, with a cable off to the right.
	g.limb(35, 27, 47, 35, 3, rubber)
	g.hull(8, 22, 40, 36, iron, 4)
	g.hull(13, 14, 35, 26, slate, 3)
	g.vent(15, 28, 20, 34, iron)
	g.vent(28, 28, 33, 34, iron)
	for _, at := range [][2]int{{10, 24}, {37, 24}, {15, 16}, {32, 16}} {
		g.rivet(at[0], at[1], steel)
	}
	g.lens(24, 30, 2.6, []ramp{amber, hot, hot, amber}[frame]) // the gauge

	// The barrel and its coil.
	g.tube(20, 2, 28, 16, steel, false)
	g.hull(18, 0, 30, 4, iron, 1.5)
	coil := []ramp{orange, tones(0xfff4d0), hot, hot}[frame]
	for _, y := range []int{5, 9, 13} {
		g.tube(18, y, 30, y+3, coil, true)
	}
	g.wear(rng, 30)
	g.outline(ink)
	if frame > 0 {
		g.glow(24, 8, 14, hex(0xffe0a0), []float64{0, 0.55, 0.35, 0.15}[frame])
	}

	p := newPic(48, 46)
	p.paste(g, 0, 10+[]int{0, 3, 4, 1}[frame])
	switch frame {
	case 1: // the flash
		p.disc(24, 9, 8, hex(0xffb030))
		p.disc(24, 9, 5.5, hex(0xfff0a0))
		p.disc(24, 9, 3, white)
		for _, d := range [][2]float64{{-11, -4}, {11, -4}, {-7, -9}, {7, -9}, {0, -9}} {
			p.limb(24, 9, 24+d[0], 9+d[1], 1.6, tones(0xfff0a0))
		}
	case 2:
		p.disc(24, 11, 5, hex(0xff9028))
		p.disc(24, 11, 2.6, hex(0xffe890))
		p.sparks(rng, 24, 8, 9, 8)
	case 3:
		p.smoke(rng, 24, 7, 4, 3)
	}
	return p
}

// health is a repair cell: a green canister between two steel caps, marked with a cross,
// that pulses.
func health(frame int) *pic {
	p := newPic(16, 16)
	p.tube(3, 3, 13, 13, green, false)
	p.hull(4, 1, 12, 4, steel, 1)
	p.hull(4, 12, 12, 15, steel, 1)
	p.fill(7, 5, 9, 11, white)
	p.fill(5, 7, 11, 9, white)
	if frame == 1 {
		p.glow(8, 8, 9, lime, 0.45)
	}
	p.outline(ink)
	return p
}

// ammo is a pack of charges: three steel pins standing in an amber box.
func ammo(frame int) *pic {
	p := newPic(16, 16)
	for _, x := range []int{3, 7, 11} {
		p.tube(x, 1, x+3, 8, steel, false)
		if frame == 1 {
			p.set(x+1, 1, white)
		}
	}
	p.hull(1, 6, 15, 15, amber, 2)
	p.stripes(2, 11, 14, 13, ink, amber[3])
	p.rivet(3, 8, amber)
	p.rivet(12, 8, amber)
	p.outline(ink)
	return p
}

// bolt is a sentry's shot: a white-hot point in a pink ball, with a ring that flickers.
func bolt(frame int) *pic {
	p := newPic(12, 12)
	if frame == 1 {
		p.disc(6, 6, 5.6, pink[1])
	}
	p.ball(6, 6, 4.4, 4.4, pink)
	p.disc(6, 6, 2.2, hex(0xffe0ff))
	p.disc(6, 6, 1.1, white)
	return p
}

// exit is the beacon that stands on the way out: a green lamp on a steel post.
func exit(frame int) *pic {
	p := newPic(14, 44)
	p.hull(2, 38, 12, 43, iron, 1.5)
	p.tube(6, 12, 8, 39, steel, false)
	p.tube(4, 22, 10, 24, iron, true)
	p.hull(3, 10, 11, 13, iron, 1)
	p.ball(7, 6, 5, 5, green)
	p.set(5, 4, white)
	p.outline(ink)
	if frame == 1 {
		p.glow(7, 6, 9, lime, 0.6)
		for _, d := range [][2]int{{0, -6}, {-6, 0}, {6, 0}, {-5, -5}, {5, -5}} {
			p.set(7+d[0], 6+d[1], lime)
		}
	}
	p.shadow(7, 42.6, 6.5, 1.2)
	return p
}

// barrel is a drum of coolant, to stand behind: teal, hooped with steel, with a label.
func barrel(int) *pic {
	rng := rand.New(rand.NewPCG(41, 0))
	p := newPic(22, 30)
	p.tube(2, 3, 20, 28, drum, false)
	for _, y := range []int{3, 12, 21, 26} {
		p.tube(1, y, 21, y+2, steel, true)
	}
	p.ball(11, 3.4, 9, 2.6, tones(0x3a98a0)) // the lid
	p.disc(7, 3, 1, ink)                     // and its bung
	p.fill(7, 15, 15, 20, amber[2])
	p.stripes(7, 15, 15, 20, amber[2], ink)
	p.wear(rng, 40)
	p.outline(ink)
	p.shadow(11, 28.6, 10.5, 1.3)
	return p
}

// crate is a shipping crate, to stand behind: olive boards in a steel frame with a brace.
func crate(int) *pic {
	rng := rand.New(rand.NewPCG(43, 0))
	p := newPic(30, 30)
	p.hull(1, 2, 29, 28, olive, 1.5)
	for y := 8; y < 26; y += 6 { // the gaps between the boards
		for x := 4; x < 26; x++ {
			p.over(x, y, olive[0])
		}
	}
	p.limb(5, 24, 25, 6, 2.4, tones(0x74683f)) // the brace
	for _, b := range [][4]int{{1, 2, 29, 5}, {1, 25, 29, 28}, {1, 2, 4, 28}, {26, 2, 29, 28}} {
		p.hull(b[0], b[1], b[2], b[3], iron, 1)
	}
	for _, at := range [][2]int{{2, 3}, {27, 3}, {2, 26}, {27, 26}} {
		p.rivet(at[0], at[1], steel)
	}
	p.wear(rng, 50)
	p.outline(ink)
	p.shadow(15, 28.6, 14.5, 1.3)
	return p
}

// lamp hangs from the ceiling: an iron shade over an amber bulb.
func lamp(int) *pic {
	p := newPic(18, 14)
	p.fill(8, 0, 10, 4, iron[1]) // the rod it hangs by
	p.ball(9, 9.5, 3.6, 3.6, tones(0xffd870))
	p.hull(2, 3, 16, 9, iron, 3)
	p.fill(3, 8, 15, 9, hex(0xffe9a0)) // the lit rim of the shade
	p.outline(ink)
	p.disc(9, 10, 2, hex(0xfff6d0))
	return p
}
