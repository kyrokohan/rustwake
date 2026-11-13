package main

import (
	"slices"

	"glonk.dev/geometry"
)

const (
	boltSpeed    = 5.0 // tiles a second
	boltDamage   = 15
	boltReach    = 0.35 // how near the player a bolt must pass to hit
	volleySpread = 0.3  // radians between the bolts of a volley
	strikeTime   = 0.15 // seconds a machine shows its strike after it attacks
	windUpTime   = 0.25 // seconds before an attack that it shows the wind-up
	dieTime      = 0.5  // seconds a machine takes to fall apart
)

// The frames of a machine's strip: four of a walk, two of an attack, four of its end.
const (
	walkFrames  = 4
	windUpFrame = 4
	strikeFrame = 5
	dieFrame    = 6
	dieFrames   = 4
	allFrames   = 10
)

// kind is what the machines of one sort have in common.
type kind struct {
	look
	health int
	speed  float64 // tiles a second
	radius float64
	reach  float64 // how near the player must be for it to attack
	every  float64 // seconds between attacks
	damage int     // of a melee hit; a ranged kind fires bolts instead
	ranged bool
	volley int // how many bolts a ranged kind fires at once, spread out; 0 is one
}

// kinds are the machines, by the rune that puts one on a map.
var kinds = map[rune]kind{
	'c': {look: strip("crawler", allFrames, 0.7, 0), health: 2, speed: 2.2, radius: 0.3, reach: 0.85, every: 0.8, damage: 10},
	's': {look: strip("sentry", allFrames, 0.95, 0), health: 3, speed: 1.2, radius: 0.22, reach: 7, every: 1.6, ranged: true},
	'd': {look: strip("drone", allFrames, 0.42, 0.3), health: 1, speed: 3.4, radius: 0.2, reach: 0.6, every: 0.5, damage: 5},
	'H': {look: strip("hauler", allFrames, 0.98, 0), health: 8, speed: 1.0, radius: 0.38, reach: 1.0, every: 1.3, damage: 25},
	'B': {look: strip("overseer", allFrames, 0.98, 0), health: 40, speed: 0.7, radius: 0.45, reach: 9, every: 1.4, ranged: true, volley: 3},
}

type enemy struct {
	kind
	pos    geometry.Vector2
	hp     int
	awake  bool             // it has seen the player
	aiming bool             // the player is in sight and in reach
	goal   geometry.Vector2 // where it last saw the player
	wait   float64          // seconds until it may attack again
	hurt   float64          // seconds left of the flash of a hit
}

// frame is the picture of its strip a machine shows at a time on the game's clock: the
// strike just after an attack, the wind-up just before one, and otherwise its walk, which
// a sleeping machine stands still in.
func (e *enemy) frame(clock float64) int {
	switch {
	case e.wait > e.every-strikeTime:
		return strikeFrame
	case e.aiming && e.wait < windUpTime:
		return windUpFrame
	case e.awake:
		return int(clock*7) % walkFrames
	}
	return 0
}

// wreck is a machine that has been shot down: it falls apart where it stood, and what is
// left stays on the floor, in nobody's way.
type wreck struct {
	look
	pos geometry.Vector2
	age float64 // seconds since it was shot down
}

// frame is the picture a wreck shows, the last of them from dieTime on.
func (w wreck) frame() int {
	return dieFrame + min(dieFrames-1, int(w.age/dieTime*dieFrames))
}

// bolt is a sentry's shot, slow enough to step out of the way of.
type bolt struct {
	pos, vel geometry.Vector2
}

// updateEnemies runs every machine: it sleeps until it sees the player, attacks when the
// player is in sight and in reach, and otherwise walks to where it last saw them.
func (g *game) updateEnemies(dt float64) {
	for i := range g.wrecks {
		g.wrecks[i].age += dt
	}
	for _, e := range g.enemies {
		e.wait -= dt
		e.hurt -= dt
		sees := g.lvl.sees(e.pos, g.pos)
		if sees {
			if !e.awake {
				e.wait = e.every / 2 // time to react, so that turning a corner is not an instant hit
			}
			e.awake, e.goal = true, g.pos
		}
		if !e.awake {
			continue
		}

		to := e.goal.Sub(e.pos)
		e.aiming = sees && to.Length() <= e.reach
		if e.aiming {
			if e.wait <= 0 {
				e.wait = e.every
				if e.ranged {
					n := max(1, e.volley)
					for i := range n {
						aim := to.Normalize().Rotate(volleySpread * (float64(i) - float64(n-1)/2))
						g.bolts = append(g.bolts, bolt{e.pos, aim.Scale(boltSpeed)})
					}
				} else {
					g.damage(e.damage)
				}
			}
			continue
		}
		// shortcut: walks straight at the goal, so a machine can be left behind a corner; give it a path over the grid when a map has mazes.
		step := to.Normalize().Scale(min(e.speed*dt, to.Length()))
		moved, _, _ := g.lvl.grid.Move(box(e.pos, e.radius), step)
		e.pos = moved.Center()
	}
}

// updateBolts flies the bolts, and ends each at a wall or at the player, whom it hurts.
// shortcut: a bolt is tested where it lands, not along its way, which holds while a frame moves it less than boltReach; sweep it if the frame rate drops below 15.
func (g *game) updateBolts(dt float64) {
	for i := range g.bolts {
		g.bolts[i].pos = g.bolts[i].pos.Add(g.bolts[i].vel.Scale(dt))
	}
	g.bolts = slices.DeleteFunc(g.bolts, func(b bolt) bool {
		if b.pos.Distance(g.pos) < boltReach {
			g.damage(boltDamage)
			return true
		}
		return g.lvl.solid(b.pos)
	})
}
