package pixel

import (
	"fmt"
	"image/color"
)

// TrianglesData specifies a list of Triangles vertices with three common properties: Position,
// Color and Texture.
type TrianglesData []struct {
	Position Vec
	Color    color.Color
	Texture  Vec
}

// Len returns the number of vertices in TrianglesData.
func (td *TrianglesData) Len() int {
	return len(*td)
}

// Draw is unimplemented for TrianglesData and panics.
func (td *TrianglesData) Draw() {
	panic(fmt.Errorf("%T.Draw: invalid operation", td))
}

// Update copies vertex properties from the supplied Triangles into this TrianglesData.
//
// TrianglesPosition, TrianglesColor and TrianglesTexture are supported.
func (td *TrianglesData) Update(t Triangles) {
	if t.Len() > td.Len() {
		*td = append(*td, make(TrianglesData, t.Len()-td.Len())...)
	}
	if t.Len() < td.Len() {
		*td = (*td)[:t.Len()]
	}

	// fast path optimization
	if t, ok := t.(*TrianglesData); ok {
		copy(*td, *t)
		return
	}
	if t, ok := t.(*TrianglesColorData); ok {
		copy(*td, *(*TrianglesData)(t))
		return
	}
	if t, ok := t.(*TrianglesTextureData); ok {
		copy(*td, *(*TrianglesData)(t))
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
	if t, ok := t.(TrianglesTexture); ok {
		for i := range *td {
			(*td)[i].Texture = t.Texture(i)
		}
	}
}

// Position returns the position property of i-th vertex.
func (td *TrianglesData) Position(i int) Vec {
	return (*td)[i].Position
}

// Color returns the color property of i-th vertex.
func (td *TrianglesData) Color(i int) color.Color {
	return (*td)[i].Color
}

// Texture returns the texture property of i-th vertex.
func (td *TrianglesData) Texture(i int) Vec {
	return (*td)[i].Texture
}

// TrianglesColorData is same as TrianglesData, except is lacks Texture property.
type TrianglesColorData TrianglesData

// Len returns the number of vertices in TrianglesColorData.
func (td *TrianglesColorData) Len() int {
	return (*TrianglesData)(td).Len()
}

// Draw is unimplemented for TrianglesColorData and panics.
func (td *TrianglesColorData) Draw() {
	(*TrianglesData)(td).Draw()
}

// Update copies vertex properties from the supplied Triangles into this TrianglesColorData.
func (td *TrianglesColorData) Update(t Triangles) {
	(*TrianglesData)(td).Update(t)
}

// Position returns the position property of i-th vertex.
func (td *TrianglesColorData) Position(i int) Vec {
	return (*TrianglesData)(td).Position(i)
}

// Color returns the color property of i-th vertex.
func (td *TrianglesColorData) Color(i int) color.Color {
	return (*TrianglesData)(td).Color(i)
}

// TrianglesTextureData is same as TrianglesData, except is lacks Color property.
type TrianglesTextureData TrianglesData

// Len returns the number of vertices in TrianglesTextureData.
func (td *TrianglesTextureData) Len() int {
	return (*TrianglesData)(td).Len()
}

// Draw is unimplemented for TrianglesTextureData and panics.
func (td *TrianglesTextureData) Draw() {
	(*TrianglesData)(td).Draw()
}

// Update copies vertex properties from the supplied Triangles into this TrianglesTextureData.
func (td *TrianglesTextureData) Update(t Triangles) {
	(*TrianglesData)(td).Update(t)
}

// Position returns the position property of i-th vertex.
func (td *TrianglesTextureData) Position(i int) Vec {
	return (*TrianglesData)(td).Position(i)
}

// Texture returns the texture property of i-th vertex.
func (td *TrianglesTextureData) Texture(i int) Vec {
	return (*TrianglesData)(td).Texture(i)
}

// TrianglesDrawer is a helper type that wraps Triangles and turns them into a Drawer.
//
// It does so by creating a separate Triangles instance for each Target. The instances are
// correctly updated alongside the wrapped Triangles.
type TrianglesDrawer struct {
	Triangles

	tris  map[Target]Triangles
	dirty bool
}

func (td *TrianglesDrawer) flush() {
	if !td.dirty {
		return
	}
	td.dirty = false

	for _, t := range td.tris {
		t.Update(td.Triangles)
	}
}

// Draw draws the wrapped Triangles onto the provided Target.
func (td *TrianglesDrawer) Draw(target Target) {
	if td.tris == nil {
		td.tris = make(map[Target]Triangles)
	}

	td.flush()

	tri := td.tris[target]
	if tri == nil {
		tri = target.MakeTriangles(td.Triangles)
		td.tris[target] = tri
	}
	tri.Draw()
}

// Update updates the wrapped Triangles with the supplied Triangles. Call only this method to
// update the wrapped Triangles, otherwise the TrianglesDrawer will not work correctly.
func (td *TrianglesDrawer) Update(t Triangles) {
	td.dirty = true
	td.Triangles.Update(t)
}
