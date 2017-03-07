package pixelgl

import (
	"fmt"
	"image/color"
	"math"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// Canvas is an off-screen rectangular BasicTarget that you can draw onto.
//
// It supports TrianglesPosition, TrianglesColor, TrianglesPicture and PictureColor.
type Canvas struct {
	f *glhf.Frame
	s *glhf.Shader

	smooth bool

	mat mgl32.Mat3
	col mgl32.Vec4

	pixels []uint8
	dirty  bool

	bounds pixel.Rect
	orig   *Canvas
}

// NewCanvas creates a new empty, fully transparent Canvas with given bounds. If the smooth flag
// set, then stretched Pictures will be smoothed and will not be drawn pixely onto this Canvas.
func NewCanvas(bounds pixel.Rect, smooth bool) *Canvas {
	c := &Canvas{
		smooth: smooth,
		mat:    mgl32.Ident3(),
		col:    mgl32.Vec4{1, 1, 1, 1},
	}
	c.orig = c

	mainthread.Call(func() {
		var err error
		c.s, err = glhf.NewShader(
			canvasVertexFormat,
			canvasUniformFormat,
			canvasVertexShader,
			canvasFragmentShader,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create Canvas, there's a bug in the shader"))
		}
	})

	c.SetBounds(bounds)

	return c
}

// MakeTriangles creates a specialized copy of the supplied Triangles that draws onto this Canvas.
//
// TrianglesPosition, TrianglesColor and TrianglesPicture are supported.
func (c *Canvas) MakeTriangles(t pixel.Triangles) pixel.TargetTriangles {
	return &canvasTriangles{
		GLTriangles: NewGLTriangles(c.s, t),
		c:           c,
	}
}

// MakePicture create a specialized copy of the supplied Picture that draws onto this Canvas.
//
// PictureColor is supported.
func (c *Canvas) MakePicture(p pixel.Picture) pixel.TargetPicture {
	if cp, ok := p.(*canvasPicture); ok {
		tp := new(canvasPicture)
		*tp = *cp
		tp.c = c
		return tp
	}

	bounds := p.Bounds()
	bx, by, bw, bh := intBounds(bounds)

	pixels := make([]uint8, 4*bw*bh)
	if p, ok := p.(pixel.PictureColor); ok {
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				at := pixel.V(
					math.Max(float64(bx+x), bounds.Pos.X()),
					math.Max(float64(by+y), bounds.Pos.Y()),
				)
				color := p.Color(at)
				pixels[(y*bw+x)*4+0] = uint8(color.R * 255)
				pixels[(y*bw+x)*4+1] = uint8(color.G * 255)
				pixels[(y*bw+x)*4+2] = uint8(color.B * 255)
				pixels[(y*bw+x)*4+3] = uint8(color.A * 255)
			}
		}
	}

	var tex *glhf.Texture
	mainthread.Call(func() {
		tex = glhf.NewTexture(bw, bh, c.smooth, pixels)
	})

	cp := &canvasPicture{
		tex: tex,
		borders: pixel.R(
			float64(bx), float64(by),
			float64(bw), float64(bh),
		),
		bounds: bounds,
		c:      c,
	}
	cp.orig = cp
	return cp
}

// SetMatrix sets a Matrix that every point will be projected by.
func (c *Canvas) SetMatrix(m pixel.Matrix) {
	for i := range m {
		c.mat[i] = float32(m[i])
	}
}

// SetColorMask sets a color that every color in triangles or a picture will be multiplied by.
func (c *Canvas) SetColorMask(col color.Color) {
	nrgba := pixel.NRGBA{R: 1, G: 1, B: 1, A: 1}
	if col != nil {
		nrgba = pixel.NRGBAModel.Convert(col).(pixel.NRGBA)
	}
	c.col = mgl32.Vec4{
		float32(nrgba.R),
		float32(nrgba.G),
		float32(nrgba.B),
		float32(nrgba.A),
	}
}

