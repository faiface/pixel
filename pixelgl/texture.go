package pixelgl

import (
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
)

// Texture is an OpenGL texture.
type Texture struct {
	tex           binder
	width, height int
}

// NewTexture creates a new texture with the specified width and height with some initial
// pixel values. The pixels must be a sequence of RGBA values.
func NewTexture(width, height int, smooth bool, pixels []uint8) *Texture {
	if len(pixels) != width*height*4 {
		panic("failed to create new texture: wrong number of pixels")
	}

	tex := &Texture{
		tex: binder{
			restoreLoc: gl.TEXTURE_BINDING_2D,
			bindFunc: func(obj uint32) {
				gl.BindTexture(gl.TEXTURE_2D, obj)
			},
		},
		width:  width,
		height: height,
	}

	gl.GenTextures(1, &tex.tex.obj)

	tex.Begin()
	defer tex.End()

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

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.MIRRORED_REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.MIRRORED_REPEAT)

	if smooth {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	} else {
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST_MIPMAP_NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	}

	gl.GenerateMipmap(gl.TEXTURE_2D)

	runtime.SetFinalizer(tex, (*Texture).delete)

	return tex
}

func (t *Texture) delete() {
	DoNoBlock(func() {
		gl.DeleteTextures(1, &t.tex.obj)
	})
}

// Width returns the width of a texture in pixels.
func (t *Texture) Width() int {
	return t.width
}

// Height returns the height of a texture in pixels.
func (t *Texture) Height() int {
	return t.height
}

// Begin binds a texture. This is necessary before using the texture.
func (t *Texture) Begin() {
	t.tex.bind()
}

// End unbinds a texture and restores the previous one.
func (t *Texture) End() {
	t.tex.restore()
}
