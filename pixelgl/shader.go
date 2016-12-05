package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// UniformFormat defines names, purposes and types of uniform variables inside a shader.
//
// Example:
//
//   UniformFormat{"transform": {Transform, Mat3}, "camera": {Camera, Mat3}}
type UniformFormat map[string]Attr

// Shader is an OpenGL shader program.
type Shader struct {
	enabled       bool
	parent        Doer
	program       uint32
	vertexFormat  VertexFormat
	uniformFormat UniformFormat
	uniforms      map[Attr]int32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're not filenames.
func NewShader(parent Doer, vertexFormat VertexFormat, uniformFormat UniformFormat, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		parent:        parent,
		vertexFormat:  vertexFormat,
		uniformFormat: uniformFormat,
		uniforms:      make(map[Attr]int32),
	}

	var err, glerr error
	parent.Do(func(ctx Context) {
		err, glerr = DoErrGLErr(func() error {
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
				shader.program = gl.CreateProgram()
				gl.AttachShader(shader.program, vshader)
				gl.AttachShader(shader.program, fshader)
				gl.LinkProgram(shader.program)

				var (
					success int32
					infoLog = make([]byte, 512)
				)
				gl.GetProgramiv(shader.program, gl.LINK_STATUS, &success)
				if success == 0 {
					gl.GetProgramInfoLog(shader.program, int32(len(infoLog)), nil, &infoLog[0])
					return fmt.Errorf("error linking shader program: %s", string(infoLog))
				}
			}

			// uniforms
			for uname, utype := range uniformFormat {
				ulocation := gl.GetUniformLocation(shader.program, gl.Str(uname+"\x00"))
				if ulocation == -1 {
					gl.DeleteProgram(shader.program)
					return fmt.Errorf("shader does not contain uniform '%s'", uname)
				}
				if _, ok := shader.uniforms[utype]; ok {
					gl.DeleteProgram(shader.program)
					return fmt.Errorf("failed to create shader: invalid uniform format: duplicate uniform attribute")
				}
				shader.uniforms[utype] = ulocation
			}

			return nil
		})
	})
	if err != nil && glerr != nil {
		return nil, errors.Wrap(glerr, err.Error())
	}
	if err != nil {
		return nil, err
	}
	if glerr != nil {
		return nil, glerr
	}

	return shader, nil
}

// Delete deletes a shader program. Don't use a shader after deletion.
func (s *Shader) Delete() {
	s.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.DeleteProgram(s.program)
		})
	})
}

// ID returns an OpenGL identifier of a shader program.
func (s *Shader) ID() uint32 {
	return s.program
}

// VertexFormat returns the vertex attribute format of this shader. Do not change it.
func (s *Shader) VertexFormat() VertexFormat {
	return s.vertexFormat
}

// UniformFormat returns the uniform attribute format of this shader. Do not change it.
func (s *Shader) UniformFormat() UniformFormat {
	return s.uniformFormat
}

// SetUniformInt sets the value of an uniform attribute Attr{Purpose: purpose, Type: Int}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformInt(purpose AttrPurpose, value int32) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Int}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.Uniform1i(s.uniforms[attr], value)
		})
	})
	return true
}

// SetUniformFloat sets the value of an uniform attribute Attr{Purpose: purpose, Type: Float}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformFloat(purpose AttrPurpose, value float32) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Float}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.Uniform1f(s.uniforms[attr], value)
		})
	})
	return true
}

// SetUniformVec2 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec2}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec2(purpose AttrPurpose, value mgl32.Vec2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec2}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.Uniform2f(s.uniforms[attr], value[0], value[1])
		})
	})
	return true
}

// SetUniformVec3 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec3}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec3(purpose AttrPurpose, value mgl32.Vec3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec3}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.Uniform3f(s.uniforms[attr], value[0], value[1], value[2])
		})
	})
	return true
}

// SetUniformVec4 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec4}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec4(purpose AttrPurpose, value mgl32.Vec4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec4}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.Uniform4f(s.uniforms[attr], value[0], value[1], value[2], value[3])
		})
	})
	return true
}

// SetUniformMat2 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat2}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat2(purpose AttrPurpose, value mgl32.Mat2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat2}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix2fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat23 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat23}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat23(purpose AttrPurpose, value mgl32.Mat2x3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat23}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix2x3fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat24 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat24}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat24(purpose AttrPurpose, value mgl32.Mat2x4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat24}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix2x4fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat3 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat3}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat3(purpose AttrPurpose, value mgl32.Mat3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat3}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix3fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat32 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat32}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat32(purpose AttrPurpose, value mgl32.Mat3x2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat32}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix3x2fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat34 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat34}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat34(purpose AttrPurpose, value mgl32.Mat3x4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat34}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix3x4fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat4 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat4}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat4(purpose AttrPurpose, value mgl32.Mat4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat4}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix4fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat42 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat42}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat42(purpose AttrPurpose, value mgl32.Mat4x2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat42}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix4x2fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// SetUniformMat43 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat43}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat43(purpose AttrPurpose, value mgl32.Mat4x3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat43}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	s.Do(func(Context) {
		DoNoBlock(func() {
			gl.UniformMatrix4x3fv(s.uniforms[attr], 1, false, &value[0])
		})
	})
	return true
}

// Do stars using a shader, executes sub, and stops using it.
func (s *Shader) Do(sub func(Context)) {
	s.parent.Do(func(ctx Context) {
		if s.enabled {
			sub(ctx.WithShader(s))
			return
		}
		DoNoBlock(func() {
			gl.UseProgram(s.program)
		})
		s.enabled = true
		sub(ctx.WithShader(s))
		s.enabled = false
		DoNoBlock(func() {
			gl.UseProgram(0)
		})
	})
}
