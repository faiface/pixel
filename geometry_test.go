package pixel_test

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"

	"github.com/faiface/pixel"
)

type rectTestTransform struct {
	name string
	f    func(pixel.Rect) pixel.Rect
}

func TestResizeRect(t *testing.T) {

	// rectangles
	squareAroundOrigin := pixel.R(-10, -10, 10, 10)
	squareAround2020 := pixel.R(10, 10, 30, 30)
	rectangleAroundOrigin := pixel.R(-20, -10, 20, 10)
	rectangleAround2020 := pixel.R(0, 10, 40, 30)

	// resize transformations
	resizeByHalfAroundCenter := rectTestTransform{"by half around center", func(rect pixel.Rect) pixel.Rect {
		return rect.Resized(rect.Center(), rect.Size().Scaled(0.5))
	}}
	resizeByHalfAroundMin := rectTestTransform{"by half around Min", func(rect pixel.Rect) pixel.Rect {
		return rect.Resized(rect.Min, rect.Size().Scaled(0.5))
	}}
	resizeByHalfAroundMax := rectTestTransform{"by half around Max", func(rect pixel.Rect) pixel.Rect {
		return rect.Resized(rect.Max, rect.Size().Scaled(0.5))
	}}
	resizeByHalfAroundMiddleOfLeftSide := rectTestTransform{"by half around middle of left side", func(rect pixel.Rect) pixel.Rect {
		return rect.Resized(pixel.V(rect.Min.X, rect.Center().Y), rect.Size().Scaled(0.5))
	}}
	resizeByHalfAroundOrigin := rectTestTransform{"by half around the origin", func(rect pixel.Rect) pixel.Rect {
		return rect.Resized(pixel.ZV, rect.Size().Scaled(0.5))
	}}

	testCases := []struct {
		input     pixel.Rect
		transform rectTestTransform
		answer    pixel.Rect
	}{
		{squareAroundOrigin, resizeByHalfAroundCenter, pixel.R(-5, -5, 5, 5)},
		{squareAround2020, resizeByHalfAroundCenter, pixel.R(15, 15, 25, 25)},
		{rectangleAroundOrigin, resizeByHalfAroundCenter, pixel.R(-10, -5, 10, 5)},
		{rectangleAround2020, resizeByHalfAroundCenter, pixel.R(10, 15, 30, 25)},

		{squareAroundOrigin, resizeByHalfAroundMin, pixel.R(-10, -10, 0, 0)},
		{squareAround2020, resizeByHalfAroundMin, pixel.R(10, 10, 20, 20)},
		{rectangleAroundOrigin, resizeByHalfAroundMin, pixel.R(-20, -10, 0, 0)},
		{rectangleAround2020, resizeByHalfAroundMin, pixel.R(0, 10, 20, 20)},

		{squareAroundOrigin, resizeByHalfAroundMax, pixel.R(0, 0, 10, 10)},
		{squareAround2020, resizeByHalfAroundMax, pixel.R(20, 20, 30, 30)},
		{rectangleAroundOrigin, resizeByHalfAroundMax, pixel.R(0, 0, 20, 10)},
		{rectangleAround2020, resizeByHalfAroundMax, pixel.R(20, 20, 40, 30)},

		{squareAroundOrigin, resizeByHalfAroundMiddleOfLeftSide, pixel.R(-10, -5, 0, 5)},
		{squareAround2020, resizeByHalfAroundMiddleOfLeftSide, pixel.R(10, 15, 20, 25)},
		{rectangleAroundOrigin, resizeByHalfAroundMiddleOfLeftSide, pixel.R(-20, -5, 0, 5)},
		{rectangleAround2020, resizeByHalfAroundMiddleOfLeftSide, pixel.R(0, 15, 20, 25)},

		{squareAroundOrigin, resizeByHalfAroundOrigin, pixel.R(-5, -5, 5, 5)},
		{squareAround2020, resizeByHalfAroundOrigin, pixel.R(5, 5, 15, 15)},
		{rectangleAroundOrigin, resizeByHalfAroundOrigin, pixel.R(-10, -5, 10, 5)},
		{rectangleAround2020, resizeByHalfAroundOrigin, pixel.R(0, 5, 20, 15)},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Resize %v %s", testCase.input, testCase.transform.name), func(t *testing.T) {
			testResult := testCase.transform.f(testCase.input)
			if testResult != testCase.answer {
				t.Errorf("Got: %v, wanted: %v\n", testResult, testCase.answer)
			}
		})
	}
}

func TestMatrix_Unproject(t *testing.T) {
	const delta = 4e-16
	t.Run("for rotated matrix", func(t *testing.T) {
		matrix := pixel.IM.Rotated(pixel.ZV, math.Pi/2)
		unprojected := matrix.Unproject(pixel.V(0, 1))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for moved matrix", func(t *testing.T) {
		matrix := pixel.IM.Moved(pixel.V(5, 5))
		unprojected := matrix.Unproject(pixel.V(0, 0))
		assert.InDelta(t, unprojected.X, -5, delta)
		assert.InDelta(t, unprojected.Y, -5, delta)
	})
	t.Run("for scaled matrix", func(t *testing.T) {
		matrix := pixel.IM.Scaled(pixel.ZV, 2)
		unprojected := matrix.Unproject(pixel.V(4, 4))
		assert.InDelta(t, unprojected.X, 2, delta)
		assert.InDelta(t, unprojected.Y, 2, delta)
	})
	t.Run("for rotated and moved matrix", func(t *testing.T) {
		matrix := pixel.IM.Rotated(pixel.ZV, math.Pi/2).Moved(pixel.V(1, 1))
		unprojected := matrix.Unproject(pixel.V(1, 2))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for singular matrix", func(t *testing.T) {
		matrix := pixel.Matrix{0, 0, 0, 0, 0, 0}
		unprojected := matrix.Unproject(pixel.ZV)
		assert.True(t, math.IsNaN(unprojected.X))
		assert.True(t, math.IsNaN(unprojected.Y))
	})
}
