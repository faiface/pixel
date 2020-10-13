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

	targets    map[Target]*drawerTarget
	allTargets []*drawerTarget
	inited     bool
}

type drawerTarget struct {
	tris  TargetTriangles
	pics  map[Picture]TargetPicture
	clean bool
}

func (d *Drawer) lazyInit() {
	if !d.inited {
		d.targets = make(map[Target]*drawerTarget)
		d.inited = true
	}
}

// Dirty marks the Triangles of this Drawer as changed. If not called, changes will not be visible
// when drawing.
func (d *Drawer) Dirty() {
	d.lazyInit()

	for _, t := range d.allTargets {
		t.clean = false
	}
}

// Draw efficiently draws Triangles with Picture onto the provided Target.
//
// If Triangles is nil, nothing will be drawn. If Picture is nil, Triangles will be drawn without a
// Picture.
func (d *Drawer) Draw(t Target) {
	d.lazyInit()

	if d.Triangles == nil {
		return
	}

	dt := d.targets[t]
	if dt == nil {
		dt = &drawerTarget{
			pics: make(map[Picture]TargetPicture),
		}
		d.targets[t] = dt
		d.allTargets = append(d.allTargets, dt)
	}

	if dt.tris == nil {
		dt.tris = t.MakeTriangles(d.Triangles)
		dt.clean = true
	}

	if !dt.clean {
		dt.tris.SetLen(d.Triangles.Len())
		dt.tris.Update(d.Triangles)
		dt.clean = true
	}

	if d.Picture == nil {
		dt.tris.Draw()
		return
	}

	pic := dt.pics[d.Picture]
	if pic == nil {
		pic = t.MakePicture(d.Picture)
		dt.pics[d.Picture] = pic
	}

	pic.Draw(dt.tris)
}
