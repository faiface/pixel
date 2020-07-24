package quadtree

import (
	"testing"

	"github.com/faiface/pixel"
)

type collider struct {
	pixel.Vec
	size float64
	vel  pixel.Vec
}

func newCol(pos pixel.Vec) collider {
	return collider{
		Vec:  pos,
		size: 10,
	}
}

func newVelCol(pos, vel pixel.Vec) collider {
	return collider{
		Vec:  pos,
		size: 10,
		vel:  vel,
	}
}

func (c *collider) GetRect() pixel.Rect {
	return pixel.R(c.X-c.size, c.Y-c.size, c.X+c.size, c.Y+c.size)
}

type collizionTest struct {
	description string
	target      collider
	other       []collider
	expected    int
}

func TestQuadtree_DetectCollizion(t *testing.T) {
	tests := []collizionTest{
		{
			description: "two colliders on same position",
			target:      newCol(pixel.ZV),
			other:       []collider{newCol(pixel.ZV)},
			expected:    2,
		},
		{
			description: "4 colliders on same position",
			target:      newCol(pixel.ZV),
			other:       []collider{newCol(pixel.ZV), newCol(pixel.ZV), newCol(pixel.ZV)},
			expected:    4,
		},
		{
			description: "two colliders apart",
			target:      newCol(pixel.ZV),
			other:       []collider{newCol(pixel.V(100, 0))},
			expected:    1,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			qt := New(pixel.R(-200, -200, 200, 200), 2, 1)
			qt.Insert(&test.target)
			for _, o := range test.other {
				qt.Insert(&o)
			}
			cols := []Collidable{}
			qt.GetColliding(test.target.GetRect(), &cols)
			if len(cols) != test.expected {
				t.Errorf("Got: %o Vanted: %o", len(cols), test.expected)
			}
		})

	}
}

func TestQuadtree_Update(t *testing.T) {
	tests := []collizionTest{
		{
			description: "two idle colliders",
			target:      newCol(pixel.ZV),
			other:       []collider{newCol(pixel.ZV)},
			expected:    2,
		},
		{
			description: "one idle, one moving apart collider",
			target:      newCol(pixel.ZV),
			other:       []collider{newVelCol(pixel.ZV, pixel.V(100, 0))},
			expected:    1,
		},
		{
			description: "two colliders apart moving to same position",
			target:      newVelCol(pixel.V(100, 0), pixel.V(-100, 0)),
			other:       []collider{newVelCol(pixel.V(0, 100), pixel.V(0, -100))},
			expected:    2,
		},
	}
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			qt := New(pixel.R(-200, -200, 200, 200), 2, 1)
			qt.Insert(&test.target)
			test.target.Vec = test.target.Vec.Add(test.target.vel)
			for _, o := range test.other {
				qt.Insert(&o)
				o.Vec = o.Vec.Add(o.vel)
			}
			qt.Update()
			cols := []Collidable{}
			qt.GetColliding(test.target.GetRect(), &cols)
			if len(cols) != test.expected {
				t.Errorf("Got: %o Vanted: %o", len(cols), test.expected)
			}
		})
	}
}
