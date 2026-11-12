package main

import (
	"math"
	"slices"
	"strings"

	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
	"glonk.dev/tilemap"
)

// levels are the maps, in the order they are played. # and R are walls, @ is where the
// player starts, facing the top of the text, and E is the exit. c is a crawler, s a sentry,
// h health and a ammo. Every row of a map is the same width, with a wall all round.
var levels = []string{`
########################
#.....#......c...#....E#
#.....#..........#.....#
#.....#...RR.....#..s..#
#..a.........RR........#
#.....#..........#.....#
#@....#..c....h..#..a..#
########################
`, `
############################
#E..R....s.....#.....c....h#
#...R..........#...........#
#...R...####...#...RRRR....#
#..............#...RRRR....#
#s......#..#.......RRRR..c.#
#...R...#a.#...#...RRRR....#
#...R...####...#...........#
#...R..........#....a......#
#h..R....c...a.#.........@.#
############################
`}

const (
	wallHeight = 1.0
	eyeHeight  = 0.5
)

var (
	floorColor   = engine.RGB(96, 88, 80)
	ceilingColor = engine.RGB(58, 62, 74)
	exitColor    = engine.RGB(60, 230, 130)
	wallColors   = map[rune]engine.Color{'#': engine.RGB(124, 134, 156), 'R': engine.RGB(176, 92, 48)}
)

// walls are the tiles that stop a mover. They are never drawn as cells, so they have none.
var walls = map[rune]tilemap.Tile{'#': {Solid: true}, 'R': {Solid: true}}

// sides are the four neighbors of a tile: the step to each, and the two ends of the edge
// the tile shares with it, as offsets from the tile's corner.
var sides = [4]struct {
	dx, dy         int
	ax, ay, bx, by float64
}{
	{1, 0, 1, 0, 1, 1},
	{-1, 0, 0, 0, 0, 1},
	{0, 1, 0, 1, 1, 1},
	{0, -1, 0, 0, 1, 0},
}

// spawn is something a map puts on a tile: a machine or a pickup, named by its rune.
type spawn struct {
	kind rune
	at   geometry.Vector2
}

// level is one map, ready to play: the grid that movers are stopped by, the mesh that is
// drawn, and what starts where.
type level struct {
	grid   *tilemap.Map
	mesh   mesh.Mesh
	start  geometry.Vector2
	exit   geometry.Vector2
	spawns []spawn
}

// parse reads a map. The rows are turned over first, so that the bottom row of the text is
// y 0: a tile's (x, y) in the grid is then its (x, z) in the world, and facing +z is facing
// the top of the text with +x on the right, as the text reads.
func parse(text string) *level {
	rows := strings.Split(strings.Trim(text, "\n"), "\n")
	slices.Reverse(rows)
	l := &level{grid: tilemap.FromLines(walls, rows...)}

	for y, row := range rows {
		for x, r := range row {
			if _, wall := walls[r]; wall {
				continue
			}
			fx, fy := float64(x), float64(y)
			mid := geometry.Vector2{X: fx + 0.5, Y: fy + 0.5}
			floor, ceiling := floorColor, ceilingColor
			switch r {
			case '.':
			case '@':
				l.start = mid
			case 'E':
				l.exit = mid
				floor, ceiling = exitColor, exitColor
			default:
				l.spawns = append(l.spawns, spawn{r, mid})
			}

			l.slab(fx, fy, 0, geometry.Vector3{Y: 1}, floor)
			l.slab(fx, fy, wallHeight, geometry.Vector3{Y: -1}, ceiling)
			for _, s := range sides {
				nx, ny := x+s.dx, y+s.dy
				if !l.grid.Solid(nx, ny) {
					continue
				}
				color := wallColors['#'] // what the edge of the world is made of
				if ny >= 0 && ny < len(rows) && nx >= 0 && nx < len(rows[ny]) {
					color = wallColors[rune(rows[ny][nx])]
				}
				l.quad(
					geometry.Vector3{X: fx + s.ax, Z: fy + s.ay}, geometry.Vector3{X: fx + s.bx, Z: fy + s.by},
					geometry.Vector3{X: fx + s.bx, Y: wallHeight, Z: fy + s.by}, geometry.Vector3{X: fx + s.ax, Y: wallHeight, Z: fy + s.ay},
					geometry.Vector3{X: float64(-s.dx), Z: float64(-s.dy)}, color)
			}
		}
	}
	return l
}

// slab adds the floor or the ceiling of the tile at (x, y): a square at height h that looks
// along n.
func (l *level) slab(x, y, h float64, n geometry.Vector3, color engine.Color) {
	l.quad(
		geometry.Vector3{X: x, Y: h, Z: y}, geometry.Vector3{X: x + 1, Y: h, Z: y},
		geometry.Vector3{X: x + 1, Y: h, Z: y + 1}, geometry.Vector3{X: x, Y: h, Z: y + 1},
		n, color)
}

// quad adds a flat face with four corners that looks along n. The corners go round the
// face, and quad puts them in the order that faces n. The face has vertices of its own, with
// normals, so that its edges stay sharp and the lamp lights every pixel of it.
func (l *level) quad(a, b, c, d, n geometry.Vector3, color engine.Color) {
	if b.Sub(a).Cross(c.Sub(a)).Dot(n) < 0 {
		b, d = d, b
	}
	i := len(l.mesh.Vertices)
	l.mesh.Vertices = append(l.mesh.Vertices, a, b, c, d)
	l.mesh.Normals = append(l.mesh.Normals, n, n, n, n)
	l.mesh.Faces = append(l.mesh.Faces, [3]int{i, i + 1, i + 2}, [3]int{i, i + 2, i + 3})
	l.mesh.FaceColors = append(l.mesh.FaceColors, color, color)
}

// sees reports whether nothing solid lies on the straight line from a to b.
// shortcut: samples the line every 0.1 tile, so it can see past the very tip of a corner; walk the grid tile by tile if that ever shows.
func (l *level) sees(a, b geometry.Vector2) bool {
	n := int(a.Distance(b)/0.1) + 1
	for i := 0; i <= n; i++ {
		if l.solid(a.Lerp(b, float64(i)/float64(n))) {
			return false
		}
	}
	return true
}

// solid reports whether the point p is inside a wall.
func (l *level) solid(p geometry.Vector2) bool {
	return l.grid.Solid(int(math.Floor(p.X)), int(math.Floor(p.Y)))
}

// box is the square a mover of the given radius fills, for Map.Move.
func box(p geometry.Vector2, radius float64) geometry.Rect {
	return geometry.NewRect(p.X-radius, p.Y-radius, 2*radius, 2*radius)
}
