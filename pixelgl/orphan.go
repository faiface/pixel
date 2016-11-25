package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// This file defines functions that can operate without a parent Doer.

// Clear clears the current context..
func Clear(r, g, b, a float64) {
	DoNoBlock(func() {
		gl.ClearColor(float32(r), float32(g), float32(b), float32(a))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}
