package pixel

import (
	"image"
	"image/draw"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
)

// Picture is a raster picture. It is usually used with sprites.
//
// A Picture is created from an image.Image, that can be either loaded from a file, or
// generated. After the creation, Pictures can be sliced (slicing creates a "sub-Picture"
// from a Picture) into smaller Pictures.
type Picture struct {
	tex    *glhf.Texture
	bounds Rect
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

	var tex *glhf.Texture
	mainthread.Call(func() {
		tex = glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			smooth,
			nrgba.Pix,
		)
	})

	return PictureFromTexture(tex)
}

// PictureFromTexture returns a new Picture that spans the whole supplied Texture.
func PictureFromTexture(tex *glhf.Texture) *Picture {
	return &Picture{
		tex:    tex,
		bounds: R(0, 0, float64(tex.Width()), float64(tex.Height())),
	}
}

// Image returns the content of the Picture as an image.NRGBA.
//
// Note, that this operation can be rather expensive.
func (p *Picture) Image() *image.NRGBA {
	bounds := p.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, int(bounds.W()), int(bounds.H())))

	mainthread.Call(func() {
		p.tex.Begin()
		nrgba.Pix = p.tex.Pixels(
			int(bounds.X()),
			int(bounds.Y()),
			int(bounds.W()),
			int(bounds.H()),
		)
		p.tex.End()
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
func (p *Picture) Texture() *glhf.Texture {
	return p.tex
}

// Slice returns a Picture within the supplied rectangle of the original picture. The original
// and the sliced Picture share the same texture.
//
// For example, suppose we have a 100x200 pixels Picture. If we slice it with rectangle (50,
// 100, 50, 100), we get the upper-right quadrant of the original Picture.
func (p *Picture) Slice(slice Rect) *Picture {
	return &Picture{
		tex:    p.tex,
		bounds: Rect{p.bounds.Pos + slice.Pos, slice.Size},
	}
}

// Bounds returns the bounding rectangle of this Picture relative to the most original picture.
//
// If the original Picture was sliced with the return value of this method, this Picture would
// be obtained.
func (p *Picture) Bounds() Rect {
	return p.bounds
}
