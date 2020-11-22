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
	clip   pixel.Rect
}

var (
	_ pixel.TrianglesPosition = (*GLTriangles)(nil)
	_ pixel.TrianglesColor    = (*GLTriangles)(nil)
	_ pixel.TrianglesPicture  = (*GLTriangles)(nil)
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
			)
			d := gt.data[i*stride : i*stride+9]
			d[0] = float32(px)
			d[1] = float32(py)
			d[2] = float32(col.R)
			d[3] = float32(col.G)
			d[4] = float32(col.B)
			d[5] = float32(col.A)
			d[6] = float32(tx)
			d[7] = float32(ty)
			d[8] = float32(in)
		}
		return
	}

	if t, ok := t.(pixel.TrianglesPosition); ok {
		for i := 0; i < length; i++ {
			px, py := t.Position(i).XY()
			gt.data[i*stride+0] = float32(px)
			gt.data[i*stride+1] = float32(py)
		}
	}
	if t, ok := t.(pixel.TrianglesColor); ok {
		for i := 0; i < length; i++ {
			col := t.Color(i)
			gt.data[i*stride+2] = float32(col.R)
			gt.data[i*stride+3] = float32(col.G)
			gt.data[i*stride+4] = float32(col.B)
			gt.data[i*stride+5] = float32(col.A)
		}
	}
	if t, ok := t.(pixel.TrianglesPicture); ok {
		for i := 0; i < length; i++ {
			pic, intensity := t.Picture(i)
			gt.data[i*stride+6] = float32(pic.X)
			gt.data[i*stride+7] = float32(pic.Y)
			gt.data[i*stride+8] = float32(intensity)
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

// Position returns the Position property of the i-th vertex.
func (gt *GLTriangles) Position(i int) pixel.Vec {
	px := gt.data[i*gt.vs.Stride()+0]
	py := gt.data[i*gt.vs.Stride()+1]
	return pixel.V(float64(px), float64(py))
}

// Color returns the Color property of the i-th vertex.
func (gt *GLTriangles) Color(i int) pixel.RGBA {
	r := gt.data[i*gt.vs.Stride()+2]
	g := gt.data[i*gt.vs.Stride()+3]
	b := gt.data[i*gt.vs.Stride()+4]
	a := gt.data[i*gt.vs.Stride()+5]
	return pixel.RGBA{
		R: float64(r),
		G: float64(g),
		B: float64(b),
		A: float64(a),
	}
}

// Picture returns the Picture property of the i-th vertex.
func (gt *GLTriangles) Picture(i int) (pic pixel.Vec, intensity float64) {
	tx := gt.data[i*gt.vs.Stride()+6]
	ty := gt.data[i*gt.vs.Stride()+7]
	intensity = float64(gt.data[i*gt.vs.Stride()+8])
	return pixel.V(float64(tx), float64(ty)), intensity
}

// SetClipRect sets the rectangle to scissor the triangles by
func (gt *GLTriangles) SetClipRect(r pixel.Rect) {
	gt.clip = r.Norm()
}

// ClipRect gets the clipping rectangle and returns true if that
//	rectangle is not the Zero Rectangle
func (gt *GLTriangles) ClipRect() (pixel.Rect, bool) {
	return gt.clip, gt.clip.Area() != 0
}
