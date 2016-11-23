package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

// Shader is an OpenGL shader program.
type Shader struct {
	parent  BeginEnder
	program uint32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're not filenames.
func NewShader(parent BeginEnder, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		parent: parent,
	}
	err, glerr := DoErrGLErr(func() error {
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

		return nil
	})
	if err != nil {
		if glerr != nil {
			err = errors.Wrap(glerr, err.Error())
		}
		return nil, err
	}
	if glerr != nil {
		return nil, glerr
	}
	return shader, nil
}

// Delete deletes a shader program. Don't use a shader after deletion.
func (s *Shader) Delete() {
	Do(func() {
		gl.DeleteProgram(s.program)
	})
}

// Begin starts using a shader program.
func (s *Shader) Begin() {
	s.parent.Begin()
	Do(func() {
		gl.UseProgram(s.program)
	})
}

// End stops using a shader program.
func (s *Shader) End() {
	Do(func() {
		gl.UseProgram(0)
	})
	s.parent.End()
}
