package pixel

import (
	"image/color"
	"math"
)

// Sprite is a drawable Picture. It's always anchored by the center of it's Picture.
type Sprite struct {
	tri    *TrianglesData
	bounds Rect
	d      Drawer
}

// NewSprite creates a Sprite from the supplied Picture.
func NewSprite(pic Picture) *Sprite {
	tri := MakeTrianglesData(6)
	s := &Sprite{
		tri: tri,
		d:   Drawer{Triangles: tri},
	}
	s.SetPicture(pic)
	return s
}

// SetPicture changes the Sprite's Picture. The new Picture may have a different size, everything
// works.
func (s *Sprite) SetPicture(pic Picture) {
	s.d.Picture = pic

	if s.bounds == pic.Bounds() {
		return
	}
	s.bounds = pic.Bounds()

	var (
		center     = s.bounds.Center()
		horizontal = V(s.bounds.W()/2, 0)
		vertical   = V(0, s.bounds.H()/2)
	)

	(*s.tri)[0].Position = -horizontal - vertical
	(*s.tri)[1].Position = +horizontal - vertical
	(*s.tri)[2].Position = +horizontal + vertical
	(*s.tri)[3].Position = -horizontal - vertical
	(*s.tri)[4].Position = +horizontal + vertical
	(*s.tri)[5].Position = -horizontal + vertical

	for i := range *s.tri {
		(*s.tri)[i].Color = NRGBA{1, 1, 1, 1}
		(*s.tri)[i].Picture = center + (*s.tri)[i].Position
		(*s.tri)[i].Intensity = 1
	}

	s.d.Dirty()
}

// Picture returns the current Sprite's Picture.
func (s *Sprite) Picture() Picture {
	return s.d.Picture
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	s.d.Draw(t)
}

// IMDraw is an immediate-like-mode shape drawer.
//
// TODO: mode doc
type IMDraw struct {
	points []point
	opts   point
	matrix Matrix
	mask   NRGBA
	tri    *TrianglesData
	d      Drawer
	tmp    []Vec
}

type point struct {
	pos       Vec
	col       NRGBA
	pic       Vec
	in        float64
	width     float64
	precision int
	endshape  EndShape
}

// EndShape specifies the shape of an end of a line or a curve.
type EndShape int

const (
	// RoundEndShape is a circular end shape.
	RoundEndShape EndShape = iota

	// SharpEndShape is a square end shape.
	SharpEndShape
)

// NewIMDraw creates a new empty IMDraw. An optional Picture can be used to draw with a Picture.
//
// If you just want to draw primitive shapes, pass nil as the Picture.
func NewIMDraw(pic Picture) *IMDraw {
	tri := &TrianglesData{}
	im := &IMDraw{
		tri: tri,
		d:   Drawer{Triangles: tri, Picture: pic},
	}
	im.Precision(64)
	im.SetMatrix(IM)
	im.SetColorMask(NRGBA{1, 1, 1, 1})
	return im
}

// Clear removes all drawn shapes from the IM. This does not remove Pushed points.
func (imd *IMDraw) Clear() {
	imd.tri.SetLen(0)
	imd.d.Dirty()
}

// Draw draws all currently drawn shapes inside the IM onto another Target.
func (imd *IMDraw) Draw(t Target) {
	imd.d.Draw(t)
}

// Push adds some points to the IM queue. All Pushed points will have the same properties except for
// the position.
func (imd *IMDraw) Push(pts ...Vec) {
	point := imd.opts
	for _, pt := range pts {
		point.pos = imd.matrix.Project(pt)
		point.col = imd.mask.Mul(imd.opts.col)
		imd.points = append(imd.points, point)
	}
}

// Color sets the color of the next Pushed points.
func (imd *IMDraw) Color(color color.Color) {
	imd.opts.col = NRGBAModel.Convert(color).(NRGBA)
}

// Picture sets the Picture coordinates of the next Pushed points.
func (imd *IMDraw) Picture(pic Vec) {
	imd.opts.pic = pic
}

// Intensity sets the picture Intensity of the next Pushed points.
func (imd *IMDraw) Intensity(in float64) {
	imd.opts.in = in
}

