package pixelgl

import (
	"fmt"
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
)

// Frame is a fixed resolution texture that you can draw on.
type Frame struct {
	fb            binder
	tex           *Texture
	width, height int
}

// NewFrame creates a new fully transparent Frame with given dimensions.
func NewFrame(width, height int, smooth bool) *Frame {
	f := &Frame{
		fb: binder{
			restoreLoc: gl.FRAMEBUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindFramebuffer(gl.FRAMEBUFFER, obj)
			},
		},
		width:  width,
		height: height,
	}

	gl.GenFramebuffers(1, &f.fb.obj)
	fmt.Println(f.fb.obj)

	f.tex = NewTexture(width, height, smooth, make([]uint8, width*height*4))

	f.fb.bind()
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, f.tex.tex.obj, 0)
	f.fb.restore()

	runtime.SetFinalizer(f, (*Frame).delete)

	return f
}

func (f *Frame) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteFramebuffers(1, &f.fb.obj)
	})
}

// Width returns the width of the Frame in pixels.
func (f *Frame) Width() int {
	return f.width
}

// Height returns the height of the Frame in pixels.
func (f *Frame) Height() int {
	return f.height
}

// Begin binds the Frame. All draw operations will target this Frame until End is called.
func (f *Frame) Begin() {
	f.fb.bind()
}

// End unbinds the Frame. All draw operations will go to whatever was bound before this Frame.
func (f *Frame) End() {
	f.fb.restore()
}

// Texture returns the Texture that this Frame draws to. The Texture changes as you use the Frame.
//
// The Texture pointer returned from this method is always the same.
func (f *Frame) Texture() *Texture {
	return f.tex
}
