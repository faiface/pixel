package pixelgl

import (
	"fmt"

	"github.com/faiface/glhf"
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel"
)

// GLTriangles are OpenGL triangles implemented using glhf.VertexSlice.
//
// Triangles returned from this function support TrianglesPosition, TrianglesColor and
// TrianglesPicture. If you need to support more, you can "override" SetLen and Update methods.
type GLTriangles struct {
	vs     *glhf.VertexSlice
	data   []float32
	shader *GLShader
}

var (
	_ pixel.TrianglesPosition = (*GLTriangles)(nil)
	_ pixel.TrianglesColor    = (*GLTriangles)(nil)
	_ pixel.TrianglesPicture  = (*GLTriangles)(nil)
	_ pixel.TrianglesClipped  = (*GLTriangles)(nil)
)

// The following is a helper so that the indices of
// 	each of these items is easier to see/debug.
const (
	triPosX = iota
	triPosY
	triColorR
	triColorG
	triColorB
	triColorA
	triPicX
	triPicY
	triIntensity
	triClipMinX
	triClipMinY
	triClipMaxX
	triClipMaxY
	trisAttrLen
)

// NewGLTriangles returns GLTriangles initialized with the data from the supplied Triangles.
//
// Only draw the Triangles using the provided Shader.
func NewGLTriangles(shader *GLShader, t pixel.Triangles) *GLTriangles {
	var gt *GLTriangles
	mainthread.Call(func() {
		gt = &GLTriangles{
			vs:     glhf.MakeVertexSlice(shader.s, 0, t.Len()),
			shader: shader,
		}
	})
	gt.SetLen(t.Len())
	gt.Update(t)
	return gt
}

// VertexSlice returns the VertexSlice of this GLTriangles.
//
// You can use it to draw them.
func (gt *GLTriangles) VertexSlice() *glhf.VertexSlice {
	return gt.vs
}

// Shader returns the GLTriangles's associated shader.
func (gt *GLTriangles) Shader() *GLShader {
	return gt.shader
}

// Len returns the number of vertices.
func (gt *GLTriangles) Len() int {
	return len(gt.data) / gt.vs.Stride()
}

// SetLen efficiently resizes GLTriangles to len.
//
// Time complexity is amortized O(1).
func (gt *GLTriangles) SetLen(length int) {
	switch {
	case length > gt.Len():
		needAppend := length - gt.Len()
		for i := 0; i < needAppend; i++ {
			gt.data = append(gt.data,
				0, 0,
				1, 1, 1, 1,
				0, 0,
				0,
				0, 0, 0, 0,
			)
		}
	case length < gt.Len():
		gt.data = gt.data[:length*gt.vs.Stride()]
	default:
		return
	}
	mainthread.Call(func() {
		gt.vs.Begin()
		gt.vs.SetLen(length)
		gt.vs.End()
	})
}

// Slice returns a sub-Triangles of this GLTriangles in range [i, j).
func (gt *GLTriangles) Slice(i, j int) pixel.Triangles {
	return &GLTriangles{
		vs:     gt.vs.Slice(i, j),
		data:   gt.data[i*gt.vs.Stride() : j*gt.vs.Stride()],
		shader: gt.shader,
	}
}

func (gt *GLTriangles) updateData(t pixel.Triangles) {
	// glTriangles short path
	if t, ok := t.(*GLTriangles); ok {
		copy(gt.data, t.data)
		return
	}

	// TrianglesData short path
	stride := gt.vs.Stride()
	length := gt.Len()
	if t, ok := t.(*pixel.TrianglesData); ok {
		for i := 0; i < length; i++ {
			var (
				px, py = (*t)[i].Position.XY()
				col    = (*t)[i].Color
				tx, ty = (*t)[i].Picture.XY()
				in     = (*t)[i].Intensity
				rec    = (*t)[i].ClipRect
			)
			d := gt.data[i*stride : i*stride+trisAttrLen]
			d[triPosX] = float32(px)
			d[triPosY] = float32(py)
			d[triColorR] = float32(col.R)
			d[triColorG] = float32(col.G)
			d[triColorB] = float32(col.B)
			d[triColorA] = float32(col.A)
			d[triPicX] = float32(tx)
			d[triPicY] = float32(ty)
			d[triIntensity] = float32(in)
			d[triClipMinX] = float32(rec.Min.X)
			d[triClipMinY] = float32(rec.Min.Y)
			d[triClipMaxX] = float32(rec.Max.X)
			d[triClipMaxY] = float32(rec.Max.Y)
		}
		return
	}

	if t, ok := t.(pixel.TrianglesPosition); ok {
		for i := 0; i < length; i++ {
			px, py := t.Position(i).XY()
			gt.data[i*stride+triPosX] = float32(px)
			gt.data[i*stride+triPosY] = float32(py)
		}
	}
	if t, ok := t.(pixel.TrianglesColor); ok {
		for i := 0; i < length; i++ {
			col := t.Color(i)
			gt.data[i*stride+triColorR] = float32(col.R)
			gt.data[i*stride+triColorG] = float32(col.G)
			gt.data[i*stride+triColorB] = float32(col.B)
			gt.data[i*stride+triColorA] = float32(col.A)
		}
	}
	if t, ok := t.(pixel.TrianglesPicture); ok {
		for i := 0; i < length; i++ {
			pic, intensity := t.Picture(i)
			gt.data[i*stride+triPicX] = float32(pic.X)
			gt.data[i*stride+triPicY] = float32(pic.Y)
			gt.data[i*stride+triIntensity] = float32(intensity)
		}
	}
	if t, ok := t.(pixel.TrianglesClipped); ok {
		for i := 0; i < length; i++ {
			rect, _ := t.ClipRect(i)
			gt.data[i*stride+triClipMinX] = float32(rect.Min.X)
			gt.data[i*stride+triClipMinY] = float32(rect.Min.Y)
			gt.data[i*stride+triClipMaxX] = float32(rect.Max.X)
			gt.data[i*stride+triClipMaxY] = float32(rect.Max.Y)
		}
	}
}

