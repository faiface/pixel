package pixel

import (
	"image/color"

	"github.com/faiface/mainthread"
	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// Canvas is basically a Picture that you can draw on.
//
// Canvas supports TrianglesPosition, TrianglesColor and TrianglesTexture.
type Canvas struct {
	f *pixelgl.Frame
	s *pixelgl.Shader

	pic *Picture
	mat mgl32.Mat3
	col mgl32.Vec4
	bnd mgl32.Vec4
}

// NewCanvas creates a new fully transparent Canvas with specified dimensions in pixels.
func NewCanvas(width, height float64, smooth bool) *Canvas {
	c := &Canvas{}
	mainthread.Call(func() {
		var err error
		c.f = pixelgl.NewFrame(int(width), int(height), smooth)
		c.s, err = pixelgl.NewShader(
			canvasVertexFormat,
			canvasUniformFormat,
			canvasVertexShader,
			canvasFragmentShader,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create canvas"))
		}
	})
	c.pic = nil
	c.mat = mgl32.Ident3()
	c.col = mgl32.Vec4{1, 1, 1, 1}
	c.bnd = mgl32.Vec4{0, 0, 1, 1}
	return c
}

// Size returns the width and the height of the Canvas in pixels.
func (c *Canvas) Size() (width, height float64) {
	return float64(c.f.Width()), float64(c.f.Height())
}

// Content returns a Picture that contains the content of this Canvas. The returned Picture changes
// as you draw onto the Canvas, so there is no real need to call this method more than once.
func (c *Canvas) Content() *Picture {
	tex := c.f.Texture()
	return &Picture{
		texture: tex,
		bounds:  R(0, 0, float64(tex.Width()), float64(tex.Height())),
	}
}

// Clear fill the whole Canvas with on specified color.
func (c *Canvas) Clear(col color.Color) {
	mainthread.CallNonBlock(func() {
		c.f.Begin()
		col := NRGBAModel.Convert(col).(NRGBA)
		pixelgl.Clear(float32(col.R), float32(col.G), float32(col.B), float32(col.A))
		c.f.End()
	})
}

// MakeTriangles returns Triangles that draw onto this Canvas.
func (c *Canvas) MakeTriangles(t Triangles) Triangles {
	tpcs := NewGLTriangles(c.s, t).(trianglesPositionColorTexture)
	return &canvasTriangles{
		c: c,
		trianglesPositionColorTexture: tpcs,
	}
}

// SetPicture sets a Picture that will be used in further draw operations.
//
// This does not set the Picture that this Canvas draws onto, don't confuse it.
func (c *Canvas) SetPicture(p *Picture) {
	if p != nil {
		min := pictureBounds(p, V(0, 0))
		max := pictureBounds(p, V(1, 1))
		c.bnd = mgl32.Vec4{
			float32(min.X()), float32(min.Y()),
			float32(max.X()), float32(max.Y()),
		}
	}
	c.pic = p
}

// SetTransform sets the transformations used in further draw operations.
func (c *Canvas) SetTransform(t ...Transform) {
	c.mat = transformToMat(t...)
}

// SetMaskColor sets the mask color used in further draw operations.
func (c *Canvas) SetMaskColor(col color.Color) {
	if col == nil {
		col = NRGBA{1, 1, 1, 1}
	}
	nrgba := NRGBAModel.Convert(col).(NRGBA)
	r := float32(nrgba.R)
	g := float32(nrgba.G)
	b := float32(nrgba.B)
	a := float32(nrgba.A)
	c.col = mgl32.Vec4{r, g, b, a}
}

type trianglesPositionColorTexture interface {
	Triangles
	Position(i int) Vec
	Color(i int) NRGBA
	Texture(i int) Vec
}

type canvasTriangles struct {
	c *Canvas
	trianglesPositionColorTexture
}

func (ct *canvasTriangles) Draw() {
	// avoid possible race condition
	pic := ct.c.pic
	mat := ct.c.mat
	col := ct.c.col
	bnd := ct.c.bnd

	mainthread.CallNonBlock(func() {
		ct.c.f.Begin()
		ct.c.s.Begin()

		ct.c.s.SetUniformAttr(canvasTransformMat3, mat)
		ct.c.s.SetUniformAttr(canvasMaskColorVec4, col)
		ct.c.s.SetUniformAttr(canvasBoundsVec4, bnd)

		if pic != nil {
			pic.Texture().Begin()
			ct.trianglesPositionColorTexture.Draw()
			pic.Texture().End()
		} else {
			ct.trianglesPositionColorTexture.Draw()
		}

		ct.c.s.End()
		ct.c.f.End()
	})
}

const (
	canvasPositionVec2 int = iota
	canvasColorVec4
	canvasTextureVec2
)

var canvasVertexFormat = pixelgl.AttrFormat{
	canvasPositionVec2: {Name: "position", Type: pixelgl.Vec2},
	canvasColorVec4:    {Name: "color", Type: pixelgl.Vec4},
	canvasTextureVec2:  {Name: "texture", Type: pixelgl.Vec2},
}

const (
	canvasMaskColorVec4 int = iota
	canvasTransformMat3
	canvasBoundsVec4
)

var canvasUniformFormat = pixelgl.AttrFormat{
	{Name: "maskColor", Type: pixelgl.Vec4},
	{Name: "transform", Type: pixelgl.Mat3},
	{Name: "bounds", Type: pixelgl.Vec4},
}

var canvasVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texture;

out vec4 Color;
out vec2 Texture;

uniform mat3 transform;

void main() {
	gl_Position = vec4((transform * vec3(position.x, position.y, 1.0)).xy, 0.0, 1.0);
	Color = color;
	Texture = texture;
}
`

var canvasFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 Texture;

out vec4 color;

uniform vec4 maskColor;
uniform vec4 bounds;
uniform sampler2D tex;

void main() {
	vec2 boundsMin = bounds.xy;
	vec2 boundsMax = bounds.zw;

	if (Texture == vec2(-1, -1)) {
		color = maskColor * Color;
	} else {
		float tx = boundsMin.x * (1 - Texture.x) + boundsMax.x * Texture.x;
		float ty = boundsMin.y * (1 - Texture.y) + boundsMax.y * Texture.y;
		color = maskColor * Color * texture(tex, vec2(tx, ty));
	}
}
`
