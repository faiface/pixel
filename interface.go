package pixel

import "github.com/faiface/pixel/pixelgl"

// Target is an OpenGL graphics destination such as a window, a canvas, and so on. Something that
// you can draw on.
type Target interface {
	pixelgl.BeginEnder
	Shader() *pixelgl.Shader

	// MakeVertexArray returns a new vertex array drawable on the Target.
	MakeVertexArray(vertices []map[pixelgl.Attr]interface{}) *pixelgl.VertexArray
}

// Drawer is anything that can be drawn. It's by no means a drawer inside your table.
//
// Drawer consists of a single methods: Draw. Draw takes a target and any number of transform
// arguments. It's up to a Drawer to make sure that it draws correctly onto the provided target.
type Drawer interface {
	Draw(target Target, t ...Transform)
}
