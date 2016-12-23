package pixel

import (
	"image/color"
	"math"

	"github.com/faiface/pixel/pixelgl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/pkg/errors"
)

// Drawer is anything that can be drawn. It's by no means a drawer inside your table.
//
// Drawer consists of a single methods: Draw. Draw methods takes any number of Transform arguments. It applies these
// transforms in the reverse order and finally draws something transformed by these transforms.
//
// Example:
//
//   // object is a drawer
//   object.Draw(pixel.Position(pixel.V(100, 100).Rotate(math.Pi / 2)))
//   camera := pixel.Camera(pixel.V(0, 0), pixel.V(500, 500), pixel.V(window.Size()))
//   object.Draw(camera, pixel.Position(0).Scale(0.5))
type Drawer interface {
	Draw(t ...Transform)
}

// Group is used to effeciently handle a collection of objects with a common parent. Usually many objects share a parent,
// using a group can significantly increase performance in these cases.
//
// To use a group, first, create a group and as it's parent use the common parent of the collection of objects:
//
//   group := pixel.NewGroup(commonParent)
//
// Then, when creating the objects, use the group as their parent, instead of the original common parent, but, don't forget
// to put everything into a With block, like this:
//
//   group.With(func() {
//       object := newArbitratyObject(group, ...) // group is the parent of the object
//   })
//
// When dealing with objects associated with a group, it's always necessary to wrap that into a With block:
//
//   group.With(func() {
//       for _, obj := range objectsWithCommonParent {
//           // do something with obj
//       }
//   })
//
// That's all!
type Group struct {
	parent  pixelgl.Doer
	context pixelgl.Context
}

// NewGroup creates a new group with the specified parent.
func NewGroup(parent pixelgl.Doer) *Group {
	return &Group{
		parent: parent,
	}
}

// With enables the parent of a group and executes sub.
func (g *Group) With(sub func()) {
	g.parent.Do(func(ctx pixelgl.Context) {
		g.context = ctx
		sub()
	})
}

// Do just passes a cached context to sub.
func (g *Group) Do(sub func(pixelgl.Context)) {
	sub(g.context)
}

// Shape is a general drawable shape constructed from vertices.
//
// Vertices are specified in the vertex array of a shape. A shape can have a picture, a color (mask) and a static
// transform.
//
// Usually you use this type only indirectly throught other specific shapes (sprites, polygons, ...) embedding it.
type Shape struct {
	parent    pixelgl.Doer
	picture   *Picture
	color     color.Color
	transform Transform
	va        *pixelgl.VertexArray
}

// NewShape creates a new shape with specified parent, picture, color, transform and vertex array.
func NewShape(parent pixelgl.Doer, picture *Picture, c color.Color, transform Transform, va *pixelgl.VertexArray) *Shape {
	return &Shape{
		parent:    parent,
		picture:   picture,
		color:     c,
		transform: transform,
		va:        va,
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

// VertexArray changes the underlying vertex array of a shape.
func (s *Shape) VertexArray() *pixelgl.VertexArray {
	return s.va
}

// Draw draws a sprite transformed by the supplied transforms applied in the reverse order.
func (s *Shape) Draw(t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat())
	}
	mat = mat.Mul3(s.transform.Mat())

	s.parent.Do(func(ctx pixelgl.Context) {
		c := NRGBAModel.Convert(s.color).(NRGBA)
		ctx.Shader().SetUniformAttr(maskColorVec4, mgl32.Vec4{float32(c.R), float32(c.G), float32(c.B), float32(c.A)})
		ctx.Shader().SetUniformAttr(transformMat3, mat)

		if s.picture != nil {
			s.picture.Texture().Do(func(pixelgl.Context) {
				s.va.Draw()
			})
		} else {
			s.va.Draw()
		}
	})
}

// MultiShape is a shape composed of several other shapes. These shapes cannot be modifies after combined into a multishape.
//
// Using a multishape can greatly increase drawing performance. However, it's only usable when the relative transformations
// of the shapes don't change (e.g. static blocks in a level).
//
// All shapes in a multishape must share the same texture (or use no texture).
type MultiShape struct {
	*Shape
}

