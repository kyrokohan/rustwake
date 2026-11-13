package main

import (
	"fmt"
	"math"

	"glonk.dev/engine"
	"glonk.dev/widget"
)

// help is the keys, shown under every menu where there is room.
var help = []string{
	"W S walk    A D strafe",
	"arrows or Q E turn",
	"Space fire   Esc pause",
}

// drawWeapon draws the weapon at the bottom middle of the view, a good quarter of its
// width, with its lower third below the edge so that it does not hide what is aimed at.
// A shot plays its three frames, one every shotFrame seconds, the weapon bobs as the
// player walks, and it sinks out of sight as the player falls.
func (g *game) drawWeapon(c *engine.Canvas) {
	const shotFrame = 0.08
	frame := 0
	if since := fireEvery - g.cool; since < 3*shotFrame { // cool counts down from fireEvery at a shot
		frame = 1 + int(since/shotFrame)
	}
	pic := weaponFrames[frame]
	w := c.Width() * 28 / 100
	h := w * pic.Height() / pic.Width()
	bob := int(math.Abs(math.Sin(g.walked*5)) * float64(h) / 14)
	drop := int(g.fallen() * float64(h))
	pic.DrawScaled(c, (c.Width()-w)/2, c.Height()-h*2/3+bob+drop, w, h)
}

// The status strip along the bottom of the screen is three rows of cells: six rows of
// pixels for the picture of it, and the middle row of cells for its numbers. A pixel is as
// wide as a cell, so the two are laid out with the same numbers.
const (
	stripRows = 3
	barAt     = 9  // where the health bar starts
	barParts  = 10 // the pieces of the health bar, each 3 pixels and a gap
	healthAt  = barAt + 4*barParts + 1
	ammoWidth = 5 // cells the ammo count has, at the right
)

var (
	stripColor = engine.RGB(28, 30, 38)
	stripEdge  = engine.RGB(96, 104, 122)
	stripText  = engine.Cell{Foreground: engine.RGB(226, 230, 240), Background: stripColor}
	stripName  = engine.Cell{Foreground: engine.RGB(128, 136, 154), Background: stripColor}
)

// drawStatus draws the picture of the status strip, before the pixels are put on the
// screen: the plate, the icons and the health bar, which turns red when health is low. A
// hit puts a red edge round the whole view.
func (g *game) drawStatus(c *engine.Canvas) {
	top := c.Height() - 2*stripRows
	c.FillRect(0, top, c.Width(), 2*stripRows, stripColor)
	c.FillRect(0, top, c.Width(), 1, stripEdge)
	healthIcon.Draw(c, 2, top+1)
	ammoIcon.Draw(c, c.Width()-ammoWidth-9, top+1)

	full := engine.RGB(72, 200, 104)
	if g.health <= maxHealth*3/10 {
		full = engine.RGB(224, 56, 56)
	}
	for i := range barParts {
		color := engine.RGB(52, 56, 68)
		if g.health > i*maxHealth/barParts {
			color = full
		}
		c.FillRect(barAt+4*i, top+2, 3, 3, color)
	}

	if g.pain > 0 {
		red := engine.Shade(engine.RGB(255, 40, 40), 0.4+0.6*g.pain/painTime)
		c.DrawRect(0, 0, c.Width(), top, red)
		c.DrawRect(1, 1, c.Width()-2, top-2, red)
	}
}

// drawHUD writes the numbers on the status strip, after the pixels are on the screen:
// the health, the ammo, and the map's number and name where there is room for them.
func (g *game) drawHUD(fb *engine.FrameBuffer) {
	y := fb.Height() - stripRows + 1
	widget.Label{Text: fmt.Sprintf("%3d", g.health), Style: stripText}.Draw(fb, healthAt, y, 3)
	widget.Label{Text: fmt.Sprintf("%d", g.ammo), Align: widget.Right, Style: stripText}.Draw(fb, fb.Width()-ammoWidth-2, y, ammoWidth)
	name := fmt.Sprintf("%d/%d %s", g.level+1, len(levels), levels[g.level].name)
	if room := fb.Width() - 2*(healthAt+5); room >= len(name) {
		widget.Label{Text: name, Align: widget.Center, Style: stripName}.Draw(fb, healthAt+5, y, room)
	}
}

// A menu is a steel plate in the middle of the view, drawn as pixels, with the view gone
// dark behind it.
const (
	plateBand  = 3                // rows of warning stripe across the top of the plate
	platePad   = 4                // pixels between the plate's edge and what is on it
	itemHeight = letterHeight + 4 // an item, and the room round it for the bar that marks it
	statHeight = letterHeight + 2
	cursorRoom = 10 // pixels at the left of the items, for the mark beside the chosen one
	smallView  = 60 // rows of view under which a menu is drawn tight: a window of 80 by 24 has 42
)

