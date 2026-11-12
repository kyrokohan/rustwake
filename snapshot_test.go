package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"testing"

	"glonk.dev/engine"
	"glonk.dev/geometry"
)

// TestSnapshot writes frames of the game as PNG files, to look at what a change to the art
// or the drawing did without a terminal:
//
//	RUSTWAKE_SNAP=/some/folder go test -run TestSnapshot .
//
// A pixel of the canvas is a square, and any other character is a mark in its own color.
func TestSnapshot(t *testing.T) {
	dir := os.Getenv("RUSTWAKE_SNAP")
	if dir == "" {
		t.Skip("RUSTWAKE_SNAP is not set")
	}
	shots := map[string]func(g *game){
		"1-title": func(g *game) {},
		"2-doorway": func(g *game) {
			g.pos, g.yaw = geometry.Vector2{X: 5.5, Y: 3.5}, math.Pi/2
		},
		"3-crawler": func(g *game) {
			g.pos, g.yaw = geometry.Vector2{X: 9.5, Y: 3.6}, math.Pi
		},
		"4-sentry-fire": func(g *game) {
			g.pos, g.yaw = geometry.Vector2{X: 18.5, Y: 2.5}, math.Pi/4
			g.flash, g.pain = 0.05, 0.1
			g.bolts = append(g.bolts, bolt{pos: geometry.Vector2{X: 19.6, Y: 3.6}})
		},
		"5-pickups-exit": func(g *game) {
			g.pos, g.yaw = geometry.Vector2{X: 18.5, Y: 1.5}, 0.35
		},
		"6-level2": func(g *game) {
			g.restart(1)
			g.yaw = -0.5
		},
	}
	for name, set := range shots {
		g := newGame()
		if name != "1-title" {
			g.screen = playing
		}
		set(g)
		fb, _ := engine.NewFrameBuffer(160, 48)
		g.Draw(fb)

		const s = 5 // the side of a pixel of the canvas, in pixels of the file
		img := image.NewRGBA(image.Rect(0, 0, fb.Width()*s, fb.Height()*2*s))
		fill := func(x0, y0, x1, y1 int, c engine.Color) {
			r, gr, b := c.RGB()
			for y := y0; y < y1; y++ {
				for x := x0; x < x1; x++ {
					img.Set(x, y, color.RGBA{r, gr, b, 255})
				}
			}
		}
		for y := range fb.Height() {
			for x := range fb.Width() {
				cell := fb.At(x, y)
				top, bottom := cell.Background, cell.Background
				switch cell.Character {
				case '▀':
					top = cell.Foreground
				case '▄':
					bottom = cell.Foreground
				case '█':
					top, bottom = cell.Foreground, cell.Foreground
				}
				fill(x*s, 2*y*s, (x+1)*s, (2*y+1)*s, top)
				fill(x*s, (2*y+1)*s, (x+1)*s, 2*(y+1)*s, bottom)
				if top == bottom && cell.Character > ' ' && cell.Character != '█' {
					fill(x*s+1, 2*y*s+3, (x+1)*s-1, 2*(y+1)*s-3, engine.RGB(210, 210, 210))
				}
			}
		}
		f, err := os.Create(filepath.Join(dir, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}
