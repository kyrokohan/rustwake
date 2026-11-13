package main

import (
	"fmt"
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
// There is the title, the start of a map in each zone, and each kind of machine seen from
// two tiles away. A pixel of the canvas is a square, and any other character is a mark.
func TestSnapshot(t *testing.T) {
	dir := os.Getenv("RUSTWAKE_SNAP")
	if dir == "" {
		t.Skip("RUSTWAKE_SNAP is not set")
	}

	shots := map[string]*game{"title": newGame(0)}
	for _, n := range []int{0, 5, 9, 11} {
		g := newGame(n)
		g.screen = playing
		shots[fmt.Sprintf("start-%02d", n+1)] = g
	}
	for kind := range kinds {
		// The first map that has the machine, with the player put in front of it.
		for n := range levels {
			g := newGame(n)
			g.screen = playing
			if faced(g, kind) {
				shots["machine-"+string(kind)] = g
				g.cool = fireEvery - 0.01 // in the middle of a shot, for the flash
				g.flash = flashTime
				break
			}
		}
	}

	for name, g := range shots {
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

// faced puts the player two tiles from the first machine of a kind on the map, looking at
// it, and reports whether there was one with room in front of it.
func faced(g *game, kind rune) bool {
	for _, e := range g.enemies {
		if e.look != kinds[kind].look {
			continue
		}
		for turn := range 4 {
			yaw := float64(turn) * math.Pi / 2
			from := e.pos.Sub(geometry.Vector2{X: math.Sin(yaw), Y: math.Cos(yaw)}.Scale(2.2))
			if !g.lvl.solid(from) && g.lvl.sees(from, e.pos) {
				g.pos, g.yaw, e.awake = from, yaw, true
				return true
			}
		}
	}
	return false
}
