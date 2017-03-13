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

// V returns a new 2d vector with the given coordinates.
func V(x, y float64) Vec {
	return Vec(complex(x, y))
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

// Unit returns a vector of length 1 with the same angle as u.
func (u Vec) Unit() Vec {
	return u / V(u.Len(), 0)
}

// Scaled returns the vector u multiplied by c.
func (u Vec) Scaled(c float64) Vec {
	return u * V(c, 0)
}

// ScaledXY returns the vector u multiplied by vector v component-wise.
func (u Vec) ScaledXY(v Vec) Vec {
	return V(u.X()*v.X(), u.Y()*v.Y())
}

// Rotated returns the vector u rotated by the given angle in radians.
func (u Vec) Rotated(angle float64) Vec {
	sin, cos := math.Sincos(angle)
	return u * V(cos, sin)
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
func (u Vec) Map(f func(float64) float64) Vec {
	return V(
		f(u.X()),
		f(u.Y()),
	)
}

// Lerp returns a linear interpolation between vectors a and b.
//
// This function basically returns a point along the line between a and b and t chooses which point.
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
	}.Norm()
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

// String returns the string representation of the rectangle.
//
//   r := pixel.R(100, 50, 200, 300)
//   r.String()     // returns "Rect(100, 50, 200, 300)"
//   fmt.Println(r) // Rect(100, 50, 200, 300)
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%v, %v, %v, %v)", r.Min.X(), r.Min.Y(), r.Max.X(), r.Max.Y())
}

// W returns the width of the rectangle.
func (r Rect) W() float64 {
	return r.Max.X() - r.Min.X()
}

// H returns the height of the rectangle.
func (r Rect) H() float64 {
	return r.Max.Y() - r.Min.Y()
}

// Center returns the position of the center of the rectangle.
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

// Resized returns the Rect resized to the given size while keeping the position of the given anchor.
//   r.Resized(r.Min, size)      // resizes while keeping the position of the lower-left corner
//   r.Resized(r.Max, size)      // same with the top-right corner
//   r.Resized(r.Center(), size) // resizes around the center
// This function does not make sense for size of zero area and will panic. Use ResizeMin in the case
// of zero area.
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
//   pixel.ZM.Move(pixel.V(100, 200)).Rotate(0, math.Pi/2)
//
// This code creates a Matrix that first moves everything by 100 units horizontaly and 200 units
// vertically and then rotates everything by 90 degrees around the origin.
type Matrix [9]float64

// ZM stands for Zero-Matrix which is the identity matrix. Does nothing, no transformation.
var ZM = Matrix(mgl64.Ident3())

// Move moves everything by the delta vector.
func (m Matrix) Move(delta Vec) Matrix {
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

// Project applies all transformations added to the Matrix to a vector u and returns the result.
func (m Matrix) Project(u Vec) Vec {
	m3 := mgl64.Mat3(m)
	proj := m3.Mul3x1(mgl64.Vec3{u.X(), u.Y(), 1})
	return V(proj.X(), proj.Y())
}

// Unproject does the inverse operation to Project.
func (m Matrix) Unproject(u Vec) Vec {
	m3 := mgl64.Mat3(m)
	inv := m3.Inv()
	unproj := inv.Mul3x1(mgl64.Vec3{u.X(), u.Y(), 1})
	return V(unproj.X(), unproj.Y())
}
