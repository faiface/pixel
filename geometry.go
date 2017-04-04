package pixel

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/go-gl/mathgl/mgl64"
)

// Vec is a 2D vector type. It is unusually implemented as complex128 for convenience. Since
// Go does not allow operator overloading, implementing vector as a struct leads to a bunch of
// methods for addition, subtraction and multiplication of vectors. With complex128, much of
// this functionality is given through operators.
//
// Create vectors with the V constructor:
//
//   u := pixel.V(1, 2)
//   v := pixel.V(8, -3)
//
// Add and subtract them using the standard + and - operators:
//
//   w := u + v
//   fmt.Println(w)     // Vec(9, -1)
//   fmt.Println(u - v) // Vec(-7, 5)
//
// Additional standard vector operations can be obtained with methods:
//
//   u := pixel.V(2, 3)
//   v := pixel.V(8, 1)
//   if u.X() < 0 {
//	     fmt.Println("this won't happen")
//   }
//   x := u.Unit().Dot(v.Unit())
type Vec complex128

// V returns a new 2D vector with the given coordinates.
func V(x, y float64) Vec {
	return Vec(complex(x, y))
}

// X returns a 2D vector with coordinates (x, 0).
func X(x float64) Vec {
	return V(x, 0)
}

// Y returns a 2D vector with coordinates (0, y).
func Y(y float64) Vec {
	return V(0, y)
}

// String returns the string representation of the vector u.
//
//   u := pixel.V(4.5, -1.3)
//   u.String()     // returns "Vec(4.5, -1.3)"
//   fmt.Println(u) // Vec(4.5, -1.3)
func (u Vec) String() string {
	return fmt.Sprintf("Vec(%v, %v)", u.X(), u.Y())
}

// X returns the x coordinate of the vector u.
func (u Vec) X() float64 {
	return real(u)
}

// Y returns the y coordinate of the vector u.
func (u Vec) Y() float64 {
	return imag(u)
}

// XY returns the components of the vector in two return values.
func (u Vec) XY() (x, y float64) {
	return real(u), imag(u)
}

// Len returns the length of the vector u.
func (u Vec) Len() float64 {
	return cmplx.Abs(complex128(u))
}

// Angle returns the angle between the vector u and the x-axis. The result is in the range [-Pi, Pi].
func (u Vec) Angle() float64 {
	return cmplx.Phase(complex128(u))
}

// Unit returns a vector of length 1 facing the direction of u (has the same angle).
func (u Vec) Unit() Vec {
	if u == 0 {
		return 1
	}
	return u / V(u.Len(), 0)
}

// Scaled returns the vector u multiplied by c.
func (u Vec) Scaled(c float64) Vec {
	return u * V(c, 0)
}

// ScaledXY returns the vector u multiplied by the vector v component-wise.
func (u Vec) ScaledXY(v Vec) Vec {
	return V(u.X()*v.X(), u.Y()*v.Y())
}

// Rotated returns the vector u rotated by the given angle in radians.
func (u Vec) Rotated(angle float64) Vec {
	sin, cos := math.Sincos(angle)
	return u * V(cos, sin)
}

// WithX return the vector u with the x coordinate changed to the given value.
func (u Vec) WithX(x float64) Vec {
	return V(x, u.Y())
}

// WithY returns the vector u with the y coordinate changed to the given value.
func (u Vec) WithY(y float64) Vec {
	return V(u.X(), y)
}

// Dot returns the dot product of vectors u and v.
func (u Vec) Dot(v Vec) float64 {
	return u.X()*v.X() + u.Y()*v.Y()
}

// Cross return the cross product of vectors u and v.
func (u Vec) Cross(v Vec) float64 {
	return u.X()*v.Y() - v.X()*u.Y()
}

// Map applies the function f to both x and y components of the vector u and returns the modified
// vector.
//
//   u := pixel.V(10.5, -1.5)
//   v := u.Map(math.Floor)   // v is Vec(10, -2), both components of u floored
func (u Vec) Map(f func(float64) float64) Vec {
	return V(
		f(u.X()),
		f(u.Y()),
	)
}

