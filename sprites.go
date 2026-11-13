package main

import (
	"bytes"
	"embed"
	"fmt"
	"math"
	"strings"

	"glonk.dev/bitmap"
	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

// assets are the game's pictures, PNG files in which a see-through pixel is a hole. They
// are drawn by tools/artgen; a PNG made by hand can take the place of any of them.
//
//go:embed assets/*.png
var assets embed.FS

// look is how a thing in the world is drawn: its picture, which holds the frames of its
// animation side by side, how tall it stands in tiles and how far off the floor. Its
// width follows from the shape of a frame.
type look struct {
	pic          *bitmap.Bitmap
	frames       int
	height, lift float64
}

var (
	boltLook    = strip("bolt", 2, 0.16, 0.42)
	exitLook    = strip("exit", 2, wallHeight, 0)
	pickupLooks = map[rune]look{
		'h': strip("health", 2, 0.3, 0),
		'a': strip("ammo", 2, 0.3, 0),
	}
	propLooks = map[rune]look{
		'o': strip("barrel", 1, 0.7, 0),
		'x': strip("crate", 1, 0.78, 0),
		'l': strip("lamp", 1, 0.2, wallHeight-0.2),
	}
	// What is drawn on the screen, not in the world: the game's name, and the marks beside
	// the numbers on the status strip.
	logo, healthIcon, ammoIcon = art("logo"), art("icon-health"), art("icon-ammo")
	// The game's letters, the mark beside the chosen item of a menu, and blood for the glass.
	font, cursor, splats = art("font"), art("cursor"), art("splat")
	// The weapon at rest, and the three frames of a shot.
	weaponFrames = [4]*bitmap.Bitmap{art("weapon-0"), art("weapon-1"), art("weapon-2"), art("weapon-3")}
)

// How the strips that are drawn on the screen are cut up.
const (
	fontFrames   = 'Z' - ' ' + 1 // a letter for every character from the space to Z, in order
	cursorFrames = 2
	splatFrames  = 4
	letterWidth  = 6 // pixels from one letter to the next: five of letter and one of gap
	letterHeight = 7
)

// art returns the picture assets/name.png. The pictures are compiled into the program, so
// one that is missing or cannot be read is a mistake in the build, and art panics.
func art(name string) *bitmap.Bitmap {
	data, err := assets.ReadFile("assets/" + name + ".png")
	if err != nil {
		panic(err)
	}
	pic, err := bitmap.Decode(bytes.NewReader(data))
	if err != nil {
		panic(fmt.Errorf("assets/%s.png: %w", name, err))
	}
	return pic
}

// strip returns the look of a thing whose picture holds the given number of frames.
func strip(name string, frames int, height, lift float64) look {
	pic := art(name)
	if pic.Width()%frames != 0 {
		panic(fmt.Sprintf("assets/%s.png is %d pixels wide, which is not %d frames", name, pic.Width(), frames))
	}
	return look{pic: pic, frames: frames, height: height, lift: lift}
}

// stand draws one frame of a thing upright at a spot on the grid, turned to the camera,
// at the given brightness. Every pixel has the thing's distance from the camera as its
// depth, which is the depth the walls are drawn with, so a wall or a nearer thing hides
// it, in whole or in part, whatever the order of drawing.
func stand(c *engine.Canvas, cam mesh.Camera, l look, frame int, at geometry.Vector2, bright float64) {
	foot := geometry.Vector3{X: at.X, Y: l.lift, Z: at.Y}
	// As Mesh.DrawShadedPixels does it: onto the grid of cells the canvas belongs to, and
	// then y doubled to rows of pixels. That is what puts a thing's feet on the floor.
	x, bottom, ok := cam.Project(foot, c.Width(), c.Height()/2)
	if !ok || l.frames == 0 {
		return
	}
	_, top, _ := cam.Project(foot.Add(geometry.Vector3{Y: l.height}), c.Width(), c.Height()/2)
	bottom, top = 2*bottom, 2*top
	across, down := l.pic.Width()/l.frames, l.pic.Height()
	h := bottom - top
	w := h * float64(across) / float64(down)
	left := x - w/2
	depth := foot.Sub(cam.Position).Dot(cam.Look()) // along the way the camera points, which tips when the player falls

	// A pixel is drawn when its middle is on the picture, which is from round to round.
	for py := max(0, int(math.Round(top))); py < min(c.Height(), int(math.Round(bottom))); py++ {
		v := int((float64(py) + 0.5 - top) / h * float64(down))
		for px := max(0, int(math.Round(left))); px < min(c.Width(), int(math.Round(left+w))); px++ {
			u := frame*across + int((float64(px)+0.5-left)/w*float64(across))
			if color := l.pic.At(u, v); color != (engine.Color{}) {
				c.SetDepth(px, py, depth, engine.Shade(color, bright))
			}
		}
	}
}

// brightness is how bright the lights make a thing at a spot on the grid, as seen from eye.
// The lights are added up as the mesh package adds them up, each filling part of what the
// others left dark, so a thing is as bright as a wall beside it that faces the camera.
func brightness(lights []mesh.Light, eye geometry.Vector3, at geometry.Vector2) float64 {
	p := geometry.Vector3{X: at.X, Y: eyeHeight, Z: at.Y}
	toward := eye.Sub(p).Normalize()
	dark := 1.0
	for _, l := range lights {
		dark *= 1 - min(1, max(0, l.Brightness(p, toward)))
	}
	return 1 - dark
}

// blit draws one frame of a strip on the screen with its top left at (x, y), each of its
// pixels a square of scale pixels, in its own colors or, given a tint, all in that one. It
// is for what is on the screen and not in the world, so it takes no notice of depth.
func blit(c *engine.Canvas, pic *bitmap.Bitmap, frames, frame, x, y, scale int, tint engine.Color) {
	across := pic.Width() / frames
	for v := range pic.Height() {
		for u := range across {
			color := pic.At(frame*across+u, v)
			if color == (engine.Color{}) {
				continue
			}
			if tint != (engine.Color{}) {
				color = tint
			}
			c.FillRect(x+u*scale, y+v*scale, scale, scale, color)
		}
	}
}

// write draws text in the game's letters with its top left at (x, y). The letters are
// capitals; a character that has no letter is left as a gap.
func write(c *engine.Canvas, x, y, scale int, color engine.Color, text string) {
	for i, r := range []rune(strings.ToUpper(text)) {
		if r >= ' ' && r <= 'Z' {
			blit(c, font, fontFrames, int(r-' '), x+i*letterWidth*scale, y, scale, color)
		}
	}
}

// span is how many pixels across write draws text.
func span(text string, scale int) int {
	return (len([]rune(text))*letterWidth - 1) * scale
}
