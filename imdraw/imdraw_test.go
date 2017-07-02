package imdraw_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

func BenchmarkPush(b *testing.B) {
	imd := imdraw.New(nil)
	for i := 0; i < b.N; i++ {
		imd.Push(pixel.V(123.1, 99.4))
	}
}

func pointLists(counts ...int) [][]pixel.Vec {
	lists := make([][]pixel.Vec, len(counts))
	for i := range lists {
		lists[i] = make([]pixel.Vec, counts[i])
		for j := range lists[i] {
			lists[i][j] = pixel.V(
				rand.Float64()*5000-2500,
				rand.Float64()*5000-2500,
			)
		}
	}
	return lists
}

func BenchmarkLine(b *testing.B) {
	lists := pointLists(2, 5, 10, 100, 1000)
	for _, pts := range lists {
		b.Run(fmt.Sprintf("%d", len(pts)), func(b *testing.B) {
			imd := imdraw.New(nil)
			for i := 0; i < b.N; i++ {
				imd.Push(pts...)
				imd.Line(1)
			}
		})
	}
}

func BenchmarkRectangle(b *testing.B) {
	lists := pointLists(2, 10, 100, 1000)
	for _, pts := range lists {
		b.Run(fmt.Sprintf("%d", len(pts)), func(b *testing.B) {
			imd := imdraw.New(nil)
			for i := 0; i < b.N; i++ {
				imd.Push(pts...)
				imd.Rectangle(0)
			}
		})
	}
}

func BenchmarkPolygon(b *testing.B) {
	lists := pointLists(3, 10, 100, 1000)
	for _, pts := range lists {
		b.Run(fmt.Sprintf("%d", len(pts)), func(b *testing.B) {
			imd := imdraw.New(nil)
			for i := 0; i < b.N; i++ {
				imd.Push(pts...)
				imd.Polygon(0)
			}
		})
	}
}

func BenchmarkEllipseFill(b *testing.B) {
	lists := pointLists(1, 10, 100, 1000)
	for _, pts := range lists {
		b.Run(fmt.Sprintf("%d", len(pts)), func(b *testing.B) {
			imd := imdraw.New(nil)
			for i := 0; i < b.N; i++ {
				imd.Push(pts...)
				imd.Ellipse(pixel.V(50, 100), 0)
			}
		})
	}
}

func BenchmarkEllipseOutline(b *testing.B) {
	lists := pointLists(1, 10, 100, 1000)
	for _, pts := range lists {
		b.Run(fmt.Sprintf("%d", len(pts)), func(b *testing.B) {
			imd := imdraw.New(nil)
			for i := 0; i < b.N; i++ {
				imd.Push(pts...)
				imd.Ellipse(pixel.V(50, 100), 1)
			}
		})
	}
}
