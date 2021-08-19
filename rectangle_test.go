package pixel_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func TestRect_Resize(t *testing.T) {
	type rectTestTransform struct {
		name string
		f    func(pixel.Rect) pixel.Rect
	}

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

func TestRect_Edges(t *testing.T) {
	type fields struct {
		Min pixel.Vec
		Max pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   [4]pixel.Line
	}{
		{
			name:   "Get edges",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			want: [4]pixel.Line{
				pixel.L(pixel.V(0, 0), pixel.V(0, 10)),
				pixel.L(pixel.V(0, 10), pixel.V(10, 10)),
				pixel.L(pixel.V(10, 10), pixel.V(10, 0)),
				pixel.L(pixel.V(10, 0), pixel.V(0, 0)),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := pixel.Rect{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			if got := r.Edges(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Edges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Vertices(t *testing.T) {
	type fields struct {
		Min pixel.Vec
		Max pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   [4]pixel.Vec
	}{
		{
			name:   "Get corners",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			want: [4]pixel.Vec{
				pixel.V(0, 0),
				pixel.V(0, 10),
				pixel.V(10, 10),
				pixel.V(10, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := pixel.Rect{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			if got := r.Vertices(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Vertices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_IntersectCircle(t *testing.T) {
	type fields struct {
		Min pixel.Vec
		Max pixel.Vec
	}
	type args struct {
		c pixel.Circle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Vec
	}{
		{
			name:   "Rect.IntersectCircle(): no overlap",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(50, 50), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): circle contains rect",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, 5), 10)},
			want:   pixel.V(-15, 0),
		},
		{
			name:   "Rect.IntersectCircle(): rect contains circle",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, 5), 1)},
			want:   pixel.V(-6, 0),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps bottom-left corner",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(-0.5, -0.5), 1)},
			want:   pixel.V(-0.2, -0.2),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps top-left corner",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(-0.5, 10.5), 1)},
			want:   pixel.V(-0.2, 0.2),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps bottom-right corner",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(10.5, -0.5), 1)},
			want:   pixel.V(0.2, -0.2),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps top-right corner",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(10.5, 10.5), 1)},
			want:   pixel.V(0.2, 0.2),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps two corners",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(0, 5), 6)},
			want:   pixel.V(6, 0),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps left edge",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(0, 5), 1)},
			want:   pixel.V(1, 0),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps bottom edge",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, 0), 1)},
			want:   pixel.V(0, 1),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps right edge",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(10, 5), 1)},
			want:   pixel.V(-1, 0),
		},
		{
			name:   "Rect.IntersectCircle(): circle overlaps top edge",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, 10), 1)},
			want:   pixel.V(0, -1),
		},
		{
			name:   "Rect.IntersectCircle(): edge is tangent of left side",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(-1, 5), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): edge is tangent of top side",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, -1), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): circle above rectangle",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, 12), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): circle below rectangle",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(5, -2), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): circle left of rectangle",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(-1, 5), 1)},
			want:   pixel.ZV,
		},
		{
			name:   "Rect.IntersectCircle(): circle right of rectangle",
			fields: fields{Min: pixel.ZV, Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(11, 5), 1)},
			want:   pixel.ZV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := pixel.Rect{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			got := r.IntersectCircle(tt.args.c)
			if !closeEnough(got.X, tt.want.X, 2) || !closeEnough(got.Y, tt.want.Y, 2) {
				t.Errorf("Rect.IntersectCircle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_IntersectionPoints(t *testing.T) {
	type fields struct {
		Min pixel.Vec
		Max pixel.Vec
	}
	type args struct {
		l pixel.Line
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []pixel.Vec
	}{
		{
			name:   "No intersection points",
			fields: fields{Min: pixel.V(1, 1), Max: pixel.V(5, 5)},
			args:   args{l: pixel.L(pixel.V(-5, 0), pixel.V(-2, 2))},
			want:   []pixel.Vec{},
		},
		{
			name:   "One intersection point",
			fields: fields{Min: pixel.V(1, 1), Max: pixel.V(5, 5)},
			args:   args{l: pixel.L(pixel.V(2, 0), pixel.V(2, 3))},
			want:   []pixel.Vec{pixel.V(2, 1)},
		},
		{
			name:   "Two intersection points",
			fields: fields{Min: pixel.V(1, 1), Max: pixel.V(5, 5)},
			args:   args{l: pixel.L(pixel.V(0, 2), pixel.V(6, 2))},
			want:   []pixel.Vec{pixel.V(1, 2), pixel.V(5, 2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := pixel.Rect{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			if got := r.IntersectionPoints(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.IntersectPoints() = %v, want %v", got, tt.want)
			}
		})
	}
}

var rectIntTests = []struct {
	name      string
	r1, r2    pixel.Rect
	want      pixel.Rect
	intersect bool
}{
	{
		name: "Nothing touching",
		r1:   pixel.R(0, 0, 10, 10),
		r2:   pixel.R(21, 21, 40, 40),
		want: pixel.ZR,
	},
	{
		name: "Edge touching",
		r1:   pixel.R(0, 0, 10, 10),
		r2:   pixel.R(10, 10, 20, 20),
		want: pixel.ZR,
	},
	{
		name:      "Bit of overlap",
		r1:        pixel.R(0, 0, 10, 10),
		r2:        pixel.R(0, 9, 20, 20),
		want:      pixel.R(0, 9, 10, 10),
		intersect: true,
	},
	{
		name:      "Fully overlapped",
		r1:        pixel.R(0, 0, 10, 10),
		r2:        pixel.R(0, 0, 10, 10),
		want:      pixel.R(0, 0, 10, 10),
		intersect: true,
	},
}

func TestRect_Intersect(t *testing.T) {
	for _, tt := range rectIntTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r1.Intersect(tt.r2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Intersects(t *testing.T) {
	for _, tt := range rectIntTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r1.Intersects(tt.r2); got != tt.intersect {
				t.Errorf("Rect.Intersects() = %v, want %v", got, tt.want)
			}
		})
	}
}
