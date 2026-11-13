package main

import (
	"math"
	"slices"
	"strings"

	"glonk.dev/bitmap"
	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
	"glonk.dev/tilemap"
)

// levels are the maps, in the order they are played. # % and & are the three walls of the
// map's zone, @ is where the player starts, facing the top of the text, and E is the exit.
// c is a crawler, s a sentry, h health and a ammo. Every row of a map is the same width,
// with a wall all round.
var levels = []string{`
########################
#.....#......c...#....E#
#.....#..........#.....#
#.....&...%%.....&..s..#
#..a.........%%........#
#.....&..........&.....#
#@....#..c....h..#..a..#
########################
`, `
############################
#E..%....s.....#.....c....h#
#...%..........#...........#
#...%...&&&&...#...%%%%....#
#..............&...%%%%....#
#s......&..&.......%%%%..c.#
#...%...&a.&...&...%%%%....#
#...%...&&&&...#...........#
#...%..........#....a......#
#h..%....c...a.#.........@.#
############################
`}

const (
	wallHeight = 1.0
	eyeHeight  = 0.5

	chunkSize  = 8   // tiles along a side of a chunk
	chunkReach = 6.0 // how far a chunk's corners are from its middle, and a little more
	fineRange  = 9.0 // how near a chunk's middle must be for its fine mesh to be drawn
)

// zone is the look of a stretch of the plant: the pictures of its three kinds of wall, of
// its floor and of its ceiling.
type zone struct {
	walls          map[rune]*bitmap.Bitmap
	floor, ceiling *bitmap.Bitmap
}

// zones are the looks, in the order the maps go through them, and exitPad is the floor
// and the ceiling of the way out in all of them.
var (
	zones   = []zone{newZone("1"), newZone("2"), newZone("3")}
	exitPad = art("floor-exit")
)

func newZone(n string) zone {
	return zone{
		walls:   map[rune]*bitmap.Bitmap{'#': art("wall-" + n + "a"), '%': art("wall-" + n + "b"), '&': art("wall-" + n + "c")},
		floor:   art("floor-" + n),
		ceiling: art("ceiling-" + n),
	}
}

// walls are the tiles that stop a mover. They are never drawn as cells, so they have none.
var walls = map[rune]tilemap.Tile{'#': {Solid: true}, '%': {Solid: true}, '&': {Solid: true}}

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

// chunk is a square piece of a map, drawn or skipped as a whole: a map in one mesh would
// be drawn whole every frame, the part behind the player with the rest. Fine has a face for
// every pixel of the pictures on its surfaces, and coarse a face for every four, for when
// the chunk is too far away to tell them apart. Their vertices are measured from middle,
// where the meshes are placed, which is what lets glonk tell that one is off the screen.
type chunk struct {
	middle       geometry.Vector2
	fine, coarse mesh.Mesh
}

// level is one map, ready to play: the grid that movers are stopped by, the chunks that
// are drawn, and what starts where.
type level struct {
	grid   *tilemap.Map
	chunks []chunk
	across int // chunks in a row of them
	start  geometry.Vector2
	exit   geometry.Vector2
	spawns []spawn
}

