package pixel

// Drawer glues all the fundamental interfaces (Target, Triangles, Picture) into a coherent and the
// only intended usage pattern.
//
// Drawer makes it possible to draw any combination of Triangles and Picture onto any Target
// efficiently.
//
// To create a Drawer, just assign it's Triangles and Picture fields:
//
//   d := pixel.Drawer{Triangles: t, Picture: p}
//
// If Triangles is nil, nothing will be drawn. If Picture is nil, Triangles will be drawn without a
// Picture.
//
// Whenever you change the Triangles, call Dirty to notify Drawer that Triangles changed. You don't
// need to notify Drawer about a change of the Picture.
//
// Note, that Drawer caches the results of MakePicture from Targets it's drawn to for each Picture
// it's set to. What it means is that using a Drawer with an unbounded number of Pictures leads to a
// memory leak, since Drawer caches them and never forgets. In such a situation, create a new Drawer
// for each Picture.
type Drawer struct {
	Triangles Triangles
	Picture   Picture

	tris   map[Target]TargetTriangles
	clean  map[Target]bool
	pics   map[targetPicturePair]TargetPicture
	dirty  bool
	inited bool
}

type targetPicturePair struct {
	Target  Target
	Picture Picture
}

func (d *Drawer) lazyInit() {
	if !d.inited {
		d.tris = make(map[Target]TargetTriangles)
		d.clean = make(map[Target]bool)
		d.pics = make(map[targetPicturePair]TargetPicture)
		d.inited = true
	}
}

// Dirty marks the Triangles of this Drawer as changed. If not called, changes will not be visible
// when drawing.
func (d *Drawer) Dirty() {
	d.lazyInit()

	d.dirty = true
}

// Draw efficiently draws Triangles with Picture onto the provided Target.
//
// If Triangles is nil, nothing will be drawn. If Picture is nil, Triangles will be drawn without a
// Picture.
func (d *Drawer) Draw(t Target) {
	d.lazyInit()

	if d.dirty {
		for t := range d.clean {
			d.clean[t] = false
		}
		d.dirty = false
	}

	if d.Triangles == nil {
		return
	}

	tri := d.tris[t]
	if tri == nil {
		tri = t.MakeTriangles(d.Triangles)
		d.tris[t] = tri
		d.clean[t] = true
	}

	if !d.clean[t] {
		tri.SetLen(d.Triangles.Len())
		tri.Update(d.Triangles)
		d.clean[t] = true
	}

	if d.Picture == nil {
		tri.Draw()
		return
	}

	pic := d.pics[targetPicturePair{t, d.Picture}]
	if pic == nil {
		pic = t.MakePicture(d.Picture)
		d.pics[targetPicturePair{t, d.Picture}] = pic
	}

	pic.Draw(tri)
}
