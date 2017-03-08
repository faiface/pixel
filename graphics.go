package pixel

// Sprite is a drawable Picture. It's always anchored by the center of it's Picture.
type Sprite struct {
	tri    *TrianglesData
	bounds Rect
	d      Drawer
}

// NewSprite creates a Sprite from the supplied Picture.
func NewSprite(pic Picture) *Sprite {
	tri := MakeTrianglesData(6)
	s := &Sprite{
		tri: tri,
		d:   Drawer{Triangles: tri},
	}
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

	var (
		center     = s.bounds.Center()
		horizontal = V(s.bounds.W()/2, 0)
		vertical   = V(0, s.bounds.H()/2)
	)

	(*s.tri)[0].Position = -horizontal - vertical
	(*s.tri)[1].Position = +horizontal - vertical
	(*s.tri)[2].Position = +horizontal + vertical
	(*s.tri)[3].Position = -horizontal - vertical
	(*s.tri)[4].Position = +horizontal + vertical
	(*s.tri)[5].Position = -horizontal + vertical

	for i := range *s.tri {
		(*s.tri)[i].Color = NRGBA{1, 1, 1, 1}
		(*s.tri)[i].Picture = center + (*s.tri)[i].Position
		(*s.tri)[i].Intensity = 1
	}

	s.d.Dirty()
}

// Picture returns the current Sprite's Picture.
func (s *Sprite) Picture() Picture {
	return s.d.Picture
}

// Draw draws the Sprite onto the provided Target.
func (s *Sprite) Draw(t Target) {
	s.d.Draw(t)
}
