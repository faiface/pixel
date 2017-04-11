package pixel

// ComposeTarget is a BasicTarget capable of Porter-Duff composition.
type ComposeTarget interface {
	BasicTarget

	// SetComposeMethod sets a Porter-Duff composition method to be used.
	SetComposeMethod(ComposeMethod)
}

// ComposeMethod is a Porter-Duff composition method.
type ComposeMethod int

// Here's the list of all available Porter-Duff composition methods. User ComposeOver for the basic
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
