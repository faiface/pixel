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
		horizontal = X(s.bounds.W() / 2)
		vertical   = Y(s.bounds.H() / 2)
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
// TODO: doc
type IMDraw struct {
	points []point
	opts   point
	matrix Matrix
	mask   NRGBA

	tri   *TrianglesData
	batch *Batch
}

var _ BasicTarget = (*IMDraw)(nil)

type point struct {
	pos       Vec
	col       NRGBA
	pic       Vec
	in        float64
	precision int
	endshape  EndShape
}

// EndShape specifies the shape of an end of a line or a curve.
type EndShape int

const (
	// NoEndShape leaves a line point with no special end shape.
	NoEndShape EndShape = iota

	// SharpEndShape is a sharp triangular end shape.
	SharpEndShape

	// RoundEndShape is a circular end shape.
	RoundEndShape
)

// NewIMDraw creates a new empty IMDraw. An optional Picture can be used to draw with a Picture.
//
// If you just want to draw primitive shapes, pass nil as the Picture.
func NewIMDraw(pic Picture) *IMDraw {
	tri := &TrianglesData{}
	im := &IMDraw{
		tri:   tri,
		batch: NewBatch(tri, pic),
	}
	im.SetMatrix(IM)
	im.SetColorMask(NRGBA{1, 1, 1, 1})
	im.Reset()
	return im
}

// Clear removes all drawn shapes from the IM. This does not remove Pushed points.
func (imd *IMDraw) Clear() {
	imd.tri.SetLen(0)
	imd.batch.Dirty()
}

// Reset restores all point properties to defaults and removes all Pushed points.
//
// This does not affect matrix and color mask set by SetMatrix and SetColorMask.
func (imd *IMDraw) Reset() {
	imd.points = nil
	imd.opts = point{}
	imd.Precision(64)
}

// Draw draws all currently drawn shapes inside the IM onto another Target.
func (imd *IMDraw) Draw(t Target) {
	imd.batch.Draw(t)
}

// Push adds some points to the IM queue. All Pushed points will have the same properties except for
// the position.
func (imd *IMDraw) Push(pts ...Vec) {
	for _, pt := range pts {
		imd.pushPt(pt, imd.opts)
	}
}

