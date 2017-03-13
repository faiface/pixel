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

// IM is an immediate-like-mode shape drawer.
//
// TODO: mode doc
type IM struct {
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

// NewIM creates a new empty IM. An optional Picture can be used to draw with a Picture.
//
// If you just want to draw primitive shapes, pass nil as the Picture.
func NewIM(pic Picture) *IM {
	tri := &TrianglesData{}
	im := &IM{
		tri: tri,
		d:   Drawer{Triangles: tri, Picture: pic},
	}
	im.Precision(64)
	im.SetMatrix(ZM)
	im.SetColorMask(NRGBA{1, 1, 1, 1})
	return im
}

// Clear removes all drawn shapes from the IM. This does not remove Pushed points.
func (im *IM) Clear() {
	im.tri.SetLen(0)
	im.d.Dirty()
}

// Draw draws all currently drawn shapes inside the IM onto another Target.
func (im *IM) Draw(t Target) {
	im.d.Draw(t)
}

// Push adds some points to the IM queue. All Pushed points will have the same properties except for
// the position.
func (im *IM) Push(pts ...Vec) {
	point := im.opts
	for _, pt := range pts {
		point.pos = im.matrix.Project(pt)
		point.col = im.mask.Mul(im.opts.col)
		im.points = append(im.points, point)
	}
}

// Color sets the color of the next Pushed points.
func (im *IM) Color(color color.Color) {
	im.opts.col = NRGBAModel.Convert(color).(NRGBA)
}

// Picture sets the Picture coordinates of the next Pushed points.
func (im *IM) Picture(pic Vec) {
	im.opts.pic = pic
}

// Intensity sets the picture Intensity of the next Pushed points.
func (im *IM) Intensity(in float64) {
	im.opts.in = in
}

// Width sets the with property of the next Pushed points.
//
// Note that this property does not apply to filled shapes.
func (im *IM) Width(w float64) {
	im.opts.width = w
}

// Precision sets the curve/circle drawing precision of the next Pushed points.
//
// It is the number of segments per 360 degrees.
func (im *IM) Precision(p int) {
	im.opts.precision = p
	if p+1 > len(im.tmp) {
		im.tmp = append(im.tmp, make([]Vec, p+1-len(im.tmp))...)
	}
	if p+1 < len(im.tmp) {
		im.tmp = im.tmp[:p+1]
	}
}

// EndShape sets the endshape of the next Pushed points.
func (im *IM) EndShape(es EndShape) {
	im.opts.endshape = es
}

// SetMatrix sets a Matrix that all further points will be transformed by.
func (im *IM) SetMatrix(m Matrix) {
	im.matrix = m
}

// SetColorMask sets a color that all futher point's color will be multiplied by.
func (im *IM) SetColorMask(color color.Color) {
	im.mask = NRGBAModel.Convert(color).(NRGBA)
}

// FillConvexPolygon takes all points Pushed into the IM's queue and fills the convex polygon formed
// by them.
func (im *IM) FillConvexPolygon() {
	points := im.points
	im.points = nil

	if len(points) < 3 {
		return
	}

	i := im.tri.Len()
	im.tri.SetLen(im.tri.Len() + 3*(len(points)-2))

	for j := 1; j+1 < len(points); j++ {
		(*im.tri)[i].Position = points[0].pos
		(*im.tri)[i].Color = points[0].col
		(*im.tri)[i].Picture = points[0].pic
		(*im.tri)[i].Intensity = points[0].in

		(*im.tri)[i+1].Position = points[j].pos
		(*im.tri)[i+1].Color = points[j].col
		(*im.tri)[i+1].Picture = points[j].pic
		(*im.tri)[i+1].Intensity = points[j].in

		(*im.tri)[i+2].Position = points[j+1].pos
		(*im.tri)[i+2].Color = points[j+1].col
		(*im.tri)[i+2].Picture = points[j+1].pic
		(*im.tri)[i+2].Intensity = points[j+1].in

		i += 3
	}

	im.d.Dirty()
}

// FillCircle draws a filled circle around each point in the IM's queue.
func (im *IM) FillCircle(radius float64) {
	im.FillEllipseArc(V(radius, radius), 0, 2*math.Pi)
}

// FillCircleArc draws a filled circle arc around each point in the IM's queue.
func (im *IM) FillCircleArc(radius, low, high float64) {
	im.FillEllipseArc(V(radius, radius), low, high)
}

// FillEllipse draws a filled ellipse around each point in the IM's queue.
func (im *IM) FillEllipse(radius Vec) {
	im.FillEllipseArc(radius, 0, 2*math.Pi)
}

// FillEllipseArc draws a filled ellipse arc around each point in the IM's queue. Low and high
// angles are in radians.
func (im *IM) FillEllipseArc(radius Vec, low, high float64) {
	points := im.points
	im.points = nil

	// normalize high
	if math.Abs(high-low) > 2*math.Pi {
		high = low + math.Mod(high-low, 2*math.Pi)
	}

	for _, pt := range points {
		im.Push(pt.pos) // center

		num := math.Ceil(math.Abs(high-low) / (2 * math.Pi) * float64(pt.precision))
		delta := (high - low) / num
		for i := range im.tmp[:int(num)+1] {
			angle := low + float64(i)*delta
			sin, cos := math.Sincos(angle)
			im.tmp[i] = pt.pos + V(
				radius.X()*cos,
				radius.Y()*sin,
			)
		}

		im.Push(im.tmp[:int(num)+1]...)
		im.FillConvexPolygon()
	}
}
