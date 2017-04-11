package pixel

import "errors"

// ComposeTarget is a BasicTarget capable of Porter-Duff composition.
type ComposeTarget interface {
	BasicTarget

	// SetComposeMethod sets a Porter-Duff composition method to be used.
	SetComposeMethod(ComposeMethod)
}

// ComposeMethod is a Porter-Duff composition method.
type ComposeMethod int

// Here's the list of all available Porter-Duff composition methods. Use ComposeOver for the basic
// alpha blending.
const (
	ComposeOver ComposeMethod = iota
	ComposeIn
	ComposeOut
	ComposeAtop
	ComposeRover
	ComposeRin
	ComposeRout
	ComposeRatop
	ComposeXor
	ComposePlus
	ComposeCopy
)

// Compose composes two colors together according to the ComposeMethod. A is the foreground, B is
// the background.
func (cm ComposeMethod) Compose(a, b RGBA) RGBA {
	var fa, fb float64

	switch cm {
	case ComposeOver:
		fa, fb = 1, 1-a.A
	case ComposeIn:
		fa, fb = b.A, 0
	case ComposeOut:
		fa, fb = 1-b.A, 0
	case ComposeAtop:
		fa, fb = b.A, 1-a.A
	case ComposeRover:
		fa, fb = 1-b.A, 1
	case ComposeRin:
		fa, fb = 0, a.A
	case ComposeRout:
		fa, fb = 0, 1-a.A
	case ComposeRatop:
		fa, fb = 1-b.A, a.A
	case ComposeXor:
		fa, fb = 1-b.A, 1-a.A
	case ComposePlus:
		fa, fb = 1, 1
	case ComposeCopy:
		fa, fb = 1, 0
	default:
		panic(errors.New("Compose: invalid ComposeMethod"))
	}

	return a.Mul(Alpha(fa)).Add(b.Mul(Alpha(fb)))
}
