package main

import (
	"bytes"
	"embed"
	"fmt"
	"math"

	"glonk.dev/bitmap"
	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

// assets are the game's pictures, PNG files in which a see-through pixel is a hole. Put
// another PNG of any size in a file's place and the game draws that instead.
//
//go:embed assets/*.png
var assets embed.FS

// look is how a thing in the world is drawn: its picture, how tall it stands in tiles and
// how far off the floor. Its width follows from the shape of the picture.
type look struct {
	pic          *bitmap.Bitmap
	height, lift float64
}

var (
	boltLook    = look{pic: art("bolt"), height: 0.15, lift: 0.4}
	exitLook    = look{pic: art("exit"), height: wallHeight}
	pickupLooks = map[rune]look{
		'h': {pic: art("health"), height: 0.3},
		'a': {pic: art("ammo"), height: 0.3},
	}
	weaponIdle, weaponFiring = art("weapon"), art("weapon-fire")
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

// stand draws a thing upright at a spot on the grid, turned to the camera, at the given
// brightness. Every pixel has the thing's distance from the camera as its depth, which is
// the depth the walls are drawn with, so a wall or a nearer thing hides it, in whole or in
// part, whatever the order of drawing.
func stand(c *engine.Canvas, cam mesh.Camera, l look, at geometry.Vector2, bright float64) {
	foot := geometry.Vector3{X: at.X, Y: l.lift, Z: at.Y}
	// As Mesh.DrawShadedPixels does it: onto the grid of cells the canvas belongs to, and
	// then y doubled to rows of pixels. That is what puts a thing's feet on the floor.
	x, bottom, ok := cam.Project(foot, c.Width(), c.Height()/2)
	if !ok || l.pic.Height() == 0 {
		return
	}
	_, top, _ := cam.Project(foot.Add(geometry.Vector3{Y: l.height}), c.Width(), c.Height()/2)
	bottom, top = 2*bottom, 2*top
	h := bottom - top
	w := h * float64(l.pic.Width()) / float64(l.pic.Height())
	left := x - w/2
	depth := foot.Sub(cam.Position).Dot(cam.Forward())

	// A pixel is drawn when its middle is on the picture, which is from round to round.
	for py := max(0, int(math.Round(top))); py < min(c.Height(), int(math.Round(bottom))); py++ {
		v := int((float64(py) + 0.5 - top) / h * float64(l.pic.Height()))
		for px := max(0, int(math.Round(left))); px < min(c.Width(), int(math.Round(left+w))); px++ {
			u := int((float64(px) + 0.5 - left) / w * float64(l.pic.Width()))
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
