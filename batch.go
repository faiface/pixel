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
	cont   TrianglesDrawer
	fixpic *GLPicture

	pic *GLPicture
	mat mgl32.Mat3
	col NRGBA
}

// NewBatch creates an empty Batch with the specified Picture and container.
//
// The container is where objects get accumulated. Batch will support precisely those vertex
// properties, that the supplied container supports.
//
// Note, that if the container does not support TrianglesColor, color masking will not work.
func NewBatch(pic *GLPicture, container Triangles) *Batch {
	return &Batch{
		cont:   TrianglesDrawer{Triangles: container},
		fixpic: pic,
	}
}

// Clear removes all objects from the Batch.
func (b *Batch) Clear() {
	b.cont.SetLen(0)
	b.cont.Dirty()
}

// Draw draws all objects that are currently in the Batch onto another Target.
func (b *Batch) Draw(t Target) {
	b.cont.Draw(t)
}

// MakeTriangles returns a specialized copy of the provided Triangles, that draws onto this Batch.
func (b *Batch) MakeTriangles(t Triangles) TargetTriangles {
	return &batchTriangles{
		Triangles: t.Copy(),
		trans:     t.Copy(),
		data:      MakeTrianglesData(t.Len()),
		batch:     b,
	}
}

// SetPicture sets the current Picture that will be used with the following draws. The original
// Picture of this Picture (the one from which p was obtained by slicing) must be same as the
// original Picture of the Batch's Picture.
func (b *Batch) SetPicture(p *GLPicture) {
	if p != nil && p.Texture() != b.fixpic.Texture() {
		panic("batch: attempted to draw with a different underlying Picture")
	}
	b.pic = p
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
	data  TrianglesData

	batch *Batch
}

func (bt *batchTriangles) Draw() {
	// need to apply transforms and mask color and picture bounds
	bt.data.Update(bt.Triangles)
	for i := range bt.data {
		transPos := bt.batch.mat.Mul3x1(mgl32.Vec3{
			float32(bt.data[i].Position.X()),
			float32(bt.data[i].Position.Y()),
			1,
		})
		bt.data[i].Position = V(float64(transPos.X()), float64(transPos.Y()))
		bt.data[i].Color = bt.data[i].Color.Mul(bt.batch.col)
		if bt.batch.pic != nil && bt.data[i].Picture != V(-1, -1) {
			bt.data[i].Picture = pictureBounds(bt.batch.pic, bt.data[i].Picture)
		}
	}
	bt.trans.Update(&bt.data)

	cont := bt.batch.cont
	cont.SetLen(cont.Len() + bt.trans.Len())
	cont.Slice(cont.Len()-bt.trans.Len(), cont.Len()).Update(bt.trans)
	cont.Dirty()
}
