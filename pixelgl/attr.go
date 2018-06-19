package pixelgl

import "github.com/go-gl/mathgl/mgl32"

// AttrType is the attribute's identifier
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
	Intp // pointers
	Floatp
	Vec2p
	Vec3p
	Vec4p
	Mat2p
	Mat23p
	Mat24p
	Mat3p
	Mat32p
	Mat34p
	Mat4p
	Mat42p
	Mat43p
)

// Returns the type identifier for any (supported) variable type
func getAttrType(v interface{}) AttrType {
	switch v.(type) {
	case int32:
		return Int
	case float32:
		return Float
	case mgl32.Vec2:
		return Vec2
	case mgl32.Vec3:
		return Vec3
	case mgl32.Vec4:
		return Vec4
	case mgl32.Mat2:
		return Mat2
	case mgl32.Mat2x3:
		return Mat23
	case mgl32.Mat2x4:
		return Mat24
	case mgl32.Mat3:
		return Mat3
	case mgl32.Mat3x2:
		return Mat32
	case mgl32.Mat3x4:
		return Mat34
	case mgl32.Mat4:
		return Mat4
	case mgl32.Mat4x2:
		return Mat42
	case mgl32.Mat4x3:
		return Mat43
	case *mgl32.Vec2:
		return Vec2p
	case *mgl32.Vec3:
		return Vec3p
	case *mgl32.Vec4:
		return Vec4p
	case *mgl32.Mat2:
		return Mat2p
	case *mgl32.Mat2x3:
		return Mat23p
	case *mgl32.Mat2x4:
		return Mat24p
	case *mgl32.Mat3:
		return Mat3p
	case *mgl32.Mat3x2:
		return Mat32p
	case *mgl32.Mat3x4:
		return Mat34p
	case *mgl32.Mat4:
		return Mat4p
	case *mgl32.Mat4x2:
		return Mat42p
	case *mgl32.Mat4x3:
		return Mat43p
	case *int32:
		return Intp
	case *float32:
		return Floatp
	default:
		panic("invalid AttrType")
	}
}
