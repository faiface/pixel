package pixel

import (
	"fmt"

	"github.com/faiface/mainthread"
	"github.com/faiface/pixel/pixelgl"
)

// NewGLTriangles returns OpenGL triangles implemented using pixelgl.VertexSlice. A few notes.
//
// Triangles returned from this function support TrianglesPosition, TrianglesColor and
// TrianglesTexture. If you need to support more, you can "override" SetLen and Update method.
//
// Draw method simply draws the underlying pixelgl.VertexSlice. It needs to be called in the main
// thread manually. Also, you need to take care of additional Target initialization or setting of
// uniform attributes.
func NewGLTriangles(shader *pixelgl.Shader, t Triangles) TargetTriangles {
	var gt *glTriangles
	mainthread.Call(func() {
		gt = &glTriangles{
			vs:     pixelgl.MakeVertexSlice(shader, 0, t.Len()),
			shader: shader,
		}
	})
	gt.SetLen(t.Len())
	gt.Update(t)
	return gt
}

type glTriangles struct {
	vs     *pixelgl.VertexSlice
	data   []float32
	shader *pixelgl.Shader
}

func (gt *glTriangles) Len() int {
	return len(gt.data) / gt.vs.Stride()
}

func (gt *glTriangles) SetLen(len int) {
	if len > gt.Len() {
		needAppend := len - gt.Len()
		for i := 0; i < needAppend; i++ {
			gt.data = append(gt.data,
				0, 0,
				1, 1, 1, 1,
				-1, -1,
			)
		}
	}
	if len < gt.Len() {
		gt.data = gt.data[:len]
	}
}

func (gt *glTriangles) Slice(i, j int) Triangles {
	return &glTriangles{
		vs:     gt.vs.Slice(i, j),
		data:   gt.data[i*gt.vs.Stride() : j*gt.vs.Stride()],
		shader: gt.shader,
	}
}

func (gt *glTriangles) updateData(t Triangles) {
	// glTriangles short path
	if t, ok := t.(*glTriangles); ok {
		copy(gt.data, t.data)
		return
	}

	// TrianglesData short path
	if t, ok := t.(*TrianglesData); ok {
		for i := 0; i < gt.Len(); i++ {
			var (
				px, py = (*t)[i].Position.XY()
				col    = (*t)[i].Color
				tx, ty = (*t)[i].Texture.XY()
			)
			gt.data[i*gt.vs.Stride()+0] = float32(px)
			gt.data[i*gt.vs.Stride()+1] = float32(py)
			gt.data[i*gt.vs.Stride()+2] = float32(col.R)
			gt.data[i*gt.vs.Stride()+3] = float32(col.G)
			gt.data[i*gt.vs.Stride()+4] = float32(col.B)
			gt.data[i*gt.vs.Stride()+5] = float32(col.A)
			gt.data[i*gt.vs.Stride()+6] = float32(tx)
			gt.data[i*gt.vs.Stride()+7] = float32(ty)
		}
		return
	}

	if t, ok := t.(TrianglesPosition); ok {
		for i := 0; i < gt.Len(); i++ {
			px, py := t.Position(i).XY()
			gt.data[i*gt.vs.Stride()+0] = float32(px)
			gt.data[i*gt.vs.Stride()+1] = float32(py)
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := 0; i < gt.Len(); i++ {
			col := t.Color(i)
			gt.data[i*gt.vs.Stride()+2] = float32(col.R)
			gt.data[i*gt.vs.Stride()+3] = float32(col.G)
			gt.data[i*gt.vs.Stride()+4] = float32(col.B)
			gt.data[i*gt.vs.Stride()+5] = float32(col.A)
		}
	}
	if t, ok := t.(TrianglesTexture); ok {
		for i := 0; i < gt.Len(); i++ {
			tx, ty := t.Texture(i).XY()
			gt.data[i*gt.vs.Stride()+6] = float32(tx)
			gt.data[i*gt.vs.Stride()+7] = float32(ty)
		}
	}
}

func (gt *glTriangles) submitData() {
	data := gt.data // avoid race condition
	mainthread.CallNonBlock(func() {
		gt.vs.Begin()
		dataLen := len(data) / gt.vs.Stride()
		gt.vs.SetLen(dataLen)
		gt.vs.SetVertexData(gt.data)
		gt.vs.End()
	})
}

func (gt *glTriangles) Update(t Triangles) {
	if gt.Len() != t.Len() {
		panic(fmt.Errorf("%T.Update: invalid triangles len", gt))
	}
	gt.updateData(t)
	gt.submitData()
}

func (gt *glTriangles) Copy() Triangles {
	return NewGLTriangles(gt.shader, gt)
}

func (gt *glTriangles) Draw() {
	gt.vs.Begin()
	gt.vs.Draw()
	gt.vs.End()
}

func (gt *glTriangles) Position(i int) Vec {
	px := gt.data[i*gt.vs.Stride()+0]
	py := gt.data[i*gt.vs.Stride()+1]
	return V(float64(px), float64(py))
}

func (gt *glTriangles) Color(i int) NRGBA {
	r := gt.data[i*gt.vs.Stride()+2]
	g := gt.data[i*gt.vs.Stride()+3]
	b := gt.data[i*gt.vs.Stride()+4]
	a := gt.data[i*gt.vs.Stride()+5]
	return NRGBA{
		R: float64(r),
		G: float64(g),
		B: float64(b),
		A: float64(a),
	}
}

func (gt *glTriangles) Texture(i int) Vec {
	tx := gt.data[i*gt.vs.Stride()+6]
	ty := gt.data[i*gt.vs.Stride()+7]
	return V(float64(tx), float64(ty))
}
