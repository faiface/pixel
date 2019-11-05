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
			closest := ZV
			closestCorner := corners[0]
			for _, c := range corners {
				cc := l.Closest(c)
				if closest == ZV || (closest.Len() > cc.Len() && r.Contains(cc)) {
					closest = cc
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

// Rect is a 2D rectangle aligned with the axes of the coordinate system. It is defined by two
// points, Min and Max.
//
// The invariant should hold, that Max's components are greater or equal than Min's components
// respectively.
type Rect struct {
	Min, Max Vec
}

// ZR is a zero rectangle.
var ZR = Rect{Min: ZV, Max: ZV}

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

// Edges will return the four lines which make up the edges of the rectangle.
func (r Rect) Edges() [4]Line {
	corners := r.Vertices()

	return [4]Line{
		{A: corners[0], B: corners[1]},
		{A: corners[1], B: corners[2]},
		{A: corners[2], B: corners[3]},
		{A: corners[3], B: corners[0]},
	}
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
// If r and s don't overlap, this function returns a zero-rectangle.
func (r Rect) Intersect(s Rect) Rect {
	t := R(
		math.Max(r.Min.X, s.Min.X),
		math.Max(r.Min.Y, s.Min.Y),
		math.Min(r.Max.X, s.Max.X),
		math.Min(r.Max.Y, s.Max.Y),
	)
	if t.Min.X >= t.Max.X || t.Min.Y >= t.Max.Y {
		return ZR
	}
	return t
}

// Intersects returns whether or not the given Rect intersects at any point with this Rect.
//
// This function is overall about 5x faster than Intersect, so it is better
// to use if you have no need for the returned Rect from Intersect.
func (r Rect) Intersects(s Rect) bool {
	return !(s.Max.X < r.Min.X ||
		s.Min.X > r.Max.X ||
		s.Max.Y < r.Min.Y ||
		s.Min.Y > r.Max.Y)
}

// IntersectCircle returns a minimal required Vector, such that moving the rect by that vector would stop the Circle
// and the Rect intersecting.  This function returns a zero-vector if the Circle and Rect do not overlap, and if only
// the perimeters touch.
//
// This function will return a non-zero vector if:
//  - The Rect contains the Circle, partially or fully
//  - The Circle contains the Rect, partially of fully
func (r Rect) IntersectCircle(c Circle) Vec {
	return c.IntersectRect(r).Scaled(-1)
}

// IntersectLine will return the shortest Vec such that if the Rect is moved by the Vec returned, the Line and Rect no
// longer intersect.
func (r Rect) IntersectLine(l Line) Vec {
	return l.IntersectRect(r).Scaled(-1)
}

// IntersectionPoints returns all the points where the Rect intersects with the line provided.  This can be zero, one or
// two points, depending on the location of the shapes.  The points of intersection will be returned in order of
// closest-to-l.A to closest-to-l.B.
func (r Rect) IntersectionPoints(l Line) []Vec {
	// Use map keys to ensure unique points
	pointMap := make(map[Vec]struct{})

	for _, edge := range r.Edges() {
		if intersect, ok := l.Intersect(edge); ok {
			pointMap[intersect] = struct{}{}
		}
	}

	points := make([]Vec, 0, len(pointMap))
	for point := range pointMap {
		points = append(points, point)
	}

	// Order the points
	if len(points) == 2 {
		if points[1].To(l.A).Len() < points[0].To(l.A).Len() {
			return []Vec{points[1], points[0]}
		}
	}

	return points
}

// Vertices returns a slice of the four corners which make up the rectangle.
func (r Rect) Vertices() [4]Vec {
	return [4]Vec{
		r.Min,
		V(r.Min.X, r.Max.Y),
		r.Max,
		V(r.Max.X, r.Min.Y),
	}
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
	return math.Pi * math.Pow(c.Radius, 2)
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

// Formula returns the values of h and k, for the equation of the circle: (x-h)^2 + (y-k)^2 = r^2
// where r is the radius of the circle.
func (c Circle) Formula() (h, k float64) {
	return c.Center.X, c.Center.Y
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

// IntersectLine will return the shortest Vec such that if the Circle is moved by the Vec returned, the Line and Rect no
// longer intersect.
func (c Circle) IntersectLine(l Line) Vec {
	return l.IntersectCircle(c).Scaled(-1)
}

// IntersectRect returns a minimal required Vector, such that moving the circle by that vector would stop the Circle
// and the Rect intersecting.  This function returns a zero-vector if the Circle and Rect do not overlap, and if only
// the perimeters touch.
//
// This function will return a non-zero vector if:
//  - The Rect contains the Circle, partially or fully
//  - The Circle contains the Rect, partially of fully
func (c Circle) IntersectRect(r Rect) Vec {
	// Checks if the c.Center is not in the diagonal quadrants of the rectangle
	if (r.Min.X <= c.Center.X && c.Center.X <= r.Max.X) || (r.Min.Y <= c.Center.Y && c.Center.Y <= r.Max.Y) {
		// 'grow' the Rect by c.Radius in each orthagonal
		grown := Rect{Min: r.Min.Sub(V(c.Radius, c.Radius)), Max: r.Max.Add(V(c.Radius, c.Radius))}
		if !grown.Contains(c.Center) {
			// c.Center not close enough to overlap, return zero-vector
			return ZV
		}

		// Get minimum distance to travel out of Rect
		rToC := r.Center().To(c.Center)
		h := c.Radius - math.Abs(rToC.X) + (r.W() / 2)
		v := c.Radius - math.Abs(rToC.Y) + (r.H() / 2)

		if rToC.X < 0 {
			h = -h
		}
		if rToC.Y < 0 {
			v = -v
		}

		// No intersect
		if h == 0 && v == 0 {
			return ZV
		}

		if math.Abs(h) > math.Abs(v) {
			// Vertical distance shorter
			return V(0, v)
		}
		return V(h, 0)
	} else {
		// The center is in the diagonal quadrants

		// Helper points to make code below easy to read.
		rectTopLeft := V(r.Min.X, r.Max.Y)
		rectBottomRight := V(r.Max.X, r.Min.Y)

		// Check for overlap.
		if !(c.Contains(r.Min) || c.Contains(r.Max) || c.Contains(rectTopLeft) || c.Contains(rectBottomRight)) {
			// No overlap.
			return ZV
		}

		var centerToCorner Vec
		if c.Center.To(r.Min).Len() <= c.Radius {
			// Closest to bottom-left
			centerToCorner = c.Center.To(r.Min)
		}
		if c.Center.To(r.Max).Len() <= c.Radius {
			// Closest to top-right
			centerToCorner = c.Center.To(r.Max)
		}
		if c.Center.To(rectTopLeft).Len() <= c.Radius {
			// Closest to top-left
			centerToCorner = c.Center.To(rectTopLeft)
		}
		if c.Center.To(rectBottomRight).Len() <= c.Radius {
			// Closest to bottom-right
			centerToCorner = c.Center.To(rectBottomRight)
		}

		cornerToCircumferenceLen := c.Radius - centerToCorner.Len()

		return centerToCorner.Unit().Scaled(cornerToCircumferenceLen)
	}
}

// IntersectionPoints returns all the points where the Circle intersects with the line provided.  This can be zero, one or
// two points, depending on the location of the shapes.  The points of intersection will be returned in order of
// closest-to-l.A to closest-to-l.B.
func (c Circle) IntersectionPoints(l Line) []Vec {
	cContainsA := c.Contains(l.A)
	cContainsB := c.Contains(l.B)

	// Special case for both endpoint being contained within the circle
	if cContainsA && cContainsB {
		return []Vec{}
	}

	// Get closest point on the line to this circles' center
	closestToCenter := l.Closest(c.Center)

	// If the distance to the closest point is greater than the radius, there are no points of intersection
	if closestToCenter.To(c.Center).Len() > c.Radius {
		return []Vec{}
	}

	// If the distance to the closest point is equal to the radius, the line is tangent and the closest point is the
	// point at which it touches the circle.
	if closestToCenter.To(c.Center).Len() == c.Radius {
		return []Vec{closestToCenter}
	}

	// Special case for endpoint being on the circles' center
	if c.Center == l.A || c.Center == l.B {
		otherEnd := l.B
		if c.Center == l.B {
			otherEnd = l.A
		}
		intersect := c.Center.Add(c.Center.To(otherEnd).Unit().Scaled(c.Radius))
		return []Vec{intersect}
	}

	// This means the distance to the closest point is less than the radius, so there is at least one intersection,
	// possibly two.

	// If one of the end points exists within the circle, there is only one intersection
	if cContainsA || cContainsB {
		containedPoint := l.A
		otherEnd := l.B
		if cContainsB {
			containedPoint = l.B
			otherEnd = l.A
		}

		// Use trigonometry to get the length of the line between the contained point and the intersection point.
		// The following is used to describe the triangle formed:
		//  - a is the side between contained point and circle center
		//  - b is the side between the center and the intersection point (radius)
		//  - c is the side between the contained point and the intersection point
		// The captials of these letters are used as the angles opposite the respective sides.
		// a and b are known
		a := containedPoint.To(c.Center).Len()
		b := c.Radius
		// B can be calculated by subtracting the angle of b (to the x-axis) from the angle of c (to the x-axis)
		B := containedPoint.To(c.Center).Angle() - containedPoint.To(otherEnd).Angle()
		// Using the Sin rule we can get A
		A := math.Asin((a * math.Sin(B)) / b)
		// Using the rule that there are 180 degrees (or Pi radians) in a triangle, we can now get C
		C := math.Pi - A + B
		// If C is zero, the line segment is in-line with the center-intersect line.
		var c float64
		if C == 0 {
			c = b - a
		} else {
			// Using the Sine rule again, we can now get c
			c = (a * math.Sin(C)) / math.Sin(A)
		}
		// Travelling from the contained point to the other end by length of a will provide the intersection point.
		return []Vec{
			containedPoint.Add(containedPoint.To(otherEnd).Unit().Scaled(c)),
		}
	}

	// Otherwise the endpoints exist outside of the circle, and the line segment intersects in two locations.
	// The vector formed by going from the closest point to the center of the circle will be perpendicular to the line;
	// this forms a right-angled triangle with the intersection points, with the radius as the hypotenuse.
	// Calculate the other triangles' sides' length.
	a := math.Sqrt(math.Pow(c.Radius, 2) - math.Pow(closestToCenter.To(c.Center).Len(), 2))

	// Travelling in both directions from the closest point by length of a will provide the two intersection points.
	first := closestToCenter.Add(closestToCenter.To(l.A).Unit().Scaled(a))
	second := closestToCenter.Add(closestToCenter.To(l.B).Unit().Scaled(a))

	if first.To(l.A).Len() < second.To(l.A).Len() {
		return []Vec{first, second}
	}
	return []Vec{second, first}
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
