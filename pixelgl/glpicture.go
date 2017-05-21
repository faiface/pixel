package pixelgl

import (
	"math"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
)

// GLPicture is a pixel.PictureColor with a Texture. All OpenGL Targets should implement and accept
// this interface, because it enables seamless drawing of one to another.
//
// Implementing this interface on an OpenGL Target enables other OpenGL Targets to efficiently draw
// that Target onto them.
type GLPicture interface {
	pixel.PictureColor
	Texture() *glhf.Texture
}

// NewGLPicture creates a new GLPicture with it's own static OpenGL texture. This function always
// allocates a new texture that cannot (shouldn't) be further modified.
func NewGLPicture(p pixel.Picture) GLPicture {
	bounds := p.Bounds()
	bx, by, bw, bh := intBounds(bounds)

	pixels := make([]uint8, 4*bw*bh)

	if pd, ok := p.(*pixel.PictureData); ok {
		// PictureData short path
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				rgba := pd.Pix[y*pd.Stride+x]
				off := (y*bw + x) * 4
				pixels[off+0] = rgba.R
				pixels[off+1] = rgba.G
				pixels[off+2] = rgba.B
				pixels[off+3] = rgba.A
			}
		}
	} else if p, ok := p.(pixel.PictureColor); ok {
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				at := pixel.V(
					math.Max(float64(bx+x), bounds.Min.X),
					math.Max(float64(by+y), bounds.Min.Y),
				)
				color := p.Color(at)
				off := (y*bw + x) * 4
				pixels[off+0] = uint8(color.R * 255)
				pixels[off+1] = uint8(color.G * 255)
				pixels[off+2] = uint8(color.B * 255)
				pixels[off+3] = uint8(color.A * 255)
			}
		}
	}

	var tex *glhf.Texture
	mainthread.Call(func() {
		tex = glhf.NewTexture(bw, bh, false, pixels)
	})

	gp := &glPicture{
		bounds: bounds,
		tex:    tex,
		pixels: pixels,
	}
	return gp
}

type glPicture struct {
	bounds pixel.Rect
	tex    *glhf.Texture
	pixels []uint8
}

func (gp *glPicture) Bounds() pixel.Rect {
	return gp.bounds
}

func (gp *glPicture) Texture() *glhf.Texture {
	return gp.tex
}

func (gp *glPicture) Color(at pixel.Vec) pixel.RGBA {
	if !gp.bounds.Contains(at) {
		return pixel.Alpha(0)
	}
	bx, by, bw, _ := intBounds(gp.bounds)
	x, y := int(at.X)-bx, int(at.Y)-by
	off := y*bw + x
	return pixel.RGBA{
		R: float64(gp.pixels[off*4+0]) / 255,
		G: float64(gp.pixels[off*4+1]) / 255,
		B: float64(gp.pixels[off*4+2]) / 255,
		A: float64(gp.pixels[off*4+3]) / 255,
	}
}
