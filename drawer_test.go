package pixel_test

import (
	"image"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkSpriteDrawBatch(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pixel.R(0, 0, 64, 64))
	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	for i := 0; i < b.N; i++ {
		sprite.Draw(batch, pixel.IM)
	}
}