// Lerp returns a linear interpolation between vectors a and b.
//
// This function basically returns a point along the line between a and b and t chooses which one.
// If t is 0, then a will be returned, if t is 1, b will be returned. Anything between 0 and 1 will
// return the appropriate point between a and b and so on.
func Lerp(a, b Vec, t float64) Vec {
	return a.Scaled(1-t) + b.Scaled(t)
}

// Rect is a 2D rectangle aligned with the axes of the coordinate system. It is defined by two
// points, Min and Max.
//
// The invariant should hold, that Max's components are greater or equal than Min's components
// respectively.
type Rect struct {
	Min, Max Vec
}

// R returns a new Rect with given the Min and Max coordinates.
func R(minX, minY, maxX, maxY float64) Rect {
	return Rect{
		Min: V(minX, minY),
		Max: V(maxX, maxY),
	}
}

// String returns the string representation of the Rect.
//
//   r := pixel.R(100, 50, 200, 300)
//   r.String()     // returns "Rect(100, 50, 200, 300)"
//   fmt.Println(r) // Rect(100, 50, 200, 300)
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%v, %v, %v, %v)", r.Min.X(), r.Min.Y(), r.Max.X(), r.Max.Y())
}

// Norm returns the Rect in normal form, such that Max is component-wise greater or equal than Min.
func (r Rect) Norm() Rect {
	return Rect{
		Min: V(
			math.Min(r.Min.X(), r.Max.X()),
			math.Min(r.Min.Y(), r.Max.Y()),
		),
		Max: V(
			math.Max(r.Min.X(), r.Max.X()),
			math.Max(r.Min.Y(), r.Max.Y()),
		),
	}
}

// W returns the width of the Rect.
func (r Rect) W() float64 {
	return r.Max.X() - r.Min.X()
}

// H returns the height of the Rect.
func (r Rect) H() float64 {
	return r.Max.Y() - r.Min.Y()
}

// Size returns the vector of width and height of the Rect.
func (r Rect) Size() Vec {
	return V(r.W(), r.H())
}

// Center returns the position of the center of the Rect.
func (r Rect) Center() Vec {
	return (r.Min + r.Max) / 2
}

// Moved returns the Rect moved (both Min and Max) by the given vector delta.
func (r Rect) Moved(delta Vec) Rect {
	return Rect{
		Min: r.Min + delta,
		Max: r.Max + delta,
	}
}

// WithMin returns the Rect with it's Min changed to the given position.
//
// Note, that the Rect is not automatically normalized.
func (r Rect) WithMin(min Vec) Rect {
	return Rect{
		Min: min,
		Max: r.Max,
	}
}

// WithMax returns the Rect with it's Max changed to the given position.
//
// Note, that the Rect is not automatically normalized.
func (r Rect) WithMax(max Vec) Rect {
	return Rect{
		Min: r.Min,
		Max: max,
	}
}

// Resized returns the Rect resized to the given size while keeping the position of the given
// anchor.
//
//   r.Resized(r.Min, size)      // resizes while keeping the position of the lower-left corner
//   r.Resized(r.Max, size)      // same with the top-right corner
//   r.Resized(r.Center(), size) // resizes around the center
//
// This function does not make sense for sizes of zero area and will panic. Use ResizedMin in the
// case of zero area.
func (r Rect) Resized(anchor, size Vec) Rect {
	if r.W()*r.H() == 0 || size.X()*size.Y() == 0 {
		panic(fmt.Errorf("(%T).Resize: zero area", r))
	}
	fraction := size.ScaledXY(V(1/r.W(), 1/r.H()))
	return Rect{
		Min: anchor + (r.Min - anchor).ScaledXY(fraction),
		Max: anchor + (r.Max - anchor).ScaledXY(fraction),
	}
}

