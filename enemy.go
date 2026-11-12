package main

import (
	"slices"

	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

const (
	boltSpeed  = 5.0 // tiles a second
	boltDamage = 15
	boltReach  = 0.35 // how near the player a bolt must pass to hit
)

// kind is what the machines of one sort have in common.
type kind struct {
	model  mesh.Mesh
	health int
	speed  float64 // tiles a second
	radius float64
	reach  float64 // how near the player must be for it to attack
	every  float64 // seconds between attacks
	damage int     // of a melee hit; a ranged kind fires a bolt instead
	ranged bool
}

// kinds are the machines, by the rune that puts one on a map.
var kinds = map[rune]kind{
	'c': {model: crawlerModel, health: 2, speed: 2.2, radius: 0.3, reach: 0.85, every: 0.8, damage: 10},
	's': {model: sentryModel, health: 3, speed: 1.2, radius: 0.22, reach: 7, every: 1.6, ranged: true},
}

type enemy struct {
	kind
	pos   geometry.Vector2
	hp    int
	awake bool             // it has seen the player
	goal  geometry.Vector2 // where it last saw the player
	wait  float64          // seconds until it may attack again
	hurt  float64          // seconds left of the flash of a hit
}

// bolt is a sentry's shot, slow enough to step out of the way of.
type bolt struct {
	pos, vel geometry.Vector2
}

// updateEnemies runs every machine: it sleeps until it sees the player, attacks when the
// player is in sight and in reach, and otherwise walks to where it last saw them.
func (g *game) updateEnemies(dt float64) {
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
		if sees && to.Length() <= e.reach {
			if e.wait <= 0 {
				e.wait = e.every
				if e.ranged {
					g.bolts = append(g.bolts, bolt{e.pos, to.Normalize().Scale(boltSpeed)})
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
