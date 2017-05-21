package pixelgl

import (
	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
)

// GLFrame is a type that helps implementing OpenGL Targets. It implements most common methods to
// avoid code redundancy. It contains an glhf.Frame that you can draw on.
type GLFrame struct {
	frame  *glhf.Frame
	bounds pixel.Rect
	pixels []uint8
	dirty  bool
}

// NewGLFrame creates a new GLFrame with the given bounds.
func NewGLFrame(bounds pixel.Rect) *GLFrame {
	gf := new(GLFrame)
	gf.SetBounds(bounds)
	return gf
}

// SetBounds resizes the GLFrame to the new bounds.
func (gf *GLFrame) SetBounds(bounds pixel.Rect) {
	if bounds == gf.Bounds() {
		return
	}

	mainthread.Call(func() {
		oldF := gf.frame

		_, _, w, h := intBounds(bounds)
		gf.frame = glhf.NewFrame(w, h, false)

		// preserve old content
		if oldF != nil {
			ox, oy, ow, oh := intBounds(bounds)
			oldF.Blit(
				gf.frame,
				ox, oy, ox+ow, oy+oh,
				ox, oy, ox+ow, oy+oh,
			)
		}
	})

	gf.bounds = bounds
	gf.pixels = nil
	gf.dirty = true
}

// Bounds returns the current GLFrame's bounds.
func (gf *GLFrame) Bounds() pixel.Rect {
	return gf.bounds
}

// Color returns the color of the pixel under the specified position.
func (gf *GLFrame) Color(at pixel.Vec) pixel.RGBA {
	if gf.dirty {
		mainthread.Call(func() {
			tex := gf.frame.Texture()
			tex.Begin()
			gf.pixels = tex.Pixels(0, 0, tex.Width(), tex.Height())
			tex.End()
		})
		gf.dirty = false
	}
	if !gf.bounds.Contains(at) {
		return pixel.Alpha(0)
	}
	bx, by, bw, _ := intBounds(gf.bounds)
	x, y := int(at.X)-bx, int(at.Y)-by
	off := y*bw + x
	return pixel.RGBA{
		R: float64(gf.pixels[off*4+0]) / 255,
		G: float64(gf.pixels[off*4+1]) / 255,
		B: float64(gf.pixels[off*4+2]) / 255,
		A: float64(gf.pixels[off*4+3]) / 255,
	}
}

// Frame returns the GLFrame's Frame that you can draw on.
func (gf *GLFrame) Frame() *glhf.Frame {
	return gf.frame
}

// Texture returns the underlying Texture of the GLFrame's Frame.
//
// Implements GLPicture interface.
func (gf *GLFrame) Texture() *glhf.Texture {
	return gf.frame.Texture()
}

// Dirty marks the GLFrame as changed. Always call this method when you draw onto the GLFrame's
// Frame.
func (gf *GLFrame) Dirty() {
	gf.dirty = true
}
