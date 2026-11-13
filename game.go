package main

import (
	"cmp"
	"math"
	"slices"

	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
	"glonk.dev/widget"
)

const lampRange = 12.0 // how far the player's lamp lights the plant, in tiles

// screen is what the game is showing: the play itself, the player's death, or a menu
// over the view.
type screen int

const (
	title screen = iota
	playing
	paused
	dying // the scene between the last hit and the game over menu
	dead
	won
)

// menus are the menu of each screen that has one. The first item always goes on with the
// game and the second always quits. The title screen has the game's name as a picture
// over its menu, and so no title on it.
var menus = map[screen]struct {
	title string
	items []string
}{
	title:  {"", []string{"Start", "Quit"}},
	paused: {"Paused", []string{"Resume", "Quit"}},
	dead:   {"Scrapped", []string{"Retry", "Quit"}},
	won:    {"Plant cleared", []string{"Play again", "Quit"}},
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
	wrecks  []wreck
	bolts   []bolt
	pickups []spawn // the health and ammo still lying on the map

	clock  float64 // seconds played, which the animations run on
	walked float64 // tiles walked, which the weapon bobs to

	menuClock float64 // seconds the menu on the screen has been up, which its mark glints on
	kills     int     // machines wrecked since the last start, for the stats
	played    float64 // seconds played since the last start, for the stats

	fall float64 // seconds since the hit that killed the player; 0 while they live
	gore gore    // what that hit put on the glass
}

// newGame returns the game main runs: map n, behind the title menu.
func newGame(n int) *game {
	g := &game{}
	g.restart(n)
	g.show(title)
	return g
}

// show changes the screen, with its menu at the first item. The menu is glonk's for the
// keys only: the game draws its own.
func (g *game) show(s screen) {
	g.screen, g.menuClock = s, 0
	g.menu = widget.Menu{Items: menus[s].items}
}

// restart begins map n with full health and the starting ammo, and the stats at nothing.
func (g *game) restart(n int) {
	g.health, g.ammo = maxHealth, startAmmo
	g.kills, g.played = 0, 0
	g.load(n)
}

// load begins map n as the player is.
func (g *game) load(n int) {
	g.level = n
	g.enter(parse(levels[n].text, zoneOf(n)))
}

// enter puts the player at the start of a map, with its machines and pickups. Health and
// ammo are kept, as they are from one map to the next.
func (g *game) enter(l *level) {
	g.lvl, g.pos, g.yaw = l, l.start, 0
	g.cool, g.flash, g.pain, g.fall = 0, 0, 0, 0
	g.enemies, g.wrecks, g.bolts, g.pickups = nil, nil, nil, nil
	for _, s := range l.spawns {
		if k, ok := kinds[s.kind]; ok {
			g.enemies = append(g.enemies, &enemy{kind: k, pos: s.at, hp: k.health})
		} else {
			g.pickups = append(g.pickups, s)
		}
	}
}

func (g *game) Update(dt float64, in *engine.Input) error {
	switch g.screen {
	case playing:
	case dying:
		g.updateDying(dt, in)
		return nil
	default:
		return g.updateMenu(dt, in)
	}
	if in.Pressed(engine.KeyEscape) {
		g.show(paused)
		return nil
	}

	g.updatePlayer(dt, in)
	g.updateEnemies(dt)
	g.updateBolts(dt)
	g.collect()
	g.clock += dt
	g.played += dt

	if g.health <= 0 {
		// The same death for the same game, so that a frame of it can be checked.
		g.screen, g.gore = dying, newGore(uint64(g.level)<<16|uint64(g.kills))
	} else if g.pos.Distance(g.lvl.exit) <= grabReach {
		if g.level+1 == len(levels) {
			g.show(won)
		} else {
			g.load(g.level + 1)
		}
	}
	return nil
}

// updateDying runs the plant on while the player goes down: the machines and the bolts
// move and the player does not. The game over menu comes when the scene is over, or at a
// press of Enter or Space once its start has been seen.
func (g *game) updateDying(dt float64, in *engine.Input) {
	g.fall += dt
	g.clock += dt
	g.pain -= dt
	g.updateEnemies(dt)
	g.updateBolts(dt)
	if g.fall >= deathTime || g.fall >= skipAfter && (in.Pressed(engine.KeyEnter) || in.Pressed(' ')) {
		g.show(dead)
	}
}

