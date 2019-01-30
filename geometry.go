package pixel

import (
	"fmt"
	"math"
)

// Clamp returns x clamped to the interval [min, max].
//
// If x is less than min, min is returned. If x is more than max, max is returned. Otherwise, x is
// returned.
func Clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

// Vec is a 2D vector type with X and Y coordinates.
//
// Create vectors with the V constructor:
//
//   u := pixel.V(1, 2)
//   v := pixel.V(8, -3)
//
// Use various methods to manipulate them:
//
//   w := u.Add(v)
//   fmt.Println(w)        // Vec(9, -1)
//   fmt.Println(u.Sub(v)) // Vec(-7, 5)
//   u = pixel.V(2, 3)
//   v = pixel.V(8, 1)
//   if u.X < 0 {
//	     fmt.Println("this won't happen")
//   }
//   x := u.Unit().Dot(v.Unit())
type Vec struct {
	X, Y float64
}

// ZV is a zero vector.
var ZV = Vec{0, 0}

// V returns a new 2D vector with the given coordinates.
func V(x, y float64) Vec {
	return Vec{x, y}
}

// Unit returns a vector of length 1 facing the given angle.
func Unit(angle float64) Vec {
	return Vec{1, 0}.Rotated(angle)
}

// String returns the string representation of the vector u.
//
//   u := pixel.V(4.5, -1.3)
//   u.String()     // returns "Vec(4.5, -1.3)"
//   fmt.Println(u) // Vec(4.5, -1.3)
func (u Vec) String() string {
	return fmt.Sprintf("Vec(%v, %v)", u.X, u.Y)
}

// XY returns the components of the vector in two return values.
func (u Vec) XY() (x, y float64) {
	return u.X, u.Y
}

// Add returns the sum of vectors u and v.
func (u Vec) Add(v Vec) Vec {
	return Vec{
		u.X + v.X,
		u.Y + v.Y,
	}
}

// Sub returns the difference betweeen vectors u and v.
func (u Vec) Sub(v Vec) Vec {
	return Vec{
		u.X - v.X,
		u.Y - v.Y,
	}
}

// Floor converts x and y to their integer equivalents.
func (u Vec) Floor() Vec {
	return Vec{
		math.Floor(u.X),
		math.Floor(u.Y),
	}
}

// To returns the vector from u to v. Equivalent to v.Sub(u).
func (u Vec) To(v Vec) Vec {
	return Vec{
		v.X - u.X,
		v.Y - u.Y,
	}
}

// Scaled returns the vector u multiplied by c.
func (u Vec) Scaled(c float64) Vec {
	return Vec{u.X * c, u.Y * c}
}

// ScaledXY returns the vector u multiplied by the vector v component-wise.
func (u Vec) ScaledXY(v Vec) Vec {
	return Vec{u.X * v.X, u.Y * v.Y}
}

// Len returns the length of the vector u.
func (u Vec) Len() float64 {
	return math.Hypot(u.X, u.Y)
}

// Angle returns the angle between the vector u and the x-axis. The result is in range [-Pi, Pi].
func (u Vec) Angle() float64 {
	return math.Atan2(u.Y, u.X)
}

// Unit returns a vector of length 1 facing the direction of u (has the same angle).
func (u Vec) Unit() Vec {
	if u.X == 0 && u.Y == 0 {
		return Vec{1, 0}
	}
	return u.Scaled(1 / u.Len())
}

// Rotated returns the vector u rotated by the given angle in radians.
func (u Vec) Rotated(angle float64) Vec {
	sin, cos := math.Sincos(angle)
	return Vec{
		u.X*cos - u.Y*sin,
		u.X*sin + u.Y*cos,
	}
}

// Normal returns a vector normal to u. Equivalent to u.Rotated(math.Pi / 2), but faster.
func (u Vec) Normal() Vec {
	return Vec{-u.Y, u.X}
}

