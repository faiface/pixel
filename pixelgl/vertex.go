package pixelgl

import (
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// VertexFormat defines a data format in a vertex buffer.
//
// Example:
//
//   VertexFormat{"position": {Position, Vec2}, "colr": {Color, Vec4}, "texCoord": {TexCoord, Vec2}}
//
// Note: vertex array currently doesn't support matrices in vertex format.
type VertexFormat []Attr

// Size calculates the total size of a single vertex in this vertex format (sum of the sizes of all vertex attributes).
func (vf VertexFormat) Size() int {
	total := 0
	for _, attr := range vf {
		total += attr.Type.Size()
	}
	return total
}

// VertexUsage specifies how often the vertex array data will be updated.
type VertexUsage int

const (
	// StaticUsage means the data never or rarely gets updated.
	StaticUsage VertexUsage = gl.STATIC_DRAW

	// DynamicUsage means the data gets updated often.
	DynamicUsage VertexUsage = gl.DYNAMIC_DRAW

	// StreamUsage means the data gets updated every frame.
	StreamUsage VertexUsage = gl.STREAM_DRAW
)

// VertexArray is an OpenGL vertex array object that also holds it's own vertex buffer object.
// From the user's points of view, VertexArray is an array of vertices that can be drawn.
type VertexArray struct {
	enabled             bool
	parent              Doer
	vao, vbo, ebo       uint32
	vertexNum, indexNum int
	format              VertexFormat
	usage               VertexUsage
	stride              int
	attrs               map[Attr]int
}

// NewVertexArray creates a new empty vertex array and wraps another Doer around it.
//
// You cannot specify vertex attributes in this constructor, only their count. Use SetVertexAttribute* methods to
// set the vertex attributes. Use indices to specify how you want to combine vertices into triangles.
func NewVertexArray(parent Doer, format VertexFormat, usage VertexUsage, vertexNum int, indices []int) (*VertexArray, error) {
	va := &VertexArray{
		parent:    parent,
		format:    format,
		usage:     usage,
		vertexNum: vertexNum,
		stride:    format.Size(),
		attrs:     make(map[Attr]int),
	}

	offset := 0
	for _, attr := range format {
		switch attr.Type {
		case Float, Vec2, Vec3, Vec4:
		default:
			return nil, errors.New("failed to create vertex array: invalid vertex format: invalid attribute type")
		}
		if _, ok := va.attrs[attr]; ok {
			return nil, errors.New("failed to create vertex array: invalid vertex format: duplicate vertex attribute")
		}
		va.attrs[attr] = offset
		offset += attr.Type.Size()
	}

	parent.Do(func(ctx Context) {
		Do(func() {
			gl.GenVertexArrays(1, &va.vao)
			gl.BindVertexArray(va.vao)

			gl.GenBuffers(1, &va.vbo)
			gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

			emptyData := make([]byte, vertexNum*va.stride)
			gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), uint32(usage))

			gl.GenBuffers(1, &va.ebo)

			offset := 0
			for i, attr := range format {
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
					uint32(i),
					size,
					gl.FLOAT,
					false,
					int32(va.stride),
					gl.PtrOffset(offset),
				)
				gl.EnableVertexAttribArray(uint32(i))
				offset += attr.Type.Size()
			}

			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, va.ebo) // need to bind EBO, so that VAO registers it

			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)

			gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
		})
	})

	va.SetIndices(indices)

	return va, nil
}

// Delete deletes a vertex array and it's associated vertex buffer. Don't use a vertex array after deletion.
func (va *VertexArray) Delete() {
	va.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.DeleteVertexArrays(1, &va.vao)
			gl.DeleteBuffers(1, &va.vbo)
		})
	})
}

// ID returns an OpenGL identifier of a vertex array.
func (va *VertexArray) ID() uint32 {
	return va.vao
}

// Count returns the number of vertices in a vertex array.
func (va *VertexArray) Count() int {
	return va.vertexNum
}

// VertexFormat returns the format of the vertices inside a vertex array.
//
// Do not change this format!
func (va *VertexArray) VertexFormat() VertexFormat {
	return va.format
}

// VertexUsage returns the usage of the verteices inside a vertex array.
func (va *VertexArray) VertexUsage() VertexUsage {
	return va.usage
}

