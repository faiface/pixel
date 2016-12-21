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

// GetPosition returns the position of a transform.
func (t Transform) GetPosition() Vec {
	return t.pos
}

// GetAnchor returns the anchor of a transform.
func (t Transform) GetAnchor() Vec {
	return t.anc
}

// GetScale returns the scale (2 dimensional) of transform.
func (t Transform) GetScale() Vec {
	return t.sca
}

// GetRotation returns the rotation of a transform (in radians).
func (t Transform) GetRotation() float64 {
	return t.rot
}

// Position sets position.
func (t Transform) Position(position Vec) Transform {
	t.pos = position
	return t
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
	t.sca = t.sca.Scaled(scale)
	return t
}

// ScaleXY scales the transform by the supplied X and Y factor. Note, that scale is applied before rotation.
//
// Note, that subsequent calls to this method accumulate the final scale factor. Scaling two times by 2 is equivalent
// to scaling once by 4.
func (t Transform) ScaleXY(scale Vec) Transform {
	t.sca = V(t.sca.X()*scale.X(), t.sca.Y()*scale.Y())
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

// Project transforms a vector by a transform.
func (t Transform) Project(v Vec) Vec {
	mat := t.Mat3()
	vec := mgl32.Vec3{float32(v.X()), float32(v.Y()), 1}
	pro := mat.Mul3x1(vec)
	return V(float64(pro.X()), float64(pro.Y()))
}

// Unproject does the inverse operation to Project.
func (t Transform) Unproject(v Vec) Vec {
	mat := t.InvMat3()
	vec := mgl32.Vec3{float32(v.X()), float32(v.Y()), 1}
	unp := mat.Mul3x1(vec)
	return V(float64(unp.X()), float64(unp.Y()))
}

// Mat3 returns a transformation matrix that satisfies previously set transform properties.
func (t Transform) Mat3() mgl32.Mat3 {
	mat := mgl32.Ident3()
	mat = mat.Mul3(mgl32.Translate2D(float32(t.pos.X()), float32(t.pos.Y())))
	mat = mat.Mul3(mgl32.Rotate3DZ(float32(t.rot)))
	mat = mat.Mul3(mgl32.Scale2D(float32(t.sca.X()), float32(t.sca.Y())))
	mat = mat.Mul3(mgl32.Translate2D(float32(-t.anc.X()), float32(-t.anc.Y())))
	return mat
}

// InvMat3 returns an inverse transformation matrix to the matrix returned by Mat3 method.
func (t Transform) InvMat3() mgl32.Mat3 {
	mat := mgl32.Ident3()
	mat = mat.Mul3(mgl32.Translate2D(float32(t.anc.X()), float32(t.anc.Y())))
	mat = mat.Mul3(mgl32.Scale2D(float32(1/t.sca.X()), float32(1/t.sca.Y())))
	mat = mat.Mul3(mgl32.Rotate3DZ(float32(-t.rot)))
	mat = mat.Mul3(mgl32.Translate2D(float32(-t.pos.X()), float32(-t.pos.Y())))
	return mat
}

// Camera is a convenience function, that returns a Transform that acts like a camera.
// Center is the position in the world coordinates, that will be projected onto the center of the screen.
// One unit in world coordinates will be projected onto zoom pixels.
//
// It is possible to apply additional rotations, scales and moves to the returned transform.
func Camera(center, zoom, screenSize Vec) Transform {
	return Position(0).Anchor(center).ScaleXY(2 * zoom).ScaleXY(V(1/screenSize.X(), 1/screenSize.Y()))
}
