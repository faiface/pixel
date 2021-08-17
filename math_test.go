package pixel_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/faiface/pixel"
)

// closeEnough will shift the decimal point by the accuracy required, truncates the results and compares them.
// Effectively this compares two floats to a given decimal point.
//  Example:
//  closeEnough(100.125342432, 100.125, 2) == true
//  closeEnough(math.Pi, 3.14, 2) == true
//  closeEnough(0.1234, 0.1245, 3) == false
func closeEnough(got, expected float64, decimalAccuracy int) bool {
	gotShifted := got * math.Pow10(decimalAccuracy)
	expectedShifted := expected * math.Pow10(decimalAccuracy)

	return math.Trunc(gotShifted) == math.Trunc(expectedShifted)
}

type clampTest struct {
	number   float64
	min      float64
	max      float64
	expected float64
}

func TestClamp(t *testing.T) {
	tests := []clampTest{
		{number: 1, min: 0, max: 5, expected: 1},
		{number: 2, min: 0, max: 5, expected: 2},
		{number: 8, min: 0, max: 5, expected: 5},
		{number: -5, min: 0, max: 5, expected: 0},
		{number: -5, min: -4, max: 5, expected: -4},
	}

	for _, tc := range tests {
		result := pixel.Clamp(tc.number, tc.min, tc.max)
		if result != tc.expected {
			t.Error(fmt.Sprintf("Clamping %v with min %v and max %v should have given %v, but gave %v", tc.number, tc.min, tc.max, tc.expected, result))
		}
	}
}
