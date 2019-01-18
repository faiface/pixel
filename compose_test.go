package pixel_test

import (
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestComposeMethod_Compose(t *testing.T) {
	type args struct {
		a pixel.RGBA
		b pixel.RGBA
	}

	a := pixel.RGBA{R: 200, G: 200, B: 200, A: .8}
	b := pixel.RGBA{R: 100, G: 100, B: 100, A: .5}
	c := pixel.RGBA{R: 200, G: 200, B: 200, A: .5}

	tests := []struct {
		name string
		cm   pixel.ComposeMethod
		args args
		want pixel.RGBA
	}{
		{
			name: "ComposeMethod.Compose: ComposeOver",
			cm:   pixel.ComposeOver,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 220, G: 220, B: 220, A: .9},
		},
		{
			name: "ComposeMethod.Compose: ComposeIn",
			cm:   pixel.ComposeIn,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 100, G: 100, B: 100, A: .4},
		},
		{
			name: "ComposeMethod.Compose: ComposeOut",
			cm:   pixel.ComposeOut,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 100, G: 100, B: 100, A: .4},
		},
		{
			name: "ComposeMethod.Compose: ComposeAtop",
			cm:   pixel.ComposeAtop,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 120, G: 120, B: 120, A: .5},
		},
		{
			name: "ComposeMethod.Compose: ComposeRover",
			cm:   pixel.ComposeRover,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 200, G: 200, B: 200, A: .9},
		},
		{
			name: "ComposeMethod.Compose: ComposeRin",
			cm:   pixel.ComposeRin,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 80, G: 80, B: 80, A: .4},
		},
		{
			name: "ComposeMethod.Compose: ComposeRout",
			cm:   pixel.ComposeRout,
			// Using 'c' here to make the "want"ed RGBA rational
			args: args{a: c, b: b},
			want: pixel.RGBA{R: 50, G: 50, B: 50, A: .25},
		},
		{
			name: "ComposeMethod.Compose: ComposeRatop",
			cm:   pixel.ComposeRatop,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 180, G: 180, B: 180, A: .8},
		},
		{
			name: "ComposeMethod.Compose: ComposeXor",
			cm:   pixel.ComposeXor,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 120, G: 120, B: 120, A: .5},
		},
		{
			name: "ComposeMethod.Compose: ComposePlus",
			cm:   pixel.ComposePlus,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 300, G: 300, B: 300, A: 1.3},
		},
		{
			name: "ComposeMethod.Compose: ComposeCopy",
			cm:   pixel.ComposeCopy,
			args: args{a: a, b: b},
			want: pixel.RGBA{R: 200, G: 200, B: 200, A: .8},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cm.Compose(tt.args.a, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ComposeMethod.Compose() = %v, want %v", got, tt.want)
			}
		})
	}
}
