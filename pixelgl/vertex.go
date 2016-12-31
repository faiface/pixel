package pixelgl

import (
	"runtime"
	"unsafe"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// VertexArray is an OpenGL vertex array object that also holds it's own vertex buffer object.
// From the user's points of view, VertexArray is an array of vertices that can be drawn.
type VertexArray struct {
	vao, vbo    binder
	numVertices int
	format      AttrFormat
	stride      int
	offset      map[string]int
}

// NewVertexArray creates a new empty vertex array.
//
// You cannot specify vertex attributes in this constructor, only their count. Use
// SetVertexAttribute* methods to set the vertex attributes.
func NewVertexArray(shader *Shader, numVertices int) (*VertexArray, error) {
	va := &VertexArray{
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
		numVertices: numVertices,
		format:      shader.VertexFormat(),
		stride:      shader.VertexFormat().Size(),
		offset:      make(map[string]int),
	}

	offset := 0
	for name, typ := range va.format {
		switch typ {
		case Float, Vec2, Vec3, Vec4:
		default:
			return nil, errors.New("failed to create vertex array: invalid attribute type")
		}
		va.offset[name] = offset
		offset += typ.Size()
	}

	gl.GenVertexArrays(1, &va.vao.obj)

	va.vao.bind()

	gl.GenBuffers(1, &va.vbo.obj)
	defer va.vbo.bind().restore()

	emptyData := make([]byte, numVertices*va.stride)
	gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), gl.DYNAMIC_DRAW)

	for name, typ := range va.format {
		loc := gl.GetAttribLocation(shader.ID(), gl.Str(name+"\x00"))

		var size int32
		switch typ {
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
			gl.PtrOffset(va.offset[name]),
		)
		gl.EnableVertexAttribArray(uint32(loc))
	}

	va.vao.restore()

	runtime.SetFinalizer(va, (*VertexArray).delete)

	return va, nil
}

func (va *VertexArray) delete() {
	DoNoBlock(func() {
		gl.DeleteVertexArrays(1, &va.vao.obj)
		gl.DeleteBuffers(1, &va.vbo.obj)
	})
}

// ID returns an OpenGL identifier of a vertex array.
func (va *VertexArray) ID() uint32 {
	return va.vao.obj
}

// NumVertices returns the number of vertices in a vertex array.
func (va *VertexArray) NumVertices() int {
	return va.numVertices
}

// VertexFormat returns the format of the vertices inside a vertex array.
//
// Do not change this format!
func (va *VertexArray) VertexFormat() AttrFormat {
	return va.format
}

// Draw draws a vertex array.
//
// The vertex array must be bound before calling this method.
func (va *VertexArray) Draw() {
	gl.DrawArrays(gl.TRIANGLES, 0, int32(va.numVertices))
}

// SetVertexAttr sets the value of the specified vertex attribute of the specified vertex.
//
// If the vertex attribute does not exist, this method returns false. If the vertex is out of
// range, this method panics.
//
// Supplied value must correspond to the type of the attribute. Correct types are these
// (righ-hand is the type of the value):
//   Attr{Type: Float}: float32
//   Attr{Type: Vec2}:  mgl32.Vec2
//   Attr{Type: Vec3}:  mgl32.Vec3
//   Attr{Type: Vec4}:  mgl32.Vec4
// No other types are supported.
//
// The vertex array must be bound before calling this method.
func (va *VertexArray) SetVertexAttr(vertex int, attr Attr, value interface{}) (ok bool) {
	if vertex < 0 || vertex >= va.numVertices {
		panic("set vertex attr: invalid vertex index")
	}

	if !va.format.Contains(attr) {
		return false
	}

	offset := va.stride*vertex + va.offset[attr.Name]

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
	default:
		panic("set vertex attr: invalid attribute type")
	}

	return true
}

// VertexAttr returns the current value of the specified vertex attribute of the specified vertex.
//
// If the vertex attribute does not exist, this method returns nil and false. If the vertex is
// out of range, this method panics.
//
// The type of the returned value follows the same rules as with SetVertexAttr.
//
// The vertex array must be bound before calling this method.
func (va *VertexArray) VertexAttr(vertex int, attr Attr) (value interface{}, ok bool) {
	if vertex < 0 || vertex >= va.numVertices {
		panic("vertex attr: invalid vertex index")
	}

	if !va.format.Contains(attr) {
		return nil, false
	}

	offset := va.stride*vertex + va.offset[attr.Name]

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
	default:
		panic("set vertex attr: invalid attribute type")
	}

	return value, true
}

