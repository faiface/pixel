package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// Init initializes OpenGL by loading the function pointers from the active OpenGL context.
// This function must be manually run inside the main thread (Do, DoErr, DoVal, etc.).
//
// It must be called under the presence of an active OpenGL context, e.g., always after calling
// window.MakeContextCurrent().  Also, always call this function when switching contexts.
func Init() {
	err := gl.Init()
	if err != nil {
		panic(err)
	}
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

// Clear clears the current framebuffer or window with the given color.
func Clear(r, g, b, a float32) {
	gl.ClearColor(r, g, b, a)
	gl.Clear(gl.COLOR_BUFFER_BIT)
}
