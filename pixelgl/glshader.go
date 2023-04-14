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
type GLShader struct {
	s      *glhf.Shader
	vf, uf glhf.AttrFormat
	vs, fs string

	uniforms []gsUniformAttr

	uniformDefaults struct {
		transform mgl32.Mat3
		colormask mgl32.Vec4
		bounds    mgl32.Vec4
		texbounds mgl32.Vec4
		cliprect  mgl32.Vec4
	}
}

type gsUniformAttr struct {
	Name      string
	Type      glhf.AttrType
	value     interface{}
	ispointer bool
}

const (
	canvasPosition int = iota
	canvasColor
	canvasTexCoords
	canvasIntensity
	canvasClip
)

var defaultCanvasVertexFormat = glhf.AttrFormat{
	canvasPosition:  glhf.Attr{Name: "aPosition", Type: glhf.Vec2},
	canvasColor:     glhf.Attr{Name: "aColor", Type: glhf.Vec4},
	canvasTexCoords: glhf.Attr{Name: "aTexCoords", Type: glhf.Vec2},
	canvasIntensity: glhf.Attr{Name: "aIntensity", Type: glhf.Float},
	canvasClip:      glhf.Attr{Name: "aClipRect", Type: glhf.Vec4},
}

// NewGLShader sets up a base shader with everything needed for a Pixel
// canvas to render correctly. The defaults can be overridden
// by simply using the SetUniform function.
func NewGLShader(fragmentShader string) *GLShader {
	gs := &GLShader{
		vf: defaultCanvasVertexFormat,
		vs: baseCanvasVertexShader,
		fs: fragmentShader,
	}

	gs.SetUniform("uTransform", &gs.uniformDefaults.transform)
	gs.SetUniform("uColorMask", &gs.uniformDefaults.colormask)
	gs.SetUniform("uBounds", &gs.uniformDefaults.bounds)
	gs.SetUniform("uTexBounds", &gs.uniformDefaults.texbounds)

	gs.Update()

	return gs
}

