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
	{name: "weapon", frames: 4, draw: weapon, apart: true},
	{name: "health", frames: 2, draw: health},
	{name: "ammo", frames: 2, draw: ammo},
	{name: "bolt", frames: 2, draw: bolt},
	{name: "exit", frames: 2, draw: exit},
	{name: "barrel", frames: 1, draw: barrel},
	{name: "crate", frames: 1, draw: crate},
	{name: "lamp", frames: 1, draw: lamp},
}

func main() {
	out := flag.String("out", "assets", "the folder the PNG files are written to")
	preview := flag.String("preview", "", "a folder for enlarged copies, to look at")
	flag.Parse()

	for _, s := range sheets {
		frames := make([]*pic, s.frames)
		for i := range frames {
			frames[i] = s.draw(i)
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
