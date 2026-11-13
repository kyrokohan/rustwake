package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand/v2"
	"os"
)

type rgb = color.NRGBA

// hex returns the solid color 0xRRGGBB.
func hex(v uint32) rgb { return rgb{R: uint8(v >> 16), G: uint8(v >> 8), B: uint8(v), A: 255} }

// mix returns the color t of the way from a to b.
func mix(a, b rgb, t float64) rgb {
	t = min(1, max(0, t))
	f := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return rgb{R: f(a.R, b.R), G: f(a.G, b.G), B: f(a.B, b.B), A: 255}
}

// ramp is five tones of one material, from its deepest shadow to its shine.
type ramp [5]rgb

// tones makes a ramp round a middle color. The shadows lean blue and the lights lean
// warm, which is what keeps a shaded shape from looking like gray plastic.
func tones(mid uint32) ramp {
	m := hex(mid)
	return ramp{
		mix(m, hex(0x120e24), 0.74),
		mix(m, hex(0x1e1a3a), 0.44),
		m,
		mix(m, hex(0xfff2d2), 0.34),
		mix(m, hex(0xfffcf0), 0.66),
	}
}

// bayer is the 4 by 4 ordered dither: the share of the way to the next tone at which each
// pixel of the pattern steps up to it.
var bayer = [4][4]float64{
	{0.5 / 16, 8.5 / 16, 2.5 / 16, 10.5 / 16},
	{12.5 / 16, 4.5 / 16, 14.5 / 16, 6.5 / 16},
	{3.5 / 16, 11.5 / 16, 1.5 / 16, 9.5 / 16},
	{15.5 / 16, 7.5 / 16, 13.5 / 16, 5.5 / 16},
}

// tone picks the tone for a brightness from 0 to 1 at a pixel, dithered between two
// neighbors so that a curve shades in steps finer than five.
func (r ramp) tone(s float64, x, y int) rgb {
	v := min(1, max(0, s)) * 4
	i := int(v)
	if i >= 4 {
		return r[4]
	}
	if v-float64(i) > bayer[y&3][x&3] {
		i++
	}
	return r[i]
}

// band picks the tone for a brightness with no dither, for flat and straight shapes, on
// which a dither only looks like dirt.
func (r ramp) band(s float64) rgb {
	return r[int(min(4, max(0, s)*5))]
}

// sun is the way to the light that every picture is shaded by: up, to the left, and
// toward the viewer.
var sun = [3]float64{-0.55, -0.65, 0.52}

// pic is a picture being drawn. A pixel with no alpha is a hole.
type pic struct {
	w, h int
	px   []rgb
}

func newPic(w, h int) *pic { return &pic{w, h, make([]rgb, w*h)} }

func (p *pic) at(x, y int) rgb {
	if x < 0 || y < 0 || x >= p.w || y >= p.h {
		return rgb{}
	}
	return p.px[y*p.w+x]
}

func (p *pic) set(x, y int, c rgb) {
	if x >= 0 && y >= 0 && x < p.w && y < p.h {
		p.px[y*p.w+x] = c
	}
}

// over sets a pixel only where something is drawn already, for marks on a shape.
func (p *pic) over(x, y int, c rgb) {
	if p.at(x, y).A != 0 {
		p.set(x, y, c)
	}
}

// fill fills a box with one color.
func (p *pic) fill(x0, y0, x1, y1 int, c rgb) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			p.set(x, y, c)
		}
	}
}

// hull fills a box with rounded corners as a plate of metal: lit at the top and the left,
// in shade at the bottom and the right, with a bright edge above and a dark one below.
func (p *pic) hull(x0, y0, x1, y1 int, r ramp, round float64) {
	inside := func(x, y int) bool {
		if x < x0 || x >= x1 || y < y0 || y >= y1 {
			return false
		}
		fx, fy := float64(x)+0.5, float64(y)+0.5
		cx := min(float64(x1)-round, max(float64(x0)+round, fx))
		cy := min(float64(y1)-round, max(float64(y0)+round, fy))
		return math.Hypot(fx-cx, fy-cy) <= round
	}
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			if !inside(x, y) {
				continue
			}
			c := r.band(0.7 - 0.36*float64(y-y0)/float64(y1-y0))
			switch {
			case !inside(x, y-1), !inside(x-1, y):
				c = r[3]
			case !inside(x, y+1), !inside(x+1, y):
				c = r[1]
			}
			p.set(x, y, c)
		}
	}
}

