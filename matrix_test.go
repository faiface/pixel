package pixel_test

import (
	"math/rand"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkMatrix(b *testing.B) {
	b.Run("Moved", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.Moved(pixel.V(4.217, -132.99))
		}
	})
	b.Run("ScaledXY", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.ScaledXY(pixel.V(-5.1, 9.3), pixel.V(2.1, 0.98))
		}
	})
	b.Run("Rotated", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.Rotated(pixel.V(-5.1, 9.3), 1.4)
		}
	})
	b.Run("Chained", func(b *testing.B) {
		var m1, m2 pixel.Matrix
		for i := range m1 {
			m1[i] = rand.Float64()
			m2[i] = rand.Float64()
		}
		for i := 0; i < b.N; i++ {
			m1 = m1.Chained(m2)
		}
	})
	b.Run("Project", func(b *testing.B) {
		var m pixel.Matrix
		for i := range m {
			m[i] = rand.Float64()
		}
		u := pixel.V(1, 1)
		for i := 0; i < b.N; i++ {
			u = m.Project(u)
		}
	})
	b.Run("Unproject", func(b *testing.B) {
	again:
		var m pixel.Matrix
		for i := range m {
			m[i] = rand.Float64()
		}
		if (m[0]*m[3])-(m[1]*m[2]) == 0 { // zero determinant, not invertible
			goto again
		}
		u := pixel.V(1, 1)
		for i := 0; i < b.N; i++ {
			u = m.Unproject(u)
		}
	})
}
