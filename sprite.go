package pixel

import "image/color"

// Sprite is a drawable frame of a Picture. It's anchored by the center of it's Picture's frame.
//
// Frame specifies a rectangular portion of the Picture that will be drawn. For example, this
// creates a Sprite that draws the whole Picture:
//
//   sprite := pixel.NewSprite(pic, pic.Bounds())
//
// Note, that Sprite caches the results of MakePicture from Targets it's drawn to for each Picture
// it's set to. What it means is that using a Sprite with an unbounded number of Pictures leads to a
// memory leak, since Sprite caches them and never forgets. In such a situation, create a new Sprite
// for each Picture.
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

// Draw draws the Sprite onto the provided Target. The Sprite will be transformed by the given Matrix.
//
// This method is equivalent to calling DrawColorMask with nil color mask.
func (s *Sprite) Draw(t Target, matrix Matrix) {
	s.DrawColorMask(t, matrix, nil)
}

// DrawColorMask draws the Sprite onto the provided Target. The Sprite will be transformed by the
// given Matrix and all of it's color will be multiplied by the given mask.
//
// If the mask is nil, a fully opaque white mask will be used, which causes no effect.
func (s *Sprite) DrawColorMask(t Target, matrix Matrix, mask color.Color) {
	dirty := false
	if matrix != s.matrix {
		s.matrix = matrix
		dirty = true
	}
	if mask == nil {
		mask = Alpha(1)
	}
	rgba := ToRGBA(mask)
	if rgba != s.mask {
		s.mask = rgba
		dirty = true
	}

	if dirty {
		s.calcData()
	}

	s.d.Draw(t)
}

func (s *Sprite) calcData() {
	var (
		center     = s.frame.Center()
		horizontal = V(s.frame.W()/2, 0)
		vertical   = V(0, s.frame.H()/2)
	)

	(*s.tri)[0].Position = Vec{}.Sub(horizontal).Sub(vertical)
	(*s.tri)[1].Position = Vec{}.Add(horizontal).Sub(vertical)
	(*s.tri)[2].Position = Vec{}.Add(horizontal).Add(vertical)
	(*s.tri)[3].Position = Vec{}.Sub(horizontal).Sub(vertical)
	(*s.tri)[4].Position = Vec{}.Add(horizontal).Add(vertical)
	(*s.tri)[5].Position = Vec{}.Sub(horizontal).Add(vertical)

	for i := range *s.tri {
		(*s.tri)[i].Color = s.mask
		(*s.tri)[i].Picture = center.Add((*s.tri)[i].Position)
		(*s.tri)[i].Intensity = 1
	}

	// matrix and mask
	for i := range *s.tri {
		(*s.tri)[i].Position = s.matrix.Project((*s.tri)[i].Position)
		(*s.tri)[i].Color = s.mask
	}

	s.d.Dirty()
}
