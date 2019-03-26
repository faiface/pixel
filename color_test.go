package pixel_test

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkColorToRGBA(b *testing.B) {
	types := []color.Color{
		color.NRGBA{R: 124, G: 14, B: 230, A: 42}, // slowest
		color.RGBA{R: 62, G: 32, B: 14, A: 63},    // faster
		pixel.RGB(0.8, 0.2, 0.5).Scaled(0.712),    // fastest
	}
	for _, col := range types {
		b.Run(fmt.Sprintf("From %T", col), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = pixel.ToRGBA(col)
			}
		})
	}
}