// Dot returns the dot product of vectors u and v.
func (u Vec) Dot(v Vec) float64 {
	return u.X*v.X + u.Y*v.Y
}

// Cross return the cross product of vectors u and v.
func (u Vec) Cross(v Vec) float64 {
	return u.X*v.Y - v.X*u.Y
}

// Project returns a projection (or component) of vector u in the direction of vector v.
//
// Behaviour is undefined if v is a zero vector.
func (u Vec) Project(v Vec) Vec {
	len := u.Dot(v) / v.Len()
	return v.Unit().Scaled(len)
}

// Map applies the function f to both x and y components of the vector u and returns the modified
// vector.
//
//   u := pixel.V(10.5, -1.5)
//   v := u.Map(math.Floor)   // v is Vec(10, -2), both components of u floored
func (u Vec) Map(f func(float64) float64) Vec {
	return Vec{
		f(u.X),
		f(u.Y),
	}
}

// Lerp returns a linear interpolation between vectors a and b.
//
// This function basically returns a point along the line between a and b and t chooses which one.
// If t is 0, then a will be returned, if t is 1, b will be returned. Anything between 0 and 1 will
// return the appropriate point between a and b and so on.
func Lerp(a, b Vec, t float64) Vec {
	return a.Scaled(1 - t).Add(b.Scaled(t))
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
//
// Note that the returned rectangle is not automatically normalized.
func R(minX, minY, maxX, maxY float64) Rect {
	return Rect{
		Min: Vec{minX, minY},
		Max: Vec{maxX, maxY},
	}
}

// String returns the string representation of the Rect.
//
//   r := pixel.R(100, 50, 200, 300)
//   r.String()     // returns "Rect(100, 50, 200, 300)"
//   fmt.Println(r) // Rect(100, 50, 200, 300)
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%v, %v, %v, %v)", r.Min.X, r.Min.Y, r.Max.X, r.Max.Y)
}

// Norm returns the Rect in normal form, such that Max is component-wise greater or equal than Min.
func (r Rect) Norm() Rect {
	return Rect{
		Min: Vec{
			math.Min(r.Min.X, r.Max.X),
			math.Min(r.Min.Y, r.Max.Y),
		},
		Max: Vec{
			math.Max(r.Min.X, r.Max.X),
			math.Max(r.Min.Y, r.Max.Y),
		},
	}
}

// W returns the width of the Rect.
func (r Rect) W() float64 {
	return r.Max.X - r.Min.X
}

// H returns the height of the Rect.
func (r Rect) H() float64 {
	return r.Max.Y - r.Min.Y
}

// Size returns the vector of width and height of the Rect.
func (r Rect) Size() Vec {
	return V(r.W(), r.H())
}

// Area returns the area of r. If r is not normalized, area may be negative.
func (r Rect) Area() float64 {
	return r.W() * r.H()
}

// Center returns the position of the center of the Rect.
func (r Rect) Center() Vec {
	return Lerp(r.Min, r.Max, 0.5)
}

// Moved returns the Rect moved (both Min and Max) by the given vector delta.
func (r Rect) Moved(delta Vec) Rect {
	return Rect{
		Min: r.Min.Add(delta),
		Max: r.Max.Add(delta),
	}
}

// Resized returns the Rect resized to the given size while keeping the position of the given
// anchor.
//
//   r.Resized(r.Min, size)      // resizes while keeping the position of the lower-left corner
//   r.Resized(r.Max, size)      // same with the top-right corner
//   r.Resized(r.Center(), size) // resizes around the center
//
// This function does not make sense for resizing a rectangle of zero area and will panic. Use
// ResizedMin in the case of zero area.
func (r Rect) Resized(anchor, size Vec) Rect {
	if r.W()*r.H() == 0 {
		panic(fmt.Errorf("(%T).Resize: zero area", r))
	}
	fraction := Vec{size.X / r.W(), size.Y / r.H()}
	return Rect{
		Min: anchor.Add(r.Min.Sub(anchor).ScaledXY(fraction)),
		Max: anchor.Add(r.Max.Sub(anchor).ScaledXY(fraction)),
	}
}

