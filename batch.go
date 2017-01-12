package pixel

import (
	"image/color"

	"github.com/go-gl/mathgl/mgl32"
)

// Batch is a Target that allows for efficient drawing of many objects with the same Picture (but
// different slices of the same Picture are allowed).
//
// To put an object into a Batch, just draw it onto it:
//   object.Draw(batch)
type Batch struct {
	cont TrianglesDrawer

	pic *Picture
	mat mgl32.Mat3
	col NRGBA
}

// NewBatch creates an empty Batch with the specified Picture and container.
//
// The container is where objects get accumulated. Batch will support precisely those vertex
// properties, that the supplied container supports.
//
// Note, that if the container does not support TrianglesColor, color masking will not work.
func NewBatch(pic *Picture, container Triangles) *Batch {
	return &Batch{
		cont: TrianglesDrawer{Triangles: container},
		pic:  pic,
	}
}

// Clear removes all objects from the Batch.
func (b *Batch) Clear() {
	b.cont.Update(&TrianglesData{})
}

// Draw draws all objects that are currently in the Batch onto another Target.
func (b *Batch) Draw(t Target) {
	t.SetPicture(b.pic)
	b.cont.Draw(t)
}

// MakeTriangles returns a specialized copy of the provided Triangles, that draws onto this Batch.
func (b *Batch) MakeTriangles(t Triangles) Triangles {
	return &batchTriangles{
		Triangles: t.Copy(),
		trans:     t.Copy(),
		batch:     b,
	}
}

// SetPicture only checks, whether the supplied Picture has the same underlying Picture as the fixed
// Picture of this Batch. If that is not true, this method panics.
func (b *Batch) SetPicture(p *Picture) {
	if p == nil {
		return
	}
	if p.Texture() != b.pic.Texture() {
		panic("batch: attempted to draw with a different Picture")
	}
}

// SetTransform sets transforms used in the following draws onto the Batch.
func (b *Batch) SetTransform(t ...Transform) {
	b.mat = transformToMat(t...)
}

// SetMaskColor sets a mask color used in the following draws onto the Batch.
func (b *Batch) SetMaskColor(c color.Color) {
	if c == nil {
		b.col = NRGBA{1, 1, 1, 1}
		return
	}
	b.col = NRGBAModel.Convert(c).(NRGBA)
}

type batchTriangles struct {
	Triangles
	trans Triangles

	batch *Batch
}

func (bt *batchTriangles) Draw() {
	// need to apply transforms and mask color and picture bounds
	trans := make(TrianglesData, bt.Len())
	trans.Update(bt.Triangles)
	for i := range trans {
		transPos := bt.batch.mat.Mul3x1(mgl32.Vec3{
			float32(trans[i].Position.X()),
			float32(trans[i].Position.Y()),
			1,
		})
		trans[i].Position = V(float64(transPos.X()), float64(transPos.Y()))
		trans[i].Color = trans[i].Color.Mul(bt.batch.col)
		//TODO: texture
	}
	bt.trans.Update(&trans)
	bt.batch.cont.Append(bt.trans)
}
