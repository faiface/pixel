package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

// Shader is an OpenGL shader program.
type Shader struct {
	parent  Doer
	program uint32
}

// NewShader creates a new shader program from the specified vertex shader and fragment shader sources.
//
// Note that vertexShader and fragmentShader parameters must contain the source code, they're not filenames.
func NewShader(parent Doer, vertexShader, fragmentShader string) (*Shader, error) {
	shader := &Shader{
		parent: parent,
	}

	errChan := make(chan error, 1)
	parent.Do(func() {
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
			errChan <- err
			return
		}
		if glerr != nil {
			errChan <- err
			return
		}
		errChan <- nil
	})
	err := <-errChan
	if err != nil {
		return nil, err
	}

	return shader, nil
}

// Delete deletes a shader program. Don't use a shader after deletion.
func (s *Shader) Delete() {
	DoNoBlock(func() {
		gl.DeleteProgram(s.program)
	})
}

// Do stars using a shader, executes sub, and stops using it.
func (s *Shader) Do(sub func()) {
	s.parent.Do(func() {
		DoNoBlock(func() {
			gl.UseProgram(s.program)
		})
		sub()
		DoNoBlock(func() {
			gl.UseProgram(0)
		})
	})
}
