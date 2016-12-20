package pixelgl

import (
	"unsafe"

	"runtime"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// VertexArray is an OpenGL vertex array object that also holds it's own vertex buffer object.
// From the user's points of view, VertexArray is an array of vertices that can be drawn.
type VertexArray struct {
	parent              Doer
	vao, vbo, ebo       binder
	vertexNum, indexNum int
	format              AttrFormat
	stride              int
	offset              map[string]int
}

// NewVertexArray creates a new empty vertex array and wraps another Doer around it.
//
// You cannot specify vertex attributes in this constructor, only their count. Use SetVertexAttribute* methods to
// set the vertex attributes. Use indices to specify how you want to combine vertices into triangles.
func NewVertexArray(parent Doer, format AttrFormat, vertexNum int, indices []int) (*VertexArray, error) {
	va := &VertexArray{
		parent: parent,
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
		ebo: binder{
			restoreLoc: gl.ELEMENT_ARRAY_BUFFER_BINDING,
			bindFunc: func(obj uint32) {
				gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, obj)
			},
		},
		vertexNum: vertexNum,
		format:    format,
		stride:    format.Size(),
		offset:    make(map[string]int),
	}

	offset := 0
	for name, typ := range format {
		switch typ {
		case Float, Vec2, Vec3, Vec4:
		default:
			return nil, errors.New("failed to create vertex array: invalid vertex format: invalid attribute type")
		}
		va.offset[name] = offset
		offset += typ.Size()
	}

	parent.Do(func(ctx Context) {
		Do(func() {
			gl.GenVertexArrays(1, &va.vao.obj)
			va.vao.bind()

			gl.GenBuffers(1, &va.vbo.obj)
			defer va.vbo.bind().restore()

			emptyData := make([]byte, vertexNum*va.stride)
			gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), gl.DYNAMIC_DRAW)

			gl.GenBuffers(1, &va.ebo.obj)
			defer va.ebo.bind().restore()

			for name, typ := range format {
				loc := gl.GetAttribLocation(ctx.Shader().ID(), gl.Str(name+"\x00"))

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
		})
	})

	va.SetIndices(indices)

	runtime.SetFinalizer(va, (*VertexArray).delete)

	return va, nil
}

func (va *VertexArray) delete() {
	DoNoBlock(func() {
		gl.DeleteVertexArrays(1, &va.vao.obj)
		gl.DeleteBuffers(1, &va.vbo.obj)
		gl.DeleteBuffers(1, &va.ebo.obj)
	})
}

// ID returns an OpenGL identifier of a vertex array.
func (va *VertexArray) ID() uint32 {
	return va.vao.obj
}

// VertexNum returns the number of vertices in a vertex array.
func (va *VertexArray) VertexNum() int {
	return va.vertexNum
}

// VertexFormat returns the format of the vertices inside a vertex array.
//
// Do not change this format!
func (va *VertexArray) VertexFormat() AttrFormat {
	return va.format
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
		va.ebo.bind()
		gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices32), gl.Ptr(indices32), gl.DYNAMIC_DRAW)
		va.ebo.restore()
	})
}

// Indices returns the current indices of triangles to be drawn.
func (va *VertexArray) Indices() []int {
	indices32 := make([]uint32, va.indexNum)
	Do(func() {
		va.ebo.bind()
		gl.GetBufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, 4*len(indices32), gl.Ptr(indices32))
		va.ebo.restore()
	})
	indices := make([]int, len(indices32))
	for i := range indices {
		indices[i] = int(indices32[i])
	}
	return indices
}

// SetVertexAttr sets the value of the specified vertex attribute of the specified vertex.
//
// If the vertex attribute does not exist, this method returns false. If the vertex is out of range,
// this method panics.
//
// Supplied value must correspond to the type of the attribute. Correct types are these (righ-hand is the type of the value):
//   Attr{Type: Float}: float32
//   Attr{Type: Vec2}:  mgl32.Vec2
//   Attr{Type: Vec3}:  mgl32.Vec3
//   Attr{Type: Vec4}:  mgl32.Vec4
// No other types are supported.
func (va *VertexArray) SetVertexAttr(vertex int, attr Attr, value interface{}) (ok bool) {
	if vertex < 0 || vertex >= va.vertexNum {
		panic("set vertex attr: invalid vertex index")
	}

	if !va.format.Contains(attr) {
		return false
	}

	DoNoBlock(func() {
		va.vbo.bind()

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

		va.vbo.restore()
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

	if !va.format.Contains(attr) {
		return nil, false
	}

	Do(func() {
		va.vbo.bind()

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

		va.vbo.restore()
	})

	return value, true
}

// SetVertex sets values of the attributes specified in the supplied map. All other attributes will be set to zero.
//
// Not existing attributes are silently skipped.
func (va *VertexArray) SetVertex(vertex int, values map[Attr]interface{}) {
	if vertex < 0 || vertex >= va.vertexNum {
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
			value := value.(float32)
			copy(data[offset/4:offset/4+attr.Type.Size()/4], []float32{value})
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

	DoNoBlock(func() {
		va.vbo.bind()

		offset := va.stride * vertex
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, len(data)*4, gl.Ptr(data))

		va.vbo.restore()
	})
}

// Set sets values of vertex attributes of all vertices as specified in the supplied slice of maps. If the length of vertices
// does not match the number of vertices in the vertex array, this method panics.
//
// Not existing attributes are silently skipped.
func (va *VertexArray) Set(vertices []map[Attr]interface{}) {
	if len(vertices) != va.vertexNum {
		panic("set vertex array: wrong number of supplied vertices")
	}

	data := make([]float32, va.vertexNum*va.format.Size()/4)

	for vertex := range vertices {
		for attr, value := range vertices[vertex] {
			if !va.format.Contains(attr) {
				continue
			}

			offset := va.stride*vertex + va.offset[attr.Name]

			switch attr.Type {
			case Float:
				value := value.(float32)
				copy(data[offset/4:offset/4+attr.Type.Size()/4], []float32{value})
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

	DoNoBlock(func() {
		va.vbo.bind()

		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(data)*4, gl.Ptr(data))

		va.vbo.restore()
	})
}

// Do binds a vertex arrray and it's associated vertex buffer, executes sub, and unbinds the vertex array and it's vertex buffer.
func (va *VertexArray) Do(sub func(Context)) {
	va.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			va.vao.bind()
		})
		sub(ctx)
		DoNoBlock(func() {
			gl.DrawElements(gl.TRIANGLES, int32(va.indexNum), gl.UNSIGNED_INT, gl.PtrOffset(0))
			va.vao.restore()
		})
	})
}
