package pixel_test

import (
	"testing"

	"github.com/faiface/pixel"
)

type rectTestTransform struct {
	name string
	f    func(pixel.Rect) pixel.Rect
}

func TestClamp(t *testing.T) {
	type args struct {
		x   float64
		min float64
		max float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "Clamp: x < min < max",
			args: args{x: 1, min: 2, max: 3},
			want: 2,
		},
		{
			name: "Clamp: min < x < max",
			args: args{x: 2, min: 1, max: 3},
			want: 2,
		},
		{
			name: "Clamp: min < max < x",
			args: args{x: 3, min: 1, max: 2},
			want: 2,
		},
		{
			name: "Clamp: x > min > max",
			args: args{x: 3, min: 2, max: 1},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.Clamp(tt.args.x, tt.args.min, tt.args.max); got != tt.want {
				t.Errorf("Clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}
