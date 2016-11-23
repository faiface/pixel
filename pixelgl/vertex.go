package pixelgl

// VertexFormat defines a data format in a vertex buffer.
//
// Example:
//
//   vf := VertexFormat{{Position, 2}, {Color, 4}, {TexCoord, 2}}
type VertexFormat []VertexAttribute

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
