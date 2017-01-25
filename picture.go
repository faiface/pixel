package pixel

import (
	"image"
	"image/draw"

	"github.com/faiface/mainthread"
	"github.com/faiface/pixel/pixelgl"
)

// Picture is a raster picture. It is usually used with sprites.
//
// A Picture is created from an image.Image, that can be either loaded from a file, or
// generated. After the creation, Pictures can be sliced (slicing creates a "sub-Picture"
// from a Picture) into smaller Pictures.
type Picture struct {
	texture *pixelgl.Texture
	bounds  Rect
}

// NewPicture creates a new Picture from an image.Image.
func NewPicture(img image.Image, smooth bool) *Picture {
	// convert the image to NRGBA format
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)

	// flip the image vertically
	tmp := make([]byte, nrgba.Stride)
	for i, j := 0, bounds.Dy()-1; i < j; i, j = i+1, j-1 {
		iSlice := nrgba.Pix[i*nrgba.Stride : (i+1)*nrgba.Stride]
		jSlice := nrgba.Pix[j*nrgba.Stride : (j+1)*nrgba.Stride]
		copy(tmp, iSlice)
		copy(iSlice, jSlice)
		copy(jSlice, tmp)
	}

	var texture *pixelgl.Texture
	mainthread.Call(func() {
		texture = pixelgl.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			smooth,
			nrgba.Pix,
		)
	})

	return &Picture{
		texture: texture,
		bounds:  R(0, 0, float64(texture.Width()), float64(texture.Height())),
	}
}

// Image returns the content of the Picture as an image.NRGBA.
func (p *Picture) Image() *image.NRGBA {
	bounds := p.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, int(bounds.W()), int(bounds.H())))

	mainthread.Call(func() {
		p.texture.Begin()
		nrgba.Pix = p.texture.Pixels(
			int(bounds.X()),
			int(bounds.Y()),
			int(bounds.W()),
			int(bounds.H()),
		)
		p.texture.End()
	})

	// flip the image vertically
	tmp := make([]byte, nrgba.Stride)
	for i, j := 0, nrgba.Bounds().Dy()-1; i < j; i, j = i+1, j-1 {
		iSlice := nrgba.Pix[i*nrgba.Stride : (i+1)*nrgba.Stride]
		jSlice := nrgba.Pix[j*nrgba.Stride : (j+1)*nrgba.Stride]
		copy(tmp, iSlice)
		copy(iSlice, jSlice)
		copy(jSlice, tmp)
	}

	return nrgba
}

// Texture returns a pointer to the underlying OpenGL texture of the Picture.
func (p *Picture) Texture() *pixelgl.Texture {
	return p.texture
}

// Slice returns a Picture within the supplied rectangle of the original picture. The original
// and the sliced Picture share the same texture.
//
// For example, suppose we have a 100x200 pixels Picture. If we slice it with rectangle (50,
// 100, 50, 100), we get the upper-right quadrant of the original Picture.
func (p *Picture) Slice(slice Rect) *Picture {
	return &Picture{
		texture: p.texture,
		bounds:  Rect{p.bounds.Pos + slice.Pos, slice.Size},
	}
}

// Bounds returns the bounding rectangle of this Picture relative to the most original picture.
//
// If the original Picture was sliced with the return value of this method, this Picture would
// be obtained.
func (p *Picture) Bounds() Rect {
	return p.bounds
}