// ball fills an ellipse shaded as a ball.
func (p *pic) ball(cx, cy, rx, ry float64, r ramp) {
	for y := int(cy - ry); y <= int(cy+ry)+1; y++ {
		for x := int(cx - rx); x <= int(cx+rx)+1; x++ {
			nx, ny := (float64(x)+0.5-cx)/rx, (float64(y)+0.5-cy)/ry
			d := nx*nx + ny*ny
			if d > 1 {
				continue
			}
			lit := nx*sun[0] + ny*sun[1] + math.Sqrt(1-d)*sun[2]
			p.set(x, y, r.tone(0.1+0.9*max(0, lit), x, y))
		}
	}
}

// disc fills a circle with one color.
func (p *pic) disc(cx, cy, radius float64, c rgb) {
	for y := int(cy - radius); y <= int(cy+radius)+1; y++ {
		for x := int(cx - radius); x <= int(cx+radius)+1; x++ {
			if math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy) <= radius {
				p.set(x, y, c)
			}
		}
	}
}

// tube fills a box shaded as a round bar: one that stands, or one that lies on its side.
func (p *pic) tube(x0, y0, x1, y1 int, r ramp, lying bool) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			n, light := 2*(float64(x-x0)+0.5)/float64(x1-x0)-1, sun[0]
			if lying {
				n, light = 2*(float64(y-y0)+0.5)/float64(y1-y0)-1, sun[1]
			}
			p.set(x, y, r.band(0.16+0.95*max(0, n*light+math.Sqrt(1-n*n)*sun[2])))
		}
	}
}

// limb draws a bar w pixels thick from one point to another, lit on the side that faces
// the light: an arm, a leg, a pipe.
func (p *pic) limb(x0, y0, x1, y1, w float64, r ramp) {
	dx, dy := x1-x0, y1-y0
	length := math.Hypot(dx, dy)
	if length == 0 {
		return
	}
	nx, ny := -dy/length, dx/length
	facing := (nx*sun[0] + ny*sun[1]) / math.Hypot(sun[0], sun[1])
	for y := int(min(y0, y1) - w); y <= int(max(y0, y1)+w); y++ {
		for x := int(min(x0, x1) - w); x <= int(max(x0, x1)+w); x++ {
			px, py := float64(x)+0.5-x0, float64(y)+0.5-y0
			along := min(length, max(0, (px*dx+py*dy)/length))
			ox, oy := px-along*dx/length, py-along*dy/length
			if math.Hypot(ox, oy) > w/2 {
				continue
			}
			side := (ox*nx + oy*ny) / (w / 2)
			p.set(x, y, r.tone(0.46+0.4*side*facing, x, y))
		}
	}
}

// stripes paints the slanted bands of a warning mark over what is drawn in a box.
func (p *pic) stripes(x0, y0, x1, y1 int, a, b rgb) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := a
			if (x+y)/3%2 == 1 {
				c = b
			}
			p.over(x, y, c)
		}
	}
}

// vent paints the slats of a grille over what is drawn in a box.
func (p *pic) vent(x0, y0, x1, y1 int, r ramp) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			c := r[0]
			if (y-y0)%2 == 1 {
				c = r[2]
			}
			p.over(x, y, c)
		}
	}
}

// rivet paints a rivet head: a bright pixel with its shadow below and to the right.
func (p *pic) rivet(x, y int, r ramp) {
	p.over(x, y, r[4])
	p.over(x+1, y+1, r[0])
}

// lens draws an eye: a dark socket, a glass ball in it and a glint on the glass.
func (p *pic) lens(cx, cy, radius float64, glass ramp) {
	p.disc(cx, cy, radius+1.6, hex(0x16141c))
	p.ball(cx, cy, radius, radius, glass)
	p.set(int(cx-radius*0.45), int(cy-radius*0.5), hex(0xffffff))
}

// glow lights what is drawn round a point with a color, strongest at the point.
func (p *pic) glow(cx, cy, radius float64, c rgb, strength float64) {
	for y := int(cy - radius); y <= int(cy+radius); y++ {
		for x := int(cx - radius); x <= int(cx+radius); x++ {
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy) / radius
			if d < 1 && p.at(x, y).A != 0 {
				p.set(x, y, mix(p.at(x, y), c, strength*(1-d)*(1-d)))
			}
		}
	}
}

// tint moves every pixel drawn t of the way to a color: white for the flash of a hit, soot
// for a wreck.
func (p *pic) tint(c rgb, t float64) {
	for i, old := range p.px {
		if old.A != 0 {
			p.px[i] = mix(old, c, t)
		}
	}
}

