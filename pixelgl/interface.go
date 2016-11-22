package pixelgl

// BeginEnder is an interface for manipulating OpenGL state.
//
// OpenGL is a state machine and as such, it is natural to manipulate it in a begin-end manner.
// This interface is intended for all OpenGL objects, that can begin being active and end being active
// such as windows, vertex arrays, vertex buffers, textures, shaders, pretty much everything.
//
// It might seem natural to use BeginEnders this way:
//
//   window.Begin()
//   shader.Begin()
//   texture.Begin()
//   vertexarray.Begin()
//   vertexarray.Draw()
//   vertexarray.End()
//   texture.End()
//   shader.End()
//   window.End()
//
// Don't do this! A better practice is to make a BeginEnder so that it wraps another BeginEnder like this:
//
//   shader = NewShader(window)
//   texture = NewTexture(shader)
//   vertexarray = NewVertexArray(texture)
//   // now, somewhere else in your code, instead of calling numerous Begin/Ends, you just call
//   vertexarray.Draw()
//
// The final single call to draw a vertex array executes all of the Begins and Ends, because the objects are
// wrapped around each other.
type BeginEnder interface {
	Begin()
	End()
}
