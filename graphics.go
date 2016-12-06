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

// Deleter is anything that can be deleted. All graphics objects that have some associated video memory
// are deleters. It is necessary to call Delete when you're done with an object, otherwise you're going
// to have video memory leaks.
type Deleter interface {
	Delete()
}

// DrawDeleter combines Drawer and Deleter interfaces.
type DrawDeleter interface {
	Drawer
	Deleter
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

// Sprite is a picture that can be drawn on the screen. Optionally it can be color masked or tranformed.
//
// Usually, you only transform objects when you're drawing them (by passing transforms to the Draw method).
// With sprites however, it can be useful to also transform them "statically". For example, sprites are
// anchor by their bottom-left corner by default. Setting a transform can change this anchor to the center,
// or wherever you want.
type Sprite struct {
	parent    pixelgl.Doer
	color     color.Color
	picture   Picture
	transform Transform
	va        *pixelgl.VertexArray
}

// NewSprite creates a new sprite with the supplied picture. The sprite's size is the size of the supplied picture.
// If you want to change the sprite's size, change it's transform.
func NewSprite(parent pixelgl.Doer, picture Picture) *Sprite {
	s := &Sprite{
		parent:    parent,
		color:     color.White,
		picture:   picture,
		transform: Position(0),
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		s.va, err = pixelgl.NewVertexArray(
			picture.Texture(),
			ctx.Shader().VertexFormat(),
			pixelgl.TriangleFanDrawMode,
			pixelgl.DynamicUsage,
			4,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create sprite"))
		}
	})

	w, h := picture.Bounds().Size.XY()
	for i, p := range []Vec{V(0, 0), V(w, 0), V(w, h), V(0, h)} {
		texCoord := V(
			(picture.Bounds().X()+p.X())/float64(picture.Texture().Width()),
			(picture.Bounds().Y()+p.Y())/float64(picture.Texture().Height()),
		)

		s.va.SetVertexAttributeVec2(
			i,
			pixelgl.Position,
			mgl32.Vec2{
				float32(p.X()),
				float32(p.Y()),
			},
		)
		s.va.SetVertexAttributeVec4(i, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})
		s.va.SetVertexAttributeVec2(
			i,
			pixelgl.TexCoord,
			mgl32.Vec2{
				float32(texCoord.X()),
				float32(texCoord.Y()),
			},
		)
	}

	return s
}

// Delete deletes a sprite. Note, that this does not delete it's picture.
func (s *Sprite) Delete() {
	s.va.Delete()
}

// Picture returns the sprite's picture.
func (s *Sprite) Picture() Picture {
	return s.picture
}

// SetColor sets a mask color of a sprite.
func (s *Sprite) SetColor(c color.Color) {
	s.color = c
}

// Color returns the mask color of a sprite. Default is white.
func (s *Sprite) Color() color.Color {
	return s.color
}

// SetTransform sets a "static" transform of a sprite. Setting a transform is equivalent to passing
// the transform as the last parameter to Draw.
//
//   sprite.SetTransform(transform)
//   sprite.Draw(camera)
//   // same as below
//   sprite.Draw(camera, tranform)
func (s *Sprite) SetTransform(t Transform) {
	s.transform = t
}

// Transform returns the static transform of a sprite.
func (s *Sprite) Transform() Transform {
	return s.transform
}

// Draw draws a sprite transformed by the supplied transforms applied in the reverse order.
func (s *Sprite) Draw(t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat3())
	}
	mat = mat.Mul3(s.transform.Mat3())

	s.parent.Do(func(ctx pixelgl.Context) {
		r, g, b, a := colorToRGBA(s.color)
		ctx.Shader().SetUniformVec4(pixelgl.MaskColor, mgl32.Vec4{r, g, b, a})
		ctx.Shader().SetUniformMat3(pixelgl.Transform, mat)
		ctx.Shader().SetUniformInt(pixelgl.IsTexture, 1)

		s.va.Draw()
	})
}

// LineColor a line shape (with sharp ends) filled with a single color.
type LineColor struct {
	parent pixelgl.Doer
	color  color.Color
	a, b   Vec
	width  float64
	va     *pixelgl.VertexArray
}

