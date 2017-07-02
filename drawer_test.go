package pixel_test

import (
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkSpriteDrawBatch(b *testing.B) {
	sprite := pixel.NewSprite(nil, pixel.R(0, 0, 64, 64))
	batch := pixel.NewBatch(&pixel.TrianglesData{}, nil)
	for i := 0; i < b.N; i++ {
		sprite.Draw(batch, pixel.IM)
	}
}
