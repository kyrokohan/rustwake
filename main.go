// Command rustwake is a first-person shooter in the terminal, drawn by glonk: clear the
// rogue machines out of a dark plant, one map at a time.
//
//	go run .
//	go run . -level 7   # start at a later level
//
// W and S walk, A and D strafe, the left and right arrows or Q and E turn, Space fires and
// Esc pauses. Walking and turning at once needs a terminal that reports key releases (the
// kitty keyboard protocol: Ghostty, kitty, WezTerm); elsewhere one key is held at a time.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"glonk.dev/engine"
)

func main() {
	start := flag.Int("level", 1, "the level to start at, from 1")
	flag.Parse()
	if *start < 1 || *start > len(levels) {
		fmt.Fprintf(os.Stderr, "there is no level %d: they run from 1 to %d\n", *start, len(levels))
		os.Exit(2)
	}

	// Width and Height of 0 fill the terminal and follow resizes.
	if err := engine.Run(context.Background(), newGame(*start-1), engine.Config{FPS: 30}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
