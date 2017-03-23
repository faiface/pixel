package pixel

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// TrianglesData specifies a list of Triangles vertices with three common properties:
// TrianglesPosition, TrianglesColor and TrianglesPicture.
type TrianglesData []struct {
	Position  Vec
	Color     NRGBA
	Picture   Vec
	Intensity float64
}

// MakeTrianglesData creates TrianglesData of length len initialized with default property values.
//
// Prefer this function to make(TrianglesData, len), because make zeros them, while this function
// does the correct intialization.
func MakeTrianglesData(len int) *TrianglesData {
	td := &TrianglesData{}
	td.SetLen(len)
	return td
}

// Len returns the number of vertices in TrianglesData.
func (td *TrianglesData) Len() int {
	return len(*td)
}

// SetLen resizes TrianglesData to len, while keeping the original content.
//
// If len is greater than TrianglesData's current length, the new data is filled with default
// values ((0, 0), white, (0, 0), 0).
func (td *TrianglesData) SetLen(len int) {
	if len > td.Len() {
		needAppend := len - td.Len()
		for i := 0; i < needAppend; i++ {
			*td = append(*td, struct {
				Position  Vec
				Color     NRGBA
				Picture   Vec
				Intensity float64
			}{V(0, 0), NRGBA{1, 1, 1, 1}, V(0, 0), 0})
		}
	}
	if len < td.Len() {
		*td = (*td)[:len]
	}
}

// Slice returns a sub-Triangles of this TrianglesData.
func (td *TrianglesData) Slice(i, j int) Triangles {
	s := TrianglesData((*td)[i:j])
	return &s
}

func (td *TrianglesData) updateData(t Triangles) {
	// fast path optimization
	if t, ok := t.(*TrianglesData); ok {
		copy(*td, *t)
		return
	}

	// slow path manual copy
	if t, ok := t.(TrianglesPosition); ok {
		for i := range *td {
			(*td)[i].Position = t.Position(i)
		}
	}
	if t, ok := t.(TrianglesColor); ok {
		for i := range *td {
			(*td)[i].Color = t.Color(i)
		}
	}
	if t, ok := t.(TrianglesPicture); ok {
		for i := range *td {
			(*td)[i].Picture, (*td)[i].Intensity = t.Picture(i)
		}
	}
}

// Update copies vertex properties from the supplied Triangles into this TrianglesData.
//
// TrianglesPosition, TrianglesColor and TrianglesTexture are supported.
func (td *TrianglesData) Update(t Triangles) {
	if td.Len() != t.Len() {
		panic(fmt.Errorf("(%T).Update: invalid triangles length", td))
	}
	td.updateData(t)
}

// Copy returns an exact independent copy of this TrianglesData.
func (td *TrianglesData) Copy() Triangles {
	copyTd := TrianglesData{}
	copyTd.SetLen(td.Len())
	copyTd.Update(td)
	return &copyTd
}

// Position returns the position property of i-th vertex.
func (td *TrianglesData) Position(i int) Vec {
	return (*td)[i].Position
}

// Color returns the color property of i-th vertex.
func (td *TrianglesData) Color(i int) NRGBA {
	return (*td)[i].Color
}

// Picture returns the picture property of i-th vertex.
func (td *TrianglesData) Picture(i int) (pic Vec, intensity float64) {
	return (*td)[i].Picture, (*td)[i].Intensity
}

// PictureData specifies an in-memory rectangular area of pixels and implements Picture and
// PictureColor.
//
// Pixels are small rectangles of unit size of form (x, y, x+1, y+1), where x and y are integers.
// PictureData contains and assigns a color to all pixels that are at least partially contained
// within it's Bounds (Rect).
//
// The struct's innards are exposed for convenience, manual modification is at your own risk.
//
// The format of the pixels is color.NRGBA and not pixel.NRGBA for a very serious reason:
// pixel.NRGBA takes up 8x more memory than color.NRGBA.
type PictureData struct {
	Pix    []color.NRGBA
	Stride int
	Rect   Rect
	Orig   *PictureData
}

// MakePictureData creates a zero-initialized PictureData covering the given rectangle.
func MakePictureData(rect Rect) *PictureData {
	w := int(math.Ceil(rect.Max.X())) - int(math.Floor(rect.Min.X()))
	h := int(math.Ceil(rect.Max.Y())) - int(math.Floor(rect.Min.Y()))
	pd := &PictureData{
		Stride: w,
		Rect:   rect,
	}
	pd.Pix = make([]color.NRGBA, w*h)
	pd.Orig = pd
	return pd
}

func verticalFlip(nrgba *image.NRGBA) {
	bounds := nrgba.Bounds()
	width := bounds.Dx()

	tmpRow := make([]uint8, width*4)
	for i, j := 0, bounds.Dy()-1; i < j; i, j = i+1, j-1 {
		iRow := nrgba.Pix[i*nrgba.Stride : i*nrgba.Stride+width*4]
		jRow := nrgba.Pix[j*nrgba.Stride : j*nrgba.Stride+width*4]

		copy(tmpRow, iRow)
		copy(iRow, jRow)
		copy(jRow, tmpRow)
	}
}