// Update copies vertex properties from the supplied Triangles into this GLTriangles.
//
// The two Triangles (gt and t) must be of the same len.
func (gt *GLTriangles) Update(t pixel.Triangles) {
	if gt.Len() != t.Len() {
		panic(fmt.Errorf("(%T).Update: invalid triangles len", gt))
	}
	gt.updateData(t)

	// Copy the verteces down to the glhf.VertexData
	gt.CopyVertices()
}

// CopyVertices copies the GLTriangle data down to the vertex data.
func (gt *GLTriangles) CopyVertices() {
	// this code is supposed to copy the vertex data and CallNonBlock the update if
	// the data is small enough, otherwise it'll block and not copy the data
	if len(gt.data) < 256 { // arbitrary heurestic constant
		data := append([]float32{}, gt.data...)
		mainthread.CallNonBlock(func() {
			gt.vs.Begin()
			gt.vs.SetVertexData(data)
			gt.vs.End()
		})
	} else {
		mainthread.Call(func() {
			gt.vs.Begin()
			gt.vs.SetVertexData(gt.data)
			gt.vs.End()
		})
	}
}

// Copy returns an independent copy of this GLTriangles.
//
// The returned Triangles are *GLTriangles as the underlying type.
func (gt *GLTriangles) Copy() pixel.Triangles {
	return NewGLTriangles(gt.shader, gt)
}

// index is a helper function that returns the index in the data
//	slice given the i-th vertex and the item index.
func (gt *GLTriangles) index(i, idx int) int {
	return i*gt.vs.Stride() + idx
}

// Position returns the Position property of the i-th vertex.
func (gt *GLTriangles) Position(i int) pixel.Vec {
	px := gt.data[gt.index(i, triPosX)]
	py := gt.data[gt.index(i, triPosY)]
	return pixel.V(float64(px), float64(py))
}

// SetPosition sets the position property of the i-th vertex.
func (gt *GLTriangles) SetPosition(i int, p pixel.Vec) {
	gt.data[gt.index(i, triPosX)] = float32(p.X)
	gt.data[gt.index(i, triPosY)] = float32(p.Y)
}

// Color returns the Color property of the i-th vertex.
func (gt *GLTriangles) Color(i int) pixel.RGBA {
	r := gt.data[gt.index(i, triColorR)]
	g := gt.data[gt.index(i, triColorG)]
	b := gt.data[gt.index(i, triColorB)]
	a := gt.data[gt.index(i, triColorA)]
	return pixel.RGBA{
		R: float64(r),
		G: float64(g),
		B: float64(b),
		A: float64(a),
	}
}

// SetColor sets the color property of the i-th vertex.
func (gt *GLTriangles) SetColor(i int, c pixel.RGBA) {
	gt.data[gt.index(i, triColorR)] = float32(c.R)
	gt.data[gt.index(i, triColorG)] = float32(c.G)
	gt.data[gt.index(i, triColorB)] = float32(c.B)
	gt.data[gt.index(i, triColorA)] = float32(c.A)
}

// Picture returns the Picture property of the i-th vertex.
func (gt *GLTriangles) Picture(i int) (pic pixel.Vec, intensity float64) {
	tx := gt.data[gt.index(i, triPicX)]
	ty := gt.data[gt.index(i, triPicY)]
	intensity = float64(gt.data[gt.index(i, triIntensity)])
	return pixel.V(float64(tx), float64(ty)), intensity
}

// SetPicture sets the picture property of the i-th vertex.
func (gt *GLTriangles) SetPicture(i int, pic pixel.Vec, intensity float64) {
	gt.data[gt.index(i, triPicX)] = float32(pic.X)
	gt.data[gt.index(i, triPicY)] = float32(pic.Y)
	gt.data[gt.index(i, triIntensity)] = float32(intensity)
}

// ClipRect returns the Clipping rectangle property of the i-th vertex.
func (gt *GLTriangles) ClipRect(i int) (rect pixel.Rect, is bool) {
	mx := gt.data[gt.index(i, triClipMinX)]
	my := gt.data[gt.index(i, triClipMinY)]
	ax := gt.data[gt.index(i, triClipMaxX)]
	ay := gt.data[gt.index(i, triClipMaxY)]
	rect = pixel.R(float64(mx), float64(my), float64(ax), float64(ay))
	is = rect.Area() != 0.0
	return
}

// SetClipRect sets the Clipping rectangle property of the i-th vertex.
func (gt *GLTriangles) SetClipRect(i int, rect pixel.Rect) {
	gt.data[gt.index(i, triClipMinX)] = float32(rect.Min.X)
	gt.data[gt.index(i, triClipMinY)] = float32(rect.Min.Y)
	gt.data[gt.index(i, triClipMaxX)] = float32(rect.Max.X)
	gt.data[gt.index(i, triClipMaxY)] = float32(rect.Max.Y)
}
