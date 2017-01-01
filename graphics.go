package pixel

import (
	"image/color"
	"math"

	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// Shape is a general drawable shape constructed from vertices.
//
// Vertices are specified in the vertex array of a shape. A shape can have a picture, a color
// (mask) and a static transform.
//
// Usually you use this type only indirectly throught other specific shapes (sprites, polygons,
// ...) embedding it.
type Shape struct {
	picture   *Picture
	color     color.Color
	transform Transform
	vertices  []map[pixelgl.Attr]interface{}
	vas       map[Target]*pixelgl.VertexArray
}

// NewShape creates a new shape with specified parent, picture, color, transform and vertex array.
func NewShape(picture *Picture, c color.Color, transform Transform, vertices []map[pixelgl.Attr]interface{}) *Shape {
	return &Shape{
		picture:   picture,
		color:     c,
		transform: transform,
		vertices:  vertices,
		vas:       make(map[Target]*pixelgl.VertexArray),
	}
}

// SetPicture changes the picture of a shape.
func (s *Shape) SetPicture(picture *Picture) {
	s.picture = picture
}

// Picture returns the current picture of a shape.
func (s *Shape) Picture() *Picture {
	return s.picture
}

// SetColor changes the color (mask) of a shape.
func (s *Shape) SetColor(c color.Color) {
	s.color = c
}

// Color returns the current color (mask) of a shape.
func (s *Shape) Color() color.Color {
	return s.color
}

// SetTransform changes the ("static") transform of a shape.
func (s *Shape) SetTransform(transform Transform) {
	s.transform = transform
}

// Transform returns the current ("static") transform of a shape.
func (s *Shape) Transform() Transform {
	return s.transform
}

// Vertices returns the vertex attribute values of all vertices in a shape.
//
// Do not change!
func (s *Shape) Vertices() []map[pixelgl.Attr]interface{} {
	return s.vertices
}

// Draw draws a sprite transformed by the supplied transforms applied in the reverse order.
func (s *Shape) Draw(target Target, t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat())
	}
	mat = mat.Mul3(s.transform.Mat())

	c := NRGBAModel.Convert(s.color).(NRGBA)

	if s.vas[target] == nil {
		s.vas[target] = target.MakeVertexArray(s.vertices)
	}
	va := s.vas[target]

	pixelgl.Do(func() {
		target.Begin()
		defer target.End()

		target.Shader().SetUniformAttr(maskColorVec4, mgl32.Vec4{float32(c.R), float32(c.G), float32(c.B), float32(c.A)})
		target.Shader().SetUniformAttr(transformMat3, mat)

		if s.picture != nil {
			s.picture.Texture().Begin()
			va.Begin()
			va.Draw()
			va.End()
			s.picture.Texture().End()
		} else {
			va.Begin()
			va.Draw()
			va.End()
		}
	})
}

// MultiShape is a shape composed of several other shapes. These shapes cannot be modifies
// after combined into a multishape.
//
// Using a multishape can greatly increase drawing performance. However, it's only usable when
// the relative transformations of the shapes don't change (e.g. static blocks in a level).
//
// All shapes in a multishape must share the same texture (or use no texture).
type MultiShape struct {
	*Shape
}

// NewMultiShape creates a new multishape from several other shapes.
//
// If two of the supplied shapes have different pictures, this function panics.
func NewMultiShape(shapes ...*Shape) *MultiShape {
	var picture *Picture
	for _, shape := range shapes {
		if picture != nil && shape.Picture() != nil && shape.Picture().Texture().ID() != picture.Texture().ID() {
			panic(errors.New("failed to create multishape: shapes have different pictures"))
		}
		if shape.Picture() != nil {
			picture = shape.Picture()
		}
	}

	var vertices []map[pixelgl.Attr]interface{}
	for _, shape := range shapes {
		shapeVertices := shape.Vertices()

		for vertex := range shapeVertices {
			if pos, ok := shapeVertices[vertex][positionVec2]; ok {
				pos := pos.(mgl32.Vec2)
				pos = shape.Transform().Mat().Mul3x1(mgl32.Vec3{pos.X(), pos.Y(), 1}).Vec2()
				shapeVertices[vertex][positionVec2] = pos
			}
			if color, ok := shapeVertices[vertex][colorVec4]; ok {
				color := color.(mgl32.Vec4)
				c := NRGBAModel.Convert(shape.Color()).(NRGBA)
				color = mgl32.Vec4{
					color[0] * float32(c.R),
					color[1] * float32(c.G),
					color[2] * float32(c.B),
					color[3] * float32(c.A),
				}
				shapeVertices[vertex][colorVec4] = color
			}
		}

		vertices = append(vertices, shapeVertices...)
	}

	return &MultiShape{NewShape(picture, color.White, Position(0), vertices)}
}

