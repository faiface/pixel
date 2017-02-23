package pixelgl

import (
	"image/color"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// Canvas is basically a Picture that you can draw onto.
//
// Canvas supports TrianglesPosition, TrianglesColor and TrianglesTexture.
type Canvas struct {
	f *glhf.Frame
	s *glhf.Shader

	copyVs *glhf.VertexSlice
	smooth bool

	drawTd pixel.TrianglesDrawer

	pic *pixel.GLPicture
	mat mgl32.Mat3
	col mgl32.Vec4
	bnd mgl32.Vec4
}

// NewCanvas creates a new fully transparent Canvas with specified dimensions in pixels.
func NewCanvas(width, height float64, smooth bool) *Canvas {
	c := &Canvas{smooth: smooth}
	mainthread.Call(func() {
		var err error
		c.f = glhf.NewFrame(int(width), int(height), smooth)
		c.s, err = glhf.NewShader(
			canvasVertexFormat,
			canvasUniformFormat,
			canvasVertexShader,
			canvasFragmentShader,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create canvas"))
		}

		c.copyVs = glhf.MakeVertexSlice(c.s, 6, 6)
		c.copyVs.Begin()
		c.copyVs.SetVertexData([]float32{
			-1, -1, 1, 1, 1, 1, 0, 0,
			1, -1, 1, 1, 1, 1, 1, 0,
			1, 1, 1, 1, 1, 1, 1, 1,
			-1, -1, 1, 1, 1, 1, 0, 0,
			1, 1, 1, 1, 1, 1, 1, 1,
			-1, 1, 1, 1, 1, 1, 0, 1,
		})
		c.copyVs.End()
	})

	white := pixel.NRGBA{R: 1, G: 1, B: 1, A: 1}
	c.drawTd = pixel.TrianglesDrawer{Triangles: &pixel.TrianglesData{
		{Position: pixel.V(-1, -1), Color: white, Picture: pixel.V(0, 0)},
		{Position: pixel.V(1, -1), Color: white, Picture: pixel.V(1, 0)},
		{Position: pixel.V(1, 1), Color: white, Picture: pixel.V(1, 1)},
		{Position: pixel.V(-1, -1), Color: white, Picture: pixel.V(0, 0)},
		{Position: pixel.V(1, 1), Color: white, Picture: pixel.V(1, 1)},
		{Position: pixel.V(-1, 1), Color: white, Picture: pixel.V(0, 1)},
	}}

	c.pic = nil
	c.mat = mgl32.Ident3()
	c.col = mgl32.Vec4{1, 1, 1, 1}
	c.bnd = mgl32.Vec4{0, 0, 1, 1}
	return c
}

// SetSize resizes the Canvas. The original content will be stretched to fit the new size.
func (c *Canvas) SetSize(width, height float64) {
	if pixel.V(width, height) == pixel.V(c.Size()) {
		return
	}
	mainthread.Call(func() {
		oldF := c.f
		c.f = glhf.NewFrame(int(width), int(height), c.smooth)

		c.f.Begin()
		c.s.Begin()

		c.s.SetUniformAttr(canvasTransformMat3, mgl32.Ident3())
		c.s.SetUniformAttr(canvasMaskColorVec4, mgl32.Vec4{1, 1, 1, 1})
		c.s.SetUniformAttr(canvasBoundsVec4, mgl32.Vec4{0, 0, 1, 1})

		oldF.Texture().Begin()
		c.copyVs.Begin()
		c.copyVs.Draw()
		c.copyVs.End()
		oldF.Texture().End()

		c.s.End()
		c.f.End()
	})
}

// Size returns the width and the height of the Canvas in pixels.
func (c *Canvas) Size() (width, height float64) {
	return float64(c.f.Width()), float64(c.f.Height())
}

// Content returns a Picture that contains the content of this Canvas. The returned Picture changes
// as you draw onto the Canvas, so there is no real need to call this method more than once (but it
// might be beneficial to your code to do so).
func (c *Canvas) Content() *pixel.GLPicture {
	return pixel.PictureFromTexture(c.f.Texture())
}

// Clear fills the whole Canvas with one specified color.
func (c *Canvas) Clear(col color.Color) {
	mainthread.CallNonBlock(func() {
		c.f.Begin()
		col := pixel.NRGBAModel.Convert(col).(pixel.NRGBA)
		glhf.Clear(float32(col.R), float32(col.G), float32(col.B), float32(col.A))
		c.f.End()
	})
}

// Draw draws the content of the Canvas onto another Target. If no transform is applied, the content
// is fully stretched to fit the Target.
func (c *Canvas) Draw(t pixel.Target) {
	c.drawTd.Draw(t)
}

// MakeTriangles returns Triangles that draw onto this Canvas.
func (c *Canvas) MakeTriangles(t pixel.Triangles) pixel.TargetTriangles {
	gt := NewGLTriangles(c.s, t).(*glTriangles)
	return &canvasTriangles{
		c:           c,
		glTriangles: gt,
	}
}

// SetPicture sets a Picture that will be used in further draw operations.
//
// This does not set the Picture that this Canvas draws onto, don't confuse it.
func (c *Canvas) SetPicture(p *pixel.GLPicture) {
	if p != nil {
		min := pictureBounds(p, pixel.V(0, 0))
		max := pictureBounds(p, pixel.V(1, 1))
		c.bnd = mgl32.Vec4{
			float32(min.X()), float32(min.Y()),
			float32(max.X()), float32(max.Y()),
		}
	}
	c.pic = p
}

// SetTransform sets the transformations used in further draw operations.
func (c *Canvas) SetTransform(t ...pixel.Transform) {
	c.mat = transformToMat(t...)
}

// SetMaskColor sets the mask color used in further draw operations.
func (c *Canvas) SetMaskColor(col color.Color) {
	if col == nil {
		col = pixel.NRGBA{R: 1, G: 1, B: 1, A: 1}
	}
	nrgba := pixel.NRGBAModel.Convert(col).(pixel.NRGBA)
	r := float32(nrgba.R)
	g := float32(nrgba.G)
	b := float32(nrgba.B)
	a := float32(nrgba.A)
	c.col = mgl32.Vec4{r, g, b, a}
}

type canvasTriangles struct {
	c *Canvas
	*glTriangles
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
			ct.glTriangles.Draw()
			pic.Texture().End()
		} else {
			ct.glTriangles.Draw()
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

var canvasVertexFormat = glhf.AttrFormat{
	canvasPositionVec2: {Name: "position", Type: glhf.Vec2},
	canvasColorVec4:    {Name: "color", Type: glhf.Vec4},
	canvasTextureVec2:  {Name: "texture", Type: glhf.Vec2},
}

const (
	canvasMaskColorVec4 int = iota
	canvasTransformMat3
	canvasBoundsVec4
)

var canvasUniformFormat = glhf.AttrFormat{
	{Name: "maskColor", Type: glhf.Vec4},
	{Name: "transform", Type: glhf.Mat3},
	{Name: "bounds", Type: glhf.Vec4},
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
