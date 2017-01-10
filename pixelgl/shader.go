package pixelgl

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// Shader is an OpenGL shader program.
type Shader struct {
	program    binder
	vertexFmt  AttrFormat
	uniformFmt AttrFormat
	uniforms   map[string]int32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader
// sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're
// not filenames.
func NewShader(vertexFmt, uniformFmt AttrFormat, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		program: binder{
			restoreLoc: gl.CURRENT_PROGRAM,
			bindFunc: func(obj uint32) {
				gl.UseProgram(obj)
			},
		},
		vertexFmt:  vertexFmt,
		uniformFmt: uniformFmt,
		uniforms:   make(map[string]int32),
	}

	var vshader, fshader uint32

	// vertex shader
	{
		vshader = gl.CreateShader(gl.VERTEX_SHADER)
		src, free := gl.Strs(vertexShader)
		defer free()
		length := int32(len(vertexShader))
		gl.ShaderSource(vshader, 1, src, &length)
		gl.CompileShader(vshader)

		var (
			success int32
			infoLog = make([]byte, 512)
		)
		gl.GetShaderiv(vshader, gl.COMPILE_STATUS, &success)
		if success == 0 {
			gl.GetShaderInfoLog(vshader, int32(len(infoLog)), nil, &infoLog[0])
			return nil, fmt.Errorf("error compiling vertex shader: %s", string(infoLog))
		}

		defer gl.DeleteShader(vshader)
	}

	// fragment shader
	{
		fshader = gl.CreateShader(gl.FRAGMENT_SHADER)
		src, free := gl.Strs(fragmentShader)
		defer free()
		length := int32(len(fragmentShader))
		gl.ShaderSource(fshader, 1, src, &length)
		gl.CompileShader(fshader)

		var (
			success int32
			infoLog = make([]byte, 512)
		)
		gl.GetShaderiv(fshader, gl.COMPILE_STATUS, &success)
		if success == 0 {
			gl.GetShaderInfoLog(fshader, int32(len(infoLog)), nil, &infoLog[0])
			return nil, fmt.Errorf("error compiling fragment shader: %s", string(infoLog))
		}

		defer gl.DeleteShader(fshader)
	}

	// shader program
	{
		shader.program.obj = gl.CreateProgram()
		gl.AttachShader(shader.program.obj, vshader)
		gl.AttachShader(shader.program.obj, fshader)
		gl.LinkProgram(shader.program.obj)

		var (
			success int32
			infoLog = make([]byte, 512)
		)
		gl.GetProgramiv(shader.program.obj, gl.LINK_STATUS, &success)
		if success == 0 {
			gl.GetProgramInfoLog(shader.program.obj, int32(len(infoLog)), nil, &infoLog[0])
			return nil, fmt.Errorf("error linking shader program: %s", string(infoLog))
		}
	}

	// uniforms
	for name := range uniformFmt {
		loc := gl.GetUniformLocation(shader.program.obj, gl.Str(name+"\x00"))
		shader.uniforms[name] = loc
	}

	runtime.SetFinalizer(shader, (*Shader).delete)

	return shader, nil
}

func (s *Shader) delete() {
	DoNoBlock(func() {
		gl.DeleteProgram(s.program.obj)
	})
}

// VertexFormat returns the vertex attribute format of this shader. Do not change it.
func (s *Shader) VertexFormat() AttrFormat {
	return s.vertexFmt
}

// UniformFormat returns the uniform attribute format of this shader. Do not change it.
func (s *Shader) UniformFormat() AttrFormat {
	return s.uniformFmt
}

// SetUniformAttr sets the value of a uniform attribute of a shader.
//
// If the attribute does not exist, this method returns false.
//
// Supplied value must correspond to the type of the attribute. Correct types are these
// (right-hand is the type of the value):
//   Attr{Type: Int}:   int32
//   Attr{Type: Float}: float32
//   Attr{Type: Vec2}:  mgl32.Vec2
//   Attr{Type: Vec3}:  mgl32.Vec3
//   Attr{Type: Vec4}:  mgl32.Vec4
//   Attr{Type: Mat2}:  mgl32.Mat2
//   Attr{Type: Mat23}: mgl32.Mat2x3
//   Attr{Type: Mat24}: mgl32.Mat2x4
//   Attr{Type: Mat3}:  mgl32.Mat3
//   Attr{Type: Mat32}: mgl32.Mat3x2
//   Attr{Type: Mat34}: mgl32.Mat3x4
//   Attr{Type: Mat4}:  mgl32.Mat4
//   Attr{Type: Mat42}: mgl32.Mat4x2
//   Attr{Type: Mat43}: mgl32.Mat4x3
// No other types are supported.
//
// The shader must be bound before calling this method.
func (s *Shader) SetUniformAttr(attr Attr, value interface{}) (ok bool) {
	if !s.uniformFmt.Contains(attr) {
		return false
	}

	switch attr.Type {
	case Int:
		value := value.(int32)
		gl.Uniform1iv(s.uniforms[attr.Name], 1, &value)
	case Float:
		value := value.(float32)
		gl.Uniform1fv(s.uniforms[attr.Name], 1, &value)
	case Vec2:
		value := value.(mgl32.Vec2)
		gl.Uniform2fv(s.uniforms[attr.Name], 1, &value[0])
	case Vec3:
		value := value.(mgl32.Vec3)
		gl.Uniform3fv(s.uniforms[attr.Name], 1, &value[0])
	case Vec4:
		value := value.(mgl32.Vec4)
		gl.Uniform4fv(s.uniforms[attr.Name], 1, &value[0])
	case Mat2:
		value := value.(mgl32.Mat2)
		gl.UniformMatrix2fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat23:
		value := value.(mgl32.Mat2x3)
		gl.UniformMatrix2x3fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat24:
		value := value.(mgl32.Mat2x4)
		gl.UniformMatrix2x4fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat3:
		value := value.(mgl32.Mat3)
		gl.UniformMatrix3fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat32:
		value := value.(mgl32.Mat3x2)
		gl.UniformMatrix3x2fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat34:
		value := value.(mgl32.Mat3x4)
		gl.UniformMatrix3x4fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat4:
		value := value.(mgl32.Mat4)
		gl.UniformMatrix4fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat42:
		value := value.(mgl32.Mat4x2)
		gl.UniformMatrix4x2fv(s.uniforms[attr.Name], 1, false, &value[0])
	case Mat43:
		value := value.(mgl32.Mat4x3)
		gl.UniformMatrix4x3fv(s.uniforms[attr.Name], 1, false, &value[0])
	default:
		panic("set uniform attr: invalid attribute type")
	}

	return true
}

// Begin binds a shader program. This is necessary before using the shader.
func (s *Shader) Begin() {
	s.program.bind()
}

// End unbinds a shader program and restores the previous one.
func (s *Shader) End() {
	s.program.restore()
}