// SetBounds resizes the Canvas to the new bounds. Old content will be preserved.
//
// If this Canvas was created using Slice-ing, then the relation between this Canvas and it's
// Original is unspecified (but Original will always return valid stuff).
func (c *Canvas) SetBounds(bounds pixel.Rect) {
	if c.Bounds() == bounds {
		return
	}

	mainthread.Call(func() {
		oldF := c.f

		_, _, w, h := intBounds(bounds)
		c.f = glhf.NewFrame(w, h, c.smooth)

		// preserve old content
		if oldF != nil {
			relBounds := c.bounds
			relBounds.Pos -= bounds.Pos
			ox, oy, ow, oh := intBounds(relBounds)
			oldF.Blit(
				c.f,
				ox, oy, ox+ow, oy+oh,
				ox, oy, ox+ow, oy+oh,
			)
		}
	})

	c.bounds = bounds
	c.orig = c // detach from the Original
	c.dirty = true
}

// Bounds returns the rectangular bounds of the Canvas.
func (c *Canvas) Bounds() pixel.Rect {
	return c.bounds
}

// SetSmooth sets whether the stretched Pictures drawn onto this Canvas should be drawn smooth or
// pixely.
func (c *Canvas) SetSmooth(smooth bool) {
	c.smooth = smooth
}

// Smooth returns whether the stretched Pictures drawn onto this Canvas are set to be drawn smooth
// or pixely.
func (c *Canvas) Smooth() bool {
	return c.smooth
}

// must be manually called inside mainthread
func (c *Canvas) setGlhfBounds() {
	bounds := c.bounds
	bounds.Pos -= c.orig.bounds.Pos
	bx, by, bw, bh := intBounds(bounds)
	glhf.Bounds(bx, by, bw, bh)
}

// Clear fill the whole Canvas with a single color.
func (c *Canvas) Clear(color color.Color) {
	c.orig.dirty = true

	nrgba := pixel.NRGBAModel.Convert(color).(pixel.NRGBA)

	mainthread.CallNonBlock(func() {
		c.setGlhfBounds()
		c.f.Begin()
		glhf.Clear(
			float32(nrgba.R),
			float32(nrgba.G),
			float32(nrgba.B),
			float32(nrgba.A),
		)
		c.f.End()
	})
}

// Slice returns a sub-Canvas with the specified Bounds.
//
// The returned value is *Canvas, the type of the return value is a general pixel.Picture just so
// that Canvas implements pixel.Picture interface.
func (c *Canvas) Slice(bounds pixel.Rect) pixel.Picture {
	sc := new(Canvas)
	*sc = *c
	sc.bounds = bounds
	return sc
}

// Original returns the most original Canvas that this Canvas was created from using Slice-ing.
//
// The returned value is *Canvas, the type of the return value is a general pixel.Picture just so
// that Canvas implements pixel.Picture interface.
func (c *Canvas) Original() pixel.Picture {
	return c.orig
}

// Color returns the color of the pixel over the given position inside the Canvas.
func (c *Canvas) Color(at pixel.Vec) pixel.NRGBA {
	if c.orig.dirty {
		mainthread.Call(func() {
			c.f.Texture.Begin()
			c.orig.pixels = c.f.Texture.Pixels(0, 0, c.f.Texture.Width(), c.f.Texture.Height())
			c.f.Texture.End()
		})
		c.orig.dirty = false
	}
	if !c.bounds.Contains(at) {
		return pixel.NRGBA{}
	}
	bx, by, bw, _ := intBounds(c.orig.bounds)
	x, y := int(at.X())-bx, int(at.Y())-by
	off := y*bw + x
	return pixel.NRGBA{
		R: float64(c.orig.pixels[off*4+0]) / 255,
		G: float64(c.orig.pixels[off*4+1]) / 255,
		B: float64(c.orig.pixels[off*4+2]) / 255,
		A: float64(c.orig.pixels[off*4+3]) / 255,
	}
}

type canvasTriangles struct {
	*GLTriangles

	c *Canvas
}

