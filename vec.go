package pixel

import (
	"fmt"
	"math"
	"math/cmplx"
)

// Vec2 is a 2d vector type. It is unusually implemented as complex128 for convenience. Since Go
// does not allow operator overloading, implementing vector as a struct leads to a bunch of methods
// for addition, subtraction and multiplication of vectors. With complex128, much of this
// functionality is given through operators.
//
// Create vectors with the V constructor:
//
//   u := pixel.V(1, 2)
//   v := pixel.V(8, -3)
//
// Add and subtract them using the standard + and - operators:
//
//   w := u + v
//   fmt.Println(w)     // Vec2(9, -1)
//   fmt.Println(u - v) // Vec2(-7, 5)
//
// Additional standard vector operations can be obtained with methods:
//
//   u := pixel.Vec(2, 3)
//   v := pixel.Vec(8, 1)
//   if u.X() < 0 {
//       fmt.Println("this won't happend")
//   }
//   x := u.Unit().Dot(v.Unit())
type Vec2 complex128

// V returns a new 2d vector with the given coordinates.
func V(x, y float64) Vec2 {
	return Vec2(complex(x, y))
}

// String returns the string representation of a vector u as "Vec2(x, y)".
func (u Vec2) String() string {
	return fmt.Sprintf("Vec2(%v, %v)", u.X(), u.Y())
}

// X returns the x coordinate of a vector u.
func (u Vec2) X() float64 {
	return real(u)
}

// Y returns the y coordinate of a vector u.
func (u Vec2) Y() float64 {
	return imag(u)
}

// Len returns the length of a vector u.
func (u Vec2) Len() float64 {
	return cmplx.Abs(complex128(u))
}

// Angle returns the angle between a vector u and the x-axis. The result is in the range [-Pi, Pi].
func (u Vec2) Angle() float64 {
	return cmplx.Phase(complex128(u))
}

// Unit returns a vector of length 1 with the same angle as u.
func (u Vec2) Unit() Vec2 {
	return u / V(u.Len(), 0)
}

// Scaled returns a vector u multiplied by k.
func (u Vec2) Scaled(k float64) Vec2 {
	return u * V(k, 0)
}

// Rotated returns a vector u rotated by the given angle in radians.
func (u Vec2) Rotated(angle float64) Vec2 {
	return u * V(math.Cos(angle), math.Sin(angle))
}

// Dot returns the dot product of vectors u and v.
func (u Vec2) Dot(v Vec2) float64 {
	return u.X()*v.X() + u.Y()*v.Y()
}

// Cross return the cross product of vectors u and v.
func (u Vec2) Cross(v Vec2) float64 {
	return u.X()*v.Y() - v.X()*u.Y()
}
