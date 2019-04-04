package pixel_test

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
	"github.com/stretchr/testify/assert"
)

// closeEnough will shift the decimal point by the accuracy required, truncates the results and compares them.
// Effectively this compares two floats to a given decimal point.
//  Example:
//  closeEnough(100.125342432, 100.125, 2) == true
//  closeEnough(math.Pi, 3.14, 2) == true
//  closeEnough(0.1234, 0.1245, 3) == false
func closeEnough(got, expected float64, decimalAccuracy int) bool {
	gotShifted := got * math.Pow10(decimalAccuracy)
	expectedShifted := expected * math.Pow10(decimalAccuracy)

	return math.Trunc(gotShifted) == math.Trunc(expectedShifted)
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

func TestMatrix_Unproject(t *testing.T) {
	const delta = 1e-15
	t.Run("for rotated matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Rotated(pixel.ZV, math.Pi/2)
		unprojected := matrix.Unproject(pixel.V(0, 1))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Moved(pixel.V(1, 2))
		unprojected := matrix.Unproject(pixel.V(2, 5))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 3, delta)
	})
	t.Run("for scaled matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Scaled(pixel.ZV, 2)
		unprojected := matrix.Unproject(pixel.V(2, 4))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 2, delta)
	})
	t.Run("for scaled, rotated and moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Scaled(pixel.ZV, 2).
			Rotated(pixel.ZV, math.Pi/2).
			Moved(pixel.V(2, 2))
		unprojected := matrix.Unproject(pixel.V(-2, 6))
		assert.InDelta(t, unprojected.X, 2, delta)
		assert.InDelta(t, unprojected.Y, 2, delta)
	})
	t.Run("for rotated and moved matrix", func(t *testing.T) {
		matrix := pixel.IM.
			Rotated(pixel.ZV, math.Pi/2).
			Moved(pixel.V(1, 1))
		unprojected := matrix.Unproject(pixel.V(1, 2))
		assert.InDelta(t, unprojected.X, 1, delta)
		assert.InDelta(t, unprojected.Y, 0, delta)
	})
	t.Run("for projected vertices using all kinds of matrices", func(t *testing.T) {
		namedMatrices := map[string]pixel.Matrix{
			"IM":                        pixel.IM,
			"Scaled":                    pixel.IM.Scaled(pixel.ZV, 0.5),
			"Scaled x 2":                pixel.IM.Scaled(pixel.ZV, 2),
			"Rotated":                   pixel.IM.Rotated(pixel.ZV, math.Pi/4),
			"Moved":                     pixel.IM.Moved(pixel.V(0.5, 1)),
			"Moved 2":                   pixel.IM.Moved(pixel.V(-1, -0.5)),
			"Scaled and Rotated":        pixel.IM.Scaled(pixel.ZV, 0.5).Rotated(pixel.ZV, math.Pi/4),
			"Scaled, Rotated and Moved": pixel.IM.Scaled(pixel.ZV, 0.5).Rotated(pixel.ZV, math.Pi/4).Moved(pixel.V(1, 2)),
			"Rotated and Moved":         pixel.IM.Rotated(pixel.ZV, math.Pi/4).Moved(pixel.V(1, 2)),
		}
		vertices := [...]pixel.Vec{
			pixel.V(0, 0),
			pixel.V(5, 0),
			pixel.V(5, 10),
			pixel.V(0, 10),
			pixel.V(-5, 10),
			pixel.V(-5, 0),
			pixel.V(-5, -10),
			pixel.V(0, -10),
			pixel.V(5, -10),
		}
		for matrixName, matrix := range namedMatrices {
			for _, vertex := range vertices {
				testCase := fmt.Sprintf("for matrix %s and vertex %v", matrixName, vertex)
				t.Run(testCase, func(t *testing.T) {
					projected := matrix.Project(vertex)
					unprojected := matrix.Unproject(projected)
					assert.InDelta(t, vertex.X, unprojected.X, delta)
					assert.InDelta(t, vertex.Y, unprojected.Y, delta)
				})
			}
		}
	})
	t.Run("for singular matrix", func(t *testing.T) {
		matrix := pixel.Matrix{0, 0, 0, 0, 0, 0}
		unprojected := matrix.Unproject(pixel.ZV)
		assert.True(t, math.IsNaN(unprojected.X))
		assert.True(t, math.IsNaN(unprojected.Y))
	})
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

