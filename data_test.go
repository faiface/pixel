package pixel_test

import (
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkTrianglesData_Len(b *testing.B) {
	tests := []struct {
		name  string
		tData *pixel.TrianglesData
	}{
		{
			name:  "Small slice",
			tData: pixel.MakeTrianglesData(10),
		},
		{
			name:  "Large slice",
			tData: pixel.MakeTrianglesData(10000),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.tData.Len()
			}
		})
	}
}

func BenchmarkTrianglesData_SetLen(b *testing.B) {
	tests := []struct {
		name        string
		tData       *pixel.TrianglesData
		nextLenFunc func(int, int) (int, int)
	}{
		{
			name:        "Stay same size",
			tData:       pixel.MakeTrianglesData(50),
			nextLenFunc: func(i, j int) (int, int) { return 50, 0 },
		},
		{
			name:  "Change size",
			tData: pixel.MakeTrianglesData(50),
			nextLenFunc: func(i, j int) (int, int) {
				// 0 is shrink
				if j == 0 {
					next := i - 1
					if next < 1 {
						return 2, 1
					}
					return next, 0
				}

				// other than 0 is grow
				next := i + 1
				if next == 100 {
					return next, 0
				}
				return next, 1
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			var newLen int
			var c int
			for i := 0; i < b.N; i++ {
				newLen, c = tt.nextLenFunc(newLen, c)
				tt.tData.SetLen(newLen)
			}
		})
	}
}

func BenchmarkTrianglesData_Slice(b *testing.B) {
	tests := []struct {
		name  string
		tData *pixel.TrianglesData
	}{
		{
			name:  "Basic slice",
			tData: pixel.MakeTrianglesData(100),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.tData.Slice(25, 50)
			}
		})
	}
}

func BenchmarkTrianglesData_Update(b *testing.B) {
	tests := []struct {
		name  string
		tData *pixel.TrianglesData
		t     pixel.Triangles
	}{
		{
			name:  "Small Triangles",
			tData: pixel.MakeTrianglesData(20),
			t:     pixel.MakeTrianglesData(20),
		},
		{
			name:  "Large Triangles",
			tData: pixel.MakeTrianglesData(10000),
			t:     pixel.MakeTrianglesData(10000),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tt.tData.Update(tt.t)
			}
		})
	}
}

func BenchmarkTrianglesData_Copy(b *testing.B) {
	tests := []struct {
		name  string
		tData *pixel.TrianglesData
	}{
		{
			name:  "Small copy",
			tData: pixel.MakeTrianglesData(20),
		},
		{
			name:  "Large copy",
			tData: pixel.MakeTrianglesData(10000),
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.tData.Copy()
			}
		})
	}
}

func BenchmarkTrianglesData_Position(b *testing.B) {
	tests := []struct {
		name     string
		tData    *pixel.TrianglesData
		position int
	}{
		{
			name:     "Getting beginning position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 2,
		},
		{
			name:     "Getting middle position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 500,
		},
		{
			name:     "Getting end position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 999,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.tData.Position(tt.position)
			}
		})
	}
}

func BenchmarkTrianglesData_Color(b *testing.B) {
	tests := []struct {
		name     string
		tData    *pixel.TrianglesData
		position int
	}{
		{
			name:     "Getting beginning position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 2,
		},
		{
			name:     "Getting middle position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 500,
		},
		{
			name:     "Getting end position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 999,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tt.tData.Color(tt.position)
			}
		})
	}
}

func BenchmarkTrianglesData_Picture(b *testing.B) {
	tests := []struct {
		name     string
		tData    *pixel.TrianglesData
		position int
	}{
		{
			name:     "Getting beginning position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 2,
		},
		{
			name:     "Getting middle position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 500,
		},
		{
			name:     "Getting end position",
			tData:    pixel.MakeTrianglesData(1000),
			position: 999,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = tt.tData.Picture(tt.position)
			}
		})
	}
}