// Sprite is a picture that can be drawn on the screen. Optionally it can be color masked
// or tranformed.
//
// Usually, you only transform objects when you're drawing them (by passing transforms to the
// Draw method).  With sprites however, it can be useful to also transform them "statically". For
// example, sprites are anchor by their bottom-left corner by default. Setting a transform can
// change this anchored to the center, or wherever you want.
type Sprite struct {
	*Shape
}

// NewSprite creates a new sprite with the supplied picture. The sprite's size is the size of
// the supplied picture.  If you want to change the sprite's size, change it's transform.
func NewSprite(picture *Picture) *Sprite {
	w, h := picture.Bounds().Size.XY()

	vertices := make([]map[pixelgl.Attr]interface{}, 4)
	for i, p := range []Vec{V(0, 0), V(w, 0), V(w, h), V(0, h)} {
		texCoord := V(
			(picture.Bounds().X()+p.X())/float64(picture.Texture().Width()),
			(picture.Bounds().Y()+p.Y())/float64(picture.Texture().Height()),
		)

		vertices[i] = map[pixelgl.Attr]interface{}{
			positionVec2: mgl32.Vec2{float32(p.X()), float32(p.Y())},
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{float32(texCoord.X()), float32(texCoord.Y())},
		}
	}

	vertices = []map[pixelgl.Attr]interface{}{
		vertices[0],
		vertices[1],
		vertices[2],
		vertices[0],
		vertices[2],
		vertices[3],
	}

	return &Sprite{NewShape(picture, color.White, Position(0), vertices)}
}

// LineColor a line shape (with sharp ends) filled with a single color.
type LineColor struct {
	*Shape
	a, b  Vec
	width float64
}

// NewLineColor creates a new line shape between points A and B filled with a single color.
func NewLineColor(c color.Color, a, b Vec, width float64) *LineColor {
	vertices := make([]map[pixelgl.Attr]interface{}, 4)
	for i := 0; i < 4; i++ {
		vertices[i] = map[pixelgl.Attr]interface{}{
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{-1, -1},
		}
	}

	vertices = []map[pixelgl.Attr]interface{}{
		vertices[0],
		vertices[1],
		vertices[2],
		vertices[1],
		vertices[2],
		vertices[3],
	}

	lc := &LineColor{NewShape(nil, c, Position(0), vertices), a, b, width}
	lc.setPoints()
	return lc
}

// setPoints updates the vertex array data according to A, B and width.
func (lc *LineColor) setPoints() {
	r := (lc.b - lc.a).Unit().Scaled(lc.width / 2).Rotated(math.Pi / 2)
	for i, p := range []Vec{lc.a - r, lc.a + r, lc.b - r, lc.b + r} {
		lc.va.SetVertexAttr(i, positionVec2, mgl32.Vec2{float32(p.X()), float32(p.Y())})
	}
}

// SetA changes the position of the first endpoint of a line.
func (lc *LineColor) SetA(a Vec) {
	lc.a = a
	lc.setPoints()
}

// A returns the current position of the first endpoint of a line.
func (lc *LineColor) A() Vec {
	return lc.a
}

// SetB changes the position of the second endpoint of a line.
func (lc *LineColor) SetB(b Vec) {
	lc.b = b
	lc.setPoints()
}

// B returns the current position of the second endpoint of a line.
func (lc *LineColor) B() Vec {
	return lc.b
}

// SetWidth changes the width of a line.
func (lc *LineColor) SetWidth(width float64) {
	lc.width = width
	lc.setPoints()
}

