package text_test

import (
	"fmt"
	"math/rand"
	"testing"
	"unicode"

	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/golang/freetype/truetype"
)

func TestClear(t *testing.T) {
	txt := text.New(pixel.ZV, text.Atlas7x13)

	if got, want := txt.Dot, pixel.ZV; !eqVectors(got, want) {
		t.Fatalf("txt.Dot = %v, want %v", got, want)
	}

	fmt.Fprint(txt, "Test\nClear")

	if got, want := txt.Dot, pixel.V(35, -13); !eqVectors(got, want) {
		t.Fatalf("txt.Dot = %v, want %v", got, want)
	}

	txt.Clear()

	if got, want := txt.Dot, pixel.ZV; !eqVectors(got, want) {
		t.Fatalf("txt.Dot = %v, want %v", got, want)
	}
}

func BenchmarkNewAtlas(b *testing.B) {
	runeSets := []struct {
		name string
		set  []rune
	}{
		{"ASCII", text.ASCII},
		{"Latin", text.RangeTable(unicode.Latin)},
	}

	ttf, _ := truetype.Parse(goregular.TTF)
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              16,
		GlyphCacheEntries: 1,
	})

	for _, runeSet := range runeSets {
		b.Run(runeSet.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = text.NewAtlas(face, runeSet.set)
			}
		})
	}
}

func BenchmarkTextWrite(b *testing.B) {
	runeSet := text.ASCII
	atlas := text.NewAtlas(basicfont.Face7x13, runeSet)

	lengths := []int{1, 10, 100, 1000}
	chunks := make([][]byte, len(lengths))
	for i := range chunks {
		chunk := make([]rune, lengths[i])
		for j := range chunk {
			chunk[j] = runeSet[rand.Intn(len(runeSet))]
		}
		chunks[i] = []byte(string(chunk))
	}

	for _, chunk := range chunks {
		b.Run(fmt.Sprintf("%d", len(chunk)), func(b *testing.B) {
			txt := text.New(pixel.ZV, atlas)
			for i := 0; i < b.N; i++ {
				txt.Write(chunk)
			}
		})
	}
}

func eqVectors(a, b pixel.Vec) bool {
	return (a.X == b.X && a.Y == b.Y)
}