func (imd *IMDraw) pushPt(pos Vec, pt point) {
	pt.pos = imd.matrix.Project(pos)
	pt.col = imd.mask.Mul(pt.col)
	imd.points = append(imd.points, pt)
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

// Precision sets the curve/circle drawing precision of the next Pushed points.
//
// It is the number of segments per 360 degrees.
func (imd *IMDraw) Precision(p int) {
	imd.opts.precision = p
}

// EndShape sets the endshape of the next Pushed points.
func (imd *IMDraw) EndShape(es EndShape) {
	imd.opts.endshape = es
}

// SetMatrix sets a Matrix that all further points will be transformed by.
func (imd *IMDraw) SetMatrix(m Matrix) {
	imd.matrix = m
	imd.batch.SetMatrix(imd.matrix)
}

// SetColorMask sets a color that all further point's color will be multiplied by.
func (imd *IMDraw) SetColorMask(color color.Color) {
	imd.mask = NRGBAModel.Convert(color).(NRGBA)
	imd.batch.SetColorMask(imd.mask)
}

// MakeTriangles returns a specialized copy of the provided Triangles that draws onto this IMDraw.
func (imd *IMDraw) MakeTriangles(t Triangles) TargetTriangles {
	return imd.batch.MakeTriangles(t)
}

// MakePicture returns a specialized copy of the provided Picture that draws onto this IMDraw.
func (imd *IMDraw) MakePicture(p Picture) TargetPicture {
	return imd.batch.MakePicture(p)
}

// Polygon draws a polygon from the Pushed points. If the thickness is 0, the convex polygon will be
// filled. Otherwise, an outline of the specified thickness will be drawn. The outline does not have
// to be convex.
//
// Note, that the filled polygon does not have to be strictly convex. The way it's drawn is that a
// triangle is drawn between each two adjacent points and the first Pushed point. You can use this
// property to draw certain kinds of concave polygons.
func (imd *IMDraw) Polygon(thickness float64) {
	if thickness == 0 {
		imd.fillPolygon()
	} else {
		imd.polyline(thickness, true)
	}
}

// Circle draws a circle of the specified radius around each Pushed point. If the thickness is 0,
// the circle will be filled, otherwise a circle outline of the specified thickness will be drawn.
func (imd *IMDraw) Circle(radius, thickness float64) {
	if thickness == 0 {
		imd.fillEllipseArc(V(radius, radius), 0, 2*math.Pi)
	} else {
		imd.outlineEllipseArc(V(radius, radius), 0, 2*math.Pi, thickness, false)
	}
}

// CircleArc draws a circle arc of the specified radius around each Pushed point. If the thickness
// is 0, the arc will be filled, otherwise will be outlined. The arc starts at the low angle and
// continues to the high angle. If low<high, the arc will be drawn counterclockwise. Otherwise it
// will be clockwise. The angles are not normalized by any means.
//
//   imd.CircleArc(40, 0, 8*math.Pi, 0)
//
// This line will fill the whole circle 4 times.
func (imd *IMDraw) CircleArc(radius, low, high, thickness float64) {
	if thickness == 0 {
		imd.fillEllipseArc(V(radius, radius), low, high)
	} else {
		imd.outlineEllipseArc(V(radius, radius), low, high, thickness, true)
	}
}

// Ellipse draws an ellipse of the specified radius in each axis around each Pushed points. If the
// thickness is 0, the ellipse will be filled, otherwise an ellipse outline of the specified
// thickness will be drawn.
func (imd *IMDraw) Ellipse(radius Vec, thickness float64) {
	if thickness == 0 {
		imd.fillEllipseArc(radius, 0, 2*math.Pi)
	} else {
		imd.outlineEllipseArc(radius, 0, 2*math.Pi, thickness, false)
	}
}

// EllipseArc draws an ellipse arc of the specified radius in each axis around each Pushed point. If
// the thickness is 0, the arc will be filled, otherwise will be outlined. The arc starts at the low
// angle and continues to the high angle. If low<high, the arc will be drawn counterclockwise.
// Otherwise it will be clockwise. The angles are not normalized by any means.
//
//   imd.EllipseArc(pixel.V(100, 50), 0, 8*math.Pi, 0)
//
// This line will fill the whole ellipse 4 times.
func (imd *IMDraw) EllipseArc(radius Vec, low, high, thickness float64) {
	if thickness == 0 {
		imd.fillEllipseArc(radius, low, high)
	} else {
		imd.outlineEllipseArc(radius, low, high, thickness, true)
	}
}

// Line draws a polyline of the specified thickness between the Pushed points.
func (imd *IMDraw) Line(thickness float64) {
	imd.polyline(thickness, false)
}

func (imd *IMDraw) getAndClearPoints() []point {
	points := imd.points
	imd.points = nil
	return points
}

func (imd *IMDraw) fillPolygon() {
	points := imd.getAndClearPoints()

	if len(points) < 3 {
		return
	}

	off := imd.tri.Len()
	imd.tri.SetLen(imd.tri.Len() + 3*(len(points)-2))

	for i := 1; i+1 < len(points); i++ {
		(*imd.tri)[off].Position = points[0].pos
		(*imd.tri)[off].Color = points[0].col
		(*imd.tri)[off].Picture = points[0].pic
		(*imd.tri)[off].Intensity = points[0].in

		(*imd.tri)[off+1].Position = points[i].pos
		(*imd.tri)[off+1].Color = points[i].col
		(*imd.tri)[off+1].Picture = points[i].pic
		(*imd.tri)[off+1].Intensity = points[i].in

		(*imd.tri)[off+2].Position = points[i+1].pos
		(*imd.tri)[off+2].Color = points[i+1].col
		(*imd.tri)[off+2].Picture = points[i+1].pic
		(*imd.tri)[off+2].Intensity = points[i+1].in

		off += 3
	}

	imd.batch.Dirty()
}

func (imd *IMDraw) fillEllipseArc(radius Vec, low, high float64) {
	points := imd.getAndClearPoints()

	for _, pt := range points {
		num := math.Ceil(math.Abs(high-low) / (2 * math.Pi) * float64(pt.precision))
		delta := (high - low) / num

		off := imd.tri.Len()
		imd.tri.SetLen(imd.tri.Len() + 3*int(num))

		for i := range (*imd.tri)[off:] {
			(*imd.tri)[off+i].Color = pt.col
			(*imd.tri)[off+i].Picture = 0
			(*imd.tri)[off+i].Intensity = 0
		}

		for i := 0.0; i < num; i++ {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			a := pt.pos + V(
				radius.X()*cos,
				radius.Y()*sin,
			)

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			b := pt.pos + V(
				radius.X()*cos,
				radius.Y()*sin,
			)

			(*imd.tri)[off+0].Position = pt.pos
			(*imd.tri)[off+1].Position = a
			(*imd.tri)[off+2].Position = b

			off += 3
		}

		imd.batch.Dirty()
	}
}

func (imd *IMDraw) outlineEllipseArc(radius Vec, low, high, thickness float64, doEndShape bool) {
	points := imd.getAndClearPoints()

	for _, pt := range points {
		num := math.Ceil(math.Abs(high-low) / (2 * math.Pi) * float64(pt.precision))
		delta := (high - low) / num

		off := imd.tri.Len()
		imd.tri.SetLen(imd.tri.Len() + 6*int(num))

		for i := range (*imd.tri)[off:] {
			(*imd.tri)[off+i].Color = pt.col
			(*imd.tri)[off+i].Picture = 0
			(*imd.tri)[off+i].Intensity = 0
		}

		for i := 0.0; i < num; i++ {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			normalSin, normalCos := V(sin, cos).ScaledXY(radius).Unit().XY()
			a := pt.pos + V(
				radius.X()*cos-thickness/2*normalCos,
				radius.Y()*sin-thickness/2*normalSin,
			)
			b := pt.pos + V(
				radius.X()*cos+thickness/2*normalCos,
				radius.Y()*sin+thickness/2*normalSin,
			)

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			normalSin, normalCos = V(sin, cos).ScaledXY(radius).Unit().XY()
			c := pt.pos + V(
				radius.X()*cos-thickness/2*normalCos,
				radius.Y()*sin-thickness/2*normalSin,
			)
			d := pt.pos + V(
				radius.X()*cos+thickness/2*normalCos,
				radius.Y()*sin+thickness/2*normalSin,
			)

			(*imd.tri)[off+0].Position = a
			(*imd.tri)[off+1].Position = b
			(*imd.tri)[off+2].Position = c
			(*imd.tri)[off+3].Position = c
			(*imd.tri)[off+4].Position = b
			(*imd.tri)[off+5].Position = d

			off += 6
		}

		imd.batch.Dirty()

		if doEndShape {
			lowSin, lowCos := math.Sincos(low)
			lowCenter := pt.pos + V(
				radius.X()*lowCos,
				radius.Y()*lowSin,
			)
			normalLowSin, normalLowCos := V(lowSin, lowCos).ScaledXY(radius).Unit().XY()
			normalLow := V(normalLowCos, normalLowSin).Angle()

			highSin, highCos := math.Sincos(high)
			highCenter := pt.pos + V(
				radius.X()*highCos,
				radius.Y()*highSin,
			)
			normalHighSin, normalHighCos := V(highSin, highCos).ScaledXY(radius).Unit().XY()
			normalHigh := V(normalHighCos, normalHighSin).Angle()

			orientation := 1.0
			if low > high {
				orientation = -1.0
			}

			switch pt.endshape {
			case NoEndShape:
				// nothing
			case SharpEndShape:
				thick := X(thickness / 2).Rotated(normalLow)
				imd.pushPt(lowCenter+thick, pt)
				imd.pushPt(lowCenter-thick, pt)
				imd.pushPt(lowCenter-thick.Rotated(math.Pi/2*orientation), pt)
				imd.fillPolygon()
				thick = X(thickness / 2).Rotated(normalHigh)
				imd.pushPt(highCenter+thick, pt)
				imd.pushPt(highCenter-thick, pt)
				imd.pushPt(highCenter+thick.Rotated(math.Pi/2*orientation), pt)
				imd.fillPolygon()
			case RoundEndShape:
				imd.pushPt(lowCenter, pt)
				imd.fillEllipseArc(V(thickness, thickness)/2, normalLow, normalLow-math.Pi*orientation)
				imd.pushPt(highCenter, pt)
				imd.fillEllipseArc(V(thickness, thickness)/2, normalHigh, normalHigh+math.Pi*orientation)
			}
		}
	}
}

func (imd *IMDraw) polyline(thickness float64, closed bool) {
	points := imd.getAndClearPoints()

	// filter identical adjacent points
	filtered := points[:0]
	for i := 0; i < len(points); i++ {
		if closed || i+1 < len(points) {
			j := (i + 1) % len(points)
			if points[i].pos != points[j].pos {
				filtered = append(filtered, points[i])
			}
		}
	}
	points = filtered

	if len(points) < 2 {
		return
	}

	// first point
	j, i := 0, 1
	normal := (points[i].pos - points[j].pos).Rotated(math.Pi / 2).Unit().Scaled(thickness / 2)

	if !closed {
		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos+normal, points[j])
			imd.pushPt(points[j].pos-normal, points[j])
			imd.pushPt(points[j].pos+normal.Rotated(math.Pi/2), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(V(thickness, thickness)/2, normal.Angle(), normal.Angle()+math.Pi)
		}
	}

	imd.pushPt(points[j].pos+normal, points[j])
	imd.pushPt(points[j].pos-normal, points[j])

	// middle points
	for i := 0; i < len(points); i++ {
		j, k := i+1, i+2

		closing := false
		if j >= len(points) {
			if !closed {
				break
			}
			j %= len(points)
			closing = true
		}
		if k >= len(points) {
			k %= len(points)
		}

		ijNormal := (points[j].pos - points[i].pos).Rotated(math.Pi / 2).Unit().Scaled(thickness / 2)
		jkNormal := (points[k].pos - points[j].pos).Rotated(math.Pi / 2).Unit().Scaled(thickness / 2)

		orientation := 1.0
		if ijNormal.Cross(jkNormal) > 0 {
			orientation = -1.0
		}

		imd.pushPt(points[j].pos-ijNormal, points[j])
		imd.pushPt(points[j].pos+ijNormal, points[j])
		imd.fillPolygon()

		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.pushPt(points[j].pos+ijNormal.Scaled(orientation), points[j])
			imd.pushPt(points[j].pos+jkNormal.Scaled(orientation), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(V(thickness, thickness)/2, ijNormal.Angle(), ijNormal.Angle()-math.Pi)
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(V(thickness, thickness)/2, jkNormal.Angle(), jkNormal.Angle()+math.Pi)
		}

		if !closing {
			imd.pushPt(points[j].pos+jkNormal, points[j])
			imd.pushPt(points[j].pos-jkNormal, points[j])
		}
	}

	// last point
	i, j = len(points)-2, len(points)-1
	normal = (points[j].pos - points[i].pos).Rotated(math.Pi / 2).Unit().Scaled(thickness / 2)

	imd.pushPt(points[j].pos-normal, points[j])
	imd.pushPt(points[j].pos+normal, points[j])
	imd.fillPolygon()

	if !closed {
		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos+normal, points[j])
			imd.pushPt(points[j].pos-normal, points[j])
			imd.pushPt(points[j].pos+normal.Rotated(-math.Pi/2), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(V(thickness, thickness)/2, normal.Angle(), normal.Angle()-math.Pi)
		}
	}
}
