package pixel_test

import (
	"fmt"
	"image/color"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkColorToRGBA(b *testing.B) {
	types := []color.Color{
		color.NRGBA{R: 124, G: 14, B: 230, A: 42}, // slowest
		color.RGBA{R: 62, G: 32, B: 14, A: 63},    // faster
		pixel.RGB(0.8, 0.2, 0.5).Scaled(0.712),    // fastest
	}
	for _, col := range types {
		b.Run(fmt.Sprintf("From %T", col), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = pixel.ToRGBA(col)
			}
		})
	}
}

func TestRGB(t *testing.T) {
	type args struct {
		r float64
		g float64
		b float64
	}
	tests := []struct {
		name string
		args args
		want pixel.RGBA
	}{
		{
			name: "RBG: create black",
			args: args{r: 0, g: 0, b: 0},
			want: pixel.RGBA{R: 0, G: 0, B: 0, A: 1},
		},
		{
			name: "RBG: create white",
			args: args{r: 1, g: 1, b: 1},
			want: pixel.RGBA{R: 1, G: 1, B: 1, A: 1},
		},
		{
			name: "RBG: create nonsense",
			args: args{r: 500, g: 500, b: 500},
			want: pixel.RGBA{R: 500, G: 500, B: 500, A: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.RGB(tt.args.r, tt.args.g, tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RGB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlpha(t *testing.T) {
	type args struct {
		a float64
	}
	tests := []struct {
		name string
		args args
		want pixel.RGBA
	}{
		{
			name: "Alpha: transparent",
			args: args{a: 0},
			want: pixel.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name: "Alpha: obaque",
			args: args{a: 1},
			want: pixel.RGBA{R: 1, G: 1, B: 1, A: 1},
		},
		{
			name: "Alpha: nonsense",
			args: args{a: 1024},
			want: pixel.RGBA{R: 1024, G: 1024, B: 1024, A: 1024},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.Alpha(tt.args.a); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Alpha() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBA_Add(t *testing.T) {
	type fields struct {
		R float64
		G float64
		B float64
		A float64
	}
	type args struct {
		d pixel.RGBA
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.RGBA
	}{
		{
			name:   "RGBA.Add: add to black",
			fields: fields{R: 0, G: 0, B: 0, A: 1},
			args:   args{d: pixel.RGBA{R: 50, G: 50, B: 50, A: 0}},
			want:   pixel.RGBA{R: 50, G: 50, B: 50, A: 1},
		},
		{
			name:   "RGBA.Add: add to white",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   args{d: pixel.RGBA{R: 1, G: 1, B: 1, A: 1}},
			want:   pixel.RGBA{R: 2, G: 2, B: 2, A: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.RGBA{
				R: tt.fields.R,
				G: tt.fields.G,
				B: tt.fields.B,
				A: tt.fields.A,
			}
			if got := c.Add(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RGBA.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBA_Sub(t *testing.T) {
	type fields struct {
		R float64
		G float64
		B float64
		A float64
	}
	type args struct {
		d pixel.RGBA
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.RGBA
	}{
		{
			name:   "RGBA.Sub: sub from white",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   args{d: pixel.RGBA{R: .5, G: .5, B: .5, A: 0}},
			want:   pixel.RGBA{R: .5, G: .5, B: .5, A: 1},
		},
		{
			name:   "RGBA.Sub: sub from black",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			args:   args{d: pixel.RGBA{R: 1, G: 1, B: 1, A: 1}},
			want:   pixel.RGBA{R: -1, G: -1, B: -1, A: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.RGBA{
				R: tt.fields.R,
				G: tt.fields.G,
				B: tt.fields.B,
				A: tt.fields.A,
			}
			if got := c.Sub(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RGBA.Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBA_Mul(t *testing.T) {
	type fields struct {
		R float64
		G float64
		B float64
		A float64
	}
	type args struct {
		d pixel.RGBA
	}

	greaterThanOne := args{d: pixel.RGBA{R: 2, G: 3, B: 4, A: 5}}
	lessThanOne := args{d: pixel.RGBA{R: .2, G: .3, B: .4, A: .5}}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.RGBA
	}{
		{
			name:   "RGBA.Mul: multiply black by >1",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			args:   greaterThanOne,
			want:   pixel.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:   "RGBA.Mul: multiply white by >1",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   greaterThanOne,
			want:   pixel.RGBA{R: 2, G: 3, B: 4, A: 5},
		},
		{
			name:   "RGBA.Mul: multiply black by <1",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			args:   lessThanOne,
			want:   pixel.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:   "RGBA.Mul: multiply white by <1",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   lessThanOne,
			want:   pixel.RGBA{R: .2, G: .3, B: .4, A: .5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.RGBA{
				R: tt.fields.R,
				G: tt.fields.G,
				B: tt.fields.B,
				A: tt.fields.A,
			}
			if got := c.Mul(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RGBA.Mul() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBA_Scaled(t *testing.T) {
	type fields struct {
		R float64
		G float64
		B float64
		A float64
	}
	type args struct {
		scale float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.RGBA
	}{
		{
			name:   "RBGA.Scaled: black <1",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			args:   args{scale: 0.5},
			want:   pixel.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:   "RBGA.Scaled: black <1",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   args{scale: 0.5},
			want:   pixel.RGBA{R: .5, G: .5, B: .5, A: .5},
		},
		{
			name:   "RBGA.Scaled: black >1",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			args:   args{scale: 2},
			want:   pixel.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
		{
			name:   "RBGA.Scaled: black >1",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			args:   args{scale: 2},
			want:   pixel.RGBA{R: 2, G: 2, B: 2, A: 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.RGBA{
				R: tt.fields.R,
				G: tt.fields.G,
				B: tt.fields.B,
				A: tt.fields.A,
			}
			if got := c.Scaled(tt.args.scale); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RGBA.Scaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRGBA_RGBA(t *testing.T) {
	type fields struct {
		R float64
		G float64
		B float64
		A float64
	}
	tests := []struct {
		name   string
		fields fields
		wantR  uint32
		wantG  uint32
		wantB  uint32
		wantA  uint32
	}{
		{
			name:   "RGBA.RGBA: black",
			fields: fields{R: 0, G: 0, B: 0, A: 0},
			wantR:  0,
			wantG:  0,
			wantB:  0,
			wantA:  0,
		},
		{
			name:   "RGBA.RGBA: white",
			fields: fields{R: 1, G: 1, B: 1, A: 1},
			wantR:  65535,
			wantG:  65535,
			wantB:  65535,
			wantA:  65535,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.RGBA{
				R: tt.fields.R,
				G: tt.fields.G,
				B: tt.fields.B,
				A: tt.fields.A,
			}
			gotR, gotG, gotB, gotA := c.RGBA()
			if gotR != tt.wantR {
				t.Errorf("RGBA.RGBA() gotR = %v, want %v", gotR, tt.wantR)
			}
			if gotG != tt.wantG {
				t.Errorf("RGBA.RGBA() gotG = %v, want %v", gotG, tt.wantG)
			}
			if gotB != tt.wantB {
				t.Errorf("RGBA.RGBA() gotB = %v, want %v", gotB, tt.wantB)
			}
			if gotA != tt.wantA {
				t.Errorf("RGBA.RGBA() gotA = %v, want %v", gotA, tt.wantA)
			}
		})
	}
}

func TestToRGBA(t *testing.T) {
	type args struct {
		c color.Color
	}
	tests := []struct {
		name string
		args args
		want pixel.RGBA
	}{
		{
			name: "ToRGBA: black",
			args: args{c: color.Black},
			want: pixel.RGBA{R: 0, G: 0, B: 0, A: 1},
		},
		{
			name: "ToRGBA: white",
			args: args{c: color.White},
			want: pixel.RGBA{R: 1, G: 1, B: 1, A: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.ToRGBA(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToRGBA() = %v, want %v", got, tt.want)
			}
		})
	}
}
