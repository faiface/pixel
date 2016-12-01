package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// This file defines functions that can operate without a parent Doer.

// Clear clears the current context..
func Clear(r, g, b, a float32) {
	DoNoBlock(func() {
		gl.ClearColor(r, g, b, a)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}