// parse reads a map, in the look of a zone. The rows are turned over first, so that the
// bottom row of the text is y 0: a tile's (x, y) in the grid is then its (x, z) in the
// world, and facing +z is facing the top of the text with +x on the right, as the text
// reads.
func parse(text string, z zone) *level {
	rows := strings.Split(strings.Trim(text, "\n"), "\n")
	slices.Reverse(rows)
	l := &level{grid: tilemap.FromLines(walls, rows...)}
	l.across = (l.grid.Width() + chunkSize - 1) / chunkSize
	l.chunks = make([]chunk, l.across*((l.grid.Height()+chunkSize-1)/chunkSize))
	for i := range l.chunks {
		c := &l.chunks[i]
		c.middle = geometry.Vector2{X: float64(i%l.across*chunkSize) + chunkSize/2, Y: float64(i/l.across*chunkSize) + chunkSize/2}
		c.fine.Transform.Position = geometry.Vector3{X: c.middle.X, Z: c.middle.Y}
		c.coarse.Transform.Position = c.fine.Transform.Position
	}

	for y, row := range rows {
		for x, r := range row {
			if _, wall := walls[r]; wall {
				continue
			}
			fx, fy := float64(x), float64(y)
			mid := geometry.Vector2{X: fx + 0.5, Y: fy + 0.5}
			floor, ceiling := z.floor, z.ceiling
			switch r {
			case '.':
			case '@':
				l.start = mid
			case 'E':
				l.exit = mid
				floor, ceiling = exitPad, exitPad
			default:
				l.spawns = append(l.spawns, spawn{r, mid})
			}

			c := &l.chunks[y/chunkSize*l.across+x/chunkSize]
			c.face(geometry.Vector3{X: fx, Z: fy + 1}, geometry.Vector3{X: 1}, geometry.Vector3{Z: -1}, geometry.Vector3{Y: 1}, floor)
			c.face(geometry.Vector3{X: fx, Y: wallHeight, Z: fy}, geometry.Vector3{X: 1}, geometry.Vector3{Z: 1}, geometry.Vector3{Y: -1}, ceiling)
			for _, s := range sides {
				nx, ny := x+s.dx, y+s.dy
				if !l.grid.Solid(nx, ny) {
					continue
				}
				wall := z.walls['#'] // what the edge of the world is made of
				if ny >= 0 && ny < len(rows) && nx >= 0 && nx < len(rows[ny]) {
					wall = z.walls[rune(rows[ny][nx])]
				}
				c.face(
					geometry.Vector3{X: fx + s.ax, Y: wallHeight, Z: fy + s.ay},
					geometry.Vector3{X: s.bx - s.ax, Z: s.by - s.ay},
					geometry.Vector3{Y: -wallHeight},
					geometry.Vector3{X: float64(-s.dx), Z: float64(-s.dy)}, wall)
			}
		}
	}
	return l
}

// face adds a flat surface that looks along n, with a picture on it, to both meshes of a
// chunk. The picture's top left pixel is at origin; across runs along its rows and down
// along its columns, each the whole length of the surface.
func (c *chunk) face(origin, across, down, n geometry.Vector3, pic *bitmap.Bitmap) {
	origin = origin.Sub(c.fine.Transform.Position)
	lay(&c.fine, origin, across, down, n, pic, 1)
	lay(&c.coarse, origin, across, down, n, pic, 2)
}

// lay adds a surface to a mesh. glonk gives a face of a mesh one color, so the surface is
// a grid of small faces that share their corners, one for each square of step by step
// pixels of the picture, in the color those pixels come to when mixed. The corners have
// normals, so the lamp lights every pixel of the screen.
func lay(m *mesh.Mesh, origin, across, down, n geometry.Vector3, pic *bitmap.Bitmap, step int) {
	w, h := pic.Width()/step, pic.Height()/step
	first := len(m.Vertices)
	for j := 0; j <= h; j++ {
		for i := 0; i <= w; i++ {
			at := origin.Add(across.Scale(float64(i) / float64(w))).Add(down.Scale(float64(j) / float64(h)))
			m.Vertices = append(m.Vertices, at)
			m.Normals = append(m.Normals, n)
		}
	}
	turned := across.Cross(down).Dot(n) < 0 // wound this way round it would look away from n
	for j := range h {
		for i := range w {
			a := first + j*(w+1) + i
			b, c, d := a+1, a+w+2, a+w+1 // the corners of the square, going round from a
			if turned {
				b, d = d, b
			}
			var red, green, blue int
			for k := range step * step {
				r, g, b := pic.At(i*step+k%step, j*step+k/step).RGB()
				red, green, blue = red+int(r), green+int(g), blue+int(b)
			}
			color := engine.RGB(uint8(red/(step*step)), uint8(green/(step*step)), uint8(blue/(step*step)))
			m.Faces = append(m.Faces, [3]int{a, b, c}, [3]int{a, c, d})
			m.FaceColors = append(m.FaceColors, color, color)
		}
	}
}

// draw draws the map as it is seen from a camera: the chunks that are not wholly behind
// it, the near ones fine and the rest coarse.
func (l *level) draw(c *engine.Canvas, cam mesh.Camera, lights ...mesh.Light) {
	eye, forward := geometry.Vector2{X: cam.Position.X, Y: cam.Position.Z}, cam.Forward()
	for i := range l.chunks {
		to := l.chunks[i].middle.Sub(eye)
		switch {
		case to.X*forward.X+to.Y*forward.Z < -chunkReach:
		case to.Length() < fineRange:
			l.chunks[i].fine.DrawShadedPixels(c, cam, lights...)
		default:
			l.chunks[i].coarse.DrawShadedPixels(c, cam, lights...)
		}
	}
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