// Width returns the current width of a line.
func (lc *LineColor) Width() float64 {
	return lc.width
}

// PolygonColor is a convex polygon shape filled with a single color.
type PolygonColor struct {
	*Shape
	points []Vec
}

// NewPolygonColor creates a new polygon shape filled with a single color. Parent is an object
// that this shape belongs to, such as a window, or a graphics effect.
func NewPolygonColor(parent pixelgl.Doer, c color.Color, points ...Vec) *PolygonColor {
	var va *pixelgl.VertexArray

	var indices []int
	for i := 2; i < len(points); i++ {
		indices = append(indices, 0, i-1, i)
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		va, err = pixelgl.NewVertexArray(
			pixelgl.ContextHolder{Context: ctx},
			ctx.Shader().VertexFormat(),
			len(points),
			indices,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create polygon"))
		}
	})

	vertices := make([]map[pixelgl.Attr]interface{}, len(points))

	for i, p := range points {
		vertices[i] = map[pixelgl.Attr]interface{}{
			positionVec2: mgl32.Vec2{float32(p.X()), float32(p.Y())},
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{-1, -1},
		}
	}

	va.SetVertices(vertices)

	return &PolygonColor{NewShape(parent, nil, c, Position(0), va), points}
}

// NumPoints returns the number of points in a polygon.
func (pc *PolygonColor) NumPoints() int {
	return len(pc.points)
}

// SetPoint changes the position of a point in a polygon.
//
// If the index is out of range, this function panics.
func (pc *PolygonColor) SetPoint(i int, point Vec) {
	pc.points[i] = point
	pc.va.SetVertexAttr(i, positionVec2, mgl32.Vec2{float32(point.X()), float32(point.Y())})
}

// Point returns the position of a point in a polygon.
//
// If the index is out of range, this function panics.
func (pc *PolygonColor) Point(i int) Vec {
	return pc.points[i]
}

// EllipseColor is an ellipse shape filled with a single color.
type EllipseColor struct {
	*Shape
	radius Vec
	fill   float64
}

// NewEllipseColor creates a new ellipse shape filled with a single color. Parent is an object
// that this shape belongs to, such as a window, or a graphics effect. Fill should be a number
// between 0 and 1 which specifies how much of the ellipse will be filled (from the outside). The
// value of 1 means that the whole ellipse is filled. The value of 0 means that none of the
// ellipse is filled (which makes the ellipse invisible).
func NewEllipseColor(parent pixelgl.Doer, c color.Color, radius Vec, fill float64) *EllipseColor {
	var va *pixelgl.VertexArray

	const n = 256

	var indices []int
	for i := 2; i < (n+1)*2; i++ {
		indices = append(indices, i-2, i-1, i)
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		va, err = pixelgl.NewVertexArray(
			pixelgl.ContextHolder{Context: ctx},
			ctx.Shader().VertexFormat(),
			(n+1)*2,
			indices,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create ellipse"))
		}
	})

	vertices := make([]map[pixelgl.Attr]interface{}, (n+1)*2)

	for k := 0; k < n+1; k++ {
		i, j := k*2, k*2+1
		angle := math.Pi * 2 * float64(k%n) / n

		vertices[i] = map[pixelgl.Attr]interface{}{
			positionVec2: mgl32.Vec2{
				float32(math.Cos(angle) * radius.X()),
				float32(math.Sin(angle) * radius.Y()),
			},
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{-1, -1},
		}

		vertices[j] = map[pixelgl.Attr]interface{}{
			positionVec2: mgl32.Vec2{
				float32(math.Cos(angle) * radius.X() * (1 - fill)),
				float32(math.Sin(angle) * radius.Y() * (1 - fill)),
			},
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{-1, -1},
		}
	}

	va.SetVertices(vertices)

	return &EllipseColor{NewShape(parent, nil, c, Position(0), va), radius, fill}
}

// Radius returns the radius of an ellipse.
func (ec *EllipseColor) Radius() Vec {
	return ec.radius
}

// Fill returns the fill ratio of an ellipse.
func (ec *EllipseColor) Fill() float64 {
	return ec.fill
}
