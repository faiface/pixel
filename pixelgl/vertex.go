package pixelgl

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/pkg/errors"
)

// VertexFormat defines a data format in a vertex buffer.
//
// Example:
//
//   VertexFormat{{Position, Vec2}, {Color, Vec4}, {TexCoord, Vec2}, {Visible, Bool}}
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
	format VertexFormat
	stride int
	count  int
	attrs  map[Attr]int
	vao    uint32
	vbo    uint32
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
		case Bool, Int, Float, Vec2, Vec3, Vec4:
		default:
			return nil, errors.New("failed to create vertex array: invalid vertex format: invalid attribute type")
		}
		if _, ok := va.attrs[attr]; ok {
			return nil, errors.New("failed to create vertex array: invalid vertex format: duplicate vertex attribute")
		}
		va.attrs[attr] = offset
		offset += attr.Type.Size()
	}

	var err error
	parent.Do(func(ctx Context) {
		err = DoGLErr(func() {
			gl.GenVertexArrays(1, &va.vao)
			gl.BindVertexArray(va.vao)

			gl.GenBuffers(1, &va.vbo)
			gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

			emptyData := make([]byte, count*va.stride)
			gl.BufferData(gl.ARRAY_BUFFER, len(emptyData), gl.Ptr(emptyData), uint32(usage))

			offset := 0
			for i, attr := range format {
				//XXX: ugly but OpenGL is so inconsistent

				var size int32
				switch attr.Type {
				case Bool, Int, Float:
					size = 1
				case Vec2:
					size = 2
				case Vec3:
					size = 3
				case Vec4:
					size = 4
				}

				var xtype uint32
				switch attr.Type {
				case Bool:
					xtype = gl.BOOL
				case Int:
					xtype = gl.INT
				case Float, Vec2, Vec3, Vec4:
					xtype = gl.DOUBLE
				}

				gl.VertexAttribPointer(
					uint32(i),
					size,
					xtype,
					false,
					int32(va.stride),
					gl.PtrOffset(offset),
				)
				gl.EnableVertexAttribArray(uint32(i))
				offset += attr.Type.Size()
			}

			gl.BindBuffer(gl.ARRAY_BUFFER, 0)
			gl.BindVertexArray(0)
		})
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a vertex array")
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

// SetVertexAttribute sets the value of the specified vertex attribute of the specified vertex.
// Argument data must be a slice/array containing the new attribute data.
//
// This function returns false if the specified attribute does not exist. Note that the function panics
// if the vertex is out of range.
func (va *VertexArray) SetVertexAttribute(vertex int, attr Attr, data interface{}) (ok bool) {
	if vertex < 0 || vertex >= va.count {
		panic("set vertex attribute error: invalid vertex index")
	}
	if _, ok := va.attrs[attr]; !ok {
		return false
	}
	DoNoBlock(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)

		offset := va.stride*vertex + va.attrs[attr]
		gl.BufferSubData(gl.ARRAY_BUFFER, offset, attr.Type.Size(), gl.Ptr(data))

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		if err := getLastGLErr(); err != nil {
			panic(errors.Wrap(err, "set attribute vertex error"))
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
