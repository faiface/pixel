package pixelgl

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

// Texture is an OpenGL texture.
type Texture struct {
	parent BeginEnder
	tex    uint32
}

// NewTexture creates a new texture with the specified width and height.
// The pixels must be a sequence of RGBA values.
func NewTexture(parent BeginEnder, width, height int, pixels []uint8) (*Texture, error) {
	texture := &Texture{parent: parent}
	err := DoGLErr(func() {
		gl.GenTextures(1, &texture.tex)
		gl.BindTexture(gl.TEXTURE_2D, texture.tex)

		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(width),
			int32(height),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(pixels),
		)
		gl.GenerateMipmap(gl.TEXTURE_2D)

		gl.BindTexture(gl.TEXTURE_2D, 0)
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a texture")
	}
	return texture, nil
}

// Delete deletes a texture. Don't use a texture after deletion.
func (t *Texture) Delete() {
	Do(func() {
		gl.DeleteTextures(1, &t.tex)
	})
}

// Begin binds a texture.
func (t *Texture) Begin() {
	t.parent.Begin()
	Do(func() {
		gl.BindTexture(gl.TEXTURE_2D, t.tex)
	})
}

// End unbinds a texture.
func (t *Texture) End() {
	Do(func() {
		gl.BindTexture(gl.TEXTURE_2D, 0)
	})
	t.parent.End()
}