// wear scratches and stains what is drawn: n pixels a little darker or a little lighter.
func (p *pic) wear(rng *rand.Rand, n int) {
	for range n {
		x, y := rng.IntN(p.w), rng.IntN(p.h)
		if old := p.at(x, y); old.A != 0 {
			to, t := hex(0x100c18), 0.28
			if rng.IntN(3) == 0 {
				to, t = hex(0xfff6e0), 0.22
			}
			p.set(x, y, mix(old, to, t))
		}
	}
}

// sparks scatters n hot points round a place, some with a short tail.
func (p *pic) sparks(rng *rand.Rand, cx, cy, spread float64, n int) {
	hot := []rgb{hex(0xffffff), hex(0xfff0a0), hex(0xffc040), hex(0xff8020)}
	for range n {
		a, d := rng.Float64()*2*math.Pi, spread*math.Sqrt(rng.Float64())
		x, y := cx+math.Cos(a)*d, cy+math.Sin(a)*d
		c := hot[rng.IntN(len(hot))]
		p.set(int(x), int(y), c)
		if rng.IntN(2) == 0 {
			p.set(int(x+math.Cos(a)*1.4), int(y+math.Sin(a)*1.4), mix(c, hex(0xff5010), 0.5))
		}
	}
}

// smoke puffs n balls of smoke round a place.
func (p *pic) smoke(rng *rand.Rand, cx, cy, spread float64, n int) {
	gray := tones(0x4a4850)
	for range n {
		radius := 1.6 + rng.Float64()*2.6
		p.ball(cx+(rng.Float64()-0.5)*2*spread, cy+(rng.Float64()-0.5)*spread, radius, radius, gray)
	}
}

// shadow darkens the floor under a thing: an ellipse drawn only where nothing is.
func (p *pic) shadow(cx, cy, rx, ry float64) {
	for y := int(cy - ry); y <= int(cy+ry)+1; y++ {
		for x := int(cx - rx); x <= int(cx+rx)+1; x++ {
			nx, ny := (float64(x)+0.5-cx)/rx, (float64(y)+0.5-cy)/ry
			if nx*nx+ny*ny <= 1 && p.at(x, y).A == 0 {
				p.set(x, y, hex(0x0c0a12))
			}
		}
	}
}

// outline draws a line round the whole shape: on every hole that touches a pixel.
func (p *pic) outline(c rgb) {
	var edge []int
	for y := range p.h {
		for x := range p.w {
			if p.at(x, y).A == 0 && p.at(x-1, y).A|p.at(x+1, y).A|p.at(x, y-1).A|p.at(x, y+1).A != 0 {
				edge = append(edge, y*p.w+x)
			}
		}
	}
	for _, i := range edge {
		p.px[i] = c
	}
}

// paste draws another picture over this one with its corner at (dx, dy), holes left out.
func (p *pic) paste(src *pic, dx, dy int) {
	for y := range src.h {
		for x := range src.w {
			if c := src.at(x, y); c.A != 0 {
				p.set(x+dx, y+dy, c)
			}
		}
	}
}

// strip lays frames side by side, which is how the game reads an animation.
func strip(frames []*pic) *pic {
	out := newPic(frames[0].w*len(frames), frames[0].h)
	for i, f := range frames {
		out.paste(f, i*f.w, 0)
	}
	return out
}

// enlarge lays frames out in rows on a backdrop, each pixel a block, for looking at.
func enlarge(frames []*pic, scale, perRow int) *pic {
	const gap = 6
	w, h := frames[0].w, frames[0].h
	rows := (len(frames) + perRow - 1) / perRow
	cols := min(perRow, len(frames))
	out := newPic(gap+cols*(w*scale+gap), gap+rows*(h*scale+gap))
	out.fill(0, 0, out.w, out.h, hex(0x34363e))
	for i, f := range frames {
		x0, y0 := gap+i%perRow*(w*scale+gap), gap+i/perRow*(h*scale+gap)
		out.fill(x0, y0, x0+w*scale, y0+h*scale, hex(0x40434c))
		for y := range h {
			for x := range w {
				if c := f.at(x, y); c.A != 0 {
					out.fill(x0+x*scale, y0+y*scale, x0+(x+1)*scale, y0+(y+1)*scale, c)
				}
			}
		}
	}
	return out
}

// save writes a picture as a PNG file.
func save(path string, p *pic) error {
	img := image.NewNRGBA(image.Rect(0, 0, p.w, p.h))
	for i, c := range p.px {
		img.SetNRGBA(i%p.w, i/p.w, c)
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
