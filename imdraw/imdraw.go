// Package imdraw implements a basic primitive geometry shape and pictured polygon drawing for Pixel
// with a nice immediate-mode-like API.
package imdraw

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
)

// IMDraw is an immediate-mode-like shape drawer and BasicTarget. IMDraw supports TrianglesPosition,
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
// Set exported fields to change properties of Pushed points:
//
//   imd.Color = pixel.RGB(1, 0, 0)
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
	Color     color.Color
	Picture   pixel.Vec
	Intensity float64
	Precision int
	EndShape  EndShape

	points []point
	pool   [][]point
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
	im.SetColorMask(pixel.Alpha(1))
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
	imd.points = imd.points[:0]
	imd.Color = pixel.Alpha(1)
	imd.Picture = pixel.ZV
	imd.Intensity = 0
	imd.Precision = 64
	imd.EndShape = NoEndShape
}

// Draw draws all currently drawn shapes inside the IM onto another Target.
//
// Note, that IMDraw's matrix and color mask have no effect here.
func (imd *IMDraw) Draw(t pixel.Target) {
	imd.batch.Draw(t)
}

// Push adds some points to the IM queue. All Pushed points will have the same properties except for
// the position.
func (imd *IMDraw) Push(pts ...pixel.Vec) {
	//Assert that Color is of type pixel.RGBA,
	if _, ok := imd.Color.(pixel.RGBA); !ok {
		//otherwise cast it
		imd.Color = pixel.ToRGBA(imd.Color)
	}
	opts := point{
		col:       imd.Color.(pixel.RGBA),
		pic:       imd.Picture,
		in:        imd.Intensity,
		precision: imd.Precision,
		endshape:  imd.EndShape,
	}
	for _, pt := range pts {
		imd.pushPt(pt, opts)
	}
}

