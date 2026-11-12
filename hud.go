package main

import (
	"fmt"

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

// weaponArt is the weapon as the player holds it, seen from behind, a pixel for each rune.
var weaponArt = []string{
	"     ####     ",
	"     #oo#     ",
	"     #oo#     ",
	"    ##oo##    ",
	"    #====#    ",
	"   ##====##   ",
	"   #======#   ",
	"  ##======##  ",
	"  #========#  ",
}

var weaponColors = map[rune]engine.Color{
	'#': engine.RGB(52, 56, 66),
	'=': engine.RGB(112, 118, 132),
	'o': engine.RGB(226, 120, 40),
}

// drawWeapon draws the weapon at the bottom middle of the view, larger in a larger window.
// While it fires, its coil is white and it kicks down.
func (g *game) drawWeapon(c *engine.Canvas) {
	size := max(1, c.Width()/70)
	x0 := (c.Width() - len(weaponArt[0])*size) / 2
	y0 := c.Height() - len(weaponArt)*size
	firing := g.flash > 0
	if firing {
		y0 += size
	}
	for y, row := range weaponArt {
		for x, r := range row {
			color, ok := weaponColors[r]
			if !ok {
				continue
			}
			if firing && r == 'o' {
				color = engine.RGB(255, 250, 220)
			}
			for dy := range size {
				for dx := range size {
					c.Set(x0+x*size+dx, y0+y*size+dy, color)
				}
			}
		}
	}
}

// drawHUD draws the panel at the bottom left: health as a number and a bar, and the ammo.
// A hit turns its border red, and puts a red border round the screen.
func (g *game) drawHUD(fb *engine.FrameBuffer) {
	style := hudStyle
	if g.pain > 0 {
		style.Foreground = engine.Palette(196)
		fb.DrawBox(0, 0, fb.Width(), fb.Height(), style.Foreground)
	}
	y := fb.Height() - 3
	widget.Panel{Title: fmt.Sprintf("level %d/%d", g.level+1, len(levels)), Style: style, Fill: hudFill}.Draw(fb, 0, y, 34, 3)
	widget.Label{Text: fmt.Sprintf("HP %3d", g.health), Style: hudStyle}.Draw(fb, 2, y+1, 6)
	widget.Bar{
		Value: float64(g.health) / maxHealth,
		Full:  engine.Cell{Character: '█', Foreground: engine.Palette(76), Background: hudStyle.Background},
		Empty: engine.Cell{Character: '░', Foreground: engine.Palette(240), Background: hudStyle.Background},
	}.Draw(fb, 9, y+1, 12)
	widget.Label{Text: fmt.Sprintf("AMMO %d", g.ammo), Style: hudStyle}.Draw(fb, 23, y+1, 9)
}

// drawMenu draws the screen's menu in a panel in the middle, with the keys under it.
func (g *game) drawMenu(fb *engine.FrameBuffer) {
	const w = 28
	h := len(g.menu.Items) + 2
	x, y := (fb.Width()-w)/2, (fb.Height()-h-len(help))/2
	widget.Panel{Title: menus[g.screen].title, Style: hudStyle, Fill: hudFill}.Draw(fb, x, y, w, h)
	g.menu.Draw(fb, x+2, y+1, w-4, h-2)
	for i, line := range help {
		widget.Label{Text: line, Align: widget.Center, Style: hudStyle}.Draw(fb, x, y+h+i, w)
	}
}
