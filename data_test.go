package pixel_test

import (
	"image"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestMakeTrianglesData(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name string
		args args
		want *pixel.TrianglesData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.MakeTrianglesData(tt.args.len); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakeTrianglesData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_Len(t *testing.T) {
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.Len(); got != tt.want {
				t.Errorf("TrianglesData.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_SetLen(t *testing.T) {
	type args struct {
		len int
	}
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.td.SetLen(tt.args.len)
		})
	}
}

func TestTrianglesData_Slice(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		args args
		want pixel.Triangles
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.Slice(tt.args.i, tt.args.j); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrianglesData.Slice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_Update(t *testing.T) {
	type args struct {
		t pixel.Triangles
	}
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.td.Update(tt.args.t)
		})
	}
}

func TestTrianglesData_Copy(t *testing.T) {
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		want pixel.Triangles
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.Copy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrianglesData.Copy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_Position(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		args args
		want pixel.Vec
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.Position(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrianglesData.Position() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_Color(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name string
		td   *pixel.TrianglesData
		args args
		want pixel.RGBA
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.td.Color(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrianglesData.Color() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrianglesData_Picture(t *testing.T) {
	type args struct {
		i int
	}
	tests := []struct {
		name          string
		td            *pixel.TrianglesData
		args          args
		wantPic       pixel.Vec
		wantIntensity float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPic, gotIntensity := tt.td.Picture(tt.args.i)
			if !reflect.DeepEqual(gotPic, tt.wantPic) {
				t.Errorf("TrianglesData.Picture() gotPic = %v, want %v", gotPic, tt.wantPic)
			}
			if gotIntensity != tt.wantIntensity {
				t.Errorf("TrianglesData.Picture() gotIntensity = %v, want %v", gotIntensity, tt.wantIntensity)
			}
		})
	}
}

func TestMakePictureData(t *testing.T) {
	type args struct {
		rect pixel.Rect
	}
	tests := []struct {
		name string
		args args
		want *pixel.PictureData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.MakePictureData(tt.args.rect); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MakePictureData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureDataFromImage(t *testing.T) {
	type args struct {
		img image.Image
	}
	tests := []struct {
		name string
		args args
		want *pixel.PictureData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.PictureDataFromImage(tt.args.img); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PictureDataFromImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureDataFromPicture(t *testing.T) {
	type args struct {
		pic pixel.Picture
	}
	tests := []struct {
		name string
		args args
		want *pixel.PictureData
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.PictureDataFromPicture(tt.args.pic); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PictureDataFromPicture() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureData_Image(t *testing.T) {
	tests := []struct {
		name string
		pd   *pixel.PictureData
		want *image.RGBA
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pd.Image(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PictureData.Image() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureData_Index(t *testing.T) {
	type args struct {
		at pixel.Vec
	}
	tests := []struct {
		name string
		pd   *pixel.PictureData
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pd.Index(tt.args.at); got != tt.want {
				t.Errorf("PictureData.Index() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureData_Bounds(t *testing.T) {
	tests := []struct {
		name string
		pd   *pixel.PictureData
		want pixel.Rect
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pd.Bounds(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PictureData.Bounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPictureData_Color(t *testing.T) {
	type args struct {
		at pixel.Vec
	}
	tests := []struct {
		name string
		pd   *pixel.PictureData
		args args
		want pixel.RGBA
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pd.Color(tt.args.at); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PictureData.Color() = %v, want %v", got, tt.want)
			}
		})
	}
}
