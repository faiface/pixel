package text_test

import (
	"testing"

	"github.com/faiface/pixel/text"
)

func TestAtlas7x13(t *testing.T) {
	if text.Atlas7x13 == nil {
		t.Fatalf("Atlas7x13 is nil")
	}

	for _, tt := range []struct {
		runes []rune
		want  bool
	}{{text.ASCII, true}, {[]rune("ÅÄÖ"), false}} {
		for _, r := range tt.runes {
			if got := text.Atlas7x13.Contains(r); got != tt.want {
				t.Fatalf("Atlas7x13.Contains('%s') = %v, want %v", string(r), got, tt.want)
			}
		}
	}
}
