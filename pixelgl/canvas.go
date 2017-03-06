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

//TODO: make Canvas a Picture

// Canvas is an off-screen rectangular BasicTarget that you can draw onto.
//
// It supports TrianglesPosition, TrianglesColor, TrianglesPicture and PictureColor.
type Canvas struct {
	f *glhf.Frame
	s *glhf.Shader

	smooth bool

	mat mgl32.Mat3
	col mgl32.Vec4

	borders pixel.Rect
	bounds  pixel.Rect
}

// NewCanvas creates a new empty, fully transparent Canvas with given bounds. If the smooth flag
// set, then stretched Pictures will be smoothed and will not be drawn pixely onto this Canvas.
func NewCanvas(bounds pixel.Rect, smooth bool) *Canvas {
	c := &Canvas{
		smooth: smooth,
		mat:    mgl32.Ident3(),
		col:    mgl32.Vec4{1, 1, 1, 1},
	}

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
	bx, by, bw, bh := discreteBounds(bounds)

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
func (c *Canvas) SetBounds(bounds pixel.Rect) {
	if c.Bounds() == bounds {
		return
	}

	mainthread.Call(func() {
		oldF := c.f

		_, _, w, h := discreteBounds(bounds)
		c.f = glhf.NewFrame(w, h, c.smooth)

		// preserve old content
		if oldF != nil {
			relBounds := c.bounds
			relBounds.Pos -= bounds.Pos
			ox, oy, ow, oh := discreteBounds(relBounds)
			oldF.Blit(
				c.f,
				ox, oy, ox+ow, oy+oh,
				ox, oy, ox+ow, oy+oh,
			)
		}
	})
	c.borders = bounds
	c.bounds = bounds
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

// Clear fill the whole Canvas with a single color.
func (c *Canvas) Clear(color color.Color) {
	nrgba := pixel.NRGBAModel.Convert(color).(pixel.NRGBA)
	mainthread.CallNonBlock(func() {
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

type canvasTriangles struct {
	*GLTriangles

	c *Canvas
}

func (ct *canvasTriangles) draw(cp *canvasPicture) {
	// save the current state vars to avoid race condition
	mat := ct.c.mat
	col := ct.c.col

	mainthread.CallNonBlock(func() {
		glhf.Bounds(0, 0, ct.c.f.Width(), ct.c.f.Height())
		ct.c.f.Begin()
		ct.c.s.Begin()

		ct.c.s.SetUniformAttr(canvasBounds, mgl32.Vec4{
			float32(cp.c.bounds.X()),
			float32(cp.c.bounds.Y()),
			float32(cp.c.bounds.W()),
			float32(cp.c.bounds.H()),
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
uniform vec4 bounds;

void main() {
	vec2 transPos = (transform * vec3(position, 1.0)).xy;
	vec2 normPos = 2 * (transPos - bounds.xy) / (bounds.zw) - vec2(1, 1);
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
