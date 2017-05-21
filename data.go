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
	Color     RGBA
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
				Color     RGBA
				Picture   Vec
				Intensity float64
			}{ZV, Alpha(1), ZV, 0})
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
func (td *TrianglesData) Color(i int) RGBA {
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
// The format of the pixels is color.RGBA and not pixel.RGBA for a very serious reason:
// pixel.RGBA takes up 8x more memory than color.RGBA.
type PictureData struct {
	Pix    []color.RGBA
	Stride int
	Rect   Rect
}

// MakePictureData creates a zero-initialized PictureData covering the given rectangle.
func MakePictureData(rect Rect) *PictureData {
	w := int(math.Ceil(rect.Max.X)) - int(math.Floor(rect.Min.X))
	h := int(math.Ceil(rect.Max.Y)) - int(math.Floor(rect.Min.Y))
	pd := &PictureData{
		Stride: w,
		Rect:   rect,
	}
	pd.Pix = make([]color.RGBA, w*h)
	return pd
}

func verticalFlip(rgba *image.RGBA) {
	bounds := rgba.Bounds()
	width := bounds.Dx()

	tmpRow := make([]uint8, width*4)
	for i, j := 0, bounds.Dy()-1; i < j; i, j = i+1, j-1 {
		iRow := rgba.Pix[i*rgba.Stride : i*rgba.Stride+width*4]
		jRow := rgba.Pix[j*rgba.Stride : j*rgba.Stride+width*4]

		copy(tmpRow, iRow)
		copy(iRow, jRow)
		copy(jRow, tmpRow)
	}
}

// PictureDataFromImage converts an image.Image into PictureData.
//
// The resulting PictureData's Bounds will be the equivalent of the supplied image.Image's Bounds.
func PictureDataFromImage(img image.Image) *PictureData {
	var rgba *image.RGBA
	if rgbaImg, ok := img.(*image.RGBA); ok {
		rgba = rgbaImg
	} else {
		rgba = image.NewRGBA(img.Bounds())
		draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	verticalFlip(rgba)

	pd := MakePictureData(R(
		float64(rgba.Bounds().Min.X),
		float64(rgba.Bounds().Min.Y),
		float64(rgba.Bounds().Max.X),
		float64(rgba.Bounds().Max.Y),
	))

	for i := range pd.Pix {
		pd.Pix[i].R = rgba.Pix[i*4+0]
		pd.Pix[i].G = rgba.Pix[i*4+1]
		pd.Pix[i].B = rgba.Pix[i*4+2]
		pd.Pix[i].A = rgba.Pix[i*4+3]
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
		for y := math.Floor(bounds.Min.Y); y < bounds.Max.Y; y++ {
			for x := math.Floor(bounds.Min.X); x < bounds.Max.X; x++ {
				// this together with the Floor is a trick to get all of the pixels
				at := V(
					math.Max(x, bounds.Min.X),
					math.Max(y, bounds.Min.Y),
				)
				col := pic.Color(at)
				pd.Pix[pd.Index(at)] = color.RGBA{
					R: uint8(col.R * 255),
					G: uint8(col.G * 255),
					B: uint8(col.B * 255),
					A: uint8(col.A * 255),
				}
			}
		}
	}

	return pd
}

// Image converts PictureData into an image.RGBA.
//
// The resulting image.RGBA's Bounds will be equivalent of the PictureData's Bounds.
func (pd *PictureData) Image() *image.RGBA {
	bounds := image.Rect(
		int(math.Floor(pd.Rect.Min.X)),
		int(math.Floor(pd.Rect.Min.Y)),
		int(math.Ceil(pd.Rect.Max.X)),
		int(math.Ceil(pd.Rect.Max.Y)),
	)
	rgba := image.NewRGBA(bounds)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			off := pd.Index(V(float64(x), float64(y)))
			rgba.Pix[i*4+0] = pd.Pix[off].R
			rgba.Pix[i*4+1] = pd.Pix[off].G
			rgba.Pix[i*4+2] = pd.Pix[off].B
			rgba.Pix[i*4+3] = pd.Pix[off].A
			i++
		}
	}

	verticalFlip(rgba)

	return rgba
}

// Index returns the index of the pixel at the specified position inside the Pix slice.
func (pd *PictureData) Index(at Vec) int {
	at = at.Sub(pd.Rect.Min.Map(math.Floor))
	x, y := int(at.X), int(at.Y)
	return y*pd.Stride + x
}

// Bounds returns the bounds of this PictureData.
func (pd *PictureData) Bounds() Rect {
	return pd.Rect
}

// Color returns the color located at the given position.
func (pd *PictureData) Color(at Vec) RGBA {
	if !pd.Rect.Contains(at) {
		return RGBA{0, 0, 0, 0}
	}
	return ToRGBA(pd.Pix[pd.Index(at)])
}