func (imd *IMDraw) pushPt(pos pixel.Vec, pt point) {
	pt.pos = pos
	imd.points = append(imd.points, pt)
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

// Rectangle draws a rectangle between each two subsequent Pushed points. Drawing a rectangle
// between two points means drawing a rectangle with sides parallel to the axes of the coordinate
// system, where the two points specify it's two opposite corners.
//
// If the thickness is 0, rectangles will be filled, otherwise will be outlined with the given
// thickness.
func (imd *IMDraw) Rectangle(thickness float64) {
	if thickness == 0 {
		imd.fillRectangle()
	} else {
		imd.outlineRectangle(thickness)
	}
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
	// use one of the existing pools so we don't reallocate as often
	if len(imd.pool) > 0 {
		pos := len(imd.pool) - 1
		imd.points = imd.pool[pos][:0]
		imd.pool = imd.pool[:pos]
	} else {
		imd.points = nil
	}
	return points
}

func (imd *IMDraw) restorePoints(points []point) {
	imd.pool = append(imd.pool, imd.points)
	imd.points = points[:0]
}

func (imd *IMDraw) applyMatrixAndMask(off int) {
	for i := range (*imd.tri)[off:] {
		(*imd.tri)[off+i].Position = imd.matrix.Project((*imd.tri)[off+i].Position)
		(*imd.tri)[off+i].Color = imd.mask.Mul((*imd.tri)[off+i].Color)
	}
}

func (imd *IMDraw) fillRectangle() {
	points := imd.getAndClearPoints()

	if len(points) < 2 {
		imd.restorePoints(points)
		return
	}

	off := imd.tri.Len()
	imd.tri.SetLen(imd.tri.Len() + 6*(len(points)-1))

	for i, j := 0, off; i+1 < len(points); i, j = i+1, j+6 {
		a, b := points[i], points[i+1]
		c := point{
			pos: pixel.V(a.pos.X, b.pos.Y),
			col: a.col.Add(b.col).Mul(pixel.Alpha(0.5)),
			pic: pixel.V(a.pic.X, b.pic.Y),
			in:  (a.in + b.in) / 2,
		}
		d := point{
			pos: pixel.V(b.pos.X, a.pos.Y),
			col: a.col.Add(b.col).Mul(pixel.Alpha(0.5)),
			pic: pixel.V(b.pic.X, a.pic.Y),
			in:  (a.in + b.in) / 2,
		}

		for k, p := range [...]point{a, b, c, a, b, d} {
			(*imd.tri)[j+k].Position = p.pos
			(*imd.tri)[j+k].Color = p.col
			(*imd.tri)[j+k].Picture = p.pic
			(*imd.tri)[j+k].Intensity = p.in
		}
	}

	imd.applyMatrixAndMask(off)
	imd.batch.Dirty()

	imd.restorePoints(points)
}

func (imd *IMDraw) outlineRectangle(thickness float64) {
	points := imd.getAndClearPoints()

	if len(points) < 2 {
		imd.restorePoints(points)
		return
	}

	for i := 0; i+1 < len(points); i++ {
		a, b := points[i], points[i+1]
		mid := a
		mid.col = a.col.Add(b.col).Mul(pixel.Alpha(0.5))
		mid.in = (a.in + b.in) / 2

		imd.pushPt(a.pos, a)
		imd.pushPt(pixel.V(a.pos.X, b.pos.Y), mid)
		imd.pushPt(b.pos, b)
		imd.pushPt(pixel.V(b.pos.X, a.pos.Y), mid)
		imd.polyline(thickness, true)
	}

	imd.restorePoints(points)
}

func (imd *IMDraw) fillPolygon() {
	points := imd.getAndClearPoints()

	if len(points) < 3 {
		imd.restorePoints(points)
		return
	}

	off := imd.tri.Len()
	imd.tri.SetLen(imd.tri.Len() + 3*(len(points)-2))

	for i, j := 1, off; i+1 < len(points); i, j = i+1, j+3 {
		for k, p := range [...]int{0, i, i + 1} {
			tri := &(*imd.tri)[j+k]
			tri.Position = points[p].pos
			tri.Color = points[p].col
			tri.Picture = points[p].pic
			tri.Intensity = points[p].in
		}
	}

	imd.applyMatrixAndMask(off)
	imd.batch.Dirty()

	imd.restorePoints(points)
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
			(*imd.tri)[off+i].Picture = pixel.ZV
			(*imd.tri)[off+i].Intensity = 0
		}

		for i, j := 0.0, off; i < num; i, j = i+1, j+3 {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			a := pt.pos.Add(pixel.V(
				radius.X*cos,
				radius.Y*sin,
			))

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			b := pt.pos.Add(pixel.V(
				radius.X*cos,
				radius.Y*sin,
			))

			(*imd.tri)[j+0].Position = pt.pos
			(*imd.tri)[j+1].Position = a
			(*imd.tri)[j+2].Position = b
		}

		imd.applyMatrixAndMask(off)
		imd.batch.Dirty()
	}

	imd.restorePoints(points)
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
			(*imd.tri)[off+i].Picture = pixel.ZV
			(*imd.tri)[off+i].Intensity = 0
		}

		for i, j := 0.0, off; i < num; i, j = i+1, j+6 {
			angle := low + i*delta
			sin, cos := math.Sincos(angle)
			normalSin, normalCos := pixel.V(sin, cos).ScaledXY(radius).Unit().XY()
			a := pt.pos.Add(pixel.V(
				radius.X*cos-thickness/2*normalCos,
				radius.Y*sin-thickness/2*normalSin,
			))
			b := pt.pos.Add(pixel.V(
				radius.X*cos+thickness/2*normalCos,
				radius.Y*sin+thickness/2*normalSin,
			))

			angle = low + (i+1)*delta
			sin, cos = math.Sincos(angle)
			normalSin, normalCos = pixel.V(sin, cos).ScaledXY(radius).Unit().XY()
			c := pt.pos.Add(pixel.V(
				radius.X*cos-thickness/2*normalCos,
				radius.Y*sin-thickness/2*normalSin,
			))
			d := pt.pos.Add(pixel.V(
				radius.X*cos+thickness/2*normalCos,
				radius.Y*sin+thickness/2*normalSin,
			))

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
			lowCenter := pt.pos.Add(pixel.V(
				radius.X*lowCos,
				radius.Y*lowSin,
			))
			normalLowSin, normalLowCos := pixel.V(lowSin, lowCos).ScaledXY(radius).Unit().XY()
			normalLow := pixel.V(normalLowCos, normalLowSin).Angle()

			highSin, highCos := math.Sincos(high)
			highCenter := pt.pos.Add(pixel.V(
				radius.X*highCos,
				radius.Y*highSin,
			))
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
				thick := pixel.V(thickness/2, 0).Rotated(normalLow)
				imd.pushPt(lowCenter.Add(thick), pt)
				imd.pushPt(lowCenter.Sub(thick), pt)
				imd.pushPt(lowCenter.Sub(thick.Normal().Scaled(orientation)), pt)
				imd.fillPolygon()
				thick = pixel.V(thickness/2, 0).Rotated(normalHigh)
				imd.pushPt(highCenter.Add(thick), pt)
				imd.pushPt(highCenter.Sub(thick), pt)
				imd.pushPt(highCenter.Add(thick.Normal().Scaled(orientation)), pt)
				imd.fillPolygon()
			case RoundEndShape:
				imd.pushPt(lowCenter, pt)
				imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), normalLow, normalLow-math.Pi*orientation)
				imd.pushPt(highCenter, pt)
				imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), normalHigh, normalHigh+math.Pi*orientation)
			}
		}
	}

	imd.restorePoints(points)
}