// SetVertex sets values of the attributes specified in the supplied map. All other attributes
// will be set to zero.
//
// Not existing attributes are silently skipped.
//
// The vertex array must be bound before calling this method.
func (va *VertexArray) SetVertex(vertex int, values map[Attr]interface{}) {
	if vertex < 0 || vertex >= va.numVertices {
		panic("set vertex: invalid vertex index")
	}

	data := make([]float32, va.format.Size()/4)

	for attr, value := range values {
		if !va.format.Contains(attr) {
			continue
		}

		offset := va.offset[attr.Name]

		switch attr.Type {
		case Float:
			data[offset/4] = value.(float32)
		case Vec2:
			value := value.(mgl32.Vec2)
			copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
		case Vec3:
			value := value.(mgl32.Vec3)
			copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
		case Vec4:
			value := value.(mgl32.Vec4)
			copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
		default:
			panic("set vertex: invalid attribute type")
		}
	}

	offset := va.stride * vertex
	gl.BufferSubData(gl.ARRAY_BUFFER, offset, len(data)*4, gl.Ptr(data))
}

// Vertex returns values of all vertex attributes of the specified vertex in a map.
//
// The vertex array must be bound before calling this method.
func (va *VertexArray) Vertex(vertex int) (values map[Attr]interface{}) {
	if vertex < 0 || vertex >= va.numVertices {
		panic("set vertex: invalid vertex index")
	}

	data := make([]float32, va.format.Size()/4)

	offset := va.stride * vertex
	gl.GetBufferSubData(gl.ARRAY_BUFFER, offset, len(data)*4, gl.Ptr(data))

	values = make(map[Attr]interface{})

	for name, typ := range va.format {
		attr := Attr{name, typ}
		offset := va.offset[attr.Name]

		switch attr.Type {
		case Float:
			values[attr] = data[offset/4]
		case Vec2:
			var value mgl32.Vec2
			copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
			values[attr] = value
		case Vec3:
			var value mgl32.Vec3
			copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
			values[attr] = value
		case Vec4:
			var value mgl32.Vec4
			copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
			values[attr] = value
		}
	}

	return values
}

// SetVertices sets values of vertex attributes of all vertices as specified in the supplied
// slice of maps. If the length of vertices does not match the number of vertices in the vertex
// array, this method panics.
//
// Not existing attributes are silently skipped.
//
// The vertex array must be bound before calling this metod.
func (va *VertexArray) SetVertices(vertices []map[Attr]interface{}) {
	if len(vertices) != va.numVertices {
		panic("set vertex array: wrong number of supplied vertices")
	}

	data := make([]float32, va.numVertices*va.format.Size()/4)

	for vertex := range vertices {
		for attr, value := range vertices[vertex] {
			if !va.format.Contains(attr) {
				continue
			}

			offset := va.stride*vertex + va.offset[attr.Name]

			switch attr.Type {
			case Float:
				data[offset/4] = value.(float32)
			case Vec2:
				value := value.(mgl32.Vec2)
				copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
			case Vec3:
				value := value.(mgl32.Vec3)
				copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
			case Vec4:
				value := value.(mgl32.Vec4)
				copy(data[offset/4:offset/4+attr.Type.Size()/4], value[:])
			default:
				panic("set vertex: invalid attribute type")
			}
		}
	}

	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))
}

// Vertices returns values of vertex attributes of all vertices in a vertex array in a slice
// of maps.
//
// The vertex array must be bound before calling this metod.
func (va *VertexArray) Vertices() (vertices []map[Attr]interface{}) {
	data := make([]float32, va.numVertices*va.format.Size()/4)

	gl.GetBufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))

	vertices = make([]map[Attr]interface{}, va.numVertices)

	for vertex := range vertices {
		values := make(map[Attr]interface{})

		for name, typ := range va.format {
			attr := Attr{name, typ}
			offset := va.stride*vertex + va.offset[attr.Name]

			switch attr.Type {
			case Float:
				values[attr] = data[offset/4]
			case Vec2:
				var value mgl32.Vec2
				copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
				values[attr] = value
			case Vec3:
				var value mgl32.Vec3
				copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
				values[attr] = value
			case Vec4:
				var value mgl32.Vec4
				copy(value[:], data[offset/4:offset/4+attr.Type.Size()/4])
				values[attr] = value
			}
		}

		vertices[vertex] = values
	}

	return vertices
}

// Begin binds a vertex array. This is neccessary before using the vertex array.
func (va *VertexArray) Begin() {
	va.vao.bind()
	va.vbo.bind()
}

// End unbinds a vertex array and restores the previous one.
func (va *VertexArray) End() {
	va.vbo.restore()
	va.vao.restore()
}
