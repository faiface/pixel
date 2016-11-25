package pixelgl

// Doer is an interface for manipulating OpenGL state.
//
// OpenGL is a state machine. Every object can 'enter' it's state and 'leave' it's state. For example,
// you can bind a buffer and unbind a buffer, bind a texture and unbind it, use shader and unuse it, and so on.
//
// This interface provides a clever and flexible way to do it. A typical workflow of an OpenGL object is that
// you enter (load, bind) that object's state, then do something with it, and then leave the state. That 'something'
// in between, let's call it sub (as in subroutine).
//
// The recommended way to implement a Doer is to wrap another Doer (vertex array wraps texture and so on), let's call
// it parent. Then the Do method will look like this:
//
//   func (o *MyObject) Do(sub func()) {
//       o.parent.Do(func() {
//	     // enter the object's state
//           sub()
//           // leave the object's state
//       })
//   }
//
// It might seem difficult to grasp this kind of recursion at first, but it's really simple. What it's basically saying
// is: "Hey parent, enter your state, then let me enter mine, then I'll do whatever I'm supposed to do in the middle.
// After that I'll leave my state and please leave your state too parent."
type Doer interface {
	Do(sub func())
}