func (imd *IMDraw) polyline(thickness float64, closed bool) {
	points := imd.getAndClearPoints()

	if len(points) == 0 {
		imd.restorePoints(points)
		return
	}
	if len(points) == 1 {
		// one point special case
		points = append(points, points[0])
	}

	// first point
	j, i := 0, 1
	ijNormal := points[0].pos.To(points[1].pos).Normal().Unit().Scaled(thickness / 2)

	if !closed {
		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos.Add(ijNormal), points[j])
			imd.pushPt(points[j].pos.Sub(ijNormal), points[j])
			imd.pushPt(points[j].pos.Add(ijNormal.Normal()), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), ijNormal.Angle(), ijNormal.Angle()+math.Pi)
		}
	}

	imd.pushPt(points[j].pos.Add(ijNormal), points[j])
	imd.pushPt(points[j].pos.Sub(ijNormal), points[j])

	// middle points
	for i := 0; i < len(points); i++ {
		j, k := i+1, i+2

		closing := false
		if j >= len(points) {
			j %= len(points)
			closing = true
		}
		if k >= len(points) {
			if !closed {
				break
			}
			k %= len(points)
		}

		jkNormal := points[j].pos.To(points[k].pos).Normal().Unit().Scaled(thickness / 2)

		orientation := 1.0
		if ijNormal.Cross(jkNormal) > 0 {
			orientation = -1.0
		}

		imd.pushPt(points[j].pos.Sub(ijNormal), points[j])
		imd.pushPt(points[j].pos.Add(ijNormal), points[j])
		imd.fillPolygon()

		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.pushPt(points[j].pos.Add(ijNormal.Scaled(orientation)), points[j])
			imd.pushPt(points[j].pos.Add(jkNormal.Scaled(orientation)), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), ijNormal.Angle(), ijNormal.Angle()-math.Pi)
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), jkNormal.Angle(), jkNormal.Angle()+math.Pi)
		}

		if !closing {
			imd.pushPt(points[j].pos.Add(jkNormal), points[j])
			imd.pushPt(points[j].pos.Sub(jkNormal), points[j])
		}
		// "next" normal becomes previous normal
		ijNormal = jkNormal
	}

	// last point
	i, j = len(points)-2, len(points)-1
	ijNormal = points[i].pos.To(points[j].pos).Normal().Unit().Scaled(thickness / 2)

	imd.pushPt(points[j].pos.Sub(ijNormal), points[j])
	imd.pushPt(points[j].pos.Add(ijNormal), points[j])
	imd.fillPolygon()

	if !closed {
		switch points[j].endshape {
		case NoEndShape:
			// nothing
		case SharpEndShape:
			imd.pushPt(points[j].pos.Add(ijNormal), points[j])
			imd.pushPt(points[j].pos.Sub(ijNormal), points[j])
			imd.pushPt(points[j].pos.Add(ijNormal.Normal().Scaled(-1)), points[j])
			imd.fillPolygon()
		case RoundEndShape:
			imd.pushPt(points[j].pos, points[j])
			imd.fillEllipseArc(pixel.V(thickness/2, thickness/2), ijNormal.Angle(), ijNormal.Angle()-math.Pi)
		}
	}

	imd.restorePoints(points)
}
