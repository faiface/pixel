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

// VertexDrawMode specifies how should the vertices be drawn.
type VertexDrawMode int

const (
	// PointsDrawMode just draws individual points
	PointsDrawMode VertexDrawMode = gl.POINTS

	// LinesDrawMode takes pairs of vertices and draws a line from each pair
	LinesDrawMode VertexDrawMode = gl.LINES

	// LineStripDrawMode takes each two subsequent vertices and draws a line from each two
	LineStripDrawMode VertexDrawMode = gl.LINE_STRIP

	// LineLoopDrawMode is same as line strip, but also draws a line between the first and the last vertex
	LineLoopDrawMode VertexDrawMode = gl.LINE_LOOP

	// TrianglesDrawMode takes triples of vertices and draws a triangle from each triple
	TrianglesDrawMode VertexDrawMode = gl.TRIANGLES

	// TriangleStripDrawMode takes each three subsequent vertices and draws a triangle from each three
	TriangleStripDrawMode VertexDrawMode = gl.TRIANGLE_STRIP

	// TriangleFanDrawMode draws triangles from the first vertex and each two subsequent: {0, 1, 2, 3} -> {0, 1, 2}, {0, 2, 3}.
	TriangleFanDrawMode VertexDrawMode = gl.TRIANGLE_FAN
)

// VertexArray is an OpenGL vertex array object that also holds it's own vertex buffer object.
// From the user's points of view, VertexArray is an array of vertices that can be drawn.
type VertexArray struct {
	parent Doer
	vao    uint32
	vbo    uint32
	format VertexFormat
	stride int
	count  int
	attrs  map[Attr]int
	mode   VertexDrawMode
}

// NewVertexArray creates a new empty vertex array and wraps another Doer around it.
func NewVertexArray(parent Doer, format VertexFormat, mode VertexDrawMode, usage VertexUsage, count int) (*VertexArray, error) {
	va := &VertexArray{
		parent: parent,
		format: format,
		count:  count,
		stride: format.Size(),
		attrs:  make(map[Attr]int),
		mode:   mode,
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

	var err, glerr error
	parent.Do(func(ctx Context) {
		err, glerr = DoErrGLErr(func() error {
			gl.GenVertexArrays(1, &va.vao)
			gl.BindVertexArray(va.vao)

			gl.GenBuffers(1, &va.vbo)
			gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

			emptyData := make([]byte, count*va.stride)
			gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), uint32(usage))

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

			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)

			return nil
		})
	})
	if err != nil && glerr != nil {
		return nil, errors.Wrap(errors.Wrap(glerr, err.Error()), "failed to create vertex array")
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to create vertex array")
	}
	if glerr != nil {
		return nil, errors.Wrap(glerr, "failed to create vertex array")
	}

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
	return va.count
}

// VertexFormat returns the format of the vertices inside a vertex array.
//
// Do not change this format!
func (va *VertexArray) VertexFormat() VertexFormat {
	return va.format
}

// SetDrawMode sets the draw mode of a vertex array. Subsequent calls to Draw will use this draw mode.
func (va *VertexArray) SetDrawMode(mode VertexDrawMode) {
	DoNoBlock(func() {
		va.mode = mode
	})
}

// DrawMode returns the most recently set draw mode of a vertex array.
func (va *VertexArray) DrawMode() VertexDrawMode {
	mode := DoVal(func() interface{} {
		return va.mode
	})
	return mode.(VertexDrawMode)
}

// Draw draws a vertex array.
func (va *VertexArray) Draw() {
	va.Do(func(Context) {})
}

// SetVertex sets the value of all attributes of a vertex.
// Argument data must be a slice/array containing the new vertex data.
func (va *VertexArray) SetVertex(vertex int, data interface{}) {
	if vertex < 0 || vertex >= va.count {
		panic("set vertex error: invalid vertex index")
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride * vertex
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, va.format.Size(), gl.Ptr(data))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set vertex error"))
		}
	})
}

func (va *VertexArray) checkVertex(vertex int) {
	if vertex < 0 || vertex >= va.count {
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

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set attribute vertex"))
		}
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

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set attribute vertex"))
		}
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

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set attribute vertex"))
		}
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

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set attribute vertex"))
		}
	})
	return true
}

// Do binds a vertex arrray and it's associated vertex buffer, executes sub, and unbinds the vertex array and it's vertex buffer.
func (va *VertexArray) Do(sub func(Context)) {
	va.parent.Do(func(ctx Context) {
		DoNoBlock(func() {
			gl.BindVertexArray(va.vao)
			gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
		})
		sub(ctx)
		DoNoBlock(func() {
			gl.DrawArrays(uint32(va.mode), 0, int32(va.count))
			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)
		})
	})
}
