package pixel

import (
	"fmt"
	"math"
)

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

// nearlyEqual compares two float64s and returns whether they are equal, accounting for rounding errors.At worst, the
// result is correct to 7 significant digits.
func nearlyEqual(a, b float64) bool {
	epsilon := 0.000001

	if a == b {
		return true
	}

	diff := math.Abs(a - b)

	if a == 0.0 || b == 0.0 || diff < math.SmallestNonzeroFloat64 {
		return diff < (epsilon * math.SmallestNonzeroFloat64)
	}

	absA := math.Abs(a)
	absB := math.Abs(b)

	return diff/math.Min(absA+absB, math.MaxFloat64) < epsilon
}

// Eq will compare two vectors and return whether they are equal accounting for rounding errors.  At worst, the result
// is correct to 7 significant digits.
func (u Vec) Eq(v Vec) bool {
	return nearlyEqual(u.X, v.X) && nearlyEqual(u.Y, v.Y)
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

// Line is a 2D line segment, between points A and B.
type Line struct {
	A, B Vec
}

// L creates and returns a new Line.
func L(from, to Vec) Line {
	return Line{
		A: from,
		B: to,
	}
}

// Bounds returns the lines bounding box.  This is in the form of a normalized Rect.
func (l Line) Bounds() Rect {
	return R(l.A.X, l.A.Y, l.B.X, l.B.Y).Norm()
}

// Center will return the point at center of the line; that is, the point equidistant from either end.
func (l Line) Center() Vec {
	return l.A.Add(l.A.To(l.B).Scaled(0.5))
}

// Closest will return the point on the line which is closest to the Vec provided.
func (l Line) Closest(v Vec) Vec {
	// between is a helper function which determines whether x is greater than min(a, b) and less than max(a, b)
	between := func(a, b, x float64) bool {
		min := math.Min(a, b)
		max := math.Max(a, b)
		return min < x && x < max
	}

	// Closest point will be on a line which perpendicular to this line.
	// If and only if the infinite perpendicular line intersects the segment.
	m, b := l.Formula()

	// Account for horizontal lines
	if m == 0 {
		x := v.X
		y := l.A.Y

		// check if the X coordinate of v is on the line
		if between(l.A.X, l.B.X, v.X) {
			return V(x, y)
		}

		// Otherwise get the closest endpoint
		if l.A.To(v).Len() < l.B.To(v).Len() {
			return l.A
		}
		return l.B
	}

	// Account for vertical lines
	if math.IsInf(math.Abs(m), 1) {
		x := l.A.X
		y := v.Y

		// check if the Y coordinate of v is on the line
		if between(l.A.Y, l.B.Y, v.Y) {
			return V(x, y)
		}

		// Otherwise get the closest endpoint
		if l.A.To(v).Len() < l.B.To(v).Len() {
			return l.A
		}
		return l.B
	}

	perpendicularM := -1 / m
	perpendicularB := v.Y - (perpendicularM * v.X)

	// Coordinates of intersect (of infinite lines)
	x := (perpendicularB - b) / (m - perpendicularM)
	y := m*x + b

	// Check if the point lies between the x and y bounds of the segment
	if !between(l.A.X, l.B.X, x) && !between(l.A.Y, l.B.Y, y) {
		// Not within bounding box
		toStart := v.To(l.A)
		toEnd := v.To(l.B)

		if toStart.Len() < toEnd.Len() {
			return l.A
		}
		return l.B
	}

	return V(x, y)
}

// Contains returns whether the provided Vec lies on the line.
func (l Line) Contains(v Vec) bool {
	return l.Closest(v).Eq(v)
}

// Formula will return the values that represent the line in the formula: y = mx + b
// This function will return math.Inf+, math.Inf- for a vertical line.
func (l Line) Formula() (m, b float64) {
	// Account for horizontal lines
	if l.B.Y == l.A.Y {
		return 0, l.A.Y
	}

	m = (l.B.Y - l.A.Y) / (l.B.X - l.A.X)
	b = l.A.Y - (m * l.A.X)

	return m, b
}

// Intersect will return the point of intersection for the two line segments.  If the line segments do not intersect,
// this function will return the zero-vector and false.
func (l Line) Intersect(k Line) (Vec, bool) {
	// Check if the lines are parallel
	lDir := l.A.To(l.B)
	kDir := k.A.To(k.B)
	if lDir.X == kDir.X && lDir.Y == kDir.Y {
		return ZV, false
	}

	// The lines intersect - but potentially not within the line segments.
	// Get the intersection point for the lines if they were infinitely long, check if the point exists on both of the
	// segments
	lm, lb := l.Formula()
	km, kb := k.Formula()

	// Account for vertical lines
	if math.IsInf(math.Abs(lm), 1) && math.IsInf(math.Abs(km), 1) {
		// Both vertical, therefore parallel
		return ZV, false
	}

	var x, y float64

	if math.IsInf(math.Abs(lm), 1) || math.IsInf(math.Abs(km), 1) {
		// One line is vertical
		intersectM := lm
		intersectB := lb
		verticalLine := k

		if math.IsInf(math.Abs(lm), 1) {
			intersectM = km
			intersectB = kb
			verticalLine = l
		}

		y = intersectM*verticalLine.A.X + intersectB
		x = verticalLine.A.X
	} else {
		// Coordinates of intersect
		x = (kb - lb) / (lm - km)
		y = lm*x + lb
	}

	if l.Contains(V(x, y)) && k.Contains(V(x, y)) {
		// The intersect point is on both line segments, they intersect.
		return V(x, y), true
	}

	return ZV, false
}

// IntersectCircle will return the shortest Vec such that moving the Line by that Vec will cause the Line and Circle
// to no longer intesect.  If they do not intersect at all, this function will return a zero-vector.
func (l Line) IntersectCircle(c Circle) Vec {
	// Get the point on the line closest to the center of the circle.
	closest := l.Closest(c.Center)
	cirToClosest := c.Center.To(closest)

	if cirToClosest.Len() >= c.Radius {
		return ZV
	}

	return cirToClosest.Scaled(cirToClosest.Len() - c.Radius)
}

// IntersectRect will return the shortest Vec such that moving the Line by that Vec will cause  the Line and Rect to
// no longer intesect.  If they do not intersect at all, this function will return a zero-vector.
func (l Line) IntersectRect(r Rect) Vec {
	// Check if either end of the line segment are within the rectangle
	if r.Contains(l.A) || r.Contains(l.B) {
		// Use the Rect.Intersect to get minimal return value
		rIntersect := l.Bounds().Intersect(r)
		if rIntersect.H() > rIntersect.W() {
			// Go vertical
			return V(0, rIntersect.H())
		}
		return V(rIntersect.W(), 0)
	}

	// Check if any of the rectangles' edges intersect with this line.
	for _, edge := range r.Edges() {
		if _, ok := l.Intersect(edge); ok {
			// Get the closest points on the line to each corner, where:
			//  - the point is contained by the rectangle
			//  - the point is not the corner itself
			corners := r.Vertices()
			var closest *Vec
			closestCorner := corners[0]
			for _, c := range corners {
				cc := l.Closest(c)
				if closest == nil || (closest.Len() > cc.Len() && r.Contains(cc)) {
					closest = &cc
					closestCorner = c
				}
			}

			return closest.To(closestCorner)
		}
	}

	// No intersect
	return ZV
}

// Len returns the length of the line segment.
func (l Line) Len() float64 {
	return l.A.To(l.B).Len()
}

// Moved will return a line moved by the delta Vec provided.
func (l Line) Moved(delta Vec) Line {
	return Line{
		A: l.A.Add(delta),
		B: l.B.Add(delta),
	}
}

// Rotated will rotate the line around the provided Vec.
func (l Line) Rotated(around Vec, angle float64) Line {
	// Move the line so we can use `Vec.Rotated`
	lineShifted := l.Moved(around.Scaled(-1))

	lineRotated := Line{
		A: lineShifted.A.Rotated(angle),
		B: lineShifted.B.Rotated(angle),
	}

	return lineRotated.Moved(around)
}

// Scaled will return the line scaled around the center point.
func (l Line) Scaled(scale float64) Line {
	return l.ScaledXY(l.Center(), scale)
}

// ScaledXY will return the line scaled around the Vec provided.
func (l Line) ScaledXY(around Vec, scale float64) Line {
	toA := around.To(l.A).Scaled(scale)
	toB := around.To(l.B).Scaled(scale)

	return Line{
		A: around.Add(toA),
		B: around.Add(toB),
	}
}

func (l Line) String() string {
	return fmt.Sprintf("Line(%v, %v)", l.A, l.B)
}