// NewMultiShape creates a new multishape from several other shapes.
//
// If two of the supplied shapes have different pictures, this function panics.
func NewMultiShape(parent pixelgl.Doer, shapes ...*Shape) *MultiShape {
	var picture *Picture
	for _, shape := range shapes {
		if picture != nil && shape.Picture() != nil && shape.Picture().Texture().ID() != picture.Texture().ID() {
			panic(errors.New("failed to create multishape: shapes have different pictures"))
		}
		if shape.Picture() != nil {
			picture = shape.Picture()
		}
	}

	var va *pixelgl.VertexArray

	var indices []int
	offset := 0
	for _, shape := range shapes {
		for _, i := range shape.va.Indices() {
			indices = append(indices, offset+i)
		}
		offset += shape.VertexArray().NumVertices()
	}

	var vertices []map[pixelgl.Attr]interface{}
	for _, shape := range shapes {
		shapeVertices := shape.VertexArray().Vertices()

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

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		va, err = pixelgl.NewVertexArray(
			pixelgl.ContextHolder{Context: ctx},
			ctx.Shader().VertexFormat(),
			len(vertices),
			indices,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create multishape"))
		}
	})

	va.SetVertices(vertices)

	return &MultiShape{NewShape(parent, picture, color.White, Position(0), va)}
}

// Sprite is a picture that can be drawn on the screen. Optionally it can be color masked or tranformed.
//
// Usually, you only transform objects when you're drawing them (by passing transforms to the Draw method).
// With sprites however, it can be useful to also transform them "statically". For example, sprites are
// anchor by their bottom-left corner by default. Setting a transform can change this anchored to the center,
// or wherever you want.
type Sprite struct {
	*Shape
}

// NewSprite creates a new sprite with the supplied picture. The sprite's size is the size of the supplied picture.
// If you want to change the sprite's size, change it's transform.
func NewSprite(parent pixelgl.Doer, picture *Picture) *Sprite {
	var va *pixelgl.VertexArray

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		va, err = pixelgl.NewVertexArray(
			pixelgl.ContextHolder{Context: ctx},
			ctx.Shader().VertexFormat(),
			4,
			[]int{0, 1, 2, 0, 2, 3},
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create sprite"))
		}
	})

	vertices := make([]map[pixelgl.Attr]interface{}, 4)

	w, h := picture.Bounds().Size.XY()
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

	va.SetVertices(vertices)

	return &Sprite{NewShape(parent, picture, color.White, Position(0), va)}
}

// LineColor a line shape (with sharp ends) filled with a single color.
type LineColor struct {
	*Shape
	a, b  Vec
	width float64
}

// NewLineColor creates a new line shape between points A and B filled with a single color. Parent is an object
// that this shape belongs to, such as a window, or a graphics effect.
func NewLineColor(parent pixelgl.Doer, c color.Color, a, b Vec, width float64) *LineColor {
	var va *pixelgl.VertexArray

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		va, err = pixelgl.NewVertexArray(
			pixelgl.ContextHolder{Context: ctx},
			ctx.Shader().VertexFormat(),
			4,
			[]int{0, 1, 2, 1, 2, 3},
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create line"))
		}
	})

	vertices := make([]map[pixelgl.Attr]interface{}, 4)

	for i := 0; i < 4; i++ {
		vertices[i] = map[pixelgl.Attr]interface{}{
			colorVec4:    mgl32.Vec4{1, 1, 1, 1},
			texCoordVec2: mgl32.Vec2{-1, -1},
		}
	}

	va.SetVertices(vertices)

	lc := &LineColor{NewShape(parent, nil, c, Position(0), va), a, b, width}
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

// NewPolygonColor creates a new polygon shape filled with a single color. Parent is an object that this shape belongs to,
// such as a window, or a graphics effect.
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

// NewEllipseColor creates a new ellipse shape filled with a single color. Parent is an object that this shape belongs to,
// such as a window, or a graphics effect. Fill should be a number between 0 and 1 which specifies how much of the ellipse will
// be filled (from the outside). The value of 1 means that the whole ellipse is filled. The value of 0 means that none of the
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