// updateMenu runs the menu of a screen while the world stands still. Behind the title the
// view turns slowly, and behind the game over the blood goes on running.
func (g *game) updateMenu(dt float64, in *engine.Input) error {
	g.menuClock += dt
	switch g.screen {
	case title:
		g.yaw += 0.15 * dt
	case dead:
		g.fall += dt
	}

	choice, chosen := g.menu.Update(in)
	switch {
	case chosen && choice == 1:
		return engine.ErrQuit
	case chosen, g.screen == paused && in.Pressed(engine.KeyEscape):
		switch g.screen {
		case title:
			g.yaw = 0 // the way the map was drawn to be walked into
		case dead:
			g.restart(g.level)
		case won:
			g.restart(0)
		}
		g.screen = playing
	}
	return nil
}

// Draw draws the view, then what is on the glass in front of it, then the status strip
// and the menu, all as pixels; and then, in cells over them, the numbers and the keys.
func (g *game) Draw(fb *engine.FrameBuffer) {
	c := fb.Canvas()
	g.drawWorld(c)
	if g.fall > 0 {
		g.gore.draw(c, g.fall)
	}
	g.drawStatus(c)
	_, menu := menus[g.screen]
	under := 0
	if menu {
		under = g.drawMenu(c)
	}
	c.Compose()

	g.drawHUD(fb)
	if menu {
		drawHints(fb, under/2+1)
	}
}

// drawWorld draws the view: the map by the light of the player's lamp, and of the muzzle
// flash while there is one, then the things in it, the crosshair and the weapon.
func (g *game) drawWorld(c *engine.Canvas) {
	cam := g.camera()
	lit := []mesh.Light{mesh.Ambient{Level: 0.2}, mesh.Point{Position: cam.Position, Range: lampRange}}
	if g.flash > 0 {
		lit = append(lit, mesh.Point{Position: cam.Position, Range: 4 * lampRange})
	}

	// The three nearest lamps light the plant too. Every light is asked about every pixel,
	// so more of them would cost more than they show.
	slices.SortFunc(g.lvl.lamps, func(a, b geometry.Vector2) int { return cmp.Compare(a.Distance(g.pos), b.Distance(g.pos)) })
	for _, at := range g.lvl.lamps[:min(3, len(g.lvl.lamps))] {
		lit = append(lit, mesh.Point{Position: geometry.Vector3{X: at.X, Y: wallHeight - 0.25, Z: at.Y}, Intensity: 0.85, Range: 6})
	}

	// The exit, the pickups, the bolts and the lamps give light of their own, so they are
	// drawn as bright as they are. The rest is lit as the walls are, but for a machine in
	// the flash of a hit.
	g.lvl.draw(c, cam, lit...)
	for _, p := range g.lvl.props {
		bright := 1.0
		if p.kind != 'l' {
			bright = brightness(lit, cam.Position, p.at)
		}
		stand(c, cam, propLooks[p.kind], 0, p.at, bright)
	}
	pulse := int(g.clock*3) % 2
	stand(c, cam, exitLook, pulse, g.lvl.exit, 1)
	for _, p := range g.pickups {
		stand(c, cam, pickupLooks[p.kind], pulse, p.at, 1)
	}
	for _, b := range g.bolts {
		stand(c, cam, boltLook, int(g.clock*12)%2, b.pos, 1)
	}
	for _, w := range g.wrecks {
		w.lift *= max(0, 1-w.age/dieTime) // what flew comes down
		stand(c, cam, w.look, w.frame(), w.pos, max(brightness(lit, cam.Position, w.pos), 1-w.age/dieTime))
	}
	for _, e := range g.enemies {
		bright := 1.0
		if e.hurt <= 0 {
			bright = brightness(lit, cam.Position, e.pos)
		}
		l := e.look
		if l.lift > 0 { // what flies bobs
			l.lift += 0.05 * math.Sin(g.clock*5+e.pos.X*3)
		}
		stand(c, cam, l, e.frame(g.clock), e.pos, bright)
	}

	if g.fall == 0 {
		c.Set(c.Width()/2, c.Height()/2, engine.RGB(120, 255, 160)) // the crosshair
	}
	g.drawWeapon(c)
}
