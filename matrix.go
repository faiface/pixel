package pixel

import (
	"fmt"
	"math"
)

// Matrix is a 2x3 affine matrix that can be used for all kinds of spatial transforms, such
// as movement, scaling and rotations.
//
// Matrix has a handful of useful methods, each of which adds a transformation to the matrix. For
// example:
//
//   pixel.IM.Moved(pixel.V(100, 200)).Rotated(pixel.ZV, math.Pi/2)
//
// This code creates a Matrix that first moves everything by 100 units horizontally and 200 units
// vertically and then rotates everything by 90 degrees around the origin.
//
// Layout is:
// [0] [2] [4]
// [1] [3] [5]
//  0   0   1  (implicit row)
type Matrix [6]float64

// IM stands for identity matrix. Does nothing, no transformation.
var IM = Matrix{1, 0, 0, 1, 0, 0}

// String returns a string representation of the Matrix.
//
//   m := pixel.IM
//   fmt.Println(m) // Matrix(1 0 0 | 0 1 0)
func (m Matrix) String() string {
	return fmt.Sprintf(
		"Matrix(%v %v %v | %v %v %v)",
		m[0], m[2], m[4],
		m[1], m[3], m[5],
	)
}

// Moved moves everything by the delta vector.
func (m Matrix) Moved(delta Vec) Matrix {
	m[4], m[5] = m[4]+delta.X, m[5]+delta.Y
	return m
}

// ScaledXY scales everything around a given point by the scale factor in each axis respectively.
func (m Matrix) ScaledXY(around Vec, scale Vec) Matrix {
	m[4], m[5] = m[4]-around.X, m[5]-around.Y
	m[0], m[2], m[4] = m[0]*scale.X, m[2]*scale.X, m[4]*scale.X
	m[1], m[3], m[5] = m[1]*scale.Y, m[3]*scale.Y, m[5]*scale.Y
	m[4], m[5] = m[4]+around.X, m[5]+around.Y
	return m
}

// Scaled scales everything around a given point by the scale factor.
func (m Matrix) Scaled(around Vec, scale float64) Matrix {
	return m.ScaledXY(around, V(scale, scale))
}

// Rotated rotates everything around a given point by the given angle in radians.
func (m Matrix) Rotated(around Vec, angle float64) Matrix {
	sint, cost := math.Sincos(angle)
	m[4], m[5] = m[4]-around.X, m[5]-around.Y
	m = m.Chained(Matrix{cost, sint, -sint, cost, 0, 0})
	m[4], m[5] = m[4]+around.X, m[5]+around.Y
	return m
}

// Chained adds another Matrix to this one. All tranformations by the next Matrix will be applied
// after the transformations of this Matrix.
func (m Matrix) Chained(next Matrix) Matrix {
	return Matrix{
		next[0]*m[0] + next[2]*m[1],
		next[1]*m[0] + next[3]*m[1],
		next[0]*m[2] + next[2]*m[3],
		next[1]*m[2] + next[3]*m[3],
		next[0]*m[4] + next[2]*m[5] + next[4],
		next[1]*m[4] + next[3]*m[5] + next[5],
	}
}

// Project applies all transformations added to the Matrix to a vector u and returns the result.
//
// Time complexity is O(1).
func (m Matrix) Project(u Vec) Vec {
	return Vec{m[0]*u.X + m[2]*u.Y + m[4], m[1]*u.X + m[3]*u.Y + m[5]}
}

// Unproject does the inverse operation to Project.
//
// Time complexity is O(1).
func (m Matrix) Unproject(u Vec) Vec {
	det := m[0]*m[3] - m[2]*m[1]
	return Vec{
		(m[3]*(u.X-m[4]) - m[2]*(u.Y-m[5])) / det,
		(-m[1]*(u.X-m[4]) + m[0]*(u.Y-m[5])) / det,
	}
}
