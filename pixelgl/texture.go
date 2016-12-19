package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"
import "runtime"

// Texture is an OpenGL texture.
type Texture struct {
	parent        Doer
	tex           binder
	width, height int
}

// NewTexture creates a new texture with the specified width and height.
// The pixels must be a sequence of RGBA values.
func NewTexture(parent Doer, width, height int, pixels []uint8) (*Texture, error) {
	texture := &Texture{
		parent: parent,
		tex: binder{
			restoreLoc: gl.TEXTURE_BINDING_2D,
			bindFunc: func(obj uint32) {
				gl.BindTexture(gl.TEXTURE_2D, obj)
			},
		},
		width:  width,
		height: height,
	}

	parent.Do(func(ctx Context) {
		Do(func() {
			gl.GenTextures(1, &texture.tex.obj)
			defer texture.tex.bind().restore()

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
		})
	})

	runtime.SetFinalizer(texture, (*Texture).delete)

	return texture, nil
}

func (t *Texture) delete() {
	DoNoBlock(func() {
		gl.DeleteTextures(1, &t.tex.obj)
	})
}

// ID returns an OpenGL identifier of a texture.
func (t *Texture) ID() uint32 {
	return t.tex.obj
}

// Width returns the width of a texture in pixels.
func (t *Texture) Width() int {
	return t.width
}

// Height returns the height of a texture in pixels.
func (t *Texture) Height() int {
	return t.height
}

// Do bind a texture, executes sub, and unbinds the texture.
func (t *Texture) Do(sub func(Context)) {
	t.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			t.tex.bind()
		})
		sub(ctx)
		DoNoBlock(func() {
			t.tex.restore()
		})
	})
}
