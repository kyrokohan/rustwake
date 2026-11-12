// Command rustwake is a first-person shooter in the terminal, drawn by glonk: clear the
// rogue machines out of a dark plant, one map at a time.
//
//	go run .
//
// W and S walk, A and D strafe, the left and right arrows or Q and E turn, Space fires and
// Esc pauses. Walking and turning at once needs a terminal that reports key releases (the
// kitty keyboard protocol: Ghostty, kitty, WezTerm); elsewhere one key is held at a time.
package main

import (
	"context"
	"fmt"
	"os"

	"glonk.dev/engine"
)

func main() {
	// Width and Height of 0 fill the terminal and follow resizes.
	if err := engine.Run(context.Background(), newGame(), engine.Config{FPS: 30}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
