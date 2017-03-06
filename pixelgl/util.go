package pixelgl

import (
	"math"

	"github.com/faiface/pixel"
)

func discreteBounds(bounds pixel.Rect) (x, y, w, h int) {
	x0 := int(math.Floor(bounds.Pos.X()))
	y0 := int(math.Floor(bounds.Pos.Y()))
	x1 := int(math.Ceil(bounds.Pos.X() + bounds.Size.X()))
	y1 := int(math.Ceil(bounds.Pos.Y() + bounds.Size.Y()))
	return x0, y0, x1 - x0, y1 - y0
}
