package main

import "math/rand/v2"

// The materials the machines are made of.
var (
	ink    = hex(0x18141e) // the line round every thing
	soot   = hex(0x1c1a1e) // what a wreck is covered in
	white  = hex(0xffffff)
	steel  = tones(0x7a8496)
	iron   = tones(0x444854)
	rubber = tones(0x2e2e3a)
	orange = tones(0xd8622a)
	amber  = tones(0xf0c040)
	cyan   = tones(0x40bcd8)
	slate  = tones(0x6c7c9c)
	red    = tones(0xe03838)
	hot    = tones(0xff7a30)
	dead   = tones(0x2c2c36) // a lens that has gone out
)

// pose is what a machine is doing in one frame of its strip. Frames 0 to 3 are a walk, 4
// and 5 an attack, and 6 to 9 its end, the last of them the wreck that stays on the floor.
type pose struct {
	step   int // 0 to 3, through the walk
	attack int // 1 winding up, 2 striking
	death  int // 1 to 4
}

func poseOf(frame int) pose {
	switch {
	case frame < 4:
		return pose{step: frame}
	case frame < 6:
		return pose{attack: frame - 3}
	}
	return pose{death: frame - 5}
}

// ending finishes a frame of a machine's end: the flash of the hit that killed it, then
// smoke and soot, more of each as it goes down, and embers in the wreck.
func (p *pic) ending(rng *rand.Rand, death int, cx, cy, size float64) {
	switch death {
	case 1:
		p.tint(white, 0.5)
		p.sparks(rng, cx, cy, size, 18)
	case 2:
		p.tint(soot, 0.3)
		p.smoke(rng, cx, cy-size*0.8, size*0.6, 3)
		p.sparks(rng, cx, cy, size*0.8, 8)
	case 3:
		p.tint(soot, 0.5)
		p.smoke(rng, cx, cy-size*0.5, size*0.7, 4)
	case 4:
		p.tint(soot, 0.66)
		p.sparks(rng, cx, cy+size*0.3, size*0.5, 4)
	}
}

// crawler is a squat floor scrubber on two treads: an orange hull, one big amber lens
// and a pair of claws under it that snap shut when it bites.
func crawler(frame int) *pic {
	po := poseOf(frame)
	rng := rand.New(rand.NewPCG(11, uint64(frame)))
	p := newPic(40, 36)

	top := 8 + po.step%2 + []int{0, 0, 3, 7, 12}[po.death] // the hull bobs as it rolls, and sinks as it dies
	alive := po.death == 0

	// The treads, seen from the front: their cleats run down as they turn.
	for _, x0 := range []int{2, 27} {
		p.hull(x0, 25, x0+11, 35, rubber, 3)
		for y := 26; y < 34; y++ {
			if (y+3-po.step%3)%3 == 0 {
				for x := x0 + 1; x < x0+10; x++ {
					p.over(x, y, rubber[3])
				}
			}
		}
	}
	p.fill(13, 27, 27, 32, iron[0]) // what is under it, in the dark

	// The claws, in front of the treads.
	tips := [2][2]float64{{12.5 + float64(po.step%2), 32}, {27.5 - float64(po.step%2), 32}}
	switch po.attack {
	case 1:
		tips = [2][2]float64{{7, 29}, {33, 29}}
	case 2:
		tips = [2][2]float64{{19, 33}, {21, 33}}
	}
	if po.death < 3 {
		p.limb(15, 26, tips[0][0], tips[0][1], 2.4, steel)
		p.limb(25, 26, tips[1][0], tips[1][1], 2.4, steel)
		p.set(int(tips[0][0]), int(tips[0][1]), steel[4])
		p.set(int(tips[1][0]), int(tips[1][1]), steel[4])
	}
	if po.attack == 2 {
		p.sparks(rng, 20, 33, 3, 7)
	}

	// The hull, with its warning lip, seams, rivets and a grille.
	p.hull(5, top, 35, 28, orange, 4)
	p.stripes(5, 25, 35, 28, amber[2], ink)
	for y := top + 3; y < 25; y++ {
		p.over(11, y, orange[1])
		p.over(28, y, orange[1])
	}
	for _, at := range [][2]int{{7, top + 2}, {31, top + 2}, {7, 22}, {31, 22}} {
		if at[1] < 24 {
			p.rivet(at[0], at[1], orange)
		}
	}
	if top < 12 {
		p.vent(7, top+6, 10, top+13, iron)
		p.vent(30, top+6, 33, top+13, iron)
	}

	// The eye.
	glass, eye := amber, float64(top)+9
	switch {
	case po.attack > 0:
		glass = hot
	case !alive:
		glass = dead
	}
	if po.death < 4 {
		p.ball(20, eye, 7.4, min(7.4, 27-eye), iron)
		p.lens(20, eye, 4.4, glass)
		if alive {
			p.glow(20, eye, 11, glass[3], 0.3)
		}
	} else {
		p.disc(20, 24, 3.6, hex(0x0c0a10)) // the socket, with the eye gone
	}

	// The aerial, with a light on its tip that blinks.
	tipX, tipY := 30.5, float64(top)-6
	if po.death >= 2 {
		tipX, tipY = 34.5, float64(top)-1 // bent over
	}
	p.limb(27, float64(top)+1, tipX, tipY, 1.5, steel)
	if alive {
		p.set(int(tipX), int(tipY)-1, []rgb{hex(0xff4040), hex(0x802020)}[po.step%2])
	}

	p.wear(rng, 24)
	p.ending(rng, po.death, 20, 18, 12)
	p.outline(ink)
	p.shadow(20, 34.6, 19, 1.5)
	return p
}

