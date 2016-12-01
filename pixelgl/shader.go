package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl64"
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
	parent   Doer
	format   UniformFormat
	program  uint32
	uniforms map[Attr]int32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're not filenames.
func NewShader(parent Doer, format UniformFormat, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		parent:   parent,
		format:   format,
		uniforms: make(map[Attr]int32),
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

			gl.DeleteShader(vshader)
			gl.DeleteShader(fshader)

			// uniforms
			for uname, utype := range format {
				ulocation := gl.GetUniformLocation(shader.program, gl.Str(uname+"\x00"))
				if ulocation == -1 {
					return fmt.Errorf("shader does not contain uniform '%s'", uname)
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

// SetUniformInt sets the value of an uniform attribute Attr{Purpose: purpose, Type: Int}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformInt(purpose AttrPurpose, value int32) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Int}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.Uniform1i(s.uniforms[attr], value)
	})
	return true
}

// SetUniformFloat sets the value of an uniform attribute Attr{Purpose: purpose, Type: Float}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformFloat(purpose AttrPurpose, value float64) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Float}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.Uniform1d(s.uniforms[attr], value)
	})
	return true
}

// SetUniformVec2 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec2}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec2(purpose AttrPurpose, value mgl64.Vec2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec2}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.Uniform2d(s.uniforms[attr], value[0], value[1])
	})
	return true
}

// SetUniformVec3 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec3}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec3(purpose AttrPurpose, value mgl64.Vec3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec3}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.Uniform3d(s.uniforms[attr], value[0], value[1], value[2])
	})
	return true
}

// SetUniformVec4 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Vec4}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformVec4(purpose AttrPurpose, value mgl64.Vec4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Vec4}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.Uniform4d(s.uniforms[attr], value[0], value[1], value[2], value[3])
	})
	return true
}

// SetUniformMat2 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat2}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat2(purpose AttrPurpose, value mgl64.Mat2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat2}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix2dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat23 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat23}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat23(purpose AttrPurpose, value mgl64.Mat2x3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat23}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix2x3dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat24 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat24}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat24(purpose AttrPurpose, value mgl64.Mat2x4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat24}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix2x4dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat3 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat3}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat3(purpose AttrPurpose, value mgl64.Mat3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat3}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix3dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat32 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat32}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat32(purpose AttrPurpose, value mgl64.Mat3x2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat32}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix3x2dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat34 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat34}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat34(purpose AttrPurpose, value mgl64.Mat3x4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat34}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix3x4dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat4 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat4}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat4(purpose AttrPurpose, value mgl64.Mat4) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat4}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix4dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat42 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat42}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat42(purpose AttrPurpose, value mgl64.Mat4x2) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat42}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix4x2dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// SetUniformMat43 sets the value of an uniform attribute Attr{Purpose: purpose, Type: Mat43}.
//
// Returns false if the attribute does not exist.
func (s *Shader) SetUniformMat43(purpose AttrPurpose, value mgl64.Mat4x3) (ok bool) {
	attr := Attr{Purpose: purpose, Type: Mat43}
	if _, ok := s.uniforms[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.UniformMatrix4x3dv(s.uniforms[attr], 1, false, &value[0])
	})
	return true
}

// Do stars using a shader, executes sub, and stops using it.
func (s *Shader) Do(sub func(Context)) {
	s.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.UseProgram(s.program)
		})
		sub(ctx.WithShader(s))
		DoNoBlock(func() {
			gl.UseProgram(0)
		})
	})
}
