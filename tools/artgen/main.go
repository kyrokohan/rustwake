// Command artgen draws RUSTWAKE's art and writes it as PNG files. It is the source of the
// pictures in assets: change a drawing here and run it again. A PNG made by hand can take
// the place of any file, as long as artgen is not run over it afterwards.
//
//	go run ./tools/artgen                    # rewrites assets/*.png
//	go run ./tools/artgen -preview /tmp/art  # and writes enlarged copies there, to look at
//
// Everything random in a drawing comes from a fixed seed, so the files are the same from
// run to run and only a changed drawing shows up as a changed file.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// sheet is one PNG file: a picture, or the frames of an animation side by side. A sheet
// that is apart has a file for each frame instead, name-0.png and so on, for the pictures
// the game draws whole on the screen.
type sheet struct {
	name   string
	frames int
	draw   func(frame int) *pic
	apart  bool
}

var sheets = []sheet{
	{name: "crawler", frames: 10, draw: crawler},
	{name: "sentry", frames: 10, draw: sentry},
	{name: "drone", frames: 10, draw: drone},
	{name: "hauler", frames: 10, draw: hauler},
	{name: "overseer", frames: 10, draw: overseer},
	{name: "weapon", frames: 4, draw: weapon, apart: true},
	{name: "health", frames: 2, draw: health},
	{name: "ammo", frames: 2, draw: ammo},
	{name: "bolt", frames: 2, draw: bolt},
	{name: "exit", frames: 2, draw: exit},
	{name: "barrel", frames: 1, draw: barrel},
	{name: "crate", frames: 1, draw: crate},
	{name: "lamp", frames: 1, draw: lamp},

	{name: "wall-1a", frames: 1, draw: wall1a},
	{name: "wall-1b", frames: 1, draw: wall1b},
	{name: "wall-1c", frames: 1, draw: wall1c},
	{name: "floor-1", frames: 1, draw: floor1},
	{name: "ceiling-1", frames: 1, draw: ceiling1},
	{name: "wall-2a", frames: 1, draw: wall2a},
	{name: "wall-2b", frames: 1, draw: wall2b},
	{name: "wall-2c", frames: 1, draw: wall2c},
	{name: "floor-2", frames: 1, draw: floor2},
	{name: "ceiling-2", frames: 1, draw: ceiling2},
	{name: "wall-3a", frames: 1, draw: wall3a},
	{name: "wall-3b", frames: 1, draw: wall3b},
	{name: "wall-3c", frames: 1, draw: wall3c},
	{name: "floor-3", frames: 1, draw: floor3},
	{name: "ceiling-3", frames: 1, draw: ceiling3},
	{name: "floor-exit", frames: 1, draw: floorExit},

	{name: "logo", frames: 1, draw: logo},
	{name: "icon-health", frames: 1, draw: iconHealth},
	{name: "icon-ammo", frames: 1, draw: iconAmmo},
	{name: "font", frames: fontFrames, draw: font},
	{name: "cursor", frames: 2, draw: cursor},
	{name: "splat", frames: 4, draw: splat},
}

func main() {
	out := flag.String("out", "assets", "the folder the PNG files are written to")
	preview := flag.String("preview", "", "a folder for enlarged copies, to look at")
	flag.Parse()

	var surfaces []*pic // the walls, floors and ceilings, for one sheet to look at
	defer func() {
		if *preview != "" {
			save(filepath.Join(*preview, "_surfaces.png"), enlarge(surfaces, 6, 8))
		}
	}()

	for _, s := range sheets {
		frames := make([]*pic, s.frames)
		for i := range frames {
			frames[i] = s.draw(i)
		}
		if frames[0].w == frames[0].h && frames[0].w >= 8 && frames[0].w <= 16 && s.frames == 1 {
			big := tiled(frames[0])
			if big.w < 32 {
				big = tiled(big)
			}
			surfaces = append(surfaces, big)
		}
		var err error
		if s.apart {
			for i, f := range frames {
				err = errors.Join(err, save(filepath.Join(*out, fmt.Sprintf("%s-%d.png", s.name, i)), f))
			}
		} else {
			err = save(filepath.Join(*out, s.name+".png"), strip(frames))
		}
		if err == nil && *preview != "" {
			err = save(filepath.Join(*preview, s.name+".png"), enlarge(frames, 6, 5))
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
