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

// Canvas is an off-screen rectangular BasicTarget and Picture at the same time, that you can draw
// onto.
//
// It supports TrianglesPosition, TrianglesColor, TrianglesPicture and PictureColor.
type Canvas struct {
	gf     *GLFrame
	shader *glhf.Shader

	cmp    pixel.ComposeMethod
	mat    mgl32.Mat3
	col    mgl32.Vec4
	smooth bool

	sprite *pixel.Sprite
}

var _ pixel.ComposeTarget = (*Canvas)(nil)

// NewCanvas creates a new empty, fully transparent Canvas with given bounds.
func NewCanvas(bounds pixel.Rect) *Canvas {
	c := &Canvas{
		gf:  NewGLFrame(bounds),
		mat: mgl32.Ident3(),
		col: mgl32.Vec4{1, 1, 1, 1},
	}

	c.SetBounds(bounds)

	var shader *glhf.Shader
	mainthread.Call(func() {
		var err error
		shader, err = glhf.NewShader(
			canvasVertexFormat,
			canvasUniformFormat,
			canvasVertexShader,
			canvasFragmentShader,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create Canvas, there's a bug in the shader"))
		}
	})
	c.shader = shader

	return c
}

// MakeTriangles creates a specialized copy of the supplied Triangles that draws onto this Canvas.
//
// TrianglesPosition, TrianglesColor and TrianglesPicture are supported.
func (c *Canvas) MakeTriangles(t pixel.Triangles) pixel.TargetTriangles {
	return &canvasTriangles{
		GLTriangles: NewGLTriangles(c.shader, t),
		dst:         c,
	}
}

// MakePicture create a specialized copy of the supplied Picture that draws onto this Canvas.
//
// PictureColor is supported.
func (c *Canvas) MakePicture(p pixel.Picture) pixel.TargetPicture {
	if cp, ok := p.(*canvasPicture); ok {
		return &canvasPicture{
			GLPicture: cp.GLPicture,
			dst:       c,
		}
	}
	if gp, ok := p.(GLPicture); ok {
		return &canvasPicture{
			GLPicture: gp,
			dst:       c,
		}
	}
	return &canvasPicture{
		GLPicture: NewGLPicture(p),
		dst:       c,
	}
}

// SetMatrix sets a Matrix that every point will be projected by.
func (c *Canvas) SetMatrix(m pixel.Matrix) {
	for i := range m {
		c.mat[i] = float32(m[i])
	}
}

// SetColorMask sets a color that every color in triangles or a picture will be multiplied by.
func (c *Canvas) SetColorMask(col color.Color) {
	rgba := pixel.Alpha(1)
	if col != nil {
		rgba = pixel.ToRGBA(col)
	}
	c.col = mgl32.Vec4{
		float32(rgba.R),
		float32(rgba.G),
		float32(rgba.B),
		float32(rgba.A),
	}
}

// SetComposeMethod sets a Porter-Duff composition method to be used in the following draws onto
// this Canvas.
func (c *Canvas) SetComposeMethod(cmp pixel.ComposeMethod) {
	c.cmp = cmp
}

// SetBounds resizes the Canvas to the new bounds. Old content will be preserved.
func (c *Canvas) SetBounds(bounds pixel.Rect) {
	c.gf.SetBounds(bounds)
	if c.sprite == nil {
		c.sprite = pixel.NewSprite(nil, pixel.Rect{})
	}
	c.sprite.Set(c, c.Bounds())
	//c.sprite.SetMatrix(pixel.IM.Moved(c.Bounds().Center()))
}

// Bounds returns the rectangular bounds of the Canvas.
func (c *Canvas) Bounds() pixel.Rect {
	return c.gf.Bounds()
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
	_, _, bw, bh := intBounds(c.gf.Bounds())
	glhf.Bounds(0, 0, bw, bh)
}

