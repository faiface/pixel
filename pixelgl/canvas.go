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

// Canvas is an off-screen rectangular BasicTarget and Picture at the same time, that you can draw
// onto.
//
// It supports TrianglesPosition, TrianglesColor, TrianglesPicture and PictureColor.
type Canvas struct {
	// these should **only** be accessed through orig
	f       *glhf.Frame
	borders pixel.Rect
	pixels  []uint8
	dirty   bool

	// these should **never** be accessed through orig
	s      *glhf.Shader
	bounds pixel.Rect
	mat    mgl32.Mat3
	col    mgl32.Vec4
	smooth bool

	orig *Canvas
}

// NewCanvas creates a new empty, fully transparent Canvas with given bounds. If the smooth flag is
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
		dst:         c,
	}
}

// MakePicture create a specialized copy of the supplied Picture that draws onto this Canvas.
//
// PictureColor is supported.
func (c *Canvas) MakePicture(p pixel.Picture) pixel.TargetPicture {
	// short paths
	if cp, ok := p.(*canvasPicture); ok {
		tp := new(canvasPicture)
		*tp = *cp
		tp.dst = c
		return tp
	}
	if ccp, ok := p.(*canvasCanvasPicture); ok {
		tp := new(canvasCanvasPicture)
		*tp = *ccp
		tp.dst = c
		return tp
	}

	// Canvas special case
	if canvas, ok := p.(*Canvas); ok {
		return &canvasCanvasPicture{
			src: canvas,
			dst: c,
		}
	}

	bounds := p.Bounds()
	bx, by, bw, bh := intBounds(bounds)

	pixels := make([]uint8, 4*bw*bh)

	if pd, ok := p.(*pixel.PictureData); ok {
		// PictureData short path
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				nrgba := pd.Pix[y*pd.Stride+x]
				off := (y*bw + x) * 4
				pixels[off+0] = nrgba.R
				pixels[off+1] = nrgba.G
				pixels[off+2] = nrgba.B
				pixels[off+3] = nrgba.A
			}
		}
	} else if p, ok := p.(pixel.PictureColor); ok {
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				at := pixel.V(
					math.Max(float64(bx+x), bounds.Min.X()),
					math.Max(float64(by+y), bounds.Min.Y()),
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
		tex = glhf.NewTexture(bw, bh, c.smooth, pixels)
	})

	cp := &canvasPicture{
		tex:    tex,
		pixels: pixels,
		bounds: pixel.R(
			float64(bx), float64(by),
			float64(bw), float64(bh),
		),
		dst: c,
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
		nrgba = pixel.ToNRGBA(col)
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
// If the new Bounds fit into the Original borders, no new Canvas will be allocated.
func (c *Canvas) SetBounds(bounds pixel.Rect) {
	c.bounds = bounds

	// if this bounds fit into the original bounds, no need to reallocate
	if c.orig.borders.Contains(bounds.Min) && c.orig.borders.Contains(bounds.Max) {
		return
	}

	mainthread.Call(func() {
		oldF := c.orig.f

		_, _, w, h := intBounds(bounds)
		c.f = glhf.NewFrame(w, h, c.smooth)

		// preserve old content
		if oldF != nil {
			relBounds := bounds
			relBounds = relBounds.Moved(-c.orig.borders.Min)
			ox, oy, ow, oh := intBounds(relBounds)
			oldF.Blit(
				c.f,
				ox, oy, ox+ow, oy+oh,
				ox, oy, ox+ow, oy+oh,
			)
		}
	})

	// detach from orig
	c.borders = bounds
	c.pixels = nil
	c.dirty = true
	c.orig = c
}

// Bounds returns the rectangular bounds of the Canvas.
func (c *Canvas) Bounds() pixel.Rect {
	return c.bounds
}

// SetSmooth sets whether stretched Pictures drawn onto this Canvas should be drawn smooth or
// pixely.
func (c *Canvas) SetSmooth(smooth bool) {
	c.smooth = smooth
}

// Smooth returns whether stretched Pictures drawn onto this Canvas are set to be drawn smooth or
// pixely.
func (c *Canvas) Smooth() bool {
	return c.smooth
}

// must be manually called inside mainthread
func (c *Canvas) setGlhfBounds() {
	bounds := c.bounds
	bounds.Moved(c.orig.borders.Min)
	bx, by, bw, bh := intBounds(bounds)
	glhf.Bounds(bx, by, bw, bh)
}

// Clear fills the whole Canvas with a single color.
func (c *Canvas) Clear(color color.Color) {
	c.orig.dirty = true

	nrgba := pixel.ToNRGBA(color)

	// color masking
	nrgba = nrgba.Mul(pixel.NRGBA{
		R: float64(c.col[0]),
		G: float64(c.col[1]),
		B: float64(c.col[2]),
		A: float64(c.col[3]),
	})

	mainthread.CallNonBlock(func() {
		c.setGlhfBounds()
		c.orig.f.Begin()
		glhf.Clear(
			float32(nrgba.R),
			float32(nrgba.G),
			float32(nrgba.B),
			float32(nrgba.A),
		)
		c.orig.f.End()
	})
}

// Color returns the color of the pixel over the given position inside the Canvas.
func (c *Canvas) Color(at pixel.Vec) pixel.NRGBA {
	if c.orig.dirty {
		mainthread.Call(func() {
			tex := c.orig.f.Texture()
			tex.Begin()
			c.orig.pixels = tex.Pixels(0, 0, tex.Width(), tex.Height())
			tex.End()
		})
		c.orig.dirty = false
	}
	if !c.bounds.Contains(at) {
		return pixel.NRGBA{}
	}
	bx, by, bw, _ := intBounds(c.orig.borders)
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

	dst *Canvas
}

func (ct *canvasTriangles) draw(tex *glhf.Texture, bounds pixel.Rect) {
	ct.dst.orig.dirty = true

	// save the current state vars to avoid race condition
	mat := ct.dst.mat
	col := ct.dst.col

	mainthread.CallNonBlock(func() {
		ct.dst.setGlhfBounds()
		ct.dst.orig.f.Begin()
		ct.dst.s.Begin()

		ct.dst.s.SetUniformAttr(canvasBounds, mgl32.Vec4{
			float32(ct.dst.bounds.Min.X()),
			float32(ct.dst.bounds.Min.Y()),
			float32(ct.dst.bounds.W()),
			float32(ct.dst.bounds.H()),
		})
		ct.dst.s.SetUniformAttr(canvasTransform, mat)
		ct.dst.s.SetUniformAttr(canvasColorMask, col)

		if tex == nil {
			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()
		} else {
			tex.Begin()

			ct.dst.s.SetUniformAttr(canvasTexBounds, mgl32.Vec4{
				float32(bounds.Min.X()),
				float32(bounds.Min.Y()),
				float32(bounds.W()),
				float32(bounds.H()),
			})

			if tex.Smooth() != ct.dst.smooth {
				tex.SetSmooth(ct.dst.smooth)
			}

			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()

			tex.End()
		}

		ct.dst.s.End()
		ct.dst.orig.f.End()
	})
}

func (ct *canvasTriangles) Draw() {
	ct.draw(nil, pixel.Rect{})
}

type canvasPicture struct {
	tex    *glhf.Texture
	pixels []uint8
	bounds pixel.Rect

	orig *canvasPicture
	dst  *Canvas
}

func (cp *canvasPicture) Bounds() pixel.Rect {
	return cp.bounds
}

func (cp *canvasPicture) Color(at pixel.Vec) pixel.NRGBA {
	if !cp.bounds.Contains(at) {
		return pixel.NRGBA{}
	}
	bx, by, bw, _ := intBounds(cp.bounds)
	x, y := int(at.X())-bx, int(at.Y())-by
	off := y*bw + x
	return pixel.NRGBA{
		R: float64(cp.pixels[off*4+0]) / 255,
		G: float64(cp.pixels[off*4+1]) / 255,
		B: float64(cp.pixels[off*4+2]) / 255,
		A: float64(cp.pixels[off*4+3]) / 255,
	}
}

func (cp *canvasPicture) Draw(t pixel.TargetTriangles) {
	ct := t.(*canvasTriangles)
	if cp.dst != ct.dst {
		panic(fmt.Errorf("(%T).Draw: TargetTriangles generated by different Canvas", cp))
	}
	ct.draw(cp.tex, cp.bounds)
}

type canvasCanvasPicture struct {
	src, dst *Canvas
}

func (ccp *canvasCanvasPicture) Bounds() pixel.Rect {
	return ccp.src.Bounds()
}

func (ccp *canvasCanvasPicture) Color(at pixel.Vec) pixel.NRGBA {
	if !ccp.Bounds().Contains(at) {
		return pixel.NRGBA{}
	}
	return ccp.src.Color(at)
}

func (ccp *canvasCanvasPicture) Draw(t pixel.TargetTriangles) {
	ct := t.(*canvasTriangles)
	if ccp.dst != ct.dst {
		panic(fmt.Errorf("(%T).Draw: TargetTriangles generated by different Canvas", ccp))
	}
	ct.draw(ccp.src.orig.f.Texture(), ccp.Bounds())
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
	canvasTexBounds
)

var canvasUniformFormat = glhf.AttrFormat{
	canvasTransform: {Name: "transform", Type: glhf.Mat3},
	canvasColorMask: {Name: "colorMask", Type: glhf.Vec4},
	canvasBounds:    {Name: "bounds", Type: glhf.Vec4},
	canvasTexBounds: {Name: "texBounds", Type: glhf.Vec4},
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
		vec2 t = (Texture - texBounds.xy) / texBounds.zw;
		color += Intensity * colorMask * Color * texture(tex, t);
	}
}
`
