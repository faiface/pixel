package pixelgl

import (
	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// GLShader is a type to assist with managing a canvas's underlying
// shader configuration. This allows for customization of shaders on
// a per canvas basis.
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

// reinitialize GLShader data and recompile the underlying gl shader object
func (gs *GLShader) update() {
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

// gets the uniform index from GLShader
func (gs *GLShader) getUniform(Name string) int {
	for i, u := range gs.uniforms {
		if u.Name == Name {
			return i
		}
	}
	return -1
}

// AddUniform appends a custom uniform name and value to the shader
//
// To add a time uniform for example:
//
// utime := float32(time.Since(starttime)).Seconds())
// mycanvas.shader.AddUniform("u_time", &utime)
//
func (gs *GLShader) AddUniform(Name string, Value interface{}) {
	Type := getAttrType(Value)
	if loc := gs.getUniform(Name); loc > -1 {
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

// Sets up a base shader with everything needed for a pixel
// canvas to render correctly. The defaults can be overridden
// by simply using AddUniform()
func baseShader(c *Canvas) {
	gs := &GLShader{
		vf: defaultCanvasVertexFormat,
		vs: defaultCanvasVertexShader,
		fs: baseCanvasFragmentShader,
	}

	gs.AddUniform("u_transform", &gs.uniformDefaults.transform)
	gs.AddUniform("u_colormask", &gs.uniformDefaults.colormask)
	gs.AddUniform("u_bounds", &gs.uniformDefaults.bounds)
	gs.AddUniform("u_texbounds", &gs.uniformDefaults.texbounds)

	c.shader = gs
}

var defaultCanvasVertexShader = `
#version 330 core

in vec2 position;
in vec4 color;
in vec2 texCoords;
in float intensity;

out vec4 Color;
out vec2 texcoords;
out float Intensity;

uniform mat3 u_transform;
uniform vec4 u_bounds;

void main() {
	vec2 transPos = (u_transform * vec3(position, 1.0)).xy;
	vec2 normPos = (transPos - u_bounds.xy) / u_bounds.zw * 2 - vec2(1, 1);
	gl_Position = vec4(normPos, 0.0, 1.0);
	Color = color;
	texcoords = texCoords;
	Intensity = intensity;
}
`

var baseCanvasFragmentShader = `
#version 330 core

in vec4 Color;
in vec2 texcoords;
in float Intensity;

out vec4 fragColor;

uniform vec4 u_colormask;
uniform vec4 u_texbounds;
uniform sampler2D u_texture;

void main() {
	if (Intensity == 0) {
		fragColor = u_colormask * Color;
	} else {
		fragColor = vec4(0, 0, 0, 0);
		fragColor += (1 - Intensity) * Color;
		vec2 t = (texcoords - u_texbounds.xy) / u_texbounds.zw;
		fragColor += Intensity * Color * texture(u_texture, t);
		fragColor *= u_colormask;
	}
}
`
