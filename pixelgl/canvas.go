package pixelgl

import (
	"fmt"
	"image/color"

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

	bounds pixel.Rect
	smooth bool

	mat mgl32.Mat3
	col mgl32.Vec4
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
	pd := pixel.PictureDataFromPicture(p)
	pixels := make([]uint8, 4*len(pd.Pix))
	for i := range pd.Pix {
		pixels[i*4+0] = uint8(pd.Pix[i].R * 255)
		pixels[i*4+1] = uint8(pd.Pix[i].G * 255)
		pixels[i*4+2] = uint8(pd.Pix[i].B * 255)
		pixels[i*4+3] = uint8(pd.Pix[i].A * 255)
	}

	var tex *glhf.Texture
	mainthread.Call(func() {
		tex = glhf.NewTexture(pd.Stride, len(pd.Pix)/pd.Stride, c.smooth, pixels)
	})

	return &canvasPicture{
		tex:    tex,
		bounds: pd.Rect,
		c:      c,
	}
}

// SetTransform sets a set of Transforms that every position in triangles will be put through.
func (c *Canvas) SetTransform(t ...pixel.Transform) {
	c.mat = transformToMat(t...)
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

		c.s.Begin()
		orig := bounds.Center()
		c.s.SetUniformAttr(canvasOrig, mgl32.Vec2{
			float32(orig.X()),
			float32(orig.Y()),
		})
		c.s.SetUniformAttr(canvasSize, mgl32.Vec2{
			float32(w),
			float32(h),
		})
		c.s.End()
	})
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
		ct.c.f.Begin()
		ct.c.s.Begin()

		ct.c.s.SetUniformAttr(canvasTransform, mat)
		ct.c.s.SetUniformAttr(canvasColorMask, col)

		if cp == nil {
			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()
		} else {
			cp.tex.Begin()

			ct.c.s.SetUniformAttr(canvasTexSize, mgl32.Vec2{
				float32(cp.tex.Width()),
				float32(cp.tex.Height()),
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
	tex    *glhf.Texture
	bounds pixel.Rect

	c *Canvas
}

func (cp *canvasPicture) Bounds() pixel.Rect {
	return cp.bounds
}

func (cp *canvasPicture) Slice(r pixel.Rect) pixel.Picture {
	return &canvasPicture{
		bounds: r,
		c:      cp.c,
	}
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
	canvasTexSize
	canvasOrig
	canvasSize
)

var canvasUniformFormat = glhf.AttrFormat{
	canvasTransform: {Name: "transform", Type: glhf.Mat3},
	canvasColorMask: {Name: "colorMask", Type: glhf.Vec4},
	canvasTexSize:   {Name: "texSize", Type: glhf.Vec2},
	canvasOrig:      {Name: "orig", Type: glhf.Vec2},
	canvasSize:      {Name: "size", Type: glhf.Vec2},
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
uniform vec2 orig;
uniform vec2 size;

void main() {
	vec2 transPos = (transform * vec3(position, 1.0)).xy;
	vec2 normPos = (transPos - orig) / (size/2);
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
uniform vec2 texSize;
uniform sampler2D tex;

void main() {
	if (Intensity == 0) {
		color = colorMask * Color;
	} else {
		vec2 t = Texture / texSize;
		color = vec4(0, 0, 0, 0);
		color += (1 - Intensity) * colorMask * Color;
		color += Intensity * colorMask * Color * texture(tex, t);
	}
}
`
