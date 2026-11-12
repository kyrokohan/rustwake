package main

import (
	"math"
	"strings"
	"testing"

	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

const frame = 1.0 / 30

// corridor has a crawler next to the player, and a second one behind a wall beyond it. The
// player starts at (2.5, 1.5) facing both.
const corridor = `
#####
#.c.#
#.#.#
#.c.#
#.@.#
#####
`

// play returns a game in the middle of the given map, as if Start had been chosen.
func play(text string) *game {
	g := &game{screen: playing, health: maxHealth, ammo: startAmmo}
	g.enter(parse(text))
	return g
}

// hold runs the game for a time with the given keys held down.
func hold(t *testing.T, g *game, seconds float64, keys ...engine.Key) {
	t.Helper()
	kb := engine.NewKeyboard()
	kb.Press(keys...)
	for range int(seconds / frame) {
		if err := g.Update(frame, kb.Input()); err != nil {
			t.Fatal(err)
		}
		kb.EndFrame()
	}
}

func TestEveryLevelIsWellFormed(t *testing.T) {
	for n, text := range levels {
		rows := strings.Split(strings.Trim(text, "\n"), "\n")
		for _, row := range rows {
			if len(row) != len(rows[0]) {
				t.Errorf("level %d: row %q is not as wide as the first", n, row)
			}
		}
		if starts, exits := strings.Count(text, "@"), strings.Count(text, "E"); starts != 1 || exits != 1 {
			t.Errorf("level %d: %d starts and %d exits, wanted one of each", n, starts, exits)
		}

		l := parse(text)
		if err := l.mesh.Validate(); err != nil {
			t.Errorf("level %d: %v", n, err)
		}
		for _, s := range l.spawns {
			_, machine := kinds[s.kind]
			_, thing := pickupModels[s.kind]
			if !machine && !thing {
				t.Errorf("level %d: nothing is spawned by %q", n, s.kind)
			}
		}
	}
}

// Inside a closed map every pixel is a wall, a floor or a ceiling. An empty one is a face
// wound the wrong way round, or a gap between two faces.
func TestTheViewHasNoHoles(t *testing.T) {
	for n, text := range levels {
		l := parse(text)
		fb, _ := engine.NewFrameBuffer(96, 32)
		for turn := range 8 {
			fb.Clear()
			c := fb.Canvas()
			cam := mesh.Camera{
				Position: geometry.Vector3{X: l.start.X, Y: eyeHeight, Z: l.start.Y},
				Yaw:      float64(turn) * math.Pi / 4,
				FOV:      fov,
			}
			l.mesh.DrawShadedPixels(c, cam, mesh.Ambient{Level: 1})
			holes := 0
			for y := range c.Height() {
				for x := range c.Width() {
					if c.At(x, y) == (engine.Color{}) {
						holes++
					}
				}
			}
			if holes > 0 {
				t.Errorf("level %d, turn %d: %d empty pixels", n, turn, holes)
			}
		}
	}
}

func TestWallsStopThePlayerAndLetThemSlide(t *testing.T) {
	const room = `
#####
#...#
#@..#
#####
`
	g := play(room)
	hold(t, g, 3, 'w')
	if want := (geometry.Vector2{X: 1.5, Y: 3 - bodyRadius}); g.pos != want {
		t.Errorf("walking into the far wall: at %v, wanted %v", g.pos, want)
	}

	g = play(room)
	g.yaw = math.Pi / 4 // into the far wall at an angle, so along it to the corner
	hold(t, g, 3, 'w')
	if want := (geometry.Vector2{X: 4 - bodyRadius, Y: 3 - bodyRadius}); g.pos != want {
		t.Errorf("sliding along the far wall: at %v, wanted %v", g.pos, want)
	}
}

func TestAShotHitsTheNearestMachineInSight(t *testing.T) {
	g := play(corridor)
	near, far := g.enemies[0], g.enemies[1] // rows are read from the bottom of the text
	if near.pos.Y > far.pos.Y {
		t.Fatalf("near is at %v and far at %v", near.pos, far.pos)
	}
	if got := g.target(); got != near {
		t.Errorf("with both in line: aimed at %v, wanted the near one", got)
	}

	g.enemies = []*enemy{far}
	if got := g.target(); got != nil {
		t.Errorf("with a wall in the way: aimed at %v, wanted nothing", got)
	}

	g.yaw = math.Pi
	g.enemies = []*enemy{near}
	if got := g.target(); got != nil {
		t.Errorf("facing away: aimed at %v, wanted nothing", got)
	}
}

func TestFiringSpendsAmmoAndKills(t *testing.T) {
	g := play(corridor)
	hold(t, g, 2*fireEvery+frame, ' ') // long enough for the two shots a crawler takes
	if g.ammo != startAmmo-2-1 && g.ammo != startAmmo-2 {
		t.Errorf("ammo is %d after about two shots from %d", g.ammo, startAmmo)
	}
	if len(g.enemies) != 1 {
		t.Errorf("%d machines left, wanted the one behind the wall", len(g.enemies))
	}
}

func TestMachinesWakeOnSightAndHurt(t *testing.T) {
	g := play(corridor)
	near, far := g.enemies[0], g.enemies[1]
	hold(t, g, 3)
	if !near.awake || far.awake {
		t.Errorf("awake: near %v, far %v; wanted only the near one, which is in sight", near.awake, far.awake)
	}
	if g.health >= maxHealth {
		t.Errorf("health is still %d after three seconds beside a crawler", g.health)
	}

	hold(t, g, 30)
	if g.screen != dead || g.health != 0 {
		t.Errorf("after standing still: screen %v and health %d, wanted dead and 0", g.screen, g.health)
	}
}

func TestPickupsAndTheExit(t *testing.T) {
	g := play("#####\n#E..#\n#h..#\n#a..#\n#@..#\n#####")
	g.health = 50
	hold(t, g, 0.5, 'w')
	if g.ammo != startAmmo+10 {
		t.Errorf("ammo is %d after walking over ammo, wanted %d", g.ammo, startAmmo+10)
	}
	hold(t, g, 0.3, 'w')
	if g.health != 75 {
		t.Errorf("health is %d after walking over health from 50, wanted 75", g.health)
	}
	hold(t, g, 1, 'w')
	if g.level != 1 {
		t.Errorf("on level %d after reaching the exit, wanted the next one", g.level)
	}

	g.level = len(levels) - 1
	g.pos = g.lvl.exit
	hold(t, g, frame)
	if g.screen != won {
		t.Errorf("screen is %v at the last exit, wanted won", g.screen)
	}
}

// TestSmoke plays every map through the real Update and Draw with keys held, and prints the
// last frame of each: go test -run TestSmoke -v.
func TestSmoke(t *testing.T) {
	g := newGame()
	kb := engine.NewKeyboard()
	kb.Press(engine.KeyEnter) // Start
	if err := g.Update(frame, kb.Input()); err != nil || g.screen != playing {
		t.Fatalf("after choosing Start: screen %v, error %v", g.screen, err)
	}
	kb.EndFrame()

	fb, _ := engine.NewFrameBuffer(80, 24)
	for n := range levels {
		g.restart(n)
		g.screen = playing
		kb.Press('w', 'e', ' ')
		for range 300 {
			if err := g.Update(frame, kb.Input()); err != nil {
				t.Fatal(err)
			}
			kb.EndFrame()
			fb.Clear()
			g.Draw(fb)
		}
		if strings.TrimSpace(fb.String()) == "" {
			t.Errorf("level %d: the frame is blank", n)
		}
		t.Logf("level %d\n%s", n, fb.String())
	}
}

func BenchmarkDraw(b *testing.B) {
	g := newGame()
	g.restart(len(levels) - 1) // the larger map
	fb, _ := engine.NewFrameBuffer(200, 60)
	for b.Loop() {
		fb.Clear()
		g.Draw(fb)
	}
}