// NewLineColor creates a new line shape between points A and B filled with a single color. Parent is an object
// that this shape belongs to, such as a window, or a graphics effect.
func NewLineColor(parent pixelgl.Doer, c color.Color, a, b Vec, width float64) *LineColor {
	lc := &LineColor{
		parent: parent,
		color:  c,
		a:      a,
		b:      b,
		width:  width,
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		lc.va, err = pixelgl.NewVertexArray(
			parent,
			ctx.Shader().VertexFormat(),
			pixelgl.TriangleStripDrawMode,
			pixelgl.DynamicUsage,
			4,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create line"))
		}
	})

	lc.va.SetVertexAttributeVec4(0, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})
	lc.va.SetVertexAttributeVec4(1, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})
	lc.va.SetVertexAttributeVec4(2, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})
	lc.va.SetVertexAttributeVec4(3, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})

	lc.setPoints()

	return lc
}

// setPoints updates the vertex array data according to A, B and width.
func (lc *LineColor) setPoints() {
	r := (lc.b - lc.a).Unit().Scaled(lc.width / 2).Rotated(math.Pi / 2)
	for i, p := range []Vec{lc.a - r, lc.a + r, lc.b - r, lc.b + r} {
		lc.va.SetVertexAttributeVec2(i, pixelgl.Position, mgl32.Vec2{float32(p.X()), float32(p.Y())})
	}
}

// SetColor changes the color of a line.
func (lc *LineColor) SetColor(c color.Color) {
	lc.color = c
}

// Color returns the current color of a line.
func (lc *LineColor) Color() color.Color {
	return lc.color
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

// Draw draws a line transformed by the supplied transforms applied in the reverse order.
func (lc *LineColor) Draw(t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat3())
	}

	lc.parent.Do(func(ctx pixelgl.Context) {
		r, g, b, a := colorToRGBA(lc.color)
		ctx.Shader().SetUniformVec4(pixelgl.MaskColor, mgl32.Vec4{r, g, b, a})
		ctx.Shader().SetUniformMat3(pixelgl.Transform, mat)
		ctx.Shader().SetUniformInt(pixelgl.IsTexture, 0)
	})

	lc.va.Draw()
}

// Delete destroys a line shape and releases it's video memory. Do not use this shape after calling Delete.
func (lc *LineColor) Delete() {
	lc.va.Delete()
}

// PolygonColor is a polygon shape filled with a single color.
type PolygonColor struct {
	parent pixelgl.Doer
	color  color.Color
	points []Vec
	va     *pixelgl.VertexArray
}

// NewPolygonColor creates a new polygon shape filled with a single color. Parent is an object that this shape belongs to,
// such as a window, or a graphics effect.
func NewPolygonColor(parent pixelgl.Doer, c color.Color, points ...Vec) *PolygonColor {
	pc := &PolygonColor{
		parent: parent,
		color:  c,
		points: points,
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		pc.va, err = pixelgl.NewVertexArray(
			parent,
			ctx.Shader().VertexFormat(),
			pixelgl.TriangleFanDrawMode,
			pixelgl.DynamicUsage,
			len(points),
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create polygon"))
		}
	})

	for i, p := range points {
		pc.va.SetVertexAttributeVec2(
			i,
			pixelgl.Position,
			mgl32.Vec2{
				float32(p.X()),
				float32(p.Y()),
			},
		)
		pc.va.SetVertexAttributeVec4(
			i,
			pixelgl.Color,
			mgl32.Vec4{1, 1, 1, 1},
		)
	}

	return pc
}

// Count returns the number of points in a polygon.
func (pc *PolygonColor) Count() int {
	return len(pc.points)
}

// SetColor changes the color of a polygon to c.
func (pc *PolygonColor) SetColor(c color.Color) {
	pc.color = c
}

// Color returns the current color of a polygon.
func (pc *PolygonColor) Color() color.Color {
	return pc.color
}

// SetPoint changes the position of a point in a polygon.
//
// If the index is out of range, this function panics.
func (pc *PolygonColor) SetPoint(i int, point Vec) {
	pc.points[i] = point
	pc.va.SetVertexAttributeVec2(i, pixelgl.Position, mgl32.Vec2{float32(point.X()), float32(point.Y())})
}

// Point returns the position of a point in a polygon.
//
// If the index is out of range, this function panics.
func (pc *PolygonColor) Point(i int) Vec {
	return pc.points[i]
}