// must be manually called inside mainthread
func setBlendFunc(cmp pixel.ComposeMethod) {
	switch cmp {
	case pixel.ComposeOver:
		glhf.BlendFunc(glhf.One, glhf.OneMinusSrcAlpha)
	case pixel.ComposeIn:
		glhf.BlendFunc(glhf.DstAlpha, glhf.Zero)
	case pixel.ComposeOut:
		glhf.BlendFunc(glhf.OneMinusDstAlpha, glhf.Zero)
	case pixel.ComposeAtop:
		glhf.BlendFunc(glhf.DstAlpha, glhf.OneMinusSrcAlpha)
	case pixel.ComposeRover:
		glhf.BlendFunc(glhf.OneMinusDstAlpha, glhf.One)
	case pixel.ComposeRin:
		glhf.BlendFunc(glhf.Zero, glhf.SrcAlpha)
	case pixel.ComposeRout:
		glhf.BlendFunc(glhf.Zero, glhf.OneMinusSrcAlpha)
	case pixel.ComposeRatop:
		glhf.BlendFunc(glhf.OneMinusDstAlpha, glhf.SrcAlpha)
	case pixel.ComposeXor:
		glhf.BlendFunc(glhf.OneMinusDstAlpha, glhf.OneMinusSrcAlpha)
	case pixel.ComposePlus:
		glhf.BlendFunc(glhf.One, glhf.One)
	case pixel.ComposeCopy:
		glhf.BlendFunc(glhf.One, glhf.Zero)
	default:
		panic(errors.New("Canvas: invalid compose method"))
	}
}

// Clear fills the whole Canvas with a single color.
func (c *Canvas) Clear(color color.Color) {
	c.gf.Dirty()

	rgba := pixel.ToRGBA(color)

	// color masking
	rgba = rgba.Mul(pixel.RGBA{
		R: float64(c.col[0]),
		G: float64(c.col[1]),
		B: float64(c.col[2]),
		A: float64(c.col[3]),
	})

	mainthread.CallNonBlock(func() {
		c.setGlhfBounds()
		c.gf.Frame().Begin()
		glhf.Clear(
			float32(rgba.R),
			float32(rgba.G),
			float32(rgba.B),
			float32(rgba.A),
		)
		c.gf.Frame().End()
	})
}

// Color returns the color of the pixel over the given position inside the Canvas.
func (c *Canvas) Color(at pixel.Vec) pixel.RGBA {
	return c.gf.Color(at)
}

// Texture returns the underlying OpenGL Texture of this Canvas.
//
// Implements GLPicture interface.
func (c *Canvas) Texture() *glhf.Texture {
	return c.gf.Texture()
}

// Frame returns the underlying OpenGL Frame of this Canvas.
func (c *Canvas) Frame() *glhf.Frame {
	return c.gf.frame
}

// SetPixels replaces the content of the Canvas with the provided pixels. The provided slice must be
// an alpha-premultiplied RGBA sequence of correct length (4 * width * height).
func (c *Canvas) SetPixels(pixels []uint8) {
	c.gf.Dirty()

	mainthread.Call(func() {
		tex := c.Texture()
		tex.Begin()
		tex.SetPixels(0, 0, tex.Width(), tex.Height(), pixels)
		tex.End()
	})
}

// Pixels returns an alpha-premultiplied RGBA sequence of the content of the Canvas.
func (c *Canvas) Pixels() []uint8 {
	var pixels []uint8

	mainthread.Call(func() {
		tex := c.Texture()
		tex.Begin()
		pixels = tex.Pixels(0, 0, tex.Width(), tex.Height())
		tex.End()
	})

	return pixels
}

// Draw draws the content of the Canvas onto another Target, transformed by the given Matrix, just
// like if it was a Sprite containing the whole Canvas.
func (c *Canvas) Draw(t pixel.Target, matrix pixel.Matrix) {
	c.sprite.Draw(t, matrix)
}

