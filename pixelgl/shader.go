package pixelgl

import (
	"fmt"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

type (
	GLShader struct {
		s      *glhf.Shader
		vf, uf glhf.AttrFormat
		vs, fs string

		uniforms []gsUniformAttr

		uniformDefaults struct {
			transform mgl32.Mat3
			colormask mgl32.Vec4
			bounds    mgl32.Vec4
			texbounds mgl32.Vec4
		}
	}

	gsUniformAttr struct {
		Name  string
		Type  AttrType
		Value interface{}
	}
)

func (gs *GLShader) compile() {
	gs.uf = nil
	for _, u := range gs.uniforms {
		gs.uf = append(gs.uf, glhf.Attr{
			Name: u.Name,
			Type: glhf.AttrType(u.Type),
		})
	}
	var shader *glhf.Shader
	mainthread.Call(func() {
		var err error
		shader, err = glhf.NewShader(
			gs.vf,
			gs.uf,
			gs.vs,
			gs.fs,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create Canvas, there's a bug in the shader"))
		}
	})

	gs.s = shader
}
func (gs *GLShader) GetUniform(Name string) int {
	for i, u := range gs.uniforms {
		if u.Name == Name {
			return i
		}
	}
	return -1
}
func (gs *GLShader) AddUniform(Name string, Value interface{}) {
	Type := getUniformType(Value)
	fmt.Println(Type)
	if loc := gs.GetUniform(Name); loc > -1 {
		gs.uniforms[loc].Name = Name
		gs.uniforms[loc].Type = Type
		gs.uniforms[loc].Value = Value
		return
	}
	gs.uniforms = append(gs.uniforms, gsUniformAttr{
		Name:  Name,
		Type:  Type,
		Value: Value,
	})
}

func (c *Canvas) setUniforms(texbounds pixel.Rect) {
	mat := c.mat
	col := c.col
	c.shader.uniformDefaults.transform = mat
	c.shader.uniformDefaults.colormask = col
	dstBounds := c.Bounds()
	c.shader.uniformDefaults.bounds = mgl32.Vec4{
		float32(dstBounds.Min.X),
		float32(dstBounds.Min.Y),
		float32(dstBounds.W()),
		float32(dstBounds.H()),
	}

	for loc, u := range c.shader.uniforms {
		c.shader.s.SetUniformAttr(loc, u.Value)
	}
}

func baseShader(c *Canvas) {
	gs := &GLShader{
		vf: defaultCanvasVertexFormat,
		vs: defaultCanvasVertexShader,
		fs: baseCanvasFragmentShader,
	}

	gs.AddUniform("transform", &gs.uniformDefaults.transform)
	gs.AddUniform("colorMask", &gs.uniformDefaults.colormask)
	gs.AddUniform("bounds", &gs.uniformDefaults.bounds)
	gs.AddUniform("texBounds", &gs.uniformDefaults.texbounds)

	c.shader = gs
}
func getUniformType(v interface{}) AttrType {
	switch v.(type) {
	case int32:
		return Int
	case float32:
		return Float
	case mgl32.Vec2:
		return Vec2
	case mgl32.Vec3:
		return Vec3
	case mgl32.Vec4:
		return Vec4
	case mgl32.Mat2:
		return Mat2
	case mgl32.Mat2x3:
		return Mat23
	case mgl32.Mat2x4:
		return Mat24
	case mgl32.Mat3:
		return Mat3
	case mgl32.Mat3x2:
		return Mat32
	case mgl32.Mat3x4:
		return Mat34
	case mgl32.Mat4:
		return Mat4
	case mgl32.Mat4x2:
		return Mat42
	case mgl32.Mat4x3:
		return Mat43
	case *mgl32.Vec2:
		return Vec2p
	case *mgl32.Vec3:
		return Vec3p
	case *mgl32.Vec4:
		return Vec4p
	case *mgl32.Mat2:
		return Mat2p
	case *mgl32.Mat2x3:
		return Mat23p
	case *mgl32.Mat2x4:
		return Mat24p
	case *mgl32.Mat3:
		return Mat3p
	case *mgl32.Mat3x2:
		return Mat32p
	case *mgl32.Mat3x4:
		return Mat34p
	case *mgl32.Mat4:
		return Mat4p
	case *mgl32.Mat4x2:
		return Mat42p
	case *mgl32.Mat4x3:
		return Mat43p
	case *int32:
		return Intp
	case *float32:
		return Floatp
	default:
		panic("invalid AttrType")
	}
}

type AttrType int

// List of all possible attribute types.
const (
	Int AttrType = iota
	Float
	Vec2
	Vec3
	Vec4
	Mat2
	Mat23
	Mat24
	Mat3
	Mat32
	Mat34
	Mat4
	Mat42
	Mat43
	Intp
	Floatp
	Vec2p
	Vec3p
	Vec4p
	Mat2p
	Mat23p
	Mat24p
	Mat3p
	Mat32p
	Mat34p
	Mat4p
	Mat42p
	Mat43p
)

var defaultCanvasVertexShader = `
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

var baseCanvasFragmentShader = `
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