// sentry is a gun turret on three legs: a slate column, a cyan dome that looks about, a
// red lens behind a dark visor and a barrel under it.
func sentry(frame int) *pic {
	po := poseOf(frame)
	rng := rand.New(rand.NewPCG(23, uint64(frame)))
	p := newPic(36, 56)
	alive := po.death == 0

	// The legs: two in front that take turns to step, and one behind.
	lift := [2]float64{}
	if alive && po.attack == 0 && po.step%2 == 1 {
		lift[po.step/2] = 2
	}
	splay := float64([]int{0, 0, 1, 3, 5}[po.death])
	p.limb(18, 40, 18, 51, 2.6, iron) // the one behind
	p.hull(15, 50, 21, 53, iron, 1.5)
	for i, side := range []float64{-1, 1} {
		footX, footY := 18+side*(12.5+splay), 53-lift[i]
		kneeX, kneeY := 18+side*(8+splay/2), 45-lift[i]/2
		p.limb(18+side*4, 40, kneeX, kneeY, 3, steel)
		p.limb(kneeX, kneeY, footX, footY, 2.6, steel)
		p.ball(kneeX, kneeY, 2, 2, iron)
		p.hull(int(footX)-3, int(footY)-1, int(footX)+4, int(footY)+2, iron, 1.5)
	}

	// The column, shorter as it comes apart, with its hips, collars and a box of lights.
	colTop := 20 + []int{0, 0, 2, 10, 24}[po.death]
	p.hull(11, 37, 25, 43, iron, 2)
	if colTop < 38 {
		p.tube(13, colTop, 23, 38, slate, false)
		p.tube(12, 35, 24, 38, steel, true)
	}
	if colTop < 26 {
		p.tube(12, colTop, 24, colTop+3, steel, true)
		p.hull(14, 25, 22, 33, iron, 1)
		p.rivet(15, 26, steel)
		p.rivet(20, 26, steel)
		if alive {
			p.set(16, 29, []rgb{hex(0x60ff80), hex(0x206030)}[po.step%2])
			p.set(19, 29, []rgb{hex(0x806020), hex(0xffc040)}[po.step%2])
		}
		p.fill(15, 31, 21, 32, iron[0])
	}

	// The dome: it sways as the sentry walks, and falls off as it dies.
	dx := float64([]int{0, 1, 0, -1}[po.step])
	cx, cy := 18+dx, 10.0
	switch po.death {
	case 2:
		cx, cy = 21, 13
	case 3:
		cx, cy = 24, 41
	case 4:
		cx, cy = 24, 45.5
	}
	if po.death < 3 {
		p.tube(16, int(cy)+6, 20, colTop+1, iron, false)               // the neck
		p.hull(int(cx)-13, int(cy)-3, int(cx)-10, int(cy)+3, steel, 1) // its ears
		p.hull(int(cx)+10, int(cy)-3, int(cx)+13, int(cy)+3, steel, 1)
	}
	p.ball(cx, cy, 11, 8.6, cyan)
	for y := int(cy) - 2; y < int(cy)+3; y++ { // the visor
		for x := int(cx) - 10; x < int(cx)+10; x++ {
			p.over(x, y, iron[0])
		}
	}
	for x := int(cx) - 9; x < int(cx)+9; x++ {
		p.over(x, int(cy)-3, cyan[1])
	}

	// The lens looks from side to side; the barrel under it charges and then fires.
	glass := red
	if !alive {
		glass = dead
	}
	look := float64([]int{-3, 0, 3, 0}[po.step])
	if po.attack > 0 || !alive {
		look = 0
	}
	p.ball(cx+look, cy+0.5, 2.4, 2.4, glass)
	if alive {
		p.set(int(cx+look)-1, int(cy), white)
	}
	p.disc(cx, cy+5.4, 2.3, iron[1])
	p.disc(cx, cy+5.4, 1.2, hex(0x08080c))
	switch po.attack {
	case 1:
		p.disc(cx, cy+5.4, 1.2, hex(0xff90ff))
		p.glow(cx, cy+3, 9, hex(0xff70ff), 0.5)
	case 2:
		p.disc(cx, cy+5.4, 3.4, hex(0xffd0ff))
		p.disc(cx, cy+5.4, 1.8, white)
		p.sparks(rng, cx, cy+5.4, 6, 10)
	}

	p.wear(rng, 28)
	p.ending(rng, po.death, 18, 26, 14)
	p.outline(ink)
	p.shadow(18, 54.4, 16, 1.4)
	return p
}
