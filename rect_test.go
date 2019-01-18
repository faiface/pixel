package pixel_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

var (
	squareAroundOrigin    = pixel.R(-10, -10, 10, 10)
	squareAround2020      = pixel.R(10, 10, 30, 30)
	rectangleAroundOrigin = pixel.R(-20, -10, 20, 10)
	rectangleAround2020   = pixel.R(0, 10, 40, 30)
)

func TestR(t *testing.T) {
	type args struct {
		minX float64
		minY float64
		maxX float64
		maxY float64
	}
	tests := []struct {
		name string
		args args
		want pixel.Rect
	}{
		{
			name: "R(): square around origin",
			args: args{minX: -10, minY: -10, maxX: 10, maxY: 10},
			want: squareAroundOrigin,
		},
		{
			name: "R(): square around 20 20",
			args: args{minX: 10, minY: 10, maxX: 30, maxY: 30},
			want: squareAround2020,
		},
		{
			name: "R(): rectangle around origin",
			args: args{minX: -20, minY: -10, maxX: 20, maxY: 10},
			want: rectangleAroundOrigin,
		},
		{
			name: "R(): square around 20 20",
			args: args{minX: 0, minY: 10, maxX: 40, maxY: 30},
			want: rectangleAround2020,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.R(tt.args.minX, tt.args.minY, tt.args.maxX, tt.args.maxY); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("R() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_String(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want string
	}{
		{
			name: "Rect.String(): square around origin",
			r:    squareAroundOrigin,
			want: "Rect(-10, -10, 10, 10)",
		},
		{
			name: "Rect.String(): square around 20 20",
			r:    squareAround2020,
			want: "Rect(10, 10, 30, 30)",
		},
		{
			name: "Rect.String(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: "Rect(-20, -10, 20, 10)",
		},
		{
			name: "Rect.String(): square around 20 20",
			r:    rectangleAround2020,
			want: "Rect(0, 10, 40, 30)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Rect.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Norm(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want pixel.Rect
	}{
		{
			name: "Rect.Norm(): square around origin",
			r:    squareAroundOrigin,
			want: squareAroundOrigin,
		},
		{
			name: "Rect.Norm(): square around 20 20",
			r:    squareAround2020,
			want: squareAround2020,
		},
		{
			name: "Rect.Norm(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: rectangleAroundOrigin,
		},
		{
			name: "Rect.Norm(): square around 20 20",
			r:    rectangleAround2020,
			want: rectangleAround2020,
		},
		{
			name: "Rect.Norm(): square around origin unnormalized",
			r:    pixel.Rect{Min: squareAroundOrigin.Max, Max: squareAroundOrigin.Min},
			want: squareAroundOrigin,
		},
		{
			name: "Rect.Norm(): square around 20 20 unnormalized",
			r:    pixel.Rect{Min: squareAround2020.Max, Max: squareAround2020.Min},
			want: squareAround2020,
		},
		{
			name: "Rect.Norm(): rectangle around origin unnormalized",
			r:    pixel.Rect{Min: rectangleAroundOrigin.Max, Max: rectangleAroundOrigin.Min},
			want: rectangleAroundOrigin,
		},
		{
			name: "Rect.Norm(): square around 20 20 unnormalized",
			r:    pixel.Rect{Min: rectangleAround2020.Max, Max: rectangleAround2020.Min},
			want: rectangleAround2020,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Norm(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Norm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_W(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want float64
	}{
		{
			name: "Rect.W(): square around origin",
			r:    squareAroundOrigin,
			want: 20,
		},
		{
			name: "Rect.W(): square around 20 20",
			r:    squareAround2020,
			want: 20,
		},
		{
			name: "Rect.W(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: 40,
		},
		{
			name: "Rect.W(): square around 20 20",
			r:    rectangleAround2020,
			want: 40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.W(); got != tt.want {
				t.Errorf("Rect.W() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_H(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want float64
	}{
		{
			name: "Rect.H(): square around origin",
			r:    squareAroundOrigin,
			want: 20,
		},
		{
			name: "Rect.H(): square around 20 20",
			r:    squareAround2020,
			want: 20,
		},
		{
			name: "Rect.H(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: 20,
		},
		{
			name: "Rect.H(): square around 20 20",
			r:    rectangleAround2020,
			want: 20,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.H(); got != tt.want {
				t.Errorf("Rect.H() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Size(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want pixel.Vec
	}{
		{
			name: "Rect.Size(): square around origin",
			r:    squareAroundOrigin,
			want: pixel.V(20, 20),
		},
		{
			name: "Rect.Size(): square around 20 20",
			r:    squareAround2020,
			want: pixel.V(20, 20),
		},
		{
			name: "Rect.Size(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: pixel.V(40, 20),
		},
		{
			name: "Rect.Size(): square around 20 20",
			r:    rectangleAround2020,
			want: pixel.V(40, 20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Size(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Area(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want float64
	}{
		{
			name: "Rect.Area(): square around origin",
			r:    squareAroundOrigin,
			want: 400,
		},
		{
			name: "Rect.Area(): square around 20 20",
			r:    squareAround2020,
			want: 400,
		},
		{
			name: "Rect.Area(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: 800,
		},
		{
			name: "Rect.Area(): square around 20 20",
			r:    rectangleAround2020,
			want: 800,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Area(); got != tt.want {
				t.Errorf("Rect.Area() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Center(t *testing.T) {
	tests := []struct {
		name string
		r    pixel.Rect
		want pixel.Vec
	}{
		{
			name: "Rect.Center(): square around origin",
			r:    squareAroundOrigin,
			want: pixel.V(0, 0),
		},
		{
			name: "Rect.Center(): square around 20 20",
			r:    squareAround2020,
			want: pixel.V(20, 20),
		},
		{
			name: "Rect.Center(): rectangle around origin",
			r:    rectangleAroundOrigin,
			want: pixel.V(0, 0),
		},
		{
			name: "Rect.Center(): square around 20 20",
			r:    rectangleAround2020,
			want: pixel.V(20, 20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Center(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Center() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Moved(t *testing.T) {
	positiveShift := pixel.V(10, 10)
	negativeShift := pixel.V(-10, -10)

	type args struct {
		delta pixel.Vec
	}
	tests := []struct {
		name string
		r    pixel.Rect
		args args
		want pixel.Rect
	}{
		{
			name: "Rect.Moved(): positive shift",
			r:    squareAroundOrigin,
			args: args{delta: positiveShift},
			want: pixel.R(0, 0, 20, 20),
		},
		{
			name: "Rect.Moved(): negative shift",
			r:    squareAroundOrigin,
			args: args{delta: negativeShift},
			want: pixel.R(-20, -20, 0, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Moved(tt.args.delta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Moved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Resized(t *testing.T) {
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

func TestRect_ResizedMin(t *testing.T) {
	grow := pixel.V(5, 5)
	shrink := pixel.V(-5, -5)

	type args struct {
		size pixel.Vec
	}
	tests := []struct {
		name string
		r    pixel.Rect
		args args
		want pixel.Rect
	}{
		{
			name: "Rect.ResizedMin(): square around origin - growing",
			r:    squareAroundOrigin,
			args: args{size: grow},
			want: pixel.R(-10, -10, -5, -5),
		},
		{
			name: "Rect.ResizedMin(): square around 20 20 - growing",
			r:    squareAround2020,
			args: args{size: grow},
			want: pixel.R(10, 10, 15, 15),
		},
		{
			name: "Rect.ResizedMin(): rectangle around origin - growing",
			r:    rectangleAroundOrigin,
			args: args{size: grow},
			want: pixel.R(-20, -10, -15, -5),
		},
		{
			name: "Rect.ResizedMin(): square around 20 20 - growing",
			r:    rectangleAround2020,
			args: args{size: grow},
			want: pixel.R(0, 10, 5, 15),
		},
		{
			name: "Rect.ResizedMin(): square around origin - growing",
			r:    squareAroundOrigin,
			args: args{size: shrink},
			want: pixel.R(-10, -10, -15, -15),
		},
		{
			name: "Rect.ResizedMin(): square around 20 20 - growing",
			r:    squareAround2020,
			args: args{size: shrink},
			want: pixel.R(10, 10, 5, 5),
		},
		{
			name: "Rect.ResizedMin(): rectangle around origin - growing",
			r:    rectangleAroundOrigin,
			args: args{size: shrink},
			want: pixel.R(-20, -10, -25, -15),
		},
		{
			name: "Rect.ResizedMin(): square around 20 20 - growing",
			r:    rectangleAround2020,
			args: args{size: shrink},
			want: pixel.R(0, 10, -5, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.ResizedMin(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.ResizedMin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Contains(t *testing.T) {
	type args struct {
		u pixel.Vec
	}
	tests := []struct {
		name string
		r    pixel.Rect
		args args
		want bool
	}{
		{
			name: "Rect.Contains(): square and contained vector",
			r:    squareAroundOrigin,
			args: args{u: pixel.V(-5, 5)},
			want: true,
		},
		{
			name: "Rect.Contains(): square and seperate vector",
			r:    squareAroundOrigin,
			args: args{u: pixel.V(50, 55)},
			want: false,
		},
		{
			name: "Rect.Contains(): square and overlapping vector",
			r:    squareAroundOrigin,
			args: args{u: pixel.V(0, 50)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Contains(tt.args.u); got != tt.want {
				t.Errorf("Rect.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Union(t *testing.T) {
	type args struct {
		s pixel.Rect
	}
	tests := []struct {
		name string
		r    pixel.Rect
		args args
		want pixel.Rect
	}{
		{
			name: "Rect.Union(): seperate squares",
			r:    squareAroundOrigin,
			args: args{s: squareAround2020},
			want: pixel.R(-10, -10, 30, 30),
		},
		{
			name: "Rect.Union(): overlapping squares",
			r:    squareAroundOrigin,
			args: args{s: pixel.R(0, 0, 20, 20)},
			want: pixel.R(-10, -10, 20, 20),
		},
		{
			name: "Rect.Union(): square within a square",
			r:    squareAroundOrigin,
			args: args{s: pixel.R(-5, -5, 5, 5)},
			want: squareAroundOrigin,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Union(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_Intersect(t *testing.T) {
	type args struct {
		s pixel.Rect
	}
	tests := []struct {
		name string
		r    pixel.Rect
		args args
		want pixel.Rect
	}{
		{
			name: "Rect.Union(): seperate squares",
			r:    squareAroundOrigin,
			args: args{s: squareAround2020},
			want: pixel.R(0, 0, 0, 0),
		},
		{
			name: "Rect.Union(): overlapping squares",
			r:    squareAroundOrigin,
			args: args{s: pixel.R(0, 0, 20, 20)},
			want: pixel.R(0, 0, 10, 10),
		},
		{
			name: "Rect.Union(): square within a square",
			r:    squareAroundOrigin,
			args: args{s: pixel.R(-5, -5, 5, 5)},
			want: pixel.R(-5, -5, 5, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Intersect(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rect.Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}
