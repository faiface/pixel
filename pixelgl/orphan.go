package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// Clear clears the current framebuffer or window with the given color.
func Clear(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}

// Viewport sets the OpenGL viewport.
func Viewport(x, y, w, h int32) {
	gl.Viewport(x, y, w, h)
}