// Width sets the with property of the next Pushed points.
//
// Note that this property does not apply to filled shapes.
func (imd *IMDraw) Width(w float64) {
	imd.opts.width = w
}

// Precision sets the curve/circle drawing precision of the next Pushed points.
//
// It is the number of segments per 360 degrees.
func (imd *IMDraw) Precision(p int) {
	imd.opts.precision = p
	if p+1 > len(imd.tmp) {
		imd.tmp = append(imd.tmp, make([]Vec, p+1-len(imd.tmp))...)
	}
	if p+1 < len(imd.tmp) {
		imd.tmp = imd.tmp[:p+1]
	}
}

// EndShape sets the endshape of the next Pushed points.
func (imd *IMDraw) EndShape(es EndShape) {
	imd.opts.endshape = es
}

// SetMatrix sets a Matrix that all further points will be transformed by.
func (imd *IMDraw) SetMatrix(m Matrix) {
	imd.matrix = m
}

// SetColorMask sets a color that all futher point's color will be multiplied by.
func (imd *IMDraw) SetColorMask(color color.Color) {
	imd.mask = NRGBAModel.Convert(color).(NRGBA)
}

// FillConvexPolygon takes all points Pushed into the IM's queue and fills the convex polygon formed
// by them.
//
// The polygon does not need to be exactly convex. The way it's drawn is that for each two adjacent
// points, a triangle is constructed from those two points and the first Pushed point. You can use
// this property to draw specific concave graphs.
func (imd *IMDraw) FillConvexPolygon() {
	points := imd.points
	imd.points = nil

	if len(points) < 3 {
		return
	}

	i := imd.tri.Len()
	imd.tri.SetLen(imd.tri.Len() + 3*(len(points)-2))

	for j := 1; j+1 < len(points); j++ {
		(*imd.tri)[i].Position = points[0].pos
		(*imd.tri)[i].Color = points[0].col
		(*imd.tri)[i].Picture = points[0].pic
		(*imd.tri)[i].Intensity = points[0].in

		(*imd.tri)[i+1].Position = points[j].pos
		(*imd.tri)[i+1].Color = points[j].col
		(*imd.tri)[i+1].Picture = points[j].pic
		(*imd.tri)[i+1].Intensity = points[j].in

		(*imd.tri)[i+2].Position = points[j+1].pos
		(*imd.tri)[i+2].Color = points[j+1].col
		(*imd.tri)[i+2].Picture = points[j+1].pic
		(*imd.tri)[i+2].Intensity = points[j+1].in

		i += 3
	}

	imd.d.Dirty()
}

// FillCircle draws a filled circle around each point in the IM's queue.
func (imd *IMDraw) FillCircle(radius float64) {
	imd.FillEllipseArc(V(radius, radius), 0, 2*math.Pi)
}

// FillCircleArc draws a filled circle arc around each point in the IM's queue.
func (imd *IMDraw) FillCircleArc(radius, low, high float64) {
	imd.FillEllipseArc(V(radius, radius), low, high)
}

// FillEllipse draws a filled ellipse around each point in the IM's queue.
func (imd *IMDraw) FillEllipse(radius Vec) {
	imd.FillEllipseArc(radius, 0, 2*math.Pi)
}

// FillEllipseArc draws a filled ellipse arc around each point in the IM's queue. Low and high
// angles are in radians.
func (imd *IMDraw) FillEllipseArc(radius Vec, low, high float64) {
	points := imd.points
	imd.points = nil

	// normalize high
	if math.Abs(high-low) > 2*math.Pi {
		high = low + math.Mod(high-low, 2*math.Pi)
	}

	for _, pt := range points {
		imd.Push(pt.pos) // center

		num := math.Ceil(math.Abs(high-low) / (2 * math.Pi) * float64(pt.precision))
		delta := (high - low) / num
		for i := range imd.tmp[:int(num)+1] {
			angle := low + float64(i)*delta
			sin, cos := math.Sincos(angle)
			imd.tmp[i] = pt.pos + V(
				radius.X()*cos,
				radius.Y()*sin,
			)
		}

		imd.Push(imd.tmp[:int(num)+1]...)
		imd.FillConvexPolygon()
	}
}