var (
	plateColor = engine.RGB(70, 78, 96)
	plateLight = engine.RGB(140, 150, 172)
	plateDark  = engine.RGB(34, 38, 50)
	plateInk   = engine.RGB(16, 14, 22)
	warnColor  = engine.RGB(240, 192, 64)
	titleColor = engine.RGB(238, 242, 250)
	itemColor  = engine.RGB(170, 178, 196)
	statColor  = engine.RGB(132, 140, 160)
)

// stats are the lines under the items of the two menus that end a game: how far it got.
func (g *game) stats() []string {
	if g.screen != dead && g.screen != won {
		return nil
	}
	return []string{
		fmt.Sprintf("level %d/%d", g.level+1, len(levels)),
		fmt.Sprintf("wrecked %d", g.kills),
		fmt.Sprintf("time %d:%02d", int(g.played)/60, int(g.played)%60),
	}
}

// drawMenu draws the screen's menu and returns the row of pixels under it. The plate is as
// large as what is on it: a title, in letters twice the size where the window is wide
// enough, the items, the chosen one on a bar with a mark that glints beside it, and the
// stats where there is room for them. On the title screen the game's name stands over it.
func (g *game) drawMenu(c *engine.Canvas) int {
	m := menus[g.screen]
	view := c.Height() - 2*stripRows // the rows of the view, above the status strip
	pad, row := platePad, itemHeight
	if view < smallView { // a small window has a tighter plate, or it would not fit
		pad, row = 2, letterHeight+2
	}
	for y := range view {
		for x := range c.Width() {
			c.Set(x, y, engine.Shade(c.At(x, y), 0.4))
		}
	}

	big := 1
	if c.Width() >= 160 && span(m.title, 2)+2*pad+2 <= c.Width()-8 {
		big = 2
	}
	w := span(m.title, big)
	for _, item := range m.items {
		w = max(w, span(item, 1)+2*cursorRoom)
	}
	h := 1 + plateBand + pad + len(m.items)*row + pad
	if m.title != "" {
		h += letterHeight*big + pad
	}
	top := 0
	if g.screen == title {
		top = drawLogo(c) + 2
	}
	stats := g.stats()
	if h+len(stats)*statHeight+pad > view-top-2 {
		stats = nil // a small window has the menu and no more
	}
	for _, line := range stats {
		w = max(w, span(line, 1))
	}
	if stats != nil {
		h += len(stats)*statHeight + pad
	}
	w += 2*pad + 2
	x, y := (c.Width()-w)/2, max(top, (view-h)/2)

	c.FillRect(x-1, y-1, w+2, h+2, plateInk)
	c.FillRect(x, y, w, h, plateColor)
	c.FillRect(x, y, w, 1, plateLight)
	c.FillRect(x, y, 1, h, plateLight)
	c.FillRect(x, y+h-1, w, 1, plateDark)
	c.FillRect(x+w-1, y, 1, h, plateDark)
	for j := range plateBand {
		for i := 1; i < w-1; i++ {
			color := warnColor
			if (i+j)/3%2 == 1 {
				color = plateInk
			}
			c.Set(x+i, y+1+j, color)
		}
	}
	for _, at := range [][2]int{{2, plateBand + 2}, {w - 4, plateBand + 2}, {2, h - 4}, {w - 4, h - 4}} { // a rivet in each corner
		c.Set(x+at[0], y+at[1], plateLight)
		c.Set(x+at[0]+1, y+at[1]+1, plateInk)
	}

	at := y + 1 + plateBand + pad
	if m.title != "" {
		write(c, (c.Width()-span(m.title, big))/2, at, big, titleColor, m.title)
		at += letterHeight*big + pad
	}
	for i, item := range m.items {
		color := itemColor
		if i == g.menu.Selected {
			color = plateInk
			c.FillRect(x+3, at-2, w-6, letterHeight+4, warnColor)
			blit(c, cursor, cursorFrames, int(g.menuClock*4)%cursorFrames, x+pad+1, at, 1, engine.Color{})
		}
		write(c, (c.Width()-span(item, 1))/2, at, 1, color, item)
		at += row
	}
	if stats != nil {
		c.FillRect(x+pad, at-1, w-2*pad, 1, plateDark)
		at += pad - 1
		for _, line := range stats {
			write(c, (c.Width()-span(line, 1))/2, at, 1, statColor, line)
			at += statHeight
		}
	}
	return y + h + 1
}

// drawLogo draws the game's name across the top of the title screen, as pixels, and
// returns the row of pixels under it.
func drawLogo(c *engine.Canvas) int {
	w := c.Width() * 6 / 10
	h := w * logo.Height() / logo.Width()
	y := c.Height() / 9
	if c.Height()-2*stripRows < smallView {
		y = 2
	}
	logo.DrawScaled(c, (c.Width()-w)/2, y, w, h)
	return y + h
}

// drawHints writes the keys under a menu, from the row of cells under the plate, where
// they fit between it and the status strip.
func drawHints(fb *engine.FrameBuffer, row int) {
	if row+len(help) > fb.Height()-stripRows {
		return
	}
	for i, line := range help {
		widget.Label{Text: line, Align: widget.Center, Style: stripText}.Draw(fb, 0, row+i, fb.Width())
	}
}