// PictureDataFromImage converts an image.Image into PictureData.
//
// The resulting PictureData's Bounds will be the equivalent of the supplied image.Image's Bounds.
func PictureDataFromImage(img image.Image) *PictureData {
	var nrgba *image.NRGBA
	if nrgbaImg, ok := img.(*image.NRGBA); ok {
		nrgba = nrgbaImg
	} else {
		nrgba = image.NewNRGBA(img.Bounds())
		draw.Draw(nrgba, nrgba.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	verticalFlip(nrgba)

	pd := MakePictureData(R(
		float64(nrgba.Bounds().Min.X),
		float64(nrgba.Bounds().Min.Y),
		float64(nrgba.Bounds().Dx()),
		float64(nrgba.Bounds().Dy()),
	))

	for i := range pd.Pix {
		pd.Pix[i].R = nrgba.Pix[i*4+0]
		pd.Pix[i].G = nrgba.Pix[i*4+1]
		pd.Pix[i].B = nrgba.Pix[i*4+2]
		pd.Pix[i].A = nrgba.Pix[i*4+3]
	}

	return pd
}

// PictureDataFromPicture converts an arbitrary Picture into PictureData (the conversion may be
// lossy, because PictureData works with unit-sized pixels).
//
// Bounds are preserved.
func PictureDataFromPicture(pic Picture) *PictureData {
	if pd, ok := pic.(*PictureData); ok {
		return pd
	}

	bounds := pic.Bounds()
	pd := MakePictureData(bounds)

	if pic, ok := pic.(PictureColor); ok {
		for y := math.Floor(bounds.Min.Y()); y < bounds.Max.Y(); y++ {
			for x := math.Floor(bounds.Min.X()); x < bounds.Max.X(); x++ {
				// this together with the Floor is a trick to get all of the pixels
				at := V(
					math.Max(x, bounds.Min.X()),
					math.Max(y, bounds.Min.Y()),
				)
				pd.SetColor(at, pic.Color(at))
			}
		}
	}

	return pd
}

// Image converts PictureData into an image.NRGBA.
//
// The resulting image.NRGBA's Bounds will be equivalent of the PictureData's Bounds.
func (pd *PictureData) Image() *image.NRGBA {
	bounds := image.Rect(
		int(math.Floor(pd.Rect.Min.X())),
		int(math.Floor(pd.Rect.Min.Y())),
		int(math.Ceil(pd.Rect.Max.X())),
		int(math.Ceil(pd.Rect.Max.Y())),
	)
	nrgba := image.NewNRGBA(bounds)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := pd.Index(V(float64(x), float64(y)))
			nrgba.Pix[i*4+0] = pd.Pix[off].R
			nrgba.Pix[i*4+1] = pd.Pix[off].G
			nrgba.Pix[i*4+2] = pd.Pix[off].B
			nrgba.Pix[i*4+3] = pd.Pix[off].A
			i++
		}
	}

	verticalFlip(nrgba)

	return nrgba
}

// Index returns the index of the pixel at the specified position inside the Pix slice.
func (pd *PictureData) Index(at Vec) int {
	at -= pd.Rect.Min.Map(math.Floor)
	x, y := int(at.X()), int(at.Y())
	return y*pd.Stride + x
}

// Bounds returns the bounds of this PictureData.
func (pd *PictureData) Bounds() Rect {
	return pd.Rect
}

// Slice returns a sub-Picture of this PictureData inside the supplied rectangle.
func (pd *PictureData) Slice(r Rect) Picture {
	return &PictureData{
		Pix:    pd.Pix[pd.Index(r.Min):],
		Stride: pd.Stride,
		Rect:   r,
		Orig:   pd.Orig,
	}
}

// Original returns the most original PictureData that this PictureData was obtained from using
// Slice-ing.
func (pd *PictureData) Original() Picture {
	return pd.Orig
}

// Color returns the color located at the given position.
func (pd *PictureData) Color(at Vec) NRGBA {
	if !pd.Rect.Contains(at) {
		return NRGBA{0, 0, 0, 0}
	}
	return ToNRGBA(pd.Pix[pd.Index(at)])
}

// SetColor changes the color located at the given position.
func (pd *PictureData) SetColor(at Vec, col color.Color) {
	if !pd.Rect.Contains(at) {
		return
	}
	nrgba := ToNRGBA(col)
	pd.Pix[pd.Index(at)] = color.NRGBA{
		R: uint8(nrgba.R * 255),
		G: uint8(nrgba.G * 255),
		B: uint8(nrgba.B * 255),
		A: uint8(nrgba.A * 255),
	}
}
