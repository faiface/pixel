package pixel_test

import (
	"image"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkSpriteDrawBatch(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pixel.R(0, 0, 64, 64))
	batch := pixel.NewBatch(&pixel.TrianglesData{}, pic)
	for i := 0; i < b.N; i++ {
		sprite.Draw(batch, pixel.IM)
	}
}

func TestDrawer_Dirty(t *testing.T) {
	tests := []struct {
		name string
		d    *pixel.Drawer
	}{
		{
			name: "Drawer.Dirty",
			d:    &pixel.Drawer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.Dirty()
		})
	}
}

type targetMock struct {
	makeTrianglesCount int
	makePictureCount   int
}

func (t *targetMock) MakePicture(_ pixel.Picture) pixel.TargetPicture {
	t.makePictureCount++
	return targetPictureMock{}
}

func (t *targetMock) MakeTriangles(_ pixel.Triangles) pixel.TargetTriangles {
	t.makeTrianglesCount++
	return targetTrianglesMock{}
}

type targetTrianglesMock struct{}

func (targetTrianglesMock) Len() int {
	return 0
}

func (targetTrianglesMock) SetLen(_ int) {

}

func (targetTrianglesMock) Slice(_, _ int) pixel.Triangles {
	return nil
}

func (targetTrianglesMock) Update(_ pixel.Triangles) {
}

func (targetTrianglesMock) Copy() pixel.Triangles {
	return nil
}

func (targetTrianglesMock) Draw() {
}

type targetPictureMock struct{}

func (targetPictureMock) Bounds() pixel.Rect {
	return pixel.R(0, 0, 0, 0)
}

func (targetPictureMock) Draw(_ pixel.TargetTriangles) {

}

func TestDrawer_Draw(t *testing.T) {
	type args struct {
		t pixel.Target
	}
	tests := []struct {
		name              string
		d                 *pixel.Drawer
		args              args
		wantPictureCount  int
		wantTriangleCount int
	}{
		{
			name:              "Drawer.Draw: nil triangles",
			d:                 &pixel.Drawer{},
			args:              args{t: &targetMock{}},
			wantPictureCount:  0,
			wantTriangleCount: 0,
		},
		{
			name:              "Drawer.Draw: non-nil triangles",
			d:                 &pixel.Drawer{Triangles: pixel.MakeTrianglesData(1)},
			args:              args{t: &targetMock{}},
			wantPictureCount:  0,
			wantTriangleCount: 1,
		},
		{
			name: "Drawer.Draw: non-nil picture",
			d: &pixel.Drawer{
				Triangles: pixel.MakeTrianglesData(1),
				Picture:   pixel.MakePictureData(pixel.R(0, 0, 0, 0)),
			},
			args:              args{t: &targetMock{}},
			wantPictureCount:  1,
			wantTriangleCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.Draw(tt.args.t)

			target := tt.args.t.(*targetMock)

			if tt.wantPictureCount != target.makePictureCount {
				t.Fatalf("MakePicture not called. Expected %d, got: %d", tt.wantPictureCount, target.makePictureCount)
			}
			if tt.wantTriangleCount != target.makeTrianglesCount {
				t.Fatalf("MakeTriangles not called. Expected %d, got: %d", tt.wantTriangleCount, target.makeTrianglesCount)
			}
		})
	}
}
