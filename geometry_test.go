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

	resizeAroundCenterAnswers := []pixel.Rect{
		pixel.R(-5, -5, 5, 5),
		pixel.R(15, 15, 25, 25),
	}

	for i, rect := range testCases {
		answer := resizeAroundCenterAnswers[i]

		// resize rectangle by 50% anchored at it's current center point
		resizedRect := rect.Resized(rect.Center(), rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}

	resizeAroundMinAnswers := []pixel.Rect{
		pixel.R(-10, -10, 0, 0),
		pixel.R(10, 10, 20, 20),
	}

	for i, rect := range testCases {
		answer := resizeAroundMinAnswers[i]

		// resize rectangle by 50% anchored at it's Min coordinate
		resizedRect := rect.Resized(rect.Min, rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}

	resizeAroundMaxAnswers := []pixel.Rect{
		pixel.R(0, 0, 10, 10),
		pixel.R(20, 20, 30, 30),
	}

	for i, rect := range testCases {
		answer := resizeAroundMaxAnswers[i]

		// resize rectangle by 50% anchored at it's Max coordinate
		resizedRect := rect.Resized(rect.Max, rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}

	resizeAroundMiddleOfLeftSideAnswers := []pixel.Rect{
		pixel.R(-10, -5, 0, 5),
		pixel.R(10, 15, 20, 25),
	}

	for i, rect := range testCases {
		answer := resizeAroundMiddleOfLeftSideAnswers[i]

		// resize rectangle by 50% anchored at the middle of it's left side
		resizedRect := rect.Resized(pixel.V(rect.Min.X, rect.Center().Y), rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}

	resizeAroundOriginAnswers := []pixel.Rect{
		pixel.R(-5, -5, 5, 5),
		pixel.R(5, 5, 15, 15),
	}

	for i, rect := range testCases {
		answer := resizeAroundOriginAnswers[i]

		// resize rectangle by 50% anchored at the origin
		resizedRect := rect.Resized(pixel.ZV, rect.Size().Scaled(0.5))

		if resizedRect != answer {
			t.Errorf("Rectangle resize was incorrect, got %v, want: %v.", resizedRect, answer)
		}
	}
}