func (ct *canvasTriangles) draw(cp *canvasPicture) {
	ct.c.orig.dirty = true

	// save the current state vars to avoid race condition
	mat := ct.c.mat
	col := ct.c.col

	mainthread.CallNonBlock(func() {
		ct.c.setGlhfBounds()
		ct.c.f.Begin()
		ct.c.s.Begin()

		ct.c.s.SetUniformAttr(canvasBounds, mgl32.Vec4{
			float32(ct.c.bounds.X()),
			float32(ct.c.bounds.Y()),
			float32(ct.c.bounds.W()),
			float32(ct.c.bounds.H()),
		})
		ct.c.s.SetUniformAttr(canvasTransform, mat)
		ct.c.s.SetUniformAttr(canvasColorMask, col)

		if cp == nil {
			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()
		} else {
			cp.tex.Begin()

			ct.c.s.SetUniformAttr(canvasTexBorders, mgl32.Vec4{
				float32(cp.borders.X()),
				float32(cp.borders.Y()),
				float32(cp.borders.W()),
				float32(cp.borders.H()),
			})
			ct.c.s.SetUniformAttr(canvasTexBounds, mgl32.Vec4{
				float32(cp.bounds.X()),
				float32(cp.bounds.Y()),
				float32(cp.bounds.W()),
				float32(cp.bounds.H()),
			})

			if cp.tex.Smooth() != ct.c.smooth {
				cp.tex.SetSmooth(ct.c.smooth)
			}

			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()

			cp.tex.End()
		}

		ct.c.s.End()
		ct.c.f.End()
	})
}

func (ct *canvasTriangles) Draw() {
	ct.draw(nil)
}

type canvasPicture struct {
	tex     *glhf.Texture
	borders pixel.Rect
	bounds  pixel.Rect

	orig *canvasPicture
	c    *Canvas
}

func (cp *canvasPicture) Bounds() pixel.Rect {
	return cp.bounds
}

func (cp *canvasPicture) Slice(r pixel.Rect) pixel.Picture {
	sp := new(canvasPicture)
	*sp = *cp
	sp.bounds = r
	return sp
}

func (cp *canvasPicture) Original() pixel.Picture {
	return cp.orig
}

func (cp *canvasPicture) Draw(t pixel.TargetTriangles) {
	ct := t.(*canvasTriangles)
	if cp.c != ct.c {
		panic(fmt.Errorf("%T.Draw: TargetTriangles generated by different Canvas", cp))
	}
	ct.draw(cp)
}

const (
	canvasPosition int = iota
	canvasColor
	canvasTexture
	canvasIntensity
)

var canvasVertexFormat = glhf.AttrFormat{
	canvasPosition:  {Name: "position", Type: glhf.Vec2},
	canvasColor:     {Name: "color", Type: glhf.Vec4},
	canvasTexture:   {Name: "texture", Type: glhf.Vec2},
	canvasIntensity: {Name: "intensity", Type: glhf.Float},
}

const (
	canvasTransform int = iota
	canvasColorMask
	canvasBounds
	canvasTexBorders
	canvasTexBounds
)

var canvasUniformFormat = glhf.AttrFormat{
	canvasTransform:  {Name: "transform", Type: glhf.Mat3},
	canvasColorMask:  {Name: "colorMask", Type: glhf.Vec4},
	canvasBounds:     {Name: "bounds", Type: glhf.Vec4},
	canvasTexBorders: {Name: "texBorders", Type: glhf.Vec4},
	canvasTexBounds:  {Name: "texBounds", Type: glhf.Vec4},
}

var canvasVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texture;
in float intensity;

out vec4 Color;
out vec2 Texture;
out float Intensity;

uniform mat3 transform;
uniform vec4 borders;
uniform vec4 bounds;

void main() {
	vec2 transPos = (transform * vec3(position, 1.0)).xy;
	vec2 normPos = (transPos - bounds.xy) / (bounds.zw) * 2 - vec2(1, 1);
	gl_Position = vec4(normPos, 0.0, 1.0);
	Color = color;
	Texture = texture;
	Intensity = intensity;
}
`

var canvasFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 Texture;
in float Intensity;

out vec4 color;

uniform vec4 colorMask;
uniform vec4 texBorders;
uniform vec4 texBounds;
uniform sampler2D tex;

void main() {
	if (Intensity == 0) {
		color = colorMask * Color;
	} else {
		color = vec4(0, 0, 0, 0);
		color += (1 - Intensity) * colorMask * Color;

		float bx = texBounds.x;
		float by = texBounds.y;
		float bw = texBounds.z;
		float bh = texBounds.w;
		if (bx <= Texture.x && Texture.x <= bx + bw && by <= Texture.y && Texture.y <= by + bh) {
			vec2 t = (Texture - texBorders.xy) / texBorders.zw;
			color += Intensity * colorMask * Color * texture(tex, t);
		}
	}
}
`
