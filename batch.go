package pixel

import (
	"image/color"
	"math"

	"github.com/go-gl/mathgl/mgl32"
)

// Batch is a Target that allows for efficient drawing of many objects with the same Picture (but
// different slices of the same Picture are allowed).
//
// To put an object into a Batch, just draw it onto it:
//   object.Draw(batch)
type Batch struct {
	cont Drawer

	mat mgl32.Mat3
	col NRGBA
}

var _ BasicTarget = (*Batch)(nil)

// NewBatch creates an empty Batch with the specified Picture and container.
//
// The container is where objects get accumulated. Batch will support precisely those vertex
// properties, that the supplied container supports.
//
// Note, that if the container does not support TrianglesColor, color masking will not work.
func NewBatch(container Triangles, pic Picture) *Batch {
	return &Batch{
		cont: Drawer{Triangles: container, Picture: pic},
	}
}

// Clear removes all objects from the Batch.
func (b *Batch) Clear() {
	b.cont.Triangles.SetLen(0)
	b.cont.Dirty()
}

// Draw draws all objects that are currently in the Batch onto another Target.
func (b *Batch) Draw(t Target) {
	b.cont.Draw(t)
}

// SetTransform sets transforms used in the following draws onto the Batch.
func (b *Batch) SetTransform(t ...Transform) {
	b.mat = transformToMat(t...)
}

// SetColorMask sets a mask color used in the following draws onto the Batch.
func (b *Batch) SetColorMask(c color.Color) {
	if c == nil {
		b.col = NRGBA{1, 1, 1, 1}
		return
	}
	b.col = NRGBAModel.Convert(c).(NRGBA)
}

// MakeTriangles returns a specialized copy of the provided Triangles that draws onto this Batch.
func (b *Batch) MakeTriangles(t Triangles) TargetTriangles {
	bt := &batchTriangles{
		Triangles: t.Copy(),
		orig:      MakeTrianglesData(t.Len()),
		trans:     MakeTrianglesData(t.Len()),
		b:         b,
	}
	bt.orig.Update(t)
	bt.trans.Update(&bt.orig)
	return bt
}

// MakePicture returns a specialized copy of the provided Picture that draws onto this Batch.
func (b *Batch) MakePicture(p Picture) TargetPicture {
	return &batchPicture{
		Picture: p,
	}
}

type batchTriangles struct {
	Triangles
	orig, trans TrianglesData

	b *Batch
}

func (bt *batchTriangles) draw(bp *batchPicture) {
	for i := range bt.trans {
		transPos := bt.b.mat.Mul3x1(mgl32.Vec3{
			float32(bt.orig[i].Position.X()),
			float32(bt.orig[i].Position.Y()),
			1,
		})
		bt.trans[i].Position = V(float64(transPos.X()), float64(transPos.Y()))
		bt.trans[i].Color = bt.orig[i].Color.Mul(bt.b.col)
		if bp == nil {
			bt.trans[i].Picture = V(math.Inf(+1), math.Inf(+1))
		}
	}

	bt.Triangles.Update(&bt.trans)

	cont := bt.b.cont.Triangles
	cont.SetLen(cont.Len() + bt.Triangles.Len())
	cont.Slice(cont.Len()-bt.Triangles.Len(), cont.Len()).Update(bt.Triangles)
	bt.b.cont.Dirty()
}

func (bt *batchTriangles) Draw() {
	bt.draw(nil)
}

type batchPicture struct {
	Picture
}

func (bp *batchPicture) Slice(r Rect) Picture {
	return &batchPicture{
		Picture: bp.Picture.Slice(r),
	}
}

func (bp *batchPicture) Draw(t TargetTriangles) {
	t.(*batchTriangles).draw(bp)
}
