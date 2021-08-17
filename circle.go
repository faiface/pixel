package pixel

import (
	"fmt"
	"math"
)

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