func TestLine_Bounds(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   pixel.Rect
	}{
		{
			name:   "Positive slope",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			want:   pixel.R(0, 0, 10, 10),
		},
		{
			name:   "Negative slope",
			fields: fields{A: pixel.V(10, 10), B: pixel.V(0, 0)},
			want:   pixel.R(0, 0, 10, 10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Bounds(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Bounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Center(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   pixel.Vec
	}{
		{
			name:   "Positive slope",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Negative slope",
			fields: fields{A: pixel.V(10, 10), B: pixel.V(0, 0)},
			want:   pixel.V(5, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Center(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Center() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Closest(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Vec
	}{
		{
			name:   "Point on line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(5, 5)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Point on next to line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(0, 10)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Point on next to vertical line",
			fields: fields{A: pixel.V(5, 0), B: pixel.V(5, 10)},
			args:   args{v: pixel.V(6, 5)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Point on next to horizontal line",
			fields: fields{A: pixel.V(0, 5), B: pixel.V(10, 5)},
			args:   args{v: pixel.V(5, 6)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Point far from line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(80, -70)},
			want:   pixel.V(5, 5),
		},
		{
			name:   "Point on inline with line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(20, 20)},
			want:   pixel.V(10, 10),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Closest(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Closest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Contains(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Point on line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(5, 5)},
			want:   true,
		},
		{
			name:   "Point on negative sloped line",
			fields: fields{A: pixel.V(0, 10), B: pixel.V(10, 0)},
			args:   args{v: pixel.V(5, 5)},
			want:   true,
		},
		{
			name:   "Point not on line",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{v: pixel.V(0, 10)},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Contains(tt.args.v); got != tt.want {
				t.Errorf("Line.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Formula(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		wantM  float64
		wantB  float64
	}{
		{
			name:   "Getting formula - 45 degs",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			wantM:  1,
			wantB:  0,
		},
		{
			name:   "Getting formula - 90 degs",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(0, 10)},
			wantM:  math.Inf(1),
			wantB:  math.NaN(),
		},
		{
			name:   "Getting formula - 0 degs",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 0)},
			wantM:  0,
			wantB:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			gotM, gotB := l.Formula()
			if gotM != tt.wantM {
				t.Errorf("Line.Formula() gotM = %v, want %v", gotM, tt.wantM)
			}
			if gotB != tt.wantB {
				if math.IsNaN(tt.wantB) && !math.IsNaN(gotB) {
					t.Errorf("Line.Formula() gotB = %v, want %v", gotB, tt.wantB)
				}
			}
		})
	}
}

func TestLine_Intersect(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		k pixel.Line
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Vec
		want1  bool
	}{
		{
			name:   "Lines intersect",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{k: pixel.L(pixel.V(0, 10), pixel.V(10, 0))},
			want:   pixel.V(5, 5),
			want1:  true,
		},
		{
			name:   "Lines intersect 2",
			fields: fields{A: pixel.V(5, 1), B: pixel.V(1, 1)},
			args:   args{k: pixel.L(pixel.V(2, 0), pixel.V(2, 3))},
			want:   pixel.V(2, 1),
			want1:  true,
		},
		{
			name:   "Line intersect with vertical",
			fields: fields{A: pixel.V(5, 0), B: pixel.V(5, 10)},
			args:   args{k: pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   pixel.V(5, 5),
			want1:  true,
		},
		{
			name:   "Line intersect with horizontal",
			fields: fields{A: pixel.V(0, 5), B: pixel.V(10, 5)},
			args:   args{k: pixel.L(pixel.V(0, 0), pixel.V(10, 10))},
			want:   pixel.V(5, 5),
			want1:  true,
		},
		{
			name:   "Lines don't intersect",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{k: pixel.L(pixel.V(0, 10), pixel.V(1, 20))},
			want:   pixel.ZV,
			want1:  false,
		},
		{
			name:   "Lines don't intersect 2",
			fields: fields{A: pixel.V(1, 1), B: pixel.V(1, 5)},
			args:   args{k: pixel.L(pixel.V(-5, 0), pixel.V(-2, 2))},
			want:   pixel.ZV,
			want1:  false,
		},
		{
			name:   "Lines don't intersect 3",
			fields: fields{A: pixel.V(2, 0), B: pixel.V(2, 3)},
			args:   args{k: pixel.L(pixel.V(1, 5), pixel.V(5, 5))},
			want:   pixel.ZV,
			want1:  false,
		},
		{
			name:   "Lines parallel",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{k: pixel.L(pixel.V(0, 1), pixel.V(10, 11))},
			want:   pixel.ZV,
			want1:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			got, got1 := l.Intersect(tt.args.k)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Intersect() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Line.Intersect() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestLine_IntersectCircle(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
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
			name:   "Cirle intersects",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(6, 4), 2)},
			want:   pixel.V(0.5857864376269049, -0.5857864376269049),
		},
		{
			name:   "Cirle doesn't intersects",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{c: pixel.C(pixel.V(0, 5), 1)},
			want:   pixel.ZV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.IntersectCircle(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.IntersectCircle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_IntersectRect(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		r pixel.Rect
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Vec
	}{
		{
			name:   "Line through rect vertically",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(0, 10)},
			args:   args{r: pixel.R(-1, 1, 5, 5)},
			want:   pixel.V(-1, 0),
		},
		{
			name:   "Line through rect horizontally",
			fields: fields{A: pixel.V(0, 1), B: pixel.V(10, 1)},
			args:   args{r: pixel.R(1, 0, 5, 5)},
			want:   pixel.V(0, -1),
		},
		{
			name:   "Line through rect diagonally bottom and left edges",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{r: pixel.R(0, 2, 3, 3)},
			want:   pixel.V(-1, 1),
		},
		{
			name:   "Line through rect diagonally top and right edges",
			fields: fields{A: pixel.V(10, 0), B: pixel.V(0, 10)},
			args:   args{r: pixel.R(5, 0, 8, 3)},
			want:   pixel.V(-2.5, -2.5),
		},
		{
			name:   "Line with not rect intersect",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{r: pixel.R(20, 20, 21, 21)},
			want:   pixel.ZV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.IntersectRect(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.IntersectRect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Len(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name:   "End right-up of start",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(3, 4)},
			want:   5,
		},
		{
			name:   "End left-up of start",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(-3, 4)},
			want:   5,
		},
		{
			name:   "End right-down of start",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(3, -4)},
			want:   5,
		},
		{
			name:   "End left-down of start",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(-3, -4)},
			want:   5,
		},
		{
			name:   "End same as start",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(0, 0)},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Len(); got != tt.want {
				t.Errorf("Line.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Rotated(t *testing.T) {
	// round returns the nearest integer, rounding ties away from zero.
	// This is required because `math.Round` wasn't introduced until Go1.10
	round := func(x float64) float64 {
		t := math.Trunc(x)
		if math.Abs(x-t) >= 0.5 {
			return t + math.Copysign(1, x)
		}
		return t
	}
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		around pixel.Vec
		angle  float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Line
	}{
		{
			name:   "Rotating around line center",
			fields: fields{A: pixel.V(1, 1), B: pixel.V(3, 3)},
			args:   args{around: pixel.V(2, 2), angle: math.Pi},
			want:   pixel.L(pixel.V(3, 3), pixel.V(1, 1)),
		},
		{
			name:   "Rotating around x-y origin",
			fields: fields{A: pixel.V(1, 1), B: pixel.V(3, 3)},
			args:   args{around: pixel.V(0, 0), angle: math.Pi},
			want:   pixel.L(pixel.V(-1, -1), pixel.V(-3, -3)),
		},
		{
			name:   "Rotating around line end",
			fields: fields{A: pixel.V(1, 1), B: pixel.V(3, 3)},
			args:   args{around: pixel.V(1, 1), angle: math.Pi},
			want:   pixel.L(pixel.V(1, 1), pixel.V(-1, -1)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			// Have to round the results, due to floating-point in accuracies.  Results are correct to approximately
			// 10 decimal places.
			got := l.Rotated(tt.args.around, tt.args.angle)
			if round(got.A.X) != tt.want.A.X ||
				round(got.B.X) != tt.want.B.X ||
				round(got.A.Y) != tt.want.A.Y ||
				round(got.B.Y) != tt.want.B.Y {
				t.Errorf("Line.Rotated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_Scaled(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		scale float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Line
	}{
		{
			name:   "Scaling by 1",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{scale: 1},
			want:   pixel.L(pixel.V(0, 0), pixel.V(10, 10)),
		},
		{
			name:   "Scaling by >1",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{scale: 2},
			want:   pixel.L(pixel.V(-5, -5), pixel.V(15, 15)),
		},
		{
			name:   "Scaling by <1",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{scale: 0.5},
			want:   pixel.L(pixel.V(2.5, 2.5), pixel.V(7.5, 7.5)),
		},
		{
			name:   "Scaling by -1",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{scale: -1},
			want:   pixel.L(pixel.V(10, 10), pixel.V(0, 0)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.Scaled(tt.args.scale); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.Scaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_ScaledXY(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	type args struct {
		around pixel.Vec
		scale  float64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   pixel.Line
	}{
		{
			name:   "Scaling by 1 around origin",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{around: pixel.ZV, scale: 1},
			want:   pixel.L(pixel.V(0, 0), pixel.V(10, 10)),
		},
		{
			name:   "Scaling by >1 around origin",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{around: pixel.ZV, scale: 2},
			want:   pixel.L(pixel.V(0, 0), pixel.V(20, 20)),
		},
		{
			name:   "Scaling by <1 around origin",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{around: pixel.ZV, scale: 0.5},
			want:   pixel.L(pixel.V(0, 0), pixel.V(5, 5)),
		},
		{
			name:   "Scaling by -1 around origin",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(10, 10)},
			args:   args{around: pixel.ZV, scale: -1},
			want:   pixel.L(pixel.V(0, 0), pixel.V(-10, -10)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.ScaledXY(tt.args.around, tt.args.scale); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Line.ScaledXY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLine_String(t *testing.T) {
	type fields struct {
		A pixel.Vec
		B pixel.Vec
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Getting string",
			fields: fields{A: pixel.V(0, 0), B: pixel.V(1, 1)},
			want:   "Line(Vec(0, 0), Vec(1, 1))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := pixel.Line{
				A: tt.fields.A,
				B: tt.fields.B,
			}
			if got := l.String(); got != tt.want {
				t.Errorf("Line.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
