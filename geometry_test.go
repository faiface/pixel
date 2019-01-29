package pixel_test

import (
	"fmt"
	"math"
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

func TestC(t *testing.T) {
	type args struct {
		radius float64
		center pixel.Vec
	}
	tests := []struct {
		name string
		args args
		want pixel.Circle
	}{
		{
			name: "C(): positive radius",
			args: args{radius: 10, center: pixel.V(0, 0)},
			want: pixel.Circle{Radius: 10, Center: pixel.V(0, 0)},
		},
		{
			name: "C(): zero radius",
			args: args{radius: 0, center: pixel.V(0, 0)},
			want: pixel.Circle{Radius: 0, Center: pixel.V(0, 0)},
		},
		{
			name: "C(): negative radius",
			args: args{radius: -5, center: pixel.V(0, 0)},
			want: pixel.Circle{Radius: -5, Center: pixel.V(0, 0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.C(tt.args.radius, tt.args.center); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("C() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_String(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Circle.String(): positive radius",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			want:   "Circle(10.00, Vec(0, 0))",
		},
		{
			name:   "Circle.String(): zero radius",
			fields: fields{radius: 0, center: pixel.V(0, 0)},
			want:   "Circle(0.00, Vec(0, 0))",
		},
		{
			name:   "Circle.String(): negative radius",
			fields: fields{radius: -5, center: pixel.V(0, 0)},
			want:   "Circle(-5.00, Vec(0, 0))",
		},
		{
			name:   "Circle.String(): irrational radius",
			fields: fields{radius: math.Pi, center: pixel.V(0, 0)},
			want:   "Circle(3.14, Vec(0, 0))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.String(); got != tt.want {
				t.Errorf("Circle.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Norm(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   pixel.Circle
	}{
		{
			name:   "Circle.Norm(): positive radius",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			want:   pixel.Circle{Radius: 10, Center: pixel.Vec{X: 0, Y: 0}},
		},
		{
			name:   "Circle.Norm(): zero radius",
			fields: fields{radius: 0, center: pixel.V(0, 0)},
			want:   pixel.Circle{Radius: 0, Center: pixel.Vec{X: 0, Y: 0}},
		},
		{
			name:   "Circle.Norm(): negative radius",
			fields: fields{radius: -5, center: pixel.V(0, 0)},
			want:   pixel.Circle{Radius: 5, Center: pixel.Vec{X: 0, Y: 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Norm(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Norm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Area(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "Circle.Area(): positive radius",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			want:   20 * math.Pi,
		},
		{
			name:   "Circle.Area(): zero radius",
			fields: fields{radius: 0, center: pixel.V(0, 0)},
			want:   0,
		},
		{
			name:   "Circle.Area(): negative radius",
			fields: fields{radius: -5, center: pixel.V(0, 0)},
			want:   -10 * math.Pi,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Area(); got != tt.want {
				t.Errorf("Circle.Area() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Moved(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	type args struct {
		delta pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Circle
	}{
		{
			name:   "Circle.Moved(): positive movement",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{delta: pixel.V(10, 20)},
			want:   pixel.Circle{Radius: 10, Center: pixel.Vec{X: 10, Y: 20}},
		},
		{
			name:   "Circle.Moved(): zero movement",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{delta: pixel.ZV},
			want:   pixel.Circle{Radius: 10, Center: pixel.Vec{X: 0, Y: 0}},
		},
		{
			name:   "Circle.Moved(): negative movement",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{delta: pixel.V(-5, -10)},
			want:   pixel.Circle{Radius: 10, Center: pixel.Vec{X: -5, Y: -10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Moved(tt.args.delta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Moved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Resized(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	type args struct {
		radiusDelta float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Circle
	}{
		{
			name:   "Circle.Resized(): positive delta",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{radiusDelta: 5},
			want:   pixel.Circle{Radius: 15, Center: pixel.Vec{X: 0, Y: 0}},
		},
		{
			name:   "Circle.Resized(): zero delta",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{radiusDelta: 0},
			want:   pixel.Circle{Radius: 10, Center: pixel.Vec{X: 0, Y: 0}},
		},
		{
			name:   "Circle.Resized(): negative delta",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{radiusDelta: -5},
			want:   pixel.Circle{Radius: 5, Center: pixel.Vec{X: 0, Y: 0}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Resized(tt.args.radiusDelta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Resized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Contains(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	type args struct {
		u pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Circle.Contains(): point on cicles' center",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{u: pixel.ZV},
			want:   true,
		},
		{
			name:   "Circle.Contains(): point offcenter",
			fields: fields{radius: 10, center: pixel.V(5, 0)},
			args:   args{u: pixel.ZV},
			want:   true,
		},
		{
			name:   "Circle.Contains(): point on circumference",
			fields: fields{radius: 10, center: pixel.V(10, 0)},
			args:   args{u: pixel.ZV},
			want:   true,
		},
		{
			name:   "Circle.Contains(): point outside circle",
			fields: fields{radius: 10, center: pixel.V(15, 0)},
			args:   args{u: pixel.ZV},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Contains(tt.args.u); got != tt.want {
				t.Errorf("Circle.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Union(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	type args struct {
		d pixel.Circle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Circle
	}{
		{
			name:   "Circle.Union(): overlapping circles",
			fields: fields{radius: 5, center: pixel.ZV},
			args:   args{d: pixel.C(5, pixel.ZV)},
			want:   pixel.C(5, pixel.ZV),
		},
		{
			name:   "Circle.Union(): separate circles",
			fields: fields{radius: 1, center: pixel.ZV},
			args:   args{d: pixel.C(1, pixel.V(0, 2))},
			want:   pixel.C(2, pixel.V(0, 1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.radius, tt.fields.center)
			if got := c.Union(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_Intersect(t *testing.T) {
	type fields struct {
		radius float64
		center pixel.Vec
	}
	type args struct {
		d pixel.Circle
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Circle
	}{
		{
			name:   "Circle.Intersect(): intersecting circles",
			fields: fields{radius: 1, center: pixel.V(0, 0)},
			args:   args{d: pixel.C(1, pixel.V(1, 0))},
			want:   pixel.C(1, pixel.V(0.5, 0)),
		},
		{
			name:   "Circle.Intersect(): non-intersecting circles",
			fields: fields{radius: 1, center: pixel.V(0, 0)},
			args:   args{d: pixel.C(1, pixel.V(3, 3))},
			want:   pixel.C(0, pixel.V(1.5, 1.5)),
		},
		{
			name:   "Circle.Intersect(): first circle encompassing second",
			fields: fields{radius: 10, center: pixel.V(0, 0)},
			args:   args{d: pixel.C(1, pixel.V(3, 3))},
			want:   pixel.C(10, pixel.V(0, 0)),
		},
		{
			name:   "Circle.Intersect(): second circle encompassing first",
			fields: fields{radius: 1, center: pixel.V(-1, -4)},
			args:   args{d: pixel.C(10, pixel.V(0, 0))},
			want:   pixel.C(10, pixel.V(0, 0)),
		},
		{
			name:   "Circle.Intersect(): matching circles",
			fields: fields{radius: 1, center: pixel.V(0, 0)},
			args:   args{d: pixel.C(1, pixel.V(0, 0))},
			want:   pixel.C(1, pixel.V(0, 0)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(
				tt.fields.radius,
				tt.fields.center,
			)
			if got := c.Intersect(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRect_IntersectsCircle(t *testing.T) {
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
		want   bool
	}{
		{
			name:   "Rect.IntersectsCircle(): no overlap",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(1, pixel.V(50, 50))},
			want:   false,
		},
		{
			name:   "Rect.IntersectsCircle(): circle contains rect",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(10, pixel.V(5, 5))},
			want:   true,
		},
		{
			name:   "Rect.IntersectsCircle(): rect contains circle",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(1, pixel.V(5, 5))},
			want:   true,
		},
		{
			name:   "Rect.IntersectsCircle(): circle overlaps one corner",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(1, pixel.V(0, 0))},
			want:   true,
		},
		{
			name:   "Rect.IntersectsCircle(): circle overlaps two corners",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(6, pixel.V(0, 5))},
			want:   true,
		},
		{
			name:   "Rect.IntersectsCircle(): circle overlaps one edge",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(1, pixel.V(0, 5))},
			want:   true,
		},
		{
			name:   "Rect.IntersectsCircle(): edge is tangent",
			fields: fields{Min: pixel.V(0, 0), Max: pixel.V(10, 10)},
			args:   args{c: pixel.C(1, pixel.V(-1, 5))},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := pixel.Rect{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			if got := r.IntersectsCircle(tt.args.c); got != tt.want {
				t.Errorf("Rect.IntersectsCircle() = %v, want %v", got, tt.want)
			}
		})
	}
}