// Update reinitialize GLShader data and recompile the underlying gl shader object
func (gs *GLShader) Update() {
	gs.uf = make([]glhf.Attr, len(gs.uniforms))
	for idx := range gs.uniforms {
		gs.uf[idx] = glhf.Attr{
			Name: gs.uniforms[idx].Name,
			Type: gs.uniforms[idx].Type,
		}
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

// SetUniform appends a custom uniform name and value to the shader.
// if the uniform already exists, it will simply be overwritten.
//
// Example:
//
//	utime := float32(time.Since(starttime)).Seconds())
//	mycanvas.shader.AddUniform("u_time", &utime)
func (gs *GLShader) SetUniform(name string, value interface{}) {
	t, p := getAttrType(value)
	if loc := gs.getUniform(name); loc > -1 {
		gs.uniforms[loc].Name = name
		gs.uniforms[loc].Type = t
		gs.uniforms[loc].ispointer = p
		gs.uniforms[loc].value = value
		return
	}
	gs.uniforms = append(gs.uniforms, gsUniformAttr{
		Name:      name,
		Type:      t,
		ispointer: p,
		value:     value,
	})
}

// Value returns the attribute's concrete value. If the stored value
// is a pointer, we return the dereferenced value.
func (gu *gsUniformAttr) Value() interface{} {
	if !gu.ispointer {
		return gu.value
	}
	switch gu.Type {
	case glhf.Vec2:
		return *gu.value.(*mgl32.Vec2)
	case glhf.Vec3:
		return *gu.value.(*mgl32.Vec3)
	case glhf.Vec4:
		return *gu.value.(*mgl32.Vec4)
	case glhf.Mat2:
		return *gu.value.(*mgl32.Mat2)
	case glhf.Mat23:
		return *gu.value.(*mgl32.Mat2x3)
	case glhf.Mat24:
		return *gu.value.(*mgl32.Mat2x4)
	case glhf.Mat3:
		return *gu.value.(*mgl32.Mat3)
	case glhf.Mat32:
		return *gu.value.(*mgl32.Mat3x2)
	case glhf.Mat34:
		return *gu.value.(*mgl32.Mat3x4)
	case glhf.Mat4:
		return *gu.value.(*mgl32.Mat4)
	case glhf.Mat42:
		return *gu.value.(*mgl32.Mat4x2)
	case glhf.Mat43:
		return *gu.value.(*mgl32.Mat4x3)
	case glhf.Int:
		return *gu.value.(*int32)
	case glhf.Float:
		return *gu.value.(*float32)
	default:
		panic("invalid attrtype")
	}
}

// Returns the type identifier for any (supported) attribute variable type
// and whether or not it is a pointer of that type.
func getAttrType(v interface{}) (glhf.AttrType, bool) {
	switch v.(type) {
	case int32:
		return glhf.Int, false
	case float32:
		return glhf.Float, false
	case mgl32.Vec2:
		return glhf.Vec2, false
	case mgl32.Vec3:
		return glhf.Vec3, false
	case mgl32.Vec4:
		return glhf.Vec4, false
	case mgl32.Mat2:
		return glhf.Mat2, false
	case mgl32.Mat2x3:
		return glhf.Mat23, false
	case mgl32.Mat2x4:
		return glhf.Mat24, false
	case mgl32.Mat3:
		return glhf.Mat3, false
	case mgl32.Mat3x2:
		return glhf.Mat32, false
	case mgl32.Mat3x4:
		return glhf.Mat34, false
	case mgl32.Mat4:
		return glhf.Mat4, false
	case mgl32.Mat4x2:
		return glhf.Mat42, false
	case mgl32.Mat4x3:
		return glhf.Mat43, false
	case *mgl32.Vec2:
		return glhf.Vec2, true
	case *mgl32.Vec3:
		return glhf.Vec3, true
	case *mgl32.Vec4:
		return glhf.Vec4, true
	case *mgl32.Mat2:
		return glhf.Mat2, true
	case *mgl32.Mat2x3:
		return glhf.Mat23, true
	case *mgl32.Mat2x4:
		return glhf.Mat24, true
	case *mgl32.Mat3:
		return glhf.Mat3, true
	case *mgl32.Mat3x2:
		return glhf.Mat32, true
	case *mgl32.Mat3x4:
		return glhf.Mat34, true
	case *mgl32.Mat4:
		return glhf.Mat4, true
	case *mgl32.Mat4x2:
		return glhf.Mat42, true
	case *mgl32.Mat4x3:
		return glhf.Mat43, true
	case *int32:
		return glhf.Int, true
	case *float32:
		return glhf.Float, true
	default:
		panic("invalid AttrType")
	}
}

var baseCanvasVertexShader = `
#version 330 core

in vec2  aPosition;
in vec4  aColor;
in vec2  aTexCoords;
in float aIntensity;
in vec4  aClipRect;
in float aIsClipped;

out vec4  vColor;
out vec2  vTexCoords;
out float vIntensity;
out vec2  vPosition;
out vec4  vClipRect;

uniform mat3 uTransform;
uniform vec4 uBounds;

void main() {
	vec2 transPos = (uTransform * vec3(aPosition, 1.0)).xy;
	vec2 normPos = (transPos - uBounds.xy) / uBounds.zw * 2 - vec2(1, 1);
	gl_Position = vec4(normPos, 0.0, 1.0);

	vColor = aColor;
	vPosition = aPosition;
	vTexCoords = aTexCoords;
	vIntensity = aIntensity;
	vClipRect = aClipRect;
}
`

var baseCanvasFragmentShader = `
#version 330 core

in vec4  vColor;
in vec2  vTexCoords;
in float vIntensity;
in vec4  vClipRect;

out vec4 fragColor;

uniform vec4 uColorMask;
uniform vec4 uTexBounds;
uniform sampler2D uTexture;

void main() {
	if ((vClipRect != vec4(0,0,0,0)) && (gl_FragCoord.x < vClipRect.x || gl_FragCoord.y < vClipRect.y || gl_FragCoord.x > vClipRect.z || gl_FragCoord.y > vClipRect.w))
		discard;

	if (vIntensity == 0) {
		fragColor = uColorMask * vColor;
	} else {
		fragColor = vec4(0, 0, 0, 0);
		fragColor += (1 - vIntensity) * vColor;
		vec2 t = (vTexCoords - uTexBounds.xy) / uTexBounds.zw;
		fragColor += vIntensity * vColor * texture(uTexture, t);
		fragColor *= uColorMask;
	}
}
`
