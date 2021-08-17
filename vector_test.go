package pixel_test

import (
	"fmt"
	"testing"

	"github.com/faiface/pixel"
)

type floorTest struct {
	input    pixel.Vec
	expected pixel.Vec
}

func TestFloor(t *testing.T) {
	tests := []floorTest{
		{input: pixel.V(4.50, 6.70), expected: pixel.V(4, 6)},
		{input: pixel.V(9.0, 6.70), expected: pixel.V(9, 6)},
	}

	for _, tc := range tests {
		result := tc.input.Floor()
		if result != tc.expected {
			t.Error(fmt.Sprintf("Expected %v but got %v", tc.expected, result))
		}
	}
}