// Draw draws a polygon transformed by the supplied transforms applied in the reverse order.
func (pc *PolygonColor) Draw(t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat3())
	}

	pc.parent.Do(func(ctx pixelgl.Context) {
		r, g, b, a := colorToRGBA(pc.color)
		ctx.Shader().SetUniformVec4(pixelgl.MaskColor, mgl32.Vec4{r, g, b, a})
		ctx.Shader().SetUniformMat3(pixelgl.Transform, mat)
		ctx.Shader().SetUniformInt(pixelgl.IsTexture, 0)
	})

	pc.va.Draw()
}

// Delete destroys a polygon shape and releases it's video memory. Do not use this shape after calling Delete.
func (pc *PolygonColor) Delete() {
	pc.va.Delete()
}

// EllipseColor is an ellipse shape filled with a single color.
type EllipseColor struct {
	parent pixelgl.Doer
	color  color.Color
	radius Vec
	fill   float64
	va     *pixelgl.VertexArray
}

// NewEllipseColor creates a new ellipse shape filled with a single color. Parent is an object that this shape belongs to,
// such as a window, or a graphics effect. Fill should be a number between 0 and 1 which specifies how much of the ellipse will
// be filled (from the outside). The value of 1 means that the whole ellipse is filled. The value of 0 means that none of the
// ellipse is filled (which makes the ellipse invisible).
func NewEllipseColor(parent pixelgl.Doer, c color.Color, radius Vec, fill float64) *EllipseColor {
	const n = 256

	ec := &EllipseColor{
		parent: parent,
		color:  c,
		radius: radius,
		fill:   fill,
	}

	parent.Do(func(ctx pixelgl.Context) {
		var err error
		ec.va, err = pixelgl.NewVertexArray(
			parent,
			ctx.Shader().VertexFormat(),
			pixelgl.TriangleStripDrawMode,
			pixelgl.DynamicUsage,
			(n+1)*2,
		)
		if err != nil {
			panic(errors.Wrap(err, "failed to create circle"))
		}
	})

	for k := 0; k < n+1; k++ {
		i, j := k*2, k*2+1
		angle := math.Pi * 2 * float64(k%n) / n
		ec.va.SetVertexAttributeVec2(
			i,
			pixelgl.Position,
			mgl32.Vec2{
				float32(math.Cos(angle)),
				float32(math.Sin(angle)),
			},
		)
		ec.va.SetVertexAttributeVec4(
			i,
			pixelgl.Color,
			mgl32.Vec4{1, 1, 1, 1},
		)
		ec.va.SetVertexAttributeVec2(
			j,
			pixelgl.Position,
			mgl32.Vec2{
				float32(math.Cos(angle) * (1 - fill)),
				float32(math.Sin(angle) * (1 - fill)),
			},
		)
		ec.va.SetVertexAttributeVec4(
			j,
			pixelgl.Color,
			mgl32.Vec4{1, 1, 1, 1},
		)
	}

	return ec
}

// SetColor changes the color of an ellipse.
func (ec *EllipseColor) SetColor(c color.Color) {
	ec.color = c
}

// Color returns the current color of an ellipse.
func (ec *EllipseColor) Color() color.Color {
	return ec.color
}

// SetRadius sets the radius (which can be different in X and Y axis) of an ellipse.
func (ec *EllipseColor) SetRadius(radius Vec) {
	ec.radius = radius
}

// Radius returns the current radius of an ellipse.
func (ec *EllipseColor) Radius() Vec {
	return ec.radius
}

// Draw dras an ellipse transformed by the supplied transforms applied in the reverse order.
func (ec *EllipseColor) Draw(t ...Transform) {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat3())
	}
	mat = mat.Mul3(mgl32.Scale2D(float32(ec.radius.X()), float32(ec.radius.Y())))

	ec.parent.Do(func(ctx pixelgl.Context) {
		r, g, b, a := colorToRGBA(ec.color)
		ctx.Shader().SetUniformVec4(pixelgl.MaskColor, mgl32.Vec4{r, g, b, a})
		ctx.Shader().SetUniformMat3(pixelgl.Transform, mat)
		ctx.Shader().SetUniformInt(pixelgl.IsTexture, 0)
	})

	ec.va.Draw()
}

// Delete destroys an ellipse shape and releases it's video memory. Do not use this shape after calling Delete.
func (ec *EllipseColor) Delete() {
	ec.va.Delete()
}
