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

var (
	leaf  = tones(0x7cc83c)
	ochre = tones(0xd8a424)
	navy  = tones(0x3a4a6c)
	ice   = tones(0xbfe6ff)
	glass = tones(0x24384c)
)

// drone is a small flier: a green pod between two rotors, one pale eye, and a saw slung
// under it that it drops onto what it catches.
func drone(frame int) *pic {
	po := poseOf(frame)
	rng := rand.New(rand.NewPCG(37, uint64(frame)))
	p := newPic(32, 30)
	alive := po.death == 0
	fall := float64([]int{0, 0, 3, 9, 14}[po.death])
	cy := 12 + fall

	// The saw, turning, and lower when it strikes.
	sawY := cy + 9 + float64([]int{0, 1, 3}[po.attack])
	if po.death < 4 {
		p.limb(16, cy+4, 16, sawY, 1.6, iron)
		p.disc(16, sawY, 4.2, steel[2])
		for i := range 8 {
			if (i+po.step)%2 == 0 {
				dx, dy := []float64{0, 3.6, 5, 3.6, 0, -3.6, -5, -3.6}[i], []float64{-5, -3.6, 0, 3.6, 5, 3.6, 0, -3.6}[i]
				p.set(int(16+dx), int(sawY+dy), steel[4])
			}
		}
		p.disc(16, sawY, 1.4, iron[1])
	} else {
		p.ball(21, 27, 5, 1.6, steel) // the saw, flat on the floor
	}
	if po.attack == 2 {
		p.sparks(rng, 16, sawY+4, 4, 9)
	}

	// The arms and the rotors: the blades show long and short by turns, which reads as a spin.
	for i, side := range []float64{-1, 1} {
		if po.death >= 2 && i == 1 {
			continue // one arm is gone
		}
		hubX, hubY := 16+side*11, cy-4
		p.limb(16+side*5, cy-1, hubX, hubY, 2.2, iron)
		p.ball(hubX, hubY, 1.8, 1.8, steel)
		span := []float64{4.6, 1.8}[(po.step+i)%2]
		if alive {
			p.limb(hubX-span, hubY-2, hubX+span, hubY-2, 1.2, tones(0xd8dce8))
		}
	}

	// The pod, with a fin on top and the eye in front.
	p.fill(15, int(cy)-8, 17, int(cy)-4, steel[2])
	p.ball(16, cy, 6.6, 5.6, leaf)
	for x := 11; x < 22; x++ {
		p.over(x, int(cy)+3, leaf[0])
	}
	eye := ice
	switch {
	case po.attack > 0:
		eye = red
	case !alive:
		eye = dead
	}
	p.lens(16, cy, 2.3, eye)
	if alive {
		p.glow(16, cy, 6, eye[3], 0.3)
	}

	p.wear(rng, 14)
	p.ending(rng, po.death, 16, cy, 9)
	p.outline(ink)
	if po.death == 4 {
		p.shadow(16, 28.6, 12, 1.2)
	}
	return p
}

// hauler is a loader on two legs: a wide yellow cab with a grille, two lamps for eyes
// behind its window, and a pair of clamps that it lifts and brings down.
func hauler(frame int) *pic {
	po := poseOf(frame)
	rng := rand.New(rand.NewPCG(53, uint64(frame)))
	p := newPic(56, 60)
	alive := po.death == 0
	sink := []int{0, 0, 3, 10, 24}[po.death]
	b := sink
	if alive && po.attack == 0 {
		b += po.step % 2
	}
	fb := float64(b)

	// The legs: each lifts its foot in turn. They fold flat under the wreck.
	if po.death < 4 {
		for i, side := range []float64{-1, 1} {
			lift := 0.0
			if alive && po.attack == 0 && po.step == 1+2*i {
				lift = 3
			}
			hipX, kneeX, footX := 28+side*8, 28+side*(11+float64(sink)/3), 28+side*12
			p.limb(hipX, 38+fb/2, kneeX, 46-lift, 6, ochre)
			p.limb(kneeX, 46-lift, footX, 55-lift, 5.4, ochre)
			p.ball(kneeX, 46-lift, 3.2, 3.2, iron)
			p.hull(int(footX)-8, 54-int(lift), int(footX)+8, 59-int(lift), iron, 2)
		}
		p.hull(17, 34+b/2, 39, 41+b/2, iron, 2)
	} else {
		p.hull(6, 54, 24, 59, iron, 2)
		p.hull(34, 54, 52, 59, iron, 2)
	}

	// The cab: a warning band, a window with the eyes behind it, and a grille.
	p.hull(10, 8+b, 46, 36+b, ochre, 5)
	p.stripes(10, 9+b, 46, 13+b, amber[2], ink)
	p.plate(18, 15+b, 38, 22+b, glass, true)
	lamps := white
	switch {
	case po.attack > 0:
		lamps = hex(0xff4030)
	case !alive:
		lamps = hex(0x30303a)
	}
	p.disc(23.5, 18.5+fb, 1.7, lamps)
	p.disc(32.5, 18.5+fb, 1.7, lamps)
	p.plate(20, 25+b, 36, 33+b, iron, true)
	p.vent(21, 26+b, 35, 32+b, iron)
	for _, at := range [][2]int{{12, 15}, {42, 15}, {12, 32}, {42, 32}} {
		p.rivet(at[0], at[1]+b, ochre)
	}

	// The stack, which puffs, and a lamp that blinks.
	p.tube(41, 1+b, 45, 9+b, iron, false)
	if alive && po.step%2 == 0 {
		p.ball(43, fb-1, 2.2, 1.8, tones(0x5a5860))
	}
	if alive {
		p.ball(14, 6+fb, 2.2, 2.2, []ramp{amber, tones(0x7a5c10)}[po.step%2])
	}

	// The arms: hanging, lifted, or brought down in front.
	for _, side := range []float64{-1, 1} {
		shX, shY := 28+side*21, 14+fb
		elX, elY, clX, clY, size := 28+side*24, 26+fb, 28+side*22, 37+fb, 4.0
		switch po.attack {
		case 1:
			elX, elY, clX, clY = 28+side*26, 6+fb, 28+side*22, -1+fb+4
		case 2:
			elX, elY, clX, clY, size = 28+side*19, 30+fb, 28+side*11, 45+fb, 5.5
		}
		if po.death >= 3 {
			elX, elY, clX, clY = 28+side*26, min(54, 24+fb), 28+side*25, 56
		}
		p.ball(shX, shY, 4.6, 4.6, iron)
		p.limb(shX, shY, elX, elY, 5, ochre)
		p.limb(elX, elY, clX, clY, 4.4, ochre)
		p.hull(int(clX-size), int(clY-size/2), int(clX+size), int(clY+size), steel, 1.5)
		p.fill(int(clX)-1, int(clY), int(clX)+1, int(clY+size), ink) // the gap between its jaws
	}
	if po.attack == 2 {
		p.sparks(rng, 28, 52, 9, 16)
	}

	p.wear(rng, 50)
	p.ending(rng, po.death, 28, 24+fb, 18)
	p.outline(ink)
	p.shadow(28, 58.5, 24, 1.5)
	return p
}

