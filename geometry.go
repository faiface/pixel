package pixel

import (
	"fmt"
	"math"
	"math/cmplx"
)

// Vec is a 2d vector type. It is unusually implemented as complex128 for convenience. Since
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

// String returns the string representation of a vector u.
//
//   u := pixel.V(4.5, -1.3)
//   u.String()     // returns "Vec(4.5, -1.3)"
//   fmt.Println(u) // Vec(4.5, -1.3)
func (u Vec) String() string {
	return fmt.Sprintf("Vec(%v, %v)", u.X(), u.Y())
}

// X returns the x coordinate of a vector u.
func (u Vec) X() float64 {
	return real(u)
}

// Y returns the y coordinate of a vector u.
func (u Vec) Y() float64 {
	return imag(u)
}

// XY returns the components of a vector in two return values.
func (u Vec) XY() (x, y float64) {
	return real(u), imag(u)
}

// Len returns the length of a vector u.
func (u Vec) Len() float64 {
	return cmplx.Abs(complex128(u))
}

// Angle returns the angle between a vector u and the x-axis. The result is in the range [-Pi, Pi].
func (u Vec) Angle() float64 {
	return cmplx.Phase(complex128(u))
}

// Unit returns a vector of length 1 with the same angle as u.
func (u Vec) Unit() Vec {
	return u / V(u.Len(), 0)
}

// Scaled returns a vector u multiplied by c.
func (u Vec) Scaled(c float64) Vec {
	return u * V(c, 0)
}

// Rotated returns a vector u rotated by the given angle in radians.
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

// Rect is a 2d rectangle aligned with the axis of the coordinate system. It has a position
// and a size.
//
// You can manipulate the position and the size using the usual vector operations.
type Rect struct {
	Pos, Size Vec
}

// R returns a new 2d rectangle with the given position (x, y) and size (w, h).
func R(x, y, w, h float64) Rect {
	return Rect{
		Pos:  V(x, y),
		Size: V(w, h),
	}
}

// String returns the string representation of a rectangle.
//
//   r := pixel.R(100, 50, 200, 300)
//   r.String()     // returns "Rect(100, 50, 200, 300)"
//   fmt.Println(r) // Rect(100, 50, 200, 300)
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%v, %v, %v, %v)", r.X(), r.Y(), r.W(), r.H())
}

// X returns the x coordinate of the position of a rectangle.
func (r Rect) X() float64 {
	return r.Pos.X()
}

// Y returns the y coordinate of the position of a rectangle
func (r Rect) Y() float64 {
	return r.Pos.Y()
}

// W returns the width of a rectangle.
func (r Rect) W() float64 {
	return r.Size.X()
}

// H returns the height of a rectangle.
func (r Rect) H() float64 {
	return r.Size.Y()
}

// XYWH returns all of the four components of a rectangle in four return values.
func (r Rect) XYWH() (x, y, w, h float64) {
	return r.X(), r.Y(), r.W(), r.H()
}

// Center returns the position of the center of a rectangle.
func (r Rect) Center() Vec {
	return r.Pos + r.Size.Scaled(0.5)
}
