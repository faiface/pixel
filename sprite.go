package pixel

import "image/color"

// Sprite is a drawable frame of a Picture. It's anchored by the center of it's Picture's frame.
//
// Frame specifies a rectangular portion of the Picture that will be drawn. For example, this
// creates a Sprite that draws the whole Picture:
//
//   sprite := pixel.NewSprite(pic, pic.Bounds())
//
// To achieve different anchoring, transformations and color masking, use SetMatrix and SetColorMask
// methods.
type Sprite struct {
	tri   *TrianglesData
	frame Rect
	d     Drawer

	matrix Matrix
	mask   RGBA
}

// NewSprite creates a Sprite from the supplied frame of a Picture.
func NewSprite(pic Picture, frame Rect) *Sprite {
	tri := MakeTrianglesData(6)
	s := &Sprite{
		tri: tri,
		d:   Drawer{Triangles: tri},
	}
	s.matrix = IM
	s.mask = Alpha(1)
	s.Set(pic, frame)
	return s
}

// Set sets a new frame of a Picture for this Sprite.
func (s *Sprite) Set(pic Picture, frame Rect) {
	s.d.Picture = pic
	if frame != s.frame {
		s.frame = frame
		s.calcData()
	}
}

// Picture returns the current Sprite's Picture.
func (s *Sprite) Picture() Picture {
	return s.d.Picture
}

// Frame returns the current Sprite's frame.
func (s *Sprite) Frame() Rect {
	return s.frame
}

// SetMatrix sets a Matrix that this Sprite will be transformed by. This overrides any previously
// set Matrix.
//
// Note, that this has nothing to do with BasicTarget's SetMatrix method. This only affects this
// Sprite and is usable with any Target.
func (s *Sprite) SetMatrix(matrix Matrix) {
	if s.matrix != matrix {
		s.matrix = matrix
		s.calcData()
	}
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
	rgba := ToRGBA(mask)
	if s.mask != rgba {
		s.mask = ToRGBA(mask)
		s.calcData()
	}
}

// ColorMask returns the currently set color mask.
func (s *Sprite) ColorMask() RGBA {
	return s.mask
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	s.d.Draw(t)
}

func (s *Sprite) calcData() {
	var (
		center     = s.frame.Center()
		horizontal = X(s.frame.W() / 2)
		vertical   = Y(s.frame.H() / 2)
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
