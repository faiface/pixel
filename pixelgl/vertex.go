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

// VertexNum returns the number of vertices in a vertex array.
func (va *VertexArray) VertexNum() int {
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

// SetVertexAttr sets the value of the specified vertex attribute of the specified vertex.
//
// If the vertex attribute does not exist, this method returns false. If the vertex is out of range,
// this method panics.
//
// Supplied value must correspond to the type of the attribute. Correct types are these (righ-hand is the type of value):
//   Attr{Type: Float}: float32
//   Attr{Type: Vec2}:  mgl32.Vec2
//   Attr{Type: Vec3}:  mgl32.Vec3
//   Attr{Type: Vec4}:  mgl32.Vec4
// No other types are supported.
func (va *VertexArray) SetVertexAttr(vertex int, attr Attr, value interface{}) (ok bool) {
	if vertex < 0 || vertex >= va.vertexNum {
		panic("set vertex attr: invalid vertex index")
	}

	if _, ok := va.attrs[attr]; !ok {
		return false
	}

	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]

		switch attr.Type {
		case Float:
			value := value.(float32)
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))
		case Vec2:
			value := value.(mgl32.Vec2)
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))
		case Vec3:
			value := value.(mgl32.Vec3)
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))
		case Vec4:
			value := value.(mgl32.Vec4)
			gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&value))
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})

	return true
}

// VertexAttr returns the current value of the specified vertex attribute of the specified vertex.
//
// If the vertex attribute does not exist, this method returns nil and false. If the vertex is out of range,
// this method panics.
//
// The type of the returned value follows the same rules as with SetVertexAttr.
func (va *VertexArray) VertexAttr(vertex int, attr Attr) (value interface{}, ok bool) {
	if vertex < 0 || vertex >= va.vertexNum {
		panic("vertex attr: invalid vertex index")
	}

	if _, ok := va.attrs[attr]; !ok {
		return nil, false
	}

	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]

		switch attr.Type {
		case Float:
			var data float32
			gl.GetBufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&data))
			value = data
		case Vec2:
			var data mgl32.Vec2
			gl.GetBufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&data))
			value = data
		case Vec3:
			var data mgl32.Vec3
			gl.GetBufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&data))
			value = data
		case Vec4:
			var data mgl32.Vec4
			gl.GetBufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), unsafe.Pointer(&data))
			value = data
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})

	return value, true
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
