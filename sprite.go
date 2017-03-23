package pixel

import "image/color"

// Sprite is a drawable Picture. It's anchored by the center of it's Picture.
//
// To achieve different anchoring, transformations and color masking, use SetMatrix and SetColorMask
// methods.
type Sprite struct {
	tri    *TrianglesData
	bounds Rect
	d      Drawer

	matrix Matrix
	mask   NRGBA
}

// NewSprite creates a Sprite from the supplied Picture.
func NewSprite(pic Picture) *Sprite {
	tri := MakeTrianglesData(6)
	s := &Sprite{
		tri: tri,
		d:   Drawer{Triangles: tri},
	}
	s.matrix = IM
	s.mask = NRGBA{1, 1, 1, 1}
	s.SetPicture(pic)
	return s
}

// SetPicture changes the Sprite's Picture. The new Picture may have a different size, everything
// works.
func (s *Sprite) SetPicture(pic Picture) {
	s.d.Picture = pic

	if s.bounds == pic.Bounds() {
		return
	}
	s.bounds = pic.Bounds()

	s.calcData()
}

// Picture returns the current Sprite's Picture.
func (s *Sprite) Picture() Picture {
	return s.d.Picture
}

// SetMatrix sets a Matrix that this Sprite will be transformed by. This overrides any previously
// set Matrix.
//
// Note, that this has nothing to do with BasicTarget's SetMatrix method. This only affects this
// Sprite and is usable with any Target.
func (s *Sprite) SetMatrix(matrix Matrix) {
	s.matrix = matrix
	s.calcData()
}

// Matrix returns the currently set Matrix.
func (s *Sprite) Matrix() Matrix {
	return s.matrix
}

// SetColorMask sets a color that this Sprite will be multiplied by. This overrides any previously
// set color mask.
//
// Note, that this has nothing to do with BasicTarget's SetColorMask method. This only affects this
// Sprite and is usable with any Target.
func (s *Sprite) SetColorMask(mask color.Color) {
	s.mask = NRGBAModel.Convert(mask).(NRGBA)
	s.calcData()
}

// ColorMask returns the currently set color mask.
func (s *Sprite) ColorMask() NRGBA {
	return s.mask
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	s.d.Draw(t)
}

func (s *Sprite) calcData() {
	var (
		center     = s.bounds.Center()
		horizontal = X(s.bounds.W() / 2)
		vertical   = Y(s.bounds.H() / 2)
	)

	(*s.tri)[0].Position = -horizontal - vertical
	(*s.tri)[1].Position = +horizontal - vertical
	(*s.tri)[2].Position = +horizontal + vertical
	(*s.tri)[3].Position = -horizontal - vertical
	(*s.tri)[4].Position = +horizontal + vertical
	(*s.tri)[5].Position = -horizontal + vertical

	for i := range *s.tri {
		(*s.tri)[i].Color = s.mask
		(*s.tri)[i].Picture = center + (*s.tri)[i].Position
		(*s.tri)[i].Intensity = 1
	}

	// matrix and mask
	for i := range *s.tri {
		(*s.tri)[i].Position = s.matrix.Project((*s.tri)[i].Position)
		(*s.tri)[i].Color = s.mask
	}

	s.d.Dirty()
}
