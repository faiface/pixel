package pixel

import (
	"image"
	"image/draw"

	"github.com/faiface/pixel/pixelgl"
	"github.com/pkg/errors"
)

// Picture is a raster picture. It is usually used with sprites.
//
// A picture is created from an image.Image, that can be either loaded from a file, or
// generated. After the creation a picture can be sliced (slicing creates a "sub-picture"
// from a picture) into smaller pictures.
type Picture struct {
	texture *pixelgl.Texture
	bounds  Rect
}

// NewPicture creates a new picture from an image.Image.
func NewPicture(img image.Image, smooth bool) *Picture {
	// convert the image to RGBA format
	rgba := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)

	texture, err := pixelgl.NewTexture(
		pixelgl.NoOpDoer,
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		smooth,
		rgba.Pix,
	)
	if err != nil {
		panic(errors.Wrap(err, "failed to create picture"))
	}

	return &Picture{
		texture: texture,
		bounds:  R(0, 0, float64(texture.Width()), float64(texture.Height())),
	}
}

// Texture returns a pointer to the underlying OpenGL texture of a picture.
//
// Note, that the parent of this texture is pixelgl.NoOpDoer.
func (p *Picture) Texture() *pixelgl.Texture {
	return p.texture
}

// Slice returns a picture within the supplied rectangle of the original picture. The original
// and the sliced picture share the same texture.
//
// For example, suppose we have a 100x200 pixels picture. If we slice it with rectangle (50,
// 100, 50, 100), we get the upper-right quadrant of the original picture.
func (p *Picture) Slice(slice Rect) *Picture {
	return &Picture{
		texture: p.texture,
		bounds:  Rect{p.bounds.Pos + slice.Pos, slice.Size},
	}
}

// Bounds returns the bounding rectangle of this picture relative to the most original picture.
//
// If the original picture gets sliced with the return value of this method, this picture will
// be obtained.
func (p *Picture) Bounds() Rect {
	return p.bounds
}
