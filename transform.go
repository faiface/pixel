package pixel

import "github.com/go-gl/mathgl/mgl32"

// Transform holds space transformation information. Concretely, a transformation is specified by position,
// anchor, scale and rotation.
//
// All points are first rotated around the anchor. Then they are multiplied by the scale. If the
// scale factor is 2, the object becomes 2x bigger. Finally, all points are moved, so that the original
// anchor is located precisely at the position.
//
// Create a Transform object with the Position function. This sets the position variable, which is the
// most important. Then use methods, like Scale and Rotate to change scale, rotation and achor. The order
// in which you apply these methods is irrelevant.
//
//   pixel.Position(pixel.V(100, 100)).Rotate(math.Pi / 3).Scale(1.5)
type Transform struct {
	pos, anc, sca Vec
	rot           float64
}

// Position creates a Transformation object with specified position. Anchor is (0, 0), rotation is 0 and scale is 1.
func Position(position Vec) Transform {
	return Transform{
		pos: position,
		sca: V(1, 1),
	}
}

// Move adds delta to position.
func (t Transform) Move(delta Vec) Transform {
	t.pos += delta
	return t
}

// Anchor sets anchor. Anchor is the rotation center and will be moved to the position.
func (t Transform) Anchor(anchor Vec) Transform {
	t.anc = anchor
	return t
}

// MoveAnchor adds delta to anchor.
func (t Transform) MoveAnchor(delta Vec) Transform {
	t.anc += delta
	return t
}

// Scale scales the transform by the supplied factor.
//
// Note, that subsequent calls to this method accumulate the final scale factor. Scaling two times by 2 is equivalent
// to scaling once by 4.
func (t Transform) Scale(scale float64) Transform {
	t.sca *= V(scale, scale)
	return t
}

// ScaleXY scales the transform by the supplied X and Y factor. Note, that scale is applied before rotation.
//
// Note, that subsequent calls to this method accumulate the final scale factor. Scaling two times by 2 is equivalent
// to scaling once by 4.
func (t Transform) ScaleXY(scale Vec) Transform {
	t.sca *= scale
	return t
}

// Rotate rotates the transform by the supplied angle in radians.
//
// Note, that subsequent calls to this method accumulate the final rotation. Rotating two times by Pi/2 is
// equivalent to rotating once by Pi.
func (t Transform) Rotate(angle float64) Transform {
	t.rot += angle
	return t
}

// Mat3 returns a transformation matrix that satisfies previously set transform properties.
func (t Transform) Mat3() mgl32.Mat3 {
	mat := mgl32.Ident3()
	mat = mat.Mul3(mgl32.Translate2D(float32(t.pos.X()), float32(t.pos.Y())))
	mat = mat.Mul3(mgl32.Rotate3DZ(float32(t.rot)))
	mat = mat.Mul3(mgl32.Scale2D(float32(t.sca.X()), float32(t.sca.Y())))
	mat = mat.Mul3(mgl32.Translate2D(float32(t.anc.X()), float32(t.anc.Y())))
	return mat
}
