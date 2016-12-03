package pixel

import (
	"image/color"

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

// PolygonColor is a polygon shape filled with a single color.
type PolygonColor struct {
	parent pixelgl.Doer
	color  color.Color
	points []Vec
	va     *pixelgl.VertexArray
}

// NewPolygonColor creates a new polygon shape filled with a single color. Parent is an object that this shape belongs to,
// such as a window, or a graphical effect.
func NewPolygonColor(parent pixelgl.Doer, c color.Color, points ...Vec) *PolygonColor {
	pc := &PolygonColor{
		parent: parent,
		color:  c,
		points: points,
	}

	var format pixelgl.VertexFormat
	parent.Do(func(ctx pixelgl.Context) {
		format = ctx.Shader().VertexFormat()
	})

	var err error
	pc.va, err = pixelgl.NewVertexArray(parent, format, pixelgl.TriangleFanDrawMode, pixelgl.DynamicUsage, len(points))
	if err != nil {
		panic(errors.Wrap(err, "failed to create polygon"))
	}

	for i, p := range points {
		pc.va.SetVertexAttributeVec2(i, pixelgl.Position, mgl32.Vec2{float32(p.X()), float32(p.Y())})
		pc.va.SetVertexAttributeVec4(i, pixelgl.Color, mgl32.Vec4{1, 1, 1, 1})
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

	var shader *pixelgl.Shader
	pc.parent.Do(func(ctx pixelgl.Context) {
		shader = ctx.Shader()
	})

	r, g, b, a := colorToRGBA(pc.color)
	shader.SetUniformVec4(pixelgl.MaskColor, mgl32.Vec4{r, g, b, a})
	shader.SetUniformMat3(pixelgl.Transform, mat)
	shader.SetUniformInt(pixelgl.IsTexture, 0)

	pc.va.Draw()
}

// Delete destroys a polygon shape and releases it's video memory. Do not use this shape after calling Delete.
func (pc *PolygonColor) Delete() {
	pc.va.Delete()
}
