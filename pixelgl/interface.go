package pixelgl

// BeginEnder is an interface for manipulating OpenGL state.
//
// OpenGL is a state machine. Every object can 'enter' it's state and 'leave' it's state. For
// example, you can bind a buffer and unbind a buffer, bind a texture and unbind it, use shader
// and unuse it, and so on.
type BeginEnder interface {
	Begin()
	End()
}