// overseer is what runs the plant: an armored blue-gray dome on a tracked base, a visor
// with three lenses that a light runs along, and a gun on each shoulder.
func overseer(frame int) *pic {
	po := poseOf(frame)
	rng := rand.New(rand.NewPCG(71, uint64(frame)))
	p := newPic(72, 70)
	alive := po.death == 0
	sink := float64([]int{0, 0, 3, 9, 20}[po.death])

	// The base: a tread across the whole front, under a skirt of armor.
	p.hull(3, 52, 69, 68, rubber, 4)
	for y := 54; y < 67; y++ {
		if (y+3-po.step%3)%3 == 0 {
			for x := 5; x < 67; x++ {
				p.over(x, y, rubber[3])
			}
		}
	}

	// The dome, lower as it dies, with its seams, rivets and grilles.
	cy, ry := 34+sink, 22-sink*0.55
	p.ball(36, cy, 27, ry, navy)
	for _, dy := range []float64{-0.45, 0.35} {
		y := int(cy + dy*ry)
		for x := 9; x < 63; x++ {
			p.over(x, y, navy[0])
			if x%6 == 0 {
				p.over(x, y+2, navy[4])
			}
		}
	}
	if po.death < 3 {
		p.vent(17, int(cy)+9, 26, int(cy)+14, iron)
		p.vent(46, int(cy)+9, 55, int(cy)+14, iron)
	}
	p.hull(7, 45, 65, 56, iron, 3)
	p.stripes(9, 51, 63, 54, amber[2], ink)
	for x := 11; x < 62; x += 7 {
		p.rivet(x, 47, steel)
	}

	// The mast, with two lights that blink in turn.
	if po.death < 3 {
		top := cy - ry
		p.limb(36, top+2, 36, top-8, 2, steel)
		p.limb(29, top-6, 43, top-6, 1.6, steel)
		if alive {
			p.set(29, int(top)-8, []rgb{hex(0xff4040), hex(0x702020)}[po.step%2])
			p.set(42, int(top)-8, []rgb{hex(0x702020), hex(0xff4040)}[po.step%2])
		}
	}

	// The guns: a box on each shoulder, its muzzle toward the player.
	for _, x := range []float64{9, 63} {
		gy := cy - 4 + []float64{0, 0, 2, 7, 12}[po.death]
		p.hull(int(x)-7, int(gy)-7, int(x)+7, int(gy)+7, iron, 3)
		p.disc(x, gy, 4, steel[1])
		p.disc(x, gy, 2.4, hex(0x08080c))
		switch po.attack {
		case 1:
			p.disc(x, gy, 2.4, hex(0xff90ff))
			p.glow(x, gy, 8, hex(0xff70ff), 0.5)
		case 2:
			p.disc(x, gy, 5.6, hex(0xffd0ff))
			p.disc(x, gy, 3, white)
			p.sparks(rng, x, gy, 8, 10)
		}
	}

	// The visor: three lenses in a slot. A light runs along them as it looks about, and all
	// three burn to fire.
	if po.death < 4 {
		p.plate(20, int(cy)-9, 52, int(cy)+2, iron, true)
		p.fill(21, int(cy)-8, 51, int(cy)+1, hex(0x0c0e16))
		for i, x := range []float64{27, 36, 45} {
			lens, size := tones(0x4c6c8c), 2.4+float64(i%2)
			switch {
			case po.attack > 0:
				lens = hot
			case !alive:
				lens = dead
			case i == []int{0, 1, 2, 1}[po.step]:
				lens = ice
			}
			p.ball(x, cy-3.5, size, size, lens)
			if alive {
				p.set(int(x)-1, int(cy)-5, white)
			}
		}
		if alive {
			p.glow(36, cy-3.5, 17, []rgb{ice[3], hot[3], hot[3]}[po.attack], 0.22)
		}
	}

	p.wear(rng, 90)
	p.ending(rng, po.death, 36, 32+sink, 24)
	if po.death == 1 || po.death == 2 { // it goes with more fire than the others
		p.sparks(rng, 36, 30, 30, 40)
	}
	p.outline(ink)
	p.shadow(36, 68.6, 34, 1.4)
	return p
}
