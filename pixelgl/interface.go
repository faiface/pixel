package pixelgl

// Doer is an interface for manipulating OpenGL state.
//
// OpenGL is a state machine. Every object can 'enter' it's state and 'leave' it's state. For
// example, you can bind a buffer and unbind a buffer, bind a texture and unbind it, use shader
// and unuse it, and so on.
//
// This interface provides a clever and flexible way to do it. A typical workflow of an OpenGL
// object is that you enter (load, bind) that object's state, then do something with it, and
// then leave the state. That 'something' in between, let's call it sub (as in subroutine).
//
// The recommended way to implement a Doer is to wrap another Doer (vertex array wraps texture
// and so on), let's call it parent. Then the Do method will look like this:
//
//   func (o *MyObject) Do(sub func(Context)) {
//       o.parent.Do(func(ctx Context) {
//	     // enter the object's state
//           sub(ctx)
//           // leave the object's state
//       })
//   }
//
// It might seem difficult to grasp this kind of recursion at first, but it's really simple. What
// it's basically saying is: "Hey parent, enter your state, then let me enter mine, then I'll
// do whatever I'm supposed to do in the middle.  After that I'll leave my state and please
// leave your state too parent."
//
// Also notice, that the functions are passing a Context around. This context contains the
// most important state variables.  Usually, you just pass it as you received it. If you want
// to pass a changed context to your child (e.g. your a shader), use ctx.With* methods.
//
// If possible and makes sense, Do method should be reentrant.
type Doer interface {
	Do(sub func(Context))
}

// Context takes state from one object to another. OpenGL is a state machine, so we have
// to approach it like that.  However, global variables are evil, so we have Context, that
// transfers important OpenGL state from one object to another.
//
// This type does *not* represent an OpenGL context in the OpenGL terminology.
type Context struct {
	shader *Shader
}

// Shader returns the current shader.
func (c Context) Shader() *Shader {
	return c.shader
}

// WithShader returns a copy of this context with the specified shader.
func (c Context) WithShader(s *Shader) Context {
	return Context{
		shader: s,
	}
}

// ContextHolder is a root Doer with no parent. It simply forwards a context to a child.
type ContextHolder struct {
	Context Context
}

// Do calls sub and passes it the held context.
func (ch ContextHolder) Do(sub func(ctx Context)) {
	sub(ch.Context)
}

type noOpDoer struct{}

func (noOpDoer) Do(sub func(ctx Context)) {
	sub(Context{})
}

// NoOpDoer is a Doer that just passes an empty context to the caller of Do.
var NoOpDoer Doer = noOpDoer{}
