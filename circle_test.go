package pixel_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

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
			args: args{radius: 10, center: pixel.ZV},
			want: pixel.Circle{Radius: 10, Center: pixel.ZV},
		},
		{
			name: "C(): zero radius",
			args: args{radius: 0, center: pixel.ZV},
			want: pixel.Circle{Radius: 0, Center: pixel.ZV},
		},
		{
			name: "C(): negative radius",
			args: args{radius: -5, center: pixel.ZV},
			want: pixel.Circle{Radius: -5, Center: pixel.ZV},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.C(tt.args.center, tt.args.radius); !reflect.DeepEqual(got, tt.want) {
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
			fields: fields{radius: 10, center: pixel.ZV},
			want:   "Circle(Vec(0, 0), 10.00)",
		},
		{
			name:   "Circle.String(): zero radius",
			fields: fields{radius: 0, center: pixel.ZV},
			want:   "Circle(Vec(0, 0), 0.00)",
		},
		{
			name:   "Circle.String(): negative radius",
			fields: fields{radius: -5, center: pixel.ZV},
			want:   "Circle(Vec(0, 0), -5.00)",
		},
		{
			name:   "Circle.String(): irrational radius",
			fields: fields{radius: math.Pi, center: pixel.ZV},
			want:   "Circle(Vec(0, 0), 3.14)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			fields: fields{radius: 10, center: pixel.ZV},
			want:   pixel.C(pixel.ZV, 10),
		},
		{
			name:   "Circle.Norm(): zero radius",
			fields: fields{radius: 0, center: pixel.ZV},
			want:   pixel.C(pixel.ZV, 0),
		},
		{
			name:   "Circle.Norm(): negative radius",
			fields: fields{radius: -5, center: pixel.ZV},
			want:   pixel.C(pixel.ZV, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			fields: fields{radius: 10, center: pixel.ZV},
			want:   100 * math.Pi,
		},
		{
			name:   "Circle.Area(): zero radius",
			fields: fields{radius: 0, center: pixel.ZV},
			want:   0,
		},
		{
			name:   "Circle.Area(): negative radius",
			fields: fields{radius: -5, center: pixel.ZV},
			want:   25 * math.Pi,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{delta: pixel.V(10, 20)},
			want:   pixel.C(pixel.V(10, 20), 10),
		},
		{
			name:   "Circle.Moved(): zero movement",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{delta: pixel.ZV},
			want:   pixel.C(pixel.V(0, 0), 10),
		},
		{
			name:   "Circle.Moved(): negative movement",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{delta: pixel.V(-5, -10)},
			want:   pixel.C(pixel.V(-5, -10), 10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{radiusDelta: 5},
			want:   pixel.C(pixel.V(0, 0), 15),
		},
		{
			name:   "Circle.Resized(): zero delta",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{radiusDelta: 0},
			want:   pixel.C(pixel.V(0, 0), 10),
		},
		{
			name:   "Circle.Resized(): negative delta",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{radiusDelta: -5},
			want:   pixel.C(pixel.V(0, 0), 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			args:   args{d: pixel.C(pixel.ZV, 5)},
			want:   pixel.C(pixel.ZV, 5),
		},
		{
			name:   "Circle.Union(): separate circles",
			fields: fields{radius: 1, center: pixel.ZV},
			args:   args{d: pixel.C(pixel.V(0, 2), 1)},
			want:   pixel.C(pixel.V(0, 1), 2),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(tt.fields.center, tt.fields.radius)
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
			fields: fields{radius: 1, center: pixel.ZV},
			args:   args{d: pixel.C(pixel.V(1, 0), 1)},
			want:   pixel.C(pixel.V(0.5, 0), 1),
		},
		{
			name:   "Circle.Intersect(): non-intersecting circles",
			fields: fields{radius: 1, center: pixel.ZV},
			args:   args{d: pixel.C(pixel.V(3, 3), 1)},
			want:   pixel.C(pixel.V(1.5, 1.5), 0),
		},
		{
			name:   "Circle.Intersect(): first circle encompassing second",
			fields: fields{radius: 10, center: pixel.ZV},
			args:   args{d: pixel.C(pixel.V(3, 3), 1)},
			want:   pixel.C(pixel.ZV, 10),
		},
		{
			name:   "Circle.Intersect(): second circle encompassing first",
			fields: fields{radius: 1, center: pixel.V(-1, -4)},
			args:   args{d: pixel.C(pixel.ZV, 10)},
			want:   pixel.C(pixel.ZV, 10),
		},
		{
			name:   "Circle.Intersect(): matching circles",
			fields: fields{radius: 1, center: pixel.ZV},
			args:   args{d: pixel.C(pixel.ZV, 1)},
			want:   pixel.C(pixel.ZV, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.C(
				tt.fields.center,
				tt.fields.radius,
			)
			if got := c.Intersect(tt.args.d); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCircle_IntersectPoints(t *testing.T) {
	type fields struct {
		Center pixel.Vec
		Radius float64
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
			name:   "Line intersects circle at two points",
			fields: fields{Center: pixel.V(2, 2), Radius: 1},
			args:   args{pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   []pixel.Vec{pixel.V(1.292, 1.292), pixel.V(2.707, 2.707)},
		},
		{
			name:   "Line intersects circle at one point",
			fields: fields{Center: pixel.V(-0.5, -0.5), Radius: 1},
			args:   args{pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   []pixel.Vec{pixel.V(0.207, 0.207)},
		},
		{
			name:   "Line endpoint is circle center",
			fields: fields{Center: pixel.V(0, 0), Radius: 1},
			args:   args{pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   []pixel.Vec{pixel.V(0.707, 0.707)},
		},
		{
			name:   "Both line endpoints within circle",
			fields: fields{Center: pixel.V(0, 0), Radius: 1},
			args:   args{pixel.L(pixel.V(0.2, 0.2), pixel.V(0.5, 0.5))},
			want:   []pixel.Vec{},
		},
		{
			name:   "Line does not intersect circle",
			fields: fields{Center: pixel.V(10, 0), Radius: 1},
			args:   args{pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   []pixel.Vec{},
		},
		{
			name:   "Horizontal line intersects circle at two points",
			fields: fields{Center: pixel.V(5, 5), Radius: 1},
			args:   args{pixel.L(pixel.V(0, 5), pixel.V(10, 5))},
			want:   []pixel.Vec{pixel.V(4, 5), pixel.V(6, 5)},
		},
		{
			name:   "Vertical line intersects circle at two points",
			fields: fields{Center: pixel.V(5, 5), Radius: 1},
			args:   args{pixel.L(pixel.V(5, 0), pixel.V(5, 10))},
			want:   []pixel.Vec{pixel.V(5, 4), pixel.V(5, 6)},
		},
		{
			name:   "Left and down line intersects circle at two points",
			fields: fields{Center: pixel.V(5, 5), Radius: 1},
			args:   args{pixel.L(pixel.V(10, 10), pixel.V(0, 0))},
			want:   []pixel.Vec{pixel.V(5.707, 5.707), pixel.V(4.292, 4.292)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := pixel.Circle{
				Center: tt.fields.Center,
				Radius: tt.fields.Radius,
			}
			got := c.IntersectionPoints(tt.args.l)
			for i, v := range got {
				if !closeEnough(v.X, tt.want[i].X, 2) || !closeEnough(v.Y, tt.want[i].Y, 2) {
					t.Errorf("Circle.IntersectPoints() = %v, want %v", v, tt.want[i])
				}
			}
		})
	}
}
