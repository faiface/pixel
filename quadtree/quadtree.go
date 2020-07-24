package quadtree

import (
	"errors"

	"github.com/faiface/pixel"
)

// Collidable is interface that stores inserted objects
type Collidable interface {
	GetRect() pixel.Rect
}

// Common part of quadtree. Commot is always coppied to children
// with change of the level
type Common struct {
	Depth int
	Level int
	Cap   int //max amount of objects per quadrant, if there is more quadrant splits
}

// Quadtree is is datastructure used for effective collision detection.
// In most cases you really only need to interact with root node.
// There are to ways to use quadtree. If objects used in quadtree hes very
// short live spam it is better to clear tree and reinset objects every frame.
// On the other hand, if objects are rather permanent use update approach.
// insert every shape just once and remove it if needed. Use Update method before
// detecting collisions or removing shapes.
type Quadtree struct {
	pixel.Rect
	tl, tr, bl, br, pr *Quadtree
	Shapes             []Collidable
	Common
	splitted bool
}

// New creates new quad tree reference.
// bounds - defines position of quad tree and its size. If shapes goes out of bounds they
// will not be assigned to quadrants and the tree will be ineffective.
// depth - resolution of quad tree. It lavais splits in half so if bounds size is 100 x 100
// and depth is 2 smallest quadrants will be 25 x 25. Making resolution too high is redundant
// if shapes cannot fit into smallest quadrants.
// cap - sets maximal capacity of quadrant before it splits to 4 smaller. Making can too big is
// inefficient. Optimal value can be 5 but its allways better to test what works the best.
func New(bounds pixel.Rect, depth, cap int) *Quadtree {
	return &Quadtree{
		Rect: bounds,
		Common: Common{
			Depth: depth,
			Cap:   cap,
		},
	}
}

// generates subquadrants, always check if quadrant is not already splitted
func (q *Quadtree) split() {
	q.splitted = true
	newCommon := q.Common
	newCommon.Level++
	halfH := q.H() / 2
	halfW := q.W() / 2
	center := q.Center()
	q.tl = &Quadtree{
		Rect: pixel.R(q.Min.X, q.Min.Y+halfH,q.Max.X-halfW, q.Max.Y),
		pr:     q,
		Common: newCommon,
	}
	q.tr = &Quadtree{
		Rect: pixel.Rect{
			Min: center,
			Max: q.Max,
		},
		pr:     q,
		Common: newCommon,
	}
	q.bl = &Quadtree{
		Rect: pixel.Rect{
			Min: q.Min,
			Max: center,
		},
		pr:     q,
		Common: newCommon,
	}
	q.br = &Quadtree{
		Rect: pixel.R(q.Min.X+halfW, q.Min.Y,q.Max.X, q.Min.Y+halfH),
		pr:     q,
		Common: newCommon,
	}
}

// returns weather shape fits into quadtree completely
func (q *Quadtree) fits(rect pixel.Rect) bool {
	return rect.Max.X > q.Min.X && rect.Max.X < q.Max.X && rect.Min.Y > q.Min.Y && rect.Max.Y < q.Max.Y
}

// finds out in witch subquadrant the shape belongs to. Shape has to overlap only with one quadrant,
// otherwise it returns nil
func (q *Quadtree) getSub(rect pixel.Rect) *Quadtree {
	vertical := q.Min.X + q.W()/2
	horizontal := q.Min.Y + q.H()/2

	if !q.fits(rect) {
		return nil
	}

	left := rect.Max.X <= vertical
	right := rect.Min.X >= vertical
	if rect.Min.Y >= horizontal {
		if left {
			return q.tl
		} else if right {
			return q.tr
		}
	} else if rect.Max.Y <= horizontal {
		if left {
			return q.bl
		} else if right {
			return q.br
		}
	}
	return nil
}

// Insert adds the shape to quad tree and assigns it to correct quadrant.
// Proper way is adding all shapes first and then detecting collisions.
func (q *Quadtree) Insert(collidable Collidable) {
	rect := collidable.GetRect()

	if q.splitted {
		fitting := q.getSub(rect)
		if fitting != nil {
			fitting.Insert(collidable)
			return
		}
		q.Shapes = append(q.Shapes, collidable)
		return
	}
	q.Shapes = append(q.Shapes, collidable)
	if q.Cap <= len(q.Shapes) && q.Level != q.Depth {

		q.split()
		new := []Collidable{}
		for _, s := range q.Shapes {
			fitting := q.getSub(s.GetRect())
			if fitting != nil {
				fitting.Insert(s)
			} else {
				new = append(new, s)
			}
		}
		q.Shapes = new
	}
}

//Update reassigns shapes to quadrants if needed
func (q *Quadtree) Update() {
	new := []Collidable{}
	if len(q.Shapes) > q.Cap && !q.splitted {
		q.split()
	}
	if q.splitted {
		q.tl.Update()
		q.tr.Update()
		q.bl.Update()
		q.br.Update()
		for _, c := range q.Shapes {
			rect := c.GetRect()
			sub := q.getSub(rect)
			if sub != nil {
				sub.Insert(c)
			} else if q.fits(rect) || q.pr == nil {
				new = append(new, c)
			} else {
				q.pr.Shapes = append(q.pr.Shapes, c)
			}
		}
	} else {
		for _, c := range q.Shapes {
			if q.fits(c.GetRect()) || q.pr == nil {
				new = append(new, c)
			} else {
				q.pr.Shapes = append(q.pr.Shapes, c)
			}
		}
	}

	q.Shapes = new
}

// GetColliding returns all coliding collidables, if rect belongs to object that is already
// inserted in tree it returns it as well
func (q *Quadtree) GetColliding(rect pixel.Rect, con *[]Collidable) {
	if q.splitted {
		if q.tl.Intersects(rect) {
			q.tl.GetColliding(rect, con)
		}
		if q.tr.Intersects(rect) {
			q.tr.GetColliding(rect, con)
		}
		if q.bl.Intersects(rect) {
			q.bl.GetColliding(rect, con)
		}
		if q.br.Intersects(rect) {
			q.br.GetColliding(rect, con)
		}
	}
	for _, c := range q.Shapes {
		if c.GetRect().Intersects(rect) {
			*con = append(*con, c)
		}
	}
}

// gets a smallest possible quadrant rect fits into.
func (q *Quadtree) getSmallestQuad(rect pixel.Rect) *Quadtree {
	current := q
	for {
		sub := current.getSub(rect)
		if sub == nil {
			break
		}
		current = sub
	}
	return current
}

// Remove removes shape from quadtree the fast wey. Always update before removing objects
// unless you are not ,moving with it.
func (q *Quadtree) Remove(c Collidable) error {
	sq := q.getSmallestQuad(c.GetRect())
	for i, o := range sq.Shapes {
		if o == c {
			last := len(sq.Shapes) - 1
			sq.Shapes[i] = nil
			sq.Shapes[i] = sq.Shapes[last]
			sq.Shapes = sq.Shapes[:last]
			return nil
		}
	}
	return errors.New("shape not found, update before removing")
}

// Clear clears the tree, use this every frame before inserting all shapes
// other wise you will run out of memory eventually and tree will not even work properly.
// You should not use this if you are using Upsate() nethod
func (q *Quadtree) Clear() {
	q.Shapes = []Collidable{}
	q.tl, q.tr, q.bl, q.br = nil, nil, nil, nil
	q.splitted = false
}
