package pixel

import (
	"github.com/faiface/mainthread"
	"github.com/faiface/pixel/pixelgl"
)

// NewGLTriangles returns OpenGL triangles implemented using pixelgl.VertexSlice. A few notes.
//
// Triangles returned from this function support TrianglesPosition, TrianglesColor and
// TrianglesTexture. If you need to support more, you can "override" Update and Append method.
//
// Draw method simply draws the underlying pixelgl.VertexSlice. It needs to be called in the main
// thread manually. Also, you need to take care of additional Target initialization or setting of
// uniform attributes.
func NewGLTriangles(shader *pixelgl.Shader, t Triangles) Triangles {
	var gt *glTriangles
	mainthread.Call(func() {
		gt = &glTriangles{
			vs:     pixelgl.MakeVertexSlice(shader, 0, t.Len()),
			shader: shader,
		}
	})
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

func (gt *glTriangles) Draw() {
	gt.vs.Begin()
	gt.vs.Draw()
	gt.vs.End()
}

func (gt *glTriangles) resize(len int) {
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

func (gt *glTriangles) updateData(offset int, t Triangles) {
	if t, ok := t.(TrianglesPosition); ok {
		for i := offset; i < offset+t.Len(); i++ {
			px, py := t.Position(i).XY()
			gt.data[i*gt.vs.Stride()+0] = float32(px)
			gt.data[i*gt.vs.Stride()+1] = float32(py)
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := offset; i < offset+t.Len(); i++ {
			col := t.Color(i)
			gt.data[i*gt.vs.Stride()+2] = float32(col.R)
			gt.data[i*gt.vs.Stride()+3] = float32(col.G)
			gt.data[i*gt.vs.Stride()+4] = float32(col.B)
			gt.data[i*gt.vs.Stride()+5] = float32(col.A)
		}
	}
	if t, ok := t.(TrianglesTexture); ok {
		for i := offset; i < offset+t.Len(); i++ {
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
		if dataLen > gt.vs.Len() {
			gt.vs.Append(make([]float32, (dataLen-gt.vs.Len())*gt.vs.Stride()))
		}
		if dataLen < gt.vs.Len() {
			gt.vs = gt.vs.Slice(0, dataLen)
		}
		gt.vs.SetVertexData(gt.data)
		gt.vs.End()
	})
}

func (gt *glTriangles) Update(t Triangles) {
	gt.resize(t.Len())
	gt.updateData(0, t)
	gt.submitData()
}

func (gt *glTriangles) Append(t Triangles) {
	gt.resize(gt.Len() + t.Len())
	gt.updateData(gt.Len()-t.Len(), t)
	gt.submitData()
}

func (gt *glTriangles) Copy() Triangles {
	return NewGLTriangles(gt.shader, gt)
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
