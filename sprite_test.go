package pixel_test

import (
	"image/color"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestNewSprite(t *testing.T) {
	type args struct {
		pic   pixel.Picture
		frame pixel.Rect
	}
	tests := []struct {
		name string
		args args
		want *pixel.Sprite
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.NewSprite(tt.args.pic, tt.args.frame); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSprite() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSprite_Set(t *testing.T) {
	type args struct {
		pic   pixel.Picture
		frame pixel.Rect
	}
	tests := []struct {
		name string
		s    *pixel.Sprite
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Set(tt.args.pic, tt.args.frame)
		})
	}
}

func TestSprite_Picture(t *testing.T) {
	tests := []struct {
		name string
		s    *pixel.Sprite
		want pixel.Picture
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Picture(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sprite.Picture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSprite_Frame(t *testing.T) {
	tests := []struct {
		name string
		s    *pixel.Sprite
		want pixel.Rect
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Frame(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sprite.Frame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSprite_Draw(t *testing.T) {
	type args struct {
		t      pixel.Target
		matrix pixel.Matrix
	}
	tests := []struct {
		name string
		s    *pixel.Sprite
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Draw(tt.args.t, tt.args.matrix)
		})
	}
}

func TestSprite_DrawColorMask(t *testing.T) {
	type args struct {
		t      pixel.Target
		matrix pixel.Matrix
		mask   color.Color
	}
	tests := []struct {
		name string
		s    *pixel.Sprite
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.DrawColorMask(tt.args.t, tt.args.matrix, tt.args.mask)
		})
	}
}