// ResizedMin returns the Rect resized to the given size while keeping the position of the Rect's
// Min.
//
// Sizes of zero area are safe here.
func (r Rect) ResizedMin(size Vec) Rect {
	return Rect{
		Min: r.Min,
		Max: r.Min + size,
	}
}

// Contains checks whether a vector u is contained within this Rect (including it's borders).
func (r Rect) Contains(u Vec) bool {
	return r.Min.X() <= u.X() && u.X() <= r.Max.X() && r.Min.Y() <= u.Y() && u.Y() <= r.Max.Y()
}

// Matrix is a 3x3 transformation matrix that can be used for all kinds of spacial transforms, such
// as movement, scaling and rotations.
//
// Matrix has a handful of useful methods, each of which adds a transformation to the matrix. For
// example:
//
//   pixel.IM.Moved(pixel.V(100, 200)).Rotated(0, math.Pi/2)
//
// This code creates a Matrix that first moves everything by 100 units horizontally and 200 units
// vertically and then rotates everything by 90 degrees around the origin.
type Matrix [9]float64

// IM stands for identity matrix. Does nothing, no transformation.
var IM = Matrix(mgl64.Ident3())

// String returns a string representation of the Matrix.
//
//   m := pixel.IM
//   fmt.Println(m) // Matrix(1 0 0 | 0 1 0 | 0 0 1)
func (m Matrix) String() string {
	return fmt.Sprintf(
		"Matrix(%v %v %v | %v %v %v | %v %v %v)",
		m[0], m[1], m[2],
		m[3], m[4], m[5],
		m[6], m[7], m[8],
	)
}

// Moved moves everything by the delta vector.
func (m Matrix) Moved(delta Vec) Matrix {
	m3 := mgl64.Mat3(m)
	m3 = mgl64.Translate2D(delta.XY()).Mul3(m3)
	return Matrix(m3)
}

// ScaledXY scales everything around a given point by the scale factor in each axis respectively.
func (m Matrix) ScaledXY(around Vec, scale Vec) Matrix {
	m3 := mgl64.Mat3(m)
	m3 = mgl64.Translate2D((-around).XY()).Mul3(m3)
	m3 = mgl64.Scale2D(scale.XY()).Mul3(m3)
	m3 = mgl64.Translate2D(around.XY()).Mul3(m3)
	return Matrix(m3)
}

// Scaled scales everything around a given point by the scale factor.
func (m Matrix) Scaled(around Vec, scale float64) Matrix {
	return m.ScaledXY(around, V(scale, scale))
}

// Rotated rotates everything around a given point by the given angle in radians.
func (m Matrix) Rotated(around Vec, angle float64) Matrix {
	m3 := mgl64.Mat3(m)
	m3 = mgl64.Translate2D((-around).XY()).Mul3(m3)
	m3 = mgl64.Rotate3DZ(angle).Mul3(m3)
	m3 = mgl64.Translate2D(around.XY()).Mul3(m3)
	return Matrix(m3)
}

// Chained adds another Matrix to this one. All tranformations by the next Matrix will be applied
// after the transformations of this Matrix.
func (m Matrix) Chained(next Matrix) Matrix {
	m3 := mgl64.Mat3(m)
	m3 = mgl64.Mat3(next).Mul3(m3)
	return Matrix(m3)
}

// Project applies all transformations added to the Matrix to a vector u and returns the result.
//
// Time complexity is O(1).
func (m Matrix) Project(u Vec) Vec {
	m3 := mgl64.Mat3(m)
	proj := m3.Mul3x1(mgl64.Vec3{u.X(), u.Y(), 1})
	return V(proj.X(), proj.Y())
}

// Unproject does the inverse operation to Project.
//
// Time complexity is O(1).
func (m Matrix) Unproject(u Vec) Vec {
	m3 := mgl64.Mat3(m)
	inv := m3.Inv()
	unproj := inv.Mul3x1(mgl64.Vec3{u.X(), u.Y(), 1})
	return V(unproj.X(), unproj.Y())
}
