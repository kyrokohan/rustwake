package main

import (
	"fmt"
	"math"

	"glonk.dev/engine"
	"glonk.dev/widget"
)

// hudStyle is the colors of the panels, their text and the menus: light on dark, so that
// the world does not show through.
var hudStyle = engine.Cell{Foreground: engine.Palette(252), Background: engine.Palette(235)}

var hudFill = engine.Cell{Background: hudStyle.Background}

// help is the keys, shown under every menu.
var help = []string{
	"W S walk    A D strafe",
	"arrows or Q E turn",
	"Space fire   Esc pause",
}

// drawWeapon draws the weapon at the bottom middle of the view, a good quarter of its
// width, with its lower third below the edge so that it does not hide what is aimed at.
// A shot plays its three frames, one every shotFrame seconds, and the weapon bobs as the
// player walks.
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
	pic.DrawScaled(c, (c.Width()-w)/2, c.Height()-h*2/3+bob, w, h)
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

// drawLogo draws the game's name across the top of the title screen, as pixels.
func drawLogo(c *engine.Canvas) {
	w := c.Width() * 6 / 10
	h := w * logo.Height() / logo.Width()
	logo.DrawScaled(c, (c.Width()-w)/2, c.Height()/7, w, h)
}

// drawMenu draws the screen's menu in a panel in the middle, with the keys under it.
func (g *game) drawMenu(fb *engine.FrameBuffer) {
	const w = 28
	h := len(g.menu.Items) + 2
	x, y := (fb.Width()-w)/2, (fb.Height()-h-len(help))/2+2 // a little low, under the name on the title screen
	widget.Panel{Title: menus[g.screen].title, Style: hudStyle, Fill: hudFill}.Draw(fb, x, y, w, h)
	g.menu.Draw(fb, x+2, y+1, w-4, h-2)
	for i, line := range help {
		widget.Label{Text: line, Align: widget.Center, Style: hudStyle}.Draw(fb, x, y+h+i, w)
	}
}
