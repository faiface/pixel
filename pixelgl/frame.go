package pixelgl

import (
	"fmt"
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
)

type Frame struct {
	fb            binder
	tex           *Texture
	width, height int
}

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

func (f *Frame) Width() int {
	return f.width
}

func (f *Frame) Height() int {
	return f.height
}

func (f *Frame) Begin() {
	f.fb.bind()
}

func (f *Frame) End() {
	f.fb.restore()
}

func (f *Frame) Texture() *Texture {
	return f.tex
}
