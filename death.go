package main

import (
	"math"
	"math/rand/v2"

	"glonk.dev/engine"
	"glonk.dev/geometry"
)

// How the player dies: a scene of deathTime seconds between the last hit and the game
// over menu, in which they go down and the glass they look through cracks and runs red.
const (
	deathTime = 2.2  // seconds from the last hit to the menu
	skipAfter = 0.8  // seconds of it that cannot be skipped
	fallTime  = 0.6  // seconds the player takes to go down
	floorEye  = 0.12 // how high the eye is when they are down
	fallPitch = 0.45 // how far the view has tipped up by then, in radians
	splatsIn  = 9    // how many times blood hits the glass
)

var (
	bloodRun  = engine.RGB(120, 8, 14)
	bloodDrop = engine.RGB(196, 22, 28)
	bloodEdge = engine.RGB(150, 6, 10)
	bloodDeep = engine.RGB(64, 0, 4) // what the view drains to
	glassLine = engine.RGB(208, 216, 228)
)

// gore is what is on the glass when the player dies: the blood, and the cracks. Places
// and sizes are shares of the screen, so that it is the same in a window of any size.
type gore struct {
	splats []bloodSplat
	cracks [][]geometry.Vector2 // each a line from where the hit landed, point by point
}

type bloodSplat struct {
	at    geometry.Vector2 // its middle, across and down the view
	size  float64          // its width, as a share of the view's
	shape int              // which of the splat pictures it is
	born  float64          // seconds after the last hit that it lands
	runs  []bloodRunning
}

// bloodRunning is a streak that blood runs down the glass in, out of a splat.
type bloodRunning struct {
	from  float64 // where across the splat it starts, from -0.3 to 0.3 of its width
	speed float64 // shares of the view's height a second
	wide  int
}

// newGore lays out a death from a seed: the same seed is the same death, so a frame of it
// can be checked.
func newGore(seed uint64) gore {
	rng := rand.New(rand.NewPCG(seed, 17))
	var g gore
	for i := range splatsIn {
		s := bloodSplat{
			at:    geometry.Vector2{X: 0.08 + 0.84*rng.Float64(), Y: 0.06 + 0.6*rng.Float64()},
			size:  0.14 + 0.2*rng.Float64(),
			shape: rng.IntN(splatFrames),
			born:  0.13 * float64(i) * (0.7 + 0.6*rng.Float64()),
		}
		for range 1 + rng.IntN(3) {
			s.runs = append(s.runs, bloodRunning{from: 0.6*rng.Float64() - 0.3, speed: 0.12 + 0.3*rng.Float64(), wide: 1 + rng.IntN(2)})
		}
		g.splats = append(g.splats, s)
	}

	hit := geometry.Vector2{X: 0.35 + 0.3*rng.Float64(), Y: 0.3 + 0.25*rng.Float64()}
	for range 7 {
		way := rng.Float64() * 2 * math.Pi
		line := []geometry.Vector2{hit}
		for range 5 {
			way += (rng.Float64() - 0.5) * 0.9 // a crack wanders as it goes
			step := geometry.Vector2{X: math.Cos(way), Y: math.Sin(way)}.Scale(0.05 + 0.07*rng.Float64())
			line = append(line, line[len(line)-1].Add(step))
		}
		g.cracks = append(g.cracks, line)
	}
	return g
}

// draw paints the death over the finished view, t seconds after the last hit. The view
// drains to dark red; the cracks spread from the hit in the first half second; each splat
// lands at its time and then runs down the glass; and a red edge closes in.
func (g gore) draw(c *engine.Canvas, t float64) {
	w, h := c.Width(), c.Height()-2*stripRows
	for y := range h {
		for x := range w {
			stain(c, x, y, bloodDeep, 0.8*min(1, t/deathTime))
		}
	}

	for _, line := range g.cracks {
		for i := range int(min(1, t/0.5) * float64(len(line)-1)) {
			a, b := line[i], line[i+1]
			c.DrawLine(int(a.X*float64(w)), int(a.Y*float64(h)), int(b.X*float64(w)), int(b.Y*float64(h)), glassLine)
		}
	}

	for _, s := range g.splats {
		age := t - s.born
		if age < 0 {
			continue
		}
		scale := max(1, int(s.size*float64(w)/float64(splats.Height())+0.5))
		size := scale * splats.Height()
		x, y := int(s.at.X*float64(w)), int(s.at.Y*float64(h))
		for _, r := range s.runs {
			thick := r.wide * max(1, scale/2)
			rx, length := x+int(r.from*float64(size)), int(r.speed*age*float64(h))
			c.FillRect(rx, y, thick, length, bloodRun)
			c.FillCircle(rx+thick/2, y+length, thick, bloodDrop) // the drop at its tip
		}
		blit(c, splats, splatFrames, s.shape, x-size/2, y-size/2, scale, engine.Color{})
	}

	// The edge: rings of red from the border in, each fainter than the one outside it, so
	// that it darkens the view and does not frame it.
	rings := int(min(1, t/deathTime) * float64(w) / 8)
	for i := range rings {
		faint := 1 - float64(i)/float64(rings)
		for x := i; x < w-i; x++ {
			stain(c, x, i, bloodEdge, 0.8*faint*faint)
			stain(c, x, h-1-i, bloodEdge, 0.8*faint*faint)
		}
		for y := i + 1; y < h-1-i; y++ {
			stain(c, i, y, bloodEdge, 0.8*faint*faint)
			stain(c, w-1-i, y, bloodEdge, 0.8*faint*faint)
		}
	}
}

// stain moves a pixel of the canvas part of the way to a color.
func stain(c *engine.Canvas, x, y int, to engine.Color, part float64) {
	r0, g0, b0 := c.At(x, y).RGB()
	r1, g1, b1 := to.RGB()
	mix := func(from, to uint8) uint8 { return uint8(float64(from) + (float64(to)-float64(from))*part) }
	c.Set(x, y, engine.RGB(mix(r0, r1), mix(g0, g1), mix(b0, b1)))
}
