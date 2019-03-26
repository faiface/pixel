package pixel_test

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestNewBatch(t *testing.T) {
	type args struct {
		container pixel.Triangles
		pic       pixel.Picture
	}
	tests := []struct {
		name string
		args args
		want *pixel.Batch
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.NewBatch(tt.args.container, tt.args.pic); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatch_Dirty(t *testing.T) {
	tests := []struct {
		name string
		b    *pixel.Batch
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Dirty()
		})
	}
}

func TestBatch_Clear(t *testing.T) {
	tests := []struct {
		name string
		b    *pixel.Batch
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Clear()
		})
	}
}

func TestBatch_Draw(t *testing.T) {
	type args struct {
		t pixel.Target
	}
	tests := []struct {
		name string
		b    *pixel.Batch
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.Draw(tt.args.t)
		})
	}
}

func TestBatch_SetMatrix(t *testing.T) {
	type args struct {
		m pixel.Matrix
	}
	tests := []struct {
		name string
		b    *pixel.Batch
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.SetMatrix(tt.args.m)
		})
	}
}

func TestBatch_SetColorMask(t *testing.T) {
	type args struct {
		c color.Color
	}
	tests := []struct {
		name string
		b    *pixel.Batch
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.b.SetColorMask(tt.args.c)
		})
	}
}

func TestBatch_MakeTriangles(t *testing.T) {
	type args struct {
		t pixel.Triangles
	}
	tests := []struct {
		name string
		b    *pixel.Batch
		args args
		want pixel.TargetTriangles
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.MakeTriangles(tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Batch.MakeTriangles() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatch_MakePicture(t *testing.T) {
	type args struct {
		p pixel.Picture
	}
	tests := []struct {
		name string
		b    *pixel.Batch
		args args
		want pixel.TargetPicture
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.b.MakePicture(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Batch.MakePicture() = %v, want %v", got, tt.want)
			}
		})
	}
}
