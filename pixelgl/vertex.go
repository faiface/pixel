package pixelgl

import (
	"fmt"
	"runtime"

	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

// VertexSlice points to a portion of (or possibly whole) vertex array. It is used as a pointer,
// contrary to Go's builtin slices. This is, so that append can be 'in-place'. That's for the good,
// because Begin/End-ing a VertexSlice would become super confusing, if append returned a new
// VertexSlice.
//
// It also implements all basic slice-like operations: appending, sub-slicing, etc.
//
// Note that you need to Begin a VertexSlice before getting or updating it's elements or drawing it.
// After you're done with it, you need to End it.
type VertexSlice struct {
	va   *vertexArray
	i, j int
}

// MakeVertexSlice allocates a new vertex array with specified capacity and returns a VertexSlice
// that points to it's first len elements.
//
// Note, that a vertex array is specialized for a specific shader and can't be used with another
// shader.
func MakeVertexSlice(shader *Shader, len, cap int) *VertexSlice {
	if len > cap {
		panic("failed to make vertex slice: len > cap")
	}
	return &VertexSlice{
		va: newVertexArray(shader, cap),
		i:  0,
		j:  len,
	}
}

// VertexFormat returns the format of vertex attributes inside the underlying vertex array of this
// VertexSlice.
func (vs *VertexSlice) VertexFormat() AttrFormat {
	return vs.va.format
}

// Stride returns the number of float32 elements occupied by one vertex.
func (vs *VertexSlice) Stride() int {
	return vs.va.stride / 4
}

// Len returns the length of the VertexSlice.
func (vs *VertexSlice) Len() int {
	return vs.j - vs.i
}

// Cap returns the capacity of an underlying vertex array.
func (vs *VertexSlice) Cap() int {
	return vs.va.cap - vs.i
}

// Slice returns a sub-slice of this VertexSlice covering the range [i, j) (relative to this
// VertexSlice).
//
// Note, that the returned VertexSlice shares an underlying vertex array with the original
// VertexSlice. Modifying the contents of one modifies corresponding contents of the other.
func (vs *VertexSlice) Slice(i, j int) *VertexSlice {
	if i < 0 || j < i || j > vs.va.cap {
		panic("failed to slice vertex slice: index out of range")
	}
	return &VertexSlice{
		va: vs.va,
		i:  vs.i + i,
		j:  vs.i + j,
	}
}

// grow returns supplied vs with length changed to len. Allocates new underlying vertex array if
// necessary. The original content is preserved.
func (vs VertexSlice) grow(len int) VertexSlice {
	if len <= vs.Cap() {
		// capacity sufficient
		return VertexSlice{
			va: vs.va,
			i:  vs.i,
			j:  vs.i + len,
		}
	}

	// grow the capacity
	newCap := vs.Cap()
	if newCap < 1024 {
		newCap += newCap
	} else {
		newCap += newCap / 4
	}
	if newCap < len {
		newCap = len
	}
	newVs := VertexSlice{
		va: newVertexArray(vs.va.shader, newCap),
		i:  0,
		j:  len,
	}
	// preserve the original content
	newVs.Begin()
	newVs.Slice(0, vs.Len()).SetVertexData(vs.VertexData())
	newVs.End()
	return newVs
}

// Append adds supplied vertices to the end of the VertexSlice. If the capacity of the VertexSlice
// is not sufficient, a new, larger underlying vertex array will be allocated. The content of the
// original VertexSlice will be copied to the new underlying vertex array.
//
// The data is in the same format as with SetVertexData.
//
// The VertexSlice is appended 'in-place', contrary Go's builtin slices.
func (vs *VertexSlice) Append(data []float32) {
	vs.End() // vs must have been Begin-ed before calling this method
	*vs = vs.grow(vs.Len() + len(data)/vs.Stride())
	vs.Begin()
	vs.Slice(vs.Len()-len(data)/vs.Stride(), vs.Len()).SetVertexData(data)
}

// SetVertexData sets the contents of the VertexSlice.
//
// The data is a slice of float32's, where each vertex attribute occupies a certain number of
// elements. Namely, Float occupies 1, Vec2 occupies 2, Vec3 occupies 3 and Vec4 occupies 4. The
// attribues in the data slice must be in the same order as in the vertex format of this Vertex
// Slice.
//
// If the length of vertices does not match the length of the VertexSlice, this methdo panics.
func (vs *VertexSlice) SetVertexData(data []float32) {
	if len(data)/vs.Stride() != vs.Len() {
		fmt.Println(len(data)/vs.Stride(), vs.Len())
		panic("set vertex data: wrong length of vertices")
	}
	vs.va.setVertexData(vs.i, vs.j, data)
}

// VertexData returns the contents of the VertexSlice.
//
// The data is in the same format as with SetVertexData.
func (vs *VertexSlice) VertexData() []float32 {
	return vs.va.vertexData(vs.i, vs.j)
}

// Draw draws the content of the VertexSlice.
func (vs *VertexSlice) Draw() {
	vs.va.draw(vs.i, vs.j)
}

// Begin binds the underlying vertex array. Calling this method is necessary before using the VertexSlice.
func (vs *VertexSlice) Begin() {
	vs.va.begin()
}

// End unbinds the underlying vertex array. Call this method when you're done with VertexSlice.
func (vs *VertexSlice) End() {
	vs.va.end()
}

type vertexArray struct {
	vao, vbo binder
	cap      int
	format   AttrFormat
	stride   int
	offset   []int
	shader   *Shader
}

const vertexArrayMinCap = 4

func newVertexArray(shader *Shader, cap int) *vertexArray {
	if cap < vertexArrayMinCap {
		cap = vertexArrayMinCap
	}

	va := &vertexArray{
		vao: binder{
			restoreLoc: gl.VERTEX_ARRAY_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindVertexArray(obj)
			},
		},
		vbo: binder{
			restoreLoc: gl.ARRAY_BUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindBuffer(gl.ARRAY_BUFFER, obj)
			},
		},
		cap:    cap,
		format: shader.VertexFormat(),
		stride: shader.VertexFormat().Size(),
		offset: make([]int, len(shader.VertexFormat())),
		shader: shader,
	}

	offset := 0
	for i, attr := range va.format {
		switch attr.Type {
		case Float, Vec2, Vec3, Vec4:
		default:
			panic(errors.New("failed to create vertex array: invalid attribute type"))
		}
		va.offset[i] = offset
		offset += attr.Type.Size()
	}

	gl.GenVertexArrays(1, &va.vao.obj)

	va.vao.bind()

	gl.GenBuffers(1, &va.vbo.obj)
	defer va.vbo.bind().restore()

	emptyData := make([]byte, cap*va.stride)
	gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), gl.DYNAMIC_DRAW)

	for i, attr := range va.format {
		loc := gl.GetAttribLocation(shader.program.obj, gl.Str(attr.Name+"\x00"))

		var size int32
		switch attr.Type {
		case Float:
			size = 1
		case Vec2:
			size = 2
		case Vec3:
			size = 3
		case Vec4:
			size = 4
		}

		gl.VertexAttribPointer(
			uint32(loc),
			size,
			gl.FLOAT,
			false,
			int32(va.stride),
			gl.PtrOffset(va.offset[i]),
		)
		gl.EnableVertexAttribArray(uint32(loc))
	}

	va.vao.restore()

	runtime.SetFinalizer(va, (*vertexArray).delete)

	return va
}

func (va *vertexArray) delete() {
	mainthread.CallNonBlock(func() {
		gl.DeleteVertexArrays(1, &va.vao.obj)
		gl.DeleteBuffers(1, &va.vbo.obj)
	})
}

func (va *vertexArray) begin() {
	va.vao.bind()
	va.vbo.bind()
}

func (va *vertexArray) end() {
	va.vbo.restore()
	va.vao.restore()
}

func (va *vertexArray) draw(i, j int) {
	gl.DrawArrays(gl.TRIANGLES, int32(i), int32(i+j))
}

func (va *vertexArray) setVertexData(i, j int, data []float32) {
	if j-i == 0 {
		// avoid setting 0 bytes of buffer data
		return
	}
	gl.BufferSubData(gl.ARRAY_BUFFER, i*va.stride, len(data)*4, gl.Ptr(data))
}

func (va *vertexArray) vertexData(i, j int) []float32 {
	if j-i == 0 {
		// avoid getting 0 bytes of buffer data
		return nil
	}
	data := make([]float32, (j-i)*va.stride/4)
	gl.GetBufferSubData(gl.ARRAY_BUFFER, i*va.stride, len(data)*4, gl.Ptr(data))
	return data
}
