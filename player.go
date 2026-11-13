package main

import (
	"math"
	"slices"
	"unicode"

	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

const (
	maxHealth  = 100
	startAmmo  = 20
	walkSpeed  = 3.5  // tiles a second
	turnSpeed  = 2.4  // radians a second
	bodyRadius = 0.25 // farther than the corners of the near plane, so the camera never looks through a wall
	fov        = 1.1  // up and down, in radians; about 95 degrees across in a window twice as wide as tall
	fireEvery  = 0.35 // seconds between shots
	flashTime  = 0.08 // seconds the muzzle flash lasts
	painTime   = 0.25 // seconds the screen shows a hit
	grabReach  = 0.5  // how near a pickup or the exit must be
)

// down reports whether a letter key is held, in either case: a terminal that cannot report
// keys exactly sends a capital with Shift or Caps Lock.
func down(in *engine.Input, letter rune) bool {
	return in.IsDown(engine.Key(letter)) || in.IsDown(engine.Key(unicode.ToUpper(letter)))
}

// camera is the player's view.
func (g *game) camera() mesh.Camera {
	return mesh.Camera{Position: geometry.Vector3{X: g.pos.X, Y: eyeHeight, Z: g.pos.Y}, Yaw: g.yaw, FOV: fov}
}

// facing is the way the player looks, on the grid.
func (g *game) facing() geometry.Vector2 {
	return geometry.Vector2{X: math.Sin(g.yaw), Y: math.Cos(g.yaw)}
}

// updatePlayer turns, walks and fires from the keys held.
func (g *game) updatePlayer(dt float64, in *engine.Input) {
	g.cool -= dt
	g.flash -= dt
	g.pain -= dt

	if down(in, 'e') || in.IsDown(engine.KeyRight) {
		g.yaw += turnSpeed * dt
	}
	if down(in, 'q') || in.IsDown(engine.KeyLeft) {
		g.yaw -= turnSpeed * dt
	}

	forward := g.facing()
	right := geometry.Vector2{X: forward.Y, Y: -forward.X}
	var move geometry.Vector2
	if down(in, 'w') {
		move = move.Add(forward)
	}
	if down(in, 's') {
		move = move.Sub(forward)
	}
	if down(in, 'd') {
		move = move.Add(right)
	}
	if down(in, 'a') {
		move = move.Sub(right)
	}
	// Move stops the box at walls and slides it along them, however long the frame took.
	moved, _, _ := g.lvl.grid.Move(box(g.pos, bodyRadius), move.Normalize().Scale(walkSpeed*dt))
	g.walked += moved.Center().Distance(g.pos)
	g.pos = moved.Center()

	if in.IsDown(' ') && g.cool <= 0 && g.ammo > 0 {
		g.ammo--
		g.cool, g.flash = fireEvery, flashTime
		if e := g.target(); e != nil {
			e.hp--
			e.hurt = 0.1
			if e.hp <= 0 {
				g.enemies = slices.DeleteFunc(g.enemies, func(o *enemy) bool { return o == e })
				g.wrecks = append(g.wrecks, wreck{look: e.look, pos: e.pos})
			}
		}
	}
}

// target returns the nearest machine on the line of fire that no wall hides, or nil.
func (g *game) target() *enemy {
	forward := g.facing()
	var best *enemy
	nearest := math.Inf(1)
	for _, e := range g.enemies {
		to := e.pos.Sub(g.pos)
		along := to.Dot(forward)
		if along <= 0 || along >= nearest || to.Perp(forward).Length() > e.radius || !g.lvl.sees(g.pos, e.pos) {
			continue
		}
		best, nearest = e, along
	}
	return best
}

// damage takes n from the player's health and shows the hit.
func (g *game) damage(n int) {
	g.health -= n
	g.pain = painTime
}

// collect picks up what the player stands on. Health is left where it lies when none is needed.
func (g *game) collect() {
	g.pickups = slices.DeleteFunc(g.pickups, func(p spawn) bool {
		if p.at.Distance(g.pos) > grabReach {
			return false
		}
		switch p.kind {
		case 'h':
			if g.health >= maxHealth {
				return false
			}
			g.health = min(maxHealth, g.health+25)
		case 'a':
			g.ammo += 10
		}
		return true
	})
}
