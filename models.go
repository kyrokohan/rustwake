package main

import (
	"glonk.dev/engine"
	"glonk.dev/geometry"
	"glonk.dev/mesh"
)

// The models stand on the floor at the origin and face +z. They have no normals, so each
// face is lit as a whole and the machines look cut from plate.
var (
	crawlerModel = join(
		part(mesh.Cube(0.6), 0, 0.3, 0, engine.RGB(214, 96, 44)),      // taller than the eye, so the crosshair lands on it
		part(mesh.Cube(0.16), 0, 0.4, 0.26, engine.RGB(255, 226, 90)), // its eye
	)
	sentryModel = join(
		part(mesh.Cylinder(0.16, 0.6, 8), 0, 0.3, 0, engine.RGB(110, 124, 150)),
		part(mesh.Sphere(0.2, 8, 5), 0, 0.72, 0, engine.RGB(70, 200, 220)),
		part(mesh.Cube(0.1), 0, 0.72, 0.17, engine.RGB(255, 60, 60)), // its eye
	)
	boltModel = join(part(mesh.Sphere(0.07, 6, 4), 0, 0.45, 0, engine.RGB(255, 120, 255)))
	exitModel = join(part(mesh.Cylinder(0.05, wallHeight, 6), 0, wallHeight/2, 0, exitColor))

	pickupModels = map[rune]mesh.Mesh{
		'h': join(part(mesh.Cube(0.2), 0, 0.25, 0, engine.RGB(80, 230, 110))),
		'a': join(part(mesh.Cylinder(0.07, 0.26, 6), 0, 0.25, 0, engine.RGB(250, 210, 70))),
	}
)

// part returns a shape moved to (x, y, z) in its model, in one color.
func part(m mesh.Mesh, x, y, z float64, color engine.Color) mesh.Mesh {
	m.Transform.Position = geometry.Vector3{X: x, Y: y, Z: z}
	m.Material.Color = color
	return m
}

// join makes one mesh of the parts, each moved to its Transform's Position and in its
// Material's color, so that a model is placed and drawn in one go.
func join(parts ...mesh.Mesh) mesh.Mesh {
	var out mesh.Mesh
	for _, p := range parts {
		base := len(out.Vertices)
		for _, v := range p.Vertices {
			out.Vertices = append(out.Vertices, v.Add(p.Transform.Position))
		}
		for _, f := range p.Faces {
			out.Faces = append(out.Faces, [3]int{f[0] + base, f[1] + base, f[2] + base})
			out.FaceColors = append(out.FaceColors, p.Material.Color)
		}
	}
	return out
}

// put draws a model standing at p, turned by yaw.
func put(c *engine.Canvas, cam mesh.Camera, m mesh.Mesh, p geometry.Vector2, yaw float64, lights ...mesh.Light) {
	m.Transform = mesh.Transform{Position: geometry.Vector3{X: p.X, Z: p.Y}, Yaw: yaw}
	m.DrawShadedPixels(c, cam, lights...)
}