// Draw draws a vertex array.
func (va *VertexArray) Draw() {
	va.Do(func(Context) {})
}

// SetIndices sets the indices of triangles to be drawn. Triangles will be formed from the vertices of the array
// as defined by these indices. The first drawn triangle is specified by the first three indices, the second by
// the fourth through sixth and so on.
func (va *VertexArray) SetIndices(indices []int) {
	if len(indices)%3 != 0 {
		panic("vertex array set indices: number of indices not divisible by 3")
	}
	indices32 := make([]uint32, len(indices))
	for i := range indices32 {
		indices32[i] = uint32(indices[i])
	}
	va.indexNum = len(indices32)
	DoNoBlock(func() {
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, va.ebo)
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices32), gl.Ptr(indices32), uint32(va.usage))
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0)
	})
}

// SetVertex sets the value of all attributes of a vertex.
// Argument data must be a slice/array containing the new vertex data.
func (va *VertexArray) SetVertex(vertex int, data interface{}) {
	if vertex < 0 || vertex >= va.vertexNum {
		panic("set vertex error: invalid vertex index")
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride * vertex
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, va.format.Size(), gl.Ptr(data))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
}

func (va *VertexArray) checkVertex(vertex int) {
	if vertex < 0 || vertex >= va.vertexNum {
		panic("invalid vertex index")
	}
}

// SetVertexAttributeFloat sets the value of a specified vertex attribute Attr{Purpose: purpose, Type: Float} of type Float
// of the specified vertex.
//
// This function returns false if the specified vertex attribute does not exist. Note that the function panics if
// the vertex if out of range.
func (va *VertexArray) SetVertexAttributeFloat(vertex int, purpose AttrPurpose, value float32) (ok bool) {
	va.checkVertex(vertex)
	attr := Attr{Purpose: purpose, Type: Float}
	if _, ok := va.attrs[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
	return true
}

// SetVertexAttributeVec2 sets the value of a specified vertex attribute Attr{Purpose: purpose, Type: Vec2} of type Vec2
// of the specified vertex.
//
// This function returns false if the specified vertex attribute does not exist. Note that the function panics if
// the vertex if out of range.
func (va *VertexArray) SetVertexAttributeVec2(vertex int, purpose AttrPurpose, value mgl32.Vec2) (ok bool) {
	va.checkVertex(vertex)
	attr := Attr{Purpose: purpose, Type: Vec2}
	if _, ok := va.attrs[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
	return true
}

// SetVertexAttributeVec3 sets the value of a specified vertex attribute Attr{Purpose: purpose, Type: Vec3} of type Vec3
// of the specified vertex.
//
// This function returns false if the specified vertex attribute does not exist. Note that the function panics if
// the vertex if out of range.
func (va *VertexArray) SetVertexAttributeVec3(vertex int, purpose AttrPurpose, value mgl32.Vec3) (ok bool) {
	va.checkVertex(vertex)
	attr := Attr{Purpose: purpose, Type: Vec3}
	if _, ok := va.attrs[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
	return true
}

// SetVertexAttributeVec4 sets the value of a specified vertex attribute Attr{Purpose: purpose, Type: Vec4} of type Vec4
// of the specified vertex.
//
// This function returns false if the specified vertex attribute does not exist. Note that the function panics if
// the vertex if out of range.
func (va *VertexArray) SetVertexAttributeVec4(vertex int, purpose AttrPurpose, value mgl32.Vec4) (ok bool) {
	va.checkVertex(vertex)
	attr := Attr{Purpose: purpose, Type: Vec4}
	if _, ok := va.attrs[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
	return true
}

// Do binds a vertex arrray and it's associated vertex buffer, executes sub, and unbinds the vertex array and it's vertex buffer.
func (va *VertexArray) Do(sub func(Context)) {
	va.parent.Do(func(ctx Context) {
		if va.enabled {
			sub(ctx)
			return
		}
		DoNoBlock(func() {
			gl.BindVertexArray(va.vao)
		})
		va.enabled = true
		sub(ctx)
		va.enabled = false
		DoNoBlock(func() {
			gl.DrawElements(gl.TRIANGLES, int32(va.indexNum), gl.UNSIGNED_INT, gl.PtrOffset(0))
			gl.BindVertexArray(0)
		})
	})
}
