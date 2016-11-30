package pixelgl

import (
	"fmt"

	"github.com/go-gl/gl/v3.3-core/gl"
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
