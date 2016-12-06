package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// Texture is an OpenGL texture.
type Texture struct {
	enabled bool
	parent  Doer
	tex     uint32
}

// NewTexture creates a new texture with the specified width and height.
// The pixels must be a sequence of RGBA values.
func NewTexture(parent Doer, width, height int, pixels []uint8) (*Texture, error) {
	texture := &Texture{parent: parent}

	parent.Do(func(ctx Context) {
		Do(func() {
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
	})

	return texture, nil
}

// Delete deletes a texture. Don't use a texture after deletion.
func (t *Texture) Delete() {
	t.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.DeleteTextures(1, &t.tex)
		})
	})
}

// ID returns an OpenGL identifier of a texture.
func (t *Texture) ID() uint32 {
	return t.tex
}

// Do bind a texture, executes sub, and unbinds the texture.
func (t *Texture) Do(sub func(Context)) {
	t.parent.Do(func(ctx Context) {
		if t.enabled {
			sub(ctx)
			return
		}
		DoNoBlock(func() {
			gl.BindTexture(gl.TEXTURE_2D, t.tex)
		})
		t.enabled = true
		sub(ctx)
		t.enabled = false
		DoNoBlock(func() {
			gl.BindTexture(gl.TEXTURE_2D, 0)
		})
	})
}
