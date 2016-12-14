package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

// UniformFormat defines names, purposes and types of uniform variables inside a shader.
//
// Example:
//
//   UniformFormat{"transform": {Transform, Mat3}, "camera": {Camera, Mat3}}
type UniformFormat map[string]Attr

// Shader is an OpenGL shader program.
type Shader struct {
	parent        Doer
	program       binder
	vertexFormat  VertexFormat
	uniformFormat UniformFormat
	uniforms      map[Attr]int32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're not filenames.
func NewShader(parent Doer, vertexFormat VertexFormat, uniformFormat UniformFormat, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		parent: parent,
		program: binder{
			restoreLoc: gl.CURRENT_PROGRAM,
			bindFunc: func(obj uint32) {
				gl.UseProgram(obj)
			},
		},
		vertexFormat:  vertexFormat,
		uniformFormat: uniformFormat,
		uniforms:      make(map[Attr]int32),
	}

	var err error
	parent.Do(func(ctx Context) {
		err = DoErr(func() error {
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
					return fmt.Errorf("error compiling vertex shader: %s", string(infoLog))
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
					return fmt.Errorf("error compiling fragment shader: %s", string(infoLog))
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
					return fmt.Errorf("error linking shader program: %s", string(infoLog))
				}
			}

			// uniforms
			for uname, utype := range uniformFormat {
				ulocation := gl.GetUniformLocation(shader.program.obj, gl.Str(uname+"\x00"))
				if ulocation == -1 {
					gl.DeleteProgram(shader.program.obj)
					return fmt.Errorf("shader does not contain uniform '%s'", uname)
				}
				if _, ok := shader.uniforms[utype]; ok {
					gl.DeleteProgram(shader.program.obj)
					return fmt.Errorf("failed to create shader: invalid uniform format: duplicate uniform attribute")
				}
				shader.uniforms[utype] = ulocation
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return shader, nil
}

// Delete deletes a shader program. Don't use a shader after deletion.
func (s *Shader) Delete() {
	s.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.DeleteProgram(s.program.obj)
		})
	})
}

// ID returns an OpenGL identifier of a shader program.
func (s *Shader) ID() uint32 {
	return s.program.obj
}

// VertexFormat returns the vertex attribute format of this shader. Do not change it.
func (s *Shader) VertexFormat() VertexFormat {
	return s.vertexFormat
}

// UniformFormat returns the uniform attribute format of this shader. Do not change it.
func (s *Shader) UniformFormat() UniformFormat {
	return s.uniformFormat
}

// SetUniformAttr sets the value of a uniform attribute of a shader.
//
// If the attribute does not exist, this method returns false.
//
// Supplied value must correspond to the type of the attribute. Correct types are these (right-hand is the type of the value):
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
func (s *Shader) SetUniformAttr(attr Attr, value interface{}) (ok bool) {
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}

	DoNoBlock(func() {
		defer s.program.bind().restore()

		switch attr.Type {
		case Int:
			value := value.(int32)
			gl.Uniform1iv(s.uniforms[attr], 1, &value)
		case Float:
			value := value.(float32)
			gl.Uniform1fv(s.uniforms[attr], 1, &value)
		case Vec2:
			value := value.(mgl32.Vec2)
			gl.Uniform2fv(s.uniforms[attr], 1, &value[0])
		case Vec3:
			value := value.(mgl32.Vec3)
			gl.Uniform3fv(s.uniforms[attr], 1, &value[0])
		case Vec4:
			value := value.(mgl32.Vec4)
			gl.Uniform4fv(s.uniforms[attr], 1, &value[0])
		case Mat2:
			value := value.(mgl32.Mat2)
			gl.UniformMatrix2fv(s.uniforms[attr], 1, false, &value[0])
		case Mat23:
			value := value.(mgl32.Mat2x3)
			gl.UniformMatrix2x3fv(s.uniforms[attr], 1, false, &value[0])
		case Mat24:
			value := value.(mgl32.Mat2x4)
			gl.UniformMatrix2x4fv(s.uniforms[attr], 1, false, &value[0])
		case Mat3:
			value := value.(mgl32.Mat3)
			gl.UniformMatrix3fv(s.uniforms[attr], 1, false, &value[0])
		case Mat32:
			value := value.(mgl32.Mat3x2)
			gl.UniformMatrix3x2fv(s.uniforms[attr], 1, false, &value[0])
		case Mat34:
			value := value.(mgl32.Mat3x4)
			gl.UniformMatrix3x4fv(s.uniforms[attr], 1, false, &value[0])
		case Mat4:
			value := value.(mgl32.Mat4)
			gl.UniformMatrix4fv(s.uniforms[attr], 1, false, &value[0])
		case Mat42:
			value := value.(mgl32.Mat4x2)
			gl.UniformMatrix4x2fv(s.uniforms[attr], 1, false, &value[0])
		case Mat43:
			value := value.(mgl32.Mat4x3)
			gl.UniformMatrix4x3fv(s.uniforms[attr], 1, false, &value[0])
		default:
			panic("set uniform attr: invalid attribute type")
		}
	})

	return true
}

// Do stars using a shader, executes sub, and stops using it.
func (s *Shader) Do(sub func(Context)) {
	s.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			s.program.bind()
		})
		sub(ctx.WithShader(s))
		DoNoBlock(func() {
			s.program.restore()
		})
	})
}
