package imdraw

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
)

// IMDraw is an immediate-like-mode shape drawer and BasicTarget. IMDraw supports TrianglesPosition,
// TrianglesColor, TrianglesPicture and PictureColor.
//
// IMDraw, other than a regular BasicTarget, is used to draw shapes. To draw shapes, you first need
// to Push some points to IMDraw:
//
//   imd := pixel.NewIMDraw(pic) // use nil pic if you only want to draw primitive shapes
//   imd.Push(pixel.V(100, 100))
//   imd.Push(pixel.V(500, 100))
//
// Once you have Pushed some points, you can use them to draw a shape, such as a line:
//
//   imd.Line(20) // draws a 20 units thick line
//
// Use various methods to change properties of Pushed points:
//
//   imd.Color(pixel.RGBA{R: 1, G: 0, B: 0, A: 1})
//   imd.Push(pixel.V(200, 200))
//   imd.Circle(400, 0)
//
// Here is the list of all available point properties (need to be set before Pushing a point):
//   - Color     - applies to all
//   - Picture   - coordinates, only applies to filled polygons
//   - Intensity - picture intensity, only applies to filled polygons
//   - Precision - curve drawing precision, only applies to circles and ellipses
//   - EndShape  - shape of the end of a line, only applies to lines and outlines
//
// And here's the list of all shapes that can be drawn (all, except for line, can be filled or
// outlined):
//   - Line
//   - Polygon
//   - Circle
//   - Circle arc
//   - Ellipse
//   - Ellipse arc
type IMDraw struct {
	points []point
	opts   point
	matrix pixel.Matrix
	mask   pixel.RGBA

	tri   *pixel.TrianglesData
	batch *pixel.Batch
}

var _ pixel.BasicTarget = (*IMDraw)(nil)

