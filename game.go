package main

import (
	"math"

	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
	"glonk.dev/widget"
)

const lampRange = 12.0 // how far the player's lamp lights the plant, in tiles

// screen is what the game is showing: the play itself, or a menu over it.
type screen int

const (
	title screen = iota
	playing
	paused
	dead
	won
)

// menus are the menu of each screen that has one. The first item always goes on with the
// game and the second always quits.
var menus = map[screen]struct {
	title string
	items []string
}{
	title:  {"RUSTWAKE", []string{"Start", "Quit"}},
	paused: {"paused", []string{"Resume", "Quit"}},
	dead:   {"you are scrap", []string{"Retry", "Quit"}},
	won:    {"the plant is quiet", []string{"Play again", "Quit"}},
}

type game struct {
	screen screen
	menu   widget.Menu

	level int // the index in levels of the map being played
	lvl   *level

	pos          geometry.Vector2 // the player, on the grid
	yaw          float64
	health, ammo int
	cool         float64 // seconds until the weapon can fire again
	flash        float64 // seconds left of the muzzle flash
	pain         float64 // seconds left of showing a hit

	enemies []*enemy
	bolts   []bolt
	pickups []spawn // the health and ammo still lying on the map
	spin    float64 // seconds played, which turns the pickups
}

// newGame returns the game main runs: the first map, behind the title menu.
func newGame() *game {
	g := &game{}
	g.restart(0)
	g.show(title)
	return g
}

// show changes the screen, with its menu at the first item.
func (g *game) show(s screen) {
	g.screen = s
	g.menu = widget.Menu{Items: menus[s].items, Style: hudStyle}
}

// restart begins map n with full health and the starting ammo.
func (g *game) restart(n int) {
	g.health, g.ammo = maxHealth, startAmmo
	g.level = n
	g.enter(parse(levels[n]))
}

// enter puts the player at the start of a map, with its machines and pickups. Health and
// ammo are kept, as they are from one map to the next.
func (g *game) enter(l *level) {
	g.lvl, g.pos, g.yaw = l, l.start, 0
	g.cool, g.flash, g.pain = 0, 0, 0
	g.enemies, g.bolts, g.pickups = nil, nil, nil
	for _, s := range l.spawns {
		if k, ok := kinds[s.kind]; ok {
			g.enemies = append(g.enemies, &enemy{kind: k, pos: s.at, hp: k.health})
		} else {
			g.pickups = append(g.pickups, s)
		}
	}
}

func (g *game) Update(dt float64, in *engine.Input) error {
	if g.screen != playing {
		return g.updateMenu(in)
	}
	if in.Pressed(engine.KeyEscape) {
		g.show(paused)
		return nil
	}

	g.updatePlayer(dt, in)
	g.updateEnemies(dt)
	g.updateBolts(dt)
	g.collect()
	g.spin += dt

	if g.health <= 0 {
		g.health = 0
		g.show(dead)
	} else if g.pos.Distance(g.lvl.exit) <= grabReach {
		if g.level+1 == len(levels) {
			g.show(won)
		} else {
			g.level++
			g.enter(parse(levels[g.level]))
		}
	}
	return nil
}

// updateMenu runs the menu of a screen while the world stands still.
func (g *game) updateMenu(in *engine.Input) error {
	choice, chosen := g.menu.Update(in)
	switch {
	case chosen && choice == 1:
		return engine.ErrQuit
	case chosen, g.screen == paused && in.Pressed(engine.KeyEscape):
		switch g.screen {
		case dead:
			g.restart(g.level)
		case won:
			g.restart(0)
		}
		g.screen = playing
	}
	return nil
}

func (g *game) Draw(fb *engine.FrameBuffer) {
	g.drawWorld(fb)
	g.drawHUD(fb)
	if g.screen != playing {
		g.drawMenu(fb)
	}
}

// drawWorld draws the view as pixels: the map by the light of the player's lamp, and of
// the muzzle flash while there is one, then the things in it, the crosshair and the weapon.
func (g *game) drawWorld(fb *engine.FrameBuffer) {
	c := fb.Canvas()
	cam := g.camera()
	lit := []mesh.Light{mesh.Ambient{Level: 0.2}, mesh.Point{Position: cam.Position, Range: lampRange}}
	if g.flash > 0 {
		lit = append(lit, mesh.Point{Position: cam.Position, Range: 4 * lampRange})
	}
	glow := mesh.Ambient{Level: 1} // for what gives light of its own

	g.lvl.mesh.DrawShadedPixels(c, cam, lit...)
	put(c, cam, exitModel, g.lvl.exit, 0, glow)
	for _, p := range g.pickups {
		put(c, cam, pickupModels[p.kind], p.at, 2*g.spin, glow)
	}
	for _, b := range g.bolts {
		put(c, cam, boltModel, b.pos, 0, glow)
	}
	for _, e := range g.enemies {
		turn := math.Atan2(g.pos.X-e.pos.X, g.pos.Y-e.pos.Y) // machines face the player
		if e.hurt > 0 {
			put(c, cam, e.model, e.pos, turn, glow)
		} else {
			put(c, cam, e.model, e.pos, turn, lit...)
		}
	}

	c.Set(c.Width()/2, c.Height()/2, engine.RGB(120, 255, 160)) // the crosshair
	g.drawWeapon(c)
	c.Compose()
}
