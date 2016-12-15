package pixelgl

// AttrFormat defines names and types of OpenGL attributes (vertex format, uniform format, etc.).
//
// Example:
//   AttrFormat{"position": Vec2, "color": Vec4, "texCoord": Vec2}
type AttrFormat map[string]AttrType

// Contains checks whether a format contains a specific attribute.
//
// It does a little more than a hard check: e.g. if you query a Vec2 attribute, but the format contains Vec3,
// Contains returns true, because Vec2 is assignable to Vec3. Specifically, Float -> Vec2 -> Vec3 -> Vec4 (transitively).
// This however does not work for matrices or ints.
func (af AttrFormat) Contains(attr Attr) bool {
	if typ, ok := af[attr.Name]; ok {
		if (Float <= typ && typ <= Vec4) && (Float <= attr.Type && attr.Type <= typ) {
			return true
		}
		return attr.Type == typ
	}
	return false
}

// Size returns the total size of all attributes of an attribute format.
func (af AttrFormat) Size() int {
	total := 0
	for _, typ := range af {
		total += typ.Size()
	}
	return total
}

// Attr represents an arbitrary OpenGL attribute, such as a vertex attribute or a shader uniform attribute.
type Attr struct {
	Name string
	Type AttrType
}

// AttrType represents the type of an OpenGL attribute.
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
	return map[AttrType]int{
		Int:   4,
		Float: 4,
		Vec2:  2 * 4,
		Vec3:  3 * 4,
		Vec4:  4 * 4,
		Mat2:  2 * 2 * 4,
		Mat23: 2 * 3 * 4,
		Mat24: 2 * 4 * 4,
		Mat3:  3 * 3 * 4,
		Mat32: 3 * 2 * 4,
		Mat34: 3 * 4 * 4,
		Mat4:  4 * 4 * 4,
		Mat42: 4 * 2 * 4,
		Mat43: 4 * 3 * 4,
	}[at]
}
