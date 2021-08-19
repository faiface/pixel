package pixel

import (
	"fmt"
	"math"
)

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

// Anchor is a vector used to define anchors, such as `Center`, `Top`, `TopRight`, etc.
type Anchor Vec

var (
	Center      = Anchor{0.5, 0.5}
	Top         = Anchor{0.5, 0}
	TopRight    = Anchor{0, 0}
	Right       = Anchor{0, 0.5}
	BottomRight = Anchor{0, 1}
	Bottom      = Anchor{0.5, 1}
	BottomLeft  = Anchor{1, 1}
	Left        = Anchor{1, 0.5}
	TopLeft     = Anchor{1, 0}
)

var anchorStrings map[Anchor]string = map[Anchor]string{
	Center:      "center",
	Top:         "top",
	TopRight:    "top-right",
	Right:       "right",
	BottomRight: "bottom-right",
	Bottom:      "bottom",
	BottomLeft:  "bottom-left",
	Left:        "left",
	TopLeft:     "top-left",
}

// String returns the string representation of an anchor.
func (anchor Anchor) String() string {
	return anchorStrings[anchor]
}

var oppositeAnchors map[Anchor]Anchor = map[Anchor]Anchor{
	Center:      Center,
	Top:         Bottom,
	Bottom:      Top,
	Right:       Left,
	Left:        Right,
	TopRight:    BottomLeft,
	BottomLeft:  TopRight,
	BottomRight: TopLeft,
	TopLeft:     BottomRight,
}

// Opposite returns the opposite position of the anchor (ie. Top -> Bottom; BottomLeft -> TopRight, etc.).
func (anchor Anchor) Opposite() Anchor {
	return oppositeAnchors[anchor]
}

// AnchorPos returns the relative position of the given anchor.
func (r Rect) AnchorPos(anchor Anchor) Vec {
	return r.Size().ScaledXY(V(0, 0).Sub(Vec(anchor)))
}

// AlignedTo returns the rect moved by the given anchor.
func (rect Rect) AlignedTo(anchor Anchor) Rect {
	return rect.Moved(rect.AnchorPos(anchor))
}

// Center returns the position of the center of the Rect.
// `rect.Center()` is equivalent to `rect.Anchor(pixel.Anchor.Center)`
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
	return !(s.Max.X <= r.Min.X ||
		s.Min.X >= r.Max.X ||
		s.Max.Y <= r.Min.Y ||
		s.Min.Y >= r.Max.Y)
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