// ResizedMin returns the Rect resized to the given size while keeping the position of the Rect's
// Min.
//
// Sizes of zero area are safe here.
func (r Rect) ResizedMin(size Vec) Rect {
	return Rect{
		Min: r.Min,
		Max: r.Min.Add(size),
	}
}

// Contains checks whether a vector u is contained within this Rect (including it's borders).
func (r Rect) Contains(u Vec) bool {
	return r.Min.X <= u.X && u.X <= r.Max.X && r.Min.Y <= u.Y && u.Y <= r.Max.Y
}

// Union returns the minimal Rect which covers both r and s. Rects r and s must be normalized.
func (r Rect) Union(s Rect) Rect {
	return R(
		math.Min(r.Min.X, s.Min.X),
		math.Min(r.Min.Y, s.Min.Y),
		math.Max(r.Max.X, s.Max.X),
		math.Max(r.Max.Y, s.Max.Y),
	)
}

// Intersect returns the maximal Rect which is covered by both r and s. Rects r and s must be normalized.
//
// If r and s don't overlap, this function returns R(0, 0, 0, 0).
func (r Rect) Intersect(s Rect) Rect {
	t := R(
		math.Max(r.Min.X, s.Min.X),
		math.Max(r.Min.Y, s.Min.Y),
		math.Min(r.Max.X, s.Max.X),
		math.Min(r.Max.Y, s.Max.Y),
	)
	if t.Min.X >= t.Max.X || t.Min.Y >= t.Max.Y {
		return Rect{}
	}
	return t
}

// IntersectsCircle returns whether the Circle and the Rect intersect.
//
// This function will return true if:
//  - The Rect contains the Circle, partially or fully
//  - The Circle contains the Rect, partially of fully
//  - An edge of the Rect is a tangent to the Circle
func (r Rect) IntersectsCircle(c Circle) bool {
	return c.IntersectsRect(r)
}

// Circle is a 2D circle. It is defined by two properties:
//  - Center vector
//  - Radius float64
type Circle struct {
	Center Vec
	Radius float64
}

// C returns a new Circle with the given radius and center coordinates.
//
// Note that a negative radius is valid.
func C(center Vec, radius float64) Circle {
	return Circle{
		Center: center,
		Radius: radius,
	}
}

// String returns the string representation of the Circle.
//
//  c := pixel.C(10.1234, pixel.ZV)
//  c.String()     // returns "Circle(10.12, Vec(0, 0))"
//  fmt.Println(c) // Circle(10.12, Vec(0, 0))
func (c Circle) String() string {
	return fmt.Sprintf("Circle(%s, %.2f)", c.Center, c.Radius)
}

// Norm returns the Circle in normalized form - this sets the radius to its absolute value.
//
// c := pixel.C(-10, pixel.ZV)
// c.Norm() // returns pixel.Circle{pixel.Vec{0, 0}, 10}
func (c Circle) Norm() Circle {
	return Circle{
		Center: c.Center,
		Radius: math.Abs(c.Radius),
	}
}

// Area returns the area of the Circle.
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * 2
}

// Moved returns the Circle moved by the given vector delta.
func (c Circle) Moved(delta Vec) Circle {
	return Circle{
		Center: c.Center.Add(delta),
		Radius: c.Radius,
	}
}

// Resized returns the Circle resized by the given delta.  The Circles center is use as the anchor.
//
// c := pixel.C(pixel.ZV, 10)
// c.Resized(-5) // returns pixel.Circle{pixel.Vec{0, 0}, 5}
// c.Resized(25) // returns pixel.Circle{pixel.Vec{0, 0}, 35}
func (c Circle) Resized(radiusDelta float64) Circle {
	return Circle{
		Center: c.Center,
		Radius: c.Radius + radiusDelta,
	}
}

