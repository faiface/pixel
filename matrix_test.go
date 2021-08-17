package pixel_test

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"github.com/faiface/pixel"
	"github.com/stretchr/testify/assert"
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

func TestMatrix_Unproject(t *testing.T) {
	const delta = 1e-15
	t.Run("for rotated matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Rotated(pixel.ZV, math.Pi/2)
		unprojected := matrix.Unproject(pixel.V(0, 1))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Moved(pixel.V(1, 2))
		unprojected := matrix.Unproject(pixel.V(2, 5))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 3, delta)
	})
	t.Run("for scaled matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Scaled(pixel.ZV, 2)
		unprojected := matrix.Unproject(pixel.V(2, 4))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 2, delta)
	})
	t.Run("for scaled, rotated and moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Scaled(pixel.ZV, 2).
			Rotated(pixel.ZV, math.Pi/2).
			Moved(pixel.V(2, 2))
		unprojected := matrix.Unproject(pixel.V(-2, 6))
		assert.InDelta(t, unprojected.X, 2, delta)
		assert.InDelta(t, unprojected.Y, 2, delta)
	})
	t.Run("for rotated and moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Rotated(pixel.ZV, math.Pi/2).
			Moved(pixel.V(1, 1))
		unprojected := matrix.Unproject(pixel.V(1, 2))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for projected vertices using all kinds of matrices", func(t *testing.T) {
		namedMatrices := map[string]pixel.Matrix{
			"IM":                        pixel.IM,
			"Scaled":                    pixel.IM.Scaled(pixel.ZV, 0.5),
			"Scaled x 2":                pixel.IM.Scaled(pixel.ZV, 2),
			"Rotated":                   pixel.IM.Rotated(pixel.ZV, math.Pi/4),
			"Moved":                     pixel.IM.Moved(pixel.V(0.5, 1)),
			"Moved 2":                   pixel.IM.Moved(pixel.V(-1, -0.5)),
			"Scaled and Rotated":        pixel.IM.Scaled(pixel.ZV, 0.5).Rotated(pixel.ZV, math.Pi/4),
			"Scaled, Rotated and Moved": pixel.IM.Scaled(pixel.ZV, 0.5).Rotated(pixel.ZV, math.Pi/4).Moved(pixel.V(1, 2)),
			"Rotated and Moved":         pixel.IM.Rotated(pixel.ZV, math.Pi/4).Moved(pixel.V(1, 2)),
		}
		vertices := [...]pixel.Vec{
			pixel.V(0, 0),
			pixel.V(5, 0),
			pixel.V(5, 10),
			pixel.V(0, 10),
			pixel.V(-5, 10),
			pixel.V(-5, 0),
			pixel.V(-5, -10),
			pixel.V(0, -10),
			pixel.V(5, -10),
		}
		for matrixName, matrix := range namedMatrices {
			for _, vertex := range vertices {
				testCase := fmt.Sprintf("for matrix %s and vertex %v", matrixName, vertex)
				t.Run(testCase, func(t *testing.T) {
					projected := matrix.Project(vertex)
					unprojected := matrix.Unproject(projected)
					assert.InDelta(t, vertex.X, unprojected.X, delta)
					assert.InDelta(t, vertex.Y, unprojected.Y, delta)
				})
			}
		}
	})
	t.Run("for singular matrix", func(t *testing.T) {
		matrix := pixel.Matrix{0, 0, 0, 0, 0, 0}
		unprojected := matrix.Unproject(pixel.ZV)
		assert.True(t, math.IsNaN(unprojected.X))
		assert.True(t, math.IsNaN(unprojected.Y))
	})
}
