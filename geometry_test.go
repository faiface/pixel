package pixel_test

import (
	"testing"

	"github.com/faiface/pixel"
)

func TestResizeRect(t *testing.T) {
	testCases := []pixel.Rect{
		pixel.R(-10, -10, 10, 10),
		pixel.R(10, 10, 30, 30),
	}

	answers := []pixel.Rect{
		pixel.R(-5, -5, 5, 5),
		pixel.R(15, 15, 25, 25),
	}

	for i, rect := range testCases {
		answer := answers[i]

		// resize rectangle by 50% anchored at it's current center point
		resizedRect := rect.Resized(rect.Center(), rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}

}