// Contains checks whether a vector `u` is contained within this Circle (including it's perimeter).
func (c Circle) Contains(u Vec) bool {
	toCenter := c.Center.To(u)
	return c.Radius >= toCenter.Len()
}

// maxCircle will return the larger circle based on the radius.
func maxCircle(c, d Circle) Circle {
	if c.Radius < d.Radius {
		return d
	}
	return c
}

// minCircle will return the smaller circle based on the radius.
func minCircle(c, d Circle) Circle {
	if c.Radius < d.Radius {
		return c
	}
	return d
}

// Union returns the minimal Circle which covers both `c` and `d`.
func (c Circle) Union(d Circle) Circle {
	biggerC := maxCircle(c.Norm(), d.Norm())
	smallerC := minCircle(c.Norm(), d.Norm())

	// Get distance between centers
	dist := c.Center.To(d.Center).Len()

	// If the bigger Circle encompasses the smaller one, we have the result
	if dist+smallerC.Radius <= biggerC.Radius {
		return biggerC
	}

	// Calculate radius for encompassing Circle
	r := (dist + biggerC.Radius + smallerC.Radius) / 2

	// Calculate center for encompassing Circle
	theta := .5 + (biggerC.Radius-smallerC.Radius)/(2*dist)
	center := Lerp(smallerC.Center, biggerC.Center, theta)

	return Circle{
		Center: center,
		Radius: r,
	}
}

// Intersect returns the maximal Circle which is covered by both `c` and `d`.
//
// If `c` and `d` don't overlap, this function returns a zero-sized circle at the centerpoint between the two Circle's
// centers.
func (c Circle) Intersect(d Circle) Circle {
	// Check if one of the circles encompasses the other; if so, return that one
	biggerC := maxCircle(c.Norm(), d.Norm())
	smallerC := minCircle(c.Norm(), d.Norm())

	if biggerC.Radius >= biggerC.Center.To(smallerC.Center).Len()+smallerC.Radius {
		return biggerC
	}

	// Calculate the midpoint between the two radii
	// Distance between centers
	dist := c.Center.To(d.Center).Len()
	// Difference between radii
	diff := dist - (c.Radius + d.Radius)
	// Distance from c.Center to the weighted midpoint
	distToMidpoint := c.Radius + 0.5*diff
	// Weighted midpoint
	center := Lerp(c.Center, d.Center, distToMidpoint/dist)

	// No need to calculate radius if the circles do not overlap
	if c.Center.To(d.Center).Len() >= c.Radius+d.Radius {
		return C(center, 0)
	}

	radius := c.Center.To(d.Center).Len() - (c.Radius + d.Radius)

	return Circle{
		Center: center,
		Radius: math.Abs(radius),
	}
}

// IntersectsRect returns whether the Circle and the Rect intersect.
//
// This function will return true if:
//  - The Rect contains the Circle, partially or fully
//  - The Circle contains the Rect, partially of fully
//  - An edge of the Rect is a tangent to the Circle
func (c Circle) IntersectsRect(r Rect) bool {
	// Checks if the c.Center is not in the diagonal quadrants of the rectangle
	if (r.Min.X <= c.Center.X && c.Center.X <= r.Max.X) || (r.Min.Y <= c.Center.Y && c.Center.Y <= r.Max.Y) {
		// 'grow' the Rect by c.Radius in each orthagonal
		return Rect{
			Min: r.Min.Sub(V(c.Radius, c.Radius)),
			Max: r.Max.Add(V(c.Radius, c.Radius)),
		}.Contains(c.Center)
	}
	// The center is in the diagonal quadrants
	return c.Center.To(r.Min).Len() <= c.Radius || c.Center.To(r.Max).Len() <= c.Radius ||
		c.Center.To(V(r.Min.X, r.Max.Y)).Len() <= c.Radius || c.Center.To(V(r.Max.X, r.Min.Y)).Len() <= c.Radius
}

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
