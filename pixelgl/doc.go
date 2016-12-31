// Package pixelgl provides some abstractions around the basic OpenGL primitives and
// operations.
//
// All calls should be wrapped inside pixelgl.Do/DoNoBlock/DoErr/DoVal.
//
// This package deliberately does not handle nor report OpenGL errors, it's up to you to
// cause none.
package pixelgl