// DrawColorMask draws the content of the Canvas onto another Target, transformed by the given
// Matrix and multiplied by the given mask, just like if it was a Sprite containing the whole Canvas.
//
// If the color mask is nil, a fully opaque white mask will be used causing no effect.
func (c *Canvas) DrawColorMask(t pixel.Target, matrix pixel.Matrix, mask color.Color) {
	c.sprite.DrawColorMask(t, matrix, mask)
}

type canvasTriangles struct {
	*GLTriangles
	dst *Canvas
}

func (ct *canvasTriangles) draw(tex *glhf.Texture, bounds pixel.Rect) {
	ct.dst.gf.Dirty()

	// save the current state vars to avoid race condition
	cmp := ct.dst.cmp
	mat := ct.dst.mat
	col := ct.dst.col
	smt := ct.dst.smooth

	mainthread.CallNonBlock(func() {
		ct.dst.setGlhfBounds()
		setBlendFunc(cmp)

		frame := ct.dst.gf.Frame()
		shader := ct.dst.shader

		frame.Begin()
		shader.Begin()

		dstBounds := ct.dst.Bounds()
		shader.SetUniformAttr(canvasBounds, mgl32.Vec4{
			float32(dstBounds.Min.X),
			float32(dstBounds.Min.Y),
			float32(dstBounds.W()),
			float32(dstBounds.H()),
		})
		shader.SetUniformAttr(canvasTransform, mat)
		shader.SetUniformAttr(canvasColorMask, col)

		if tex == nil {
			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()
		} else {
			tex.Begin()

			bx, by, bw, bh := intBounds(bounds)
			shader.SetUniformAttr(canvasTexBounds, mgl32.Vec4{
				float32(bx),
				float32(by),
				float32(bw),
				float32(bh),
			})

			if tex.Smooth() != smt {
				tex.SetSmooth(smt)
			}

			ct.vs.Begin()
			ct.vs.Draw()
			ct.vs.End()

			tex.End()
		}

		shader.End()
		frame.End()
	})
}

func (ct *canvasTriangles) Draw() {
	ct.draw(nil, pixel.Rect{})
}

type canvasPicture struct {
	GLPicture
	dst *Canvas
}

func (cp *canvasPicture) Draw(t pixel.TargetTriangles) {
	ct := t.(*canvasTriangles)
	if cp.dst != ct.dst {
		panic(fmt.Errorf("(%T).Draw: TargetTriangles generated by different Canvas", cp))
	}
	ct.draw(cp.GLPicture.Texture(), cp.GLPicture.Bounds())
}

const (
	canvasPosition int = iota
	canvasColor
	canvasTexCoords
	canvasIntensity
)

var canvasVertexFormat = glhf.AttrFormat{
	canvasPosition:  {Name: "position", Type: glhf.Vec2},
	canvasColor:     {Name: "color", Type: glhf.Vec4},
	canvasTexCoords: {Name: "texCoords", Type: glhf.Vec2},
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
in vec2 texCoords;
in float intensity;

out vec4 Color;
out vec2 TexCoords;
out float Intensity;

uniform mat3 transform;
uniform vec4 bounds;

void main() {
	vec2 transPos = (transform * vec3(position, 1.0)).xy;
	vec2 normPos = (transPos - bounds.xy) / bounds.zw * 2 - vec2(1, 1);
	gl_Position = vec4(normPos, 0.0, 1.0);
	Color = color;
	TexCoords = texCoords;
	Intensity = intensity;
}
`

var canvasFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 TexCoords;
in float Intensity;

out vec4 color;

uniform vec4 colorMask;
uniform vec4 texBounds;
uniform sampler2D tex;

void main() {
	if (Intensity == 0) {
		color = colorMask * Color;
	} else {
		color = vec4(0, 0, 0, 0);
		color += (1 - Intensity) * Color;
		vec2 t = (TexCoords - texBounds.xy) / texBounds.zw;
		color += Intensity * Color * texture(tex, t);
		color *= colorMask;
	}
}
`