type point struct {
	pos       pixel.Vec
	col       pixel.RGBA
	pic       pixel.Vec
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

// New creates a new empty IMDraw. An optional Picture can be used to draw with a Picture.
//
// If you just want to draw primitive shapes, pass nil as the Picture.
func New(pic pixel.Picture) *IMDraw {
	tri := &pixel.TrianglesData{}
	im := &IMDraw{
		tri:   tri,
		batch: pixel.NewBatch(tri, pic),
	}
	im.SetMatrix(pixel.IM)
	im.SetColorMask(pixel.RGBA{R: 1, G: 1, B: 1, A: 1})
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
func (imd *IMDraw) Draw(t pixel.Target) {
	imd.batch.Draw(t)
}

// Push adds some points to the IM queue. All Pushed points will have the same properties except for
// the position.
func (imd *IMDraw) Push(pts ...pixel.Vec) {
	for _, pt := range pts {
		imd.pushPt(pt, imd.opts)
	}
}

func (imd *IMDraw) pushPt(pos pixel.Vec, pt point) {
	pt.pos = pos
	imd.points = append(imd.points, pt)
}

// Color sets the color of the next Pushed points.
func (imd *IMDraw) Color(color color.Color) {
	imd.opts.col = pixel.ToRGBA(color)
}

// Picture sets the Picture coordinates of the next Pushed points.
func (imd *IMDraw) Picture(pic pixel.Vec) {
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
func (imd *IMDraw) SetMatrix(m pixel.Matrix) {
	imd.matrix = m
	imd.batch.SetMatrix(imd.matrix)
}

// SetColorMask sets a color that all further point's color will be multiplied by.
func (imd *IMDraw) SetColorMask(color color.Color) {
	imd.mask = pixel.ToRGBA(color)
	imd.batch.SetColorMask(imd.mask)
}

// MakeTriangles returns a specialized copy of the provided Triangles that draws onto this IMDraw.
func (imd *IMDraw) MakeTriangles(t pixel.Triangles) pixel.TargetTriangles {
	return imd.batch.MakeTriangles(t)
}

// MakePicture returns a specialized copy of the provided Picture that draws onto this IMDraw.
func (imd *IMDraw) MakePicture(p pixel.Picture) pixel.TargetPicture {
	return imd.batch.MakePicture(p)
}

// Line draws a polyline of the specified thickness between the Pushed points.
func (imd *IMDraw) Line(thickness float64) {
	imd.polyline(thickness, false)
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
		imd.fillEllipseArc(pixel.V(radius, radius), 0, 2*math.Pi)
	} else {
		imd.outlineEllipseArc(pixel.V(radius, radius), 0, 2*math.Pi, thickness, false)
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
		imd.fillEllipseArc(pixel.V(radius, radius), low, high)
	} else {
		imd.outlineEllipseArc(pixel.V(radius, radius), low, high, thickness, true)
	}
}

// Ellipse draws an ellipse of the specified radius in each axis around each Pushed points. If the
// thickness is 0, the ellipse will be filled, otherwise an ellipse outline of the specified
// thickness will be drawn.
func (imd *IMDraw) Ellipse(radius pixel.Vec, thickness float64) {
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
func (imd *IMDraw) EllipseArc(radius pixel.Vec, low, high, thickness float64) {
	if thickness == 0 {
		imd.fillEllipseArc(radius, low, high)
	} else {
		imd.outlineEllipseArc(radius, low, high, thickness, true)
	}
}

func (imd *IMDraw) getAndClearPoints() []point {
	points := imd.points
	imd.points = nil
	return points
}

func (imd *IMDraw) applyMatrixAndMask(off int) {
	for i := range (*imd.tri)[off:] {
		(*imd.tri)[off+i].Position = imd.matrix.Project((*imd.tri)[off+i].Position)
		(*imd.tri)[off+i].Color = imd.mask.Mul((*imd.tri)[off+i].Color)
	}
}

func (imd *IMDraw) fillPolygon() {
	points := imd.getAndClearPoints()

	if len(points) < 3 {
		return
	}

	off := imd.tri.Len()
	imd.tri.SetLen(imd.tri.Len() + 3*(len(points)-2))

	for i, j := 1, off; i+1 < len(points); i, j = i+1, j+3 {
		(*imd.tri)[j+0].Position = points[0].pos
		(*imd.tri)[j+0].Color = points[0].col
		(*imd.tri)[j+0].Picture = points[0].pic
		(*imd.tri)[j+0].Intensity = points[0].in

		(*imd.tri)[j+1].Position = points[i].pos
		(*imd.tri)[j+1].Color = points[i].col
		(*imd.tri)[j+1].Picture = points[i].pic
		(*imd.tri)[j+1].Intensity = points[i].in

		(*imd.tri)[j+2].Position = points[i+1].pos
		(*imd.tri)[j+2].Color = points[i+1].col
		(*imd.tri)[j+2].Picture = points[i+1].pic
		(*imd.tri)[j+2].Intensity = points[i+1].in
	}

	imd.applyMatrixAndMask(off)
	imd.batch.Dirty()
}

func (imd *IMDraw) fillEllipseArc(radius pixel.Vec, low, high float64) {
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

		for i, j := 0.0, off; i < num; i, j = i+1, j+3 {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			a := pt.pos + pixel.V(
				radius.X()*cos,
				radius.Y()*sin,
			)

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			b := pt.pos + pixel.V(
				radius.X()*cos,
				radius.Y()*sin,
			)

			(*imd.tri)[j+0].Position = pt.pos
			(*imd.tri)[j+1].Position = a
			(*imd.tri)[j+2].Position = b
		}

		imd.applyMatrixAndMask(off)
		imd.batch.Dirty()
	}
}

func (imd *IMDraw) outlineEllipseArc(radius pixel.Vec, low, high, thickness float64, doEndShape bool) {
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

		for i, j := 0.0, off; i < num; i, j = i+1, j+6 {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			normalSin, normalCos := pixel.V(sin, cos).ScaledXY(radius).Unit().XY()
			a := pt.pos + pixel.V(
				radius.X()*cos-thickness/2*normalCos,
				radius.Y()*sin-thickness/2*normalSin,
			)
			b := pt.pos + pixel.V(
				radius.X()*cos+thickness/2*normalCos,
				radius.Y()*sin+thickness/2*normalSin,
			)

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			normalSin, normalCos = pixel.V(sin, cos).ScaledXY(radius).Unit().XY()
			c := pt.pos + pixel.V(
				radius.X()*cos-thickness/2*normalCos,
				radius.Y()*sin-thickness/2*normalSin,
			)
			d := pt.pos + pixel.V(
				radius.X()*cos+thickness/2*normalCos,
				radius.Y()*sin+thickness/2*normalSin,
			)

			(*imd.tri)[j+0].Position = a
			(*imd.tri)[j+1].Position = b
			(*imd.tri)[j+2].Position = c
			(*imd.tri)[j+3].Position = c
			(*imd.tri)[j+4].Position = b
			(*imd.tri)[j+5].Position = d
		}

		imd.applyMatrixAndMask(off)
		imd.batch.Dirty()

		if doEndShape {
			lowSin, lowCos := math.Sincos(low)
			lowCenter := pt.pos + pixel.V(
				radius.X()*lowCos,
				radius.Y()*lowSin,
			)
			normalLowSin, normalLowCos := pixel.V(lowSin, lowCos).ScaledXY(radius).Unit().XY()
			normalLow := pixel.V(normalLowCos, normalLowSin).Angle()

			highSin, highCos := math.Sincos(high)
			highCenter := pt.pos + pixel.V(
				radius.X()*highCos,
				radius.Y()*highSin,
			)
			normalHighSin, normalHighCos := pixel.V(highSin, highCos).ScaledXY(radius).Unit().XY()
			normalHigh := pixel.V(normalHighCos, normalHighSin).Angle()

			orientation := 1.0
			if low > high {
				orientation = -1.0
			}

			switch pt.endshape {
			case NoEndShape:
				// nothing
			case SharpEndShape:
				thick := pixel.X(thickness / 2).Rotated(normalLow)
				imd.pushPt(lowCenter+thick, pt)
				imd.pushPt(lowCenter-thick, pt)
				imd.pushPt(lowCenter-thick.Rotated(math.Pi/2*orientation), pt)
				imd.fillPolygon()
				thick = pixel.X(thickness / 2).Rotated(normalHigh)
				imd.pushPt(highCenter+thick, pt)
				imd.pushPt(highCenter-thick, pt)
				imd.pushPt(highCenter+thick.Rotated(math.Pi/2*orientation), pt)
				imd.fillPolygon()
			case RoundEndShape:
				imd.pushPt(lowCenter, pt)
				imd.fillEllipseArc(pixel.V(thickness, thickness)/2, normalLow, normalLow-math.Pi*orientation)
				imd.pushPt(highCenter, pt)
				imd.fillEllipseArc(pixel.V(thickness, thickness)/2, normalHigh, normalHigh+math.Pi*orientation)
			}
		}
	}
}

func (imd *IMDraw) polyline(thickness float64, closed bool) {
	points := imd.getAndClearPoints()

	if len(points) == 0 {
		return
	}
	if len(points) == 1 {
		// one point special case
		points = append(points, points[0])
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
			imd.fillEllipseArc(pixel.V(thickness, thickness)/2, normal.Angle(), normal.Angle()+math.Pi)
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
			imd.fillEllipseArc(pixel.V(thickness, thickness)/2, ijNormal.Angle(), ijNormal.Angle()-math.Pi)
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(pixel.V(thickness, thickness)/2, jkNormal.Angle(), jkNormal.Angle()+math.Pi)
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
			imd.fillEllipseArc(pixel.V(thickness, thickness)/2, normal.Angle(), normal.Angle()-math.Pi)
		}
	}
}
