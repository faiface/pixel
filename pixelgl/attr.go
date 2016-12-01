package pixelgl

// Attr represents an arbitrary OpenGL attribute, such as a vertex attribute or a shader uniform attribute.
type Attr struct {
	Purpose AttrPurpose
	Type    AttrType
}

// AttrPurpose specified a purpose of an attribute. Feel free to create your own purposes for your own needs.
type AttrPurpose int

const (
	// Position of a vertex
	Position AttrPurpose = iota
	// Color of a vertex
	Color
	// TexCoord are texture coordinates
	TexCoord
	// Transform is an object transformation matrix
	Transform
	// Camera is a camera view matrix
	Camera
)

// AttrType represents the type of an OpenGL attribute.
//
// Important note: Int is 32-bit and Float is 64-bit.
type AttrType int

// List of all possible attribute types.
const (
	Int AttrType = iota
	Float
	Vec2
	Vec3
	Vec4
	Mat2
	Mat23
	Mat24
	Mat3
	Mat32
	Mat34
	Mat4
	Mat42
	Mat43
)

// Size returns the size of a type in bytes.
func (at AttrType) Size() int {
	sizeOf := map[AttrType]int{
		Int:   4,
		Float: 8,
		Vec2:  2 * 8,
		Vec3:  3 * 8,
		Vec4:  4 * 8,
		Mat2:  2 * 2 * 8,
		Mat23: 2 * 3 * 8,
		Mat24: 2 * 4 * 8,
		Mat3:  3 * 3 * 8,
		Mat32: 3 * 2 * 8,
		Mat34: 3 * 4 * 8,
		Mat4:  4 * 4 * 8,
		Mat42: 4 * 2 * 8,
		Mat43: 4 * 3 * 8,
	}
	return sizeOf[at]
}
