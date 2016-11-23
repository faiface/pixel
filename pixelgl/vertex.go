package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

// VertexFormat defines a data format in a vertex buffer.
//
// Example:
//
//   vf := VertexFormat{{Position, 2}, {Color, 4}, {TexCoord, 2}}
type VertexFormat []VertexAttribute

// Size returns the total size of all vertex attributes in a vertex format.
func (vf VertexFormat) Size() int {
	size := 0
	for _, va := range vf {
		size += va.Size
	}
	return size
}

// VertexAttribute specifies a single attribute in a vertex buffer.
// All vertex attributes are composed of float64s.
//
// A vertex attribute has a Purpose (such as Position, Color, etc.) and Size. Size specifies
// the number of float64s the vertex attribute is composed of.
type VertexAttribute struct {
	Purpose VertexAttributePurpose
	Size    int
}

// VertexAttributePurpose clarifies the purpose of a vertex attribute. This can be a color, position, texture
// coordinates or anything else.
//
// VertexAttributePurpose may be used to correctly assign data to a vertex buffer.
type VertexAttributePurpose int

// Position, Color and TexCoord are the standard vertex attributes.
//
// Feel free to define more vertex attribute purposes (e.g. in an effects library).
const (
	Position VertexAttributePurpose = iota
	Color
	TexCoord
	NumStandardVertexAttrib
)

// VertexUsage specifies how often the vertex array data will be updated.
type VertexUsage int

// Possible VertexUsage values are:
//
// 1. StaticUsage - never or rarely updated
//
// 2. DynamicUsage - often updated
//
// 3. StreamUsage - updated every frame
const (
	StaticUsage  VertexUsage = gl.STATIC_DRAW
	DynamicUsage             = gl.DYNAMIC_DRAW
	StreamUsage              = gl.STREAM_DRAW
)

// VertexDrawMode specifies how should the vertices be drawn.
type VertexDrawMode int

// Possible VertexDrawMode values are:
//
// 1. PointsDrawMode - just draw individual PointsDrawMode
//
// 2. LinesDrawMode - take pairs of vertices and draw a line from each pair
//
// 3. LineStripDrawMode - take each two subsequent vertices and draw a line from each two
//
// 4. LineLoopDrawMode - same as line strip, but also draw a line between the first and the last vertex
//
// 5. TrianglesDrawMode - take triples of vertices and draw a triangle from each triple
//
// 6. TriangleStripDrawMode - take each three subsequent vertices and draw a triangle from each three
//
// 7. TriangleFanDrawMode - take each two subsequent vertices excluding the first vertex and draw a triangle from the first vertex and the two
const (
	PointsDrawMode        VertexDrawMode = gl.POINTS
	LinesDrawMode                        = gl.LINES
	LineStripDrawMode                    = gl.LINE_STRIP
	LineLoopDrawMode                     = gl.LINE_LOOP
	TrianglesDrawMode                    = gl.TRIANGLES
	TriangleStripDrawMode                = gl.TRIANGLE_STRIP
	TriangleFanDrawMode                  = gl.TRIANGLE_FAN
)

// VertexArray is an OpenGL vertex array object that also holds it's own vertex buffer object.
// From the user's points of view, VertexArray is an array of vertices that can be drawn.
type VertexArray struct {
	parent BeginEnder
	format VertexFormat
	vao    uint32
	vbo    uint32
	mode   VertexDrawMode
	count  int
}

// NewVertexArray creates a new vertex array and wrap another BeginEnder around it.
func NewVertexArray(parent BeginEnder, format VertexFormat, mode VertexDrawMode, usage VertexUsage, data []float64) *VertexArray {
	va := &VertexArray{
		parent: parent,
		format: format,
		mode:   mode,
	}
	Do(func() {
		gl.GenVertexArrays(1, &va.vao)
		gl.BindVertexArray(va.vao)

		gl.GenBuffers(1, &va.vbo)
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
		gl.BufferData(gl.ARRAY_BUFFER, 8*len(data), gl.Ptr(data), uint32(usage))

		stride := format.Size()
		va.count = len(data) / stride

		offset := 0
		for i, attr := range format {
			gl.VertexAttribPointer(
				uint32(i),
				int32(attr.Size),
				gl.DOUBLE,
				false,
				int32(stride),
				gl.PtrOffset(8*offset),
			)
			gl.EnableVertexAttribArray(uint32(i))
			offset += attr.Size
		}

		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	})
	return va
}

// VertexFormat returns the format of the vertices inside a vertex array.
//
// Do not change this format!
func (va *VertexArray) VertexFormat() VertexFormat {
	return va.format
}

// SetDrawMode sets the draw mode of a vertex array. Subsequent calls to Draw will use this draw mode.
func (va *VertexArray) SetDrawMode(mode VertexDrawMode) {
	Do(func() {
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
	va.Begin()
	va.End()
}

// UpdateData overwrites the current vertex array data starting at the index offset.
//
// Offset is not a number of bytes, instead, it's an index in the array.
func (va *VertexArray) UpdateData(offset int, data []float64) {
	Do(func() {
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 8*offset, 8*len(data), gl.Ptr(data))
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	})
}

// Begin binds a vertex array and it's associated vertex buffer.
func (va *VertexArray) Begin() {
	va.parent.Begin()
	Do(func() {
		gl.BindVertexArray(va.vao)
		gl.BindBuffer(gl.ARRAY_BUFFER, va.vbo)
	})
}

// End draws a vertex array and unbinds it alongside with it's associated vertex buffer.
func (va *VertexArray) End() {
	Do(func() {
		gl.DrawArrays(uint32(va.mode), 0, int32(va.count))
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)
		gl.BindVertexArray(0)
	})
	va.parent.End()
}
