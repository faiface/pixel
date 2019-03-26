package pixel_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

// closeEnough is a test helper function to establish if vectors are "close enough".  This is to resolve floating point
// errors, specifically when dealing with `math.Pi`
func closeEnough(u, v pixel.Vec) bool {
	uX, uY := math.Round(u.X), math.Round(u.Y)
	vX, vY := math.Round(v.X), math.Round(v.Y)
	return uX == vX && uY == vY
}

func TestV(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		args args
		want pixel.Vec
	}{
		{
			name: "V(): both 0",
			args: args{x: 0, y: 0},
			want: pixel.ZV,
		},
		{
			name: "V(): x < y",
			args: args{x: 0, y: 10},
			want: pixel.Vec{X: 0, Y: 10},
		},
		{
			name: "V(): x > y",
			args: args{x: 10, y: 0},
			want: pixel.Vec{X: 10, Y: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.V(tt.args.x, tt.args.y); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("V() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit(t *testing.T) {
	type args struct {
		angle float64
	}
	tests := []struct {
		name string
		args args
		want pixel.Vec
	}{
		{
			name: "Unit(): 0 radians",
			args: args{angle: 0},
			want: pixel.V(1, 0),
		},
		{
			name: "Unit(): pi radians",
			args: args{angle: math.Pi},
			want: pixel.V(-1, 0),
		},
		{
			name: "Unit(): 10 * pi radians",
			args: args{angle: 10 * math.Pi},
			want: pixel.V(1, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.Unit(tt.args.angle); !closeEnough(got, tt.want) {
				t.Errorf("Unit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_String(t *testing.T) {
	tests := []struct {
		name string
		u    pixel.Vec
		want string
	}{
		{
			name: "Vec.String(): both 0",
			u:    pixel.Vec{X: 0, Y: 0},
			want: "Vec(0, 0)",
		},
		{
			name: "Vec.String(): x < y",
			u:    pixel.Vec{X: 0, Y: 10},
			want: "Vec(0, 10)",
		},
		{
			name: "Vec.String(): x > y",
			u:    pixel.Vec{X: 10, Y: 0},
			want: "Vec(10, 0)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.String(); got != tt.want {
				t.Errorf("Vec.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_XY(t *testing.T) {
	tests := []struct {
		name  string
		u     pixel.Vec
		wantX float64
		wantY float64
	}{
		{
			name:  "Vec.XY(): both 0",
			u:     pixel.Vec{X: 0, Y: 0},
			wantX: 0,
			wantY: 0,
		},
		{
			name:  "Vec.XY(): x < y",
			u:     pixel.Vec{X: 0, Y: 10},
			wantX: 0,
			wantY: 10,
		},
		{
			name:  "Vec.XY(): x > y",
			u:     pixel.Vec{X: 10, Y: 0},
			wantX: 10,
			wantY: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotX, gotY := tt.u.XY()
			if gotX != tt.wantX {
				t.Errorf("Vec.XY() gotX = %v, want %v", gotX, tt.wantX)
			}
			if gotY != tt.wantY {
				t.Errorf("Vec.XY() gotY = %v, want %v", gotY, tt.wantY)
			}
		})
	}
}

func TestVec_Add(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Add(): positive vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(10, 10)},
			want: pixel.V(10, 20),
		},
		{
			name: "Vec.Add(): zero vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.ZV},
			want: pixel.V(0, 10),
		},
		{
			name: "Vec.Add(): negative vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(-20, -30)},
			want: pixel.V(-20, -20),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Add(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Sub(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Sub(): positive vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(10, 10)},
			want: pixel.V(-10, 0),
		},
		{
			name: "Vec.Sub(): zero vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.ZV},
			want: pixel.V(0, 10),
		},
		{
			name: "Vec.Sub(): negative vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(-20, -30)},
			want: pixel.V(20, 40),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Sub(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Sub() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_To(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.To(): positive vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(10, 10)},
			want: pixel.V(10, 0),
		},
		{
			name: "Vec.To(): zero vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.ZV},
			want: pixel.V(0, -10),
		},
		{
			name: "Vec.To(): negative vector",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(-20, -30)},
			want: pixel.V(-20, -40),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.To(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.To() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Scaled(t *testing.T) {
	type args struct {
		c float64
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Scaled(): positive scale",
			u:    pixel.V(0, 10),
			args: args{c: 10},
			want: pixel.V(0, 100),
		},
		{
			name: "Vec.Scaled(): zero scale",
			u:    pixel.V(0, 10),
			args: args{c: 0},
			want: pixel.ZV,
		},
		{
			name: "Vec.Scaled(): identity scale",
			u:    pixel.V(0, 10),
			args: args{c: 1},
			want: pixel.V(0, 10),
		},
		{
			name: "Vec.Scaled(): negative scale",
			u:    pixel.V(0, 10),
			args: args{c: -10},
			want: pixel.V(0, -100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Scaled(tt.args.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Scaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_ScaledXY(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.ScaledXY(): positive scale",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(10, 20)},
			want: pixel.V(0, 200),
		},
		{
			name: "Vec.ScaledXY(): zero scale",
			u:    pixel.V(0, 10),
			args: args{v: pixel.ZV},
			want: pixel.ZV,
		},
		{
			name: "Vec.ScaledXY(): identity scale",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(1, 1)},
			want: pixel.V(0, 10),
		},
		{
			name: "Vec.ScaledXY(): negative scale",
			u:    pixel.V(0, 10),
			args: args{v: pixel.V(-5, -10)},
			want: pixel.V(0, -100),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.ScaledXY(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.ScaledXY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Len(t *testing.T) {
	tests := []struct {
		name string
		u    pixel.Vec
		want float64
	}{
		{
			name: "Vec.Len(): positive vector",
			u:    pixel.V(40, 30),
			want: 50,
		},
		{
			name: "Vec.Len(): zero vector",
			u:    pixel.ZV,
			want: 0,
		},
		{
			name: "Vec.Len(): negative vector",
			u:    pixel.V(-5, -12),
			want: 13,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Len(); got != tt.want {
				t.Errorf("Vec.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Angle(t *testing.T) {
	tests := []struct {
		name string
		u    pixel.Vec
		want float64
	}{
		{
			name: "Vec.Angle(): positive vector",
			u:    pixel.V(0, 30),
			want: math.Pi / 2,
		},
		{
			name: "Vec.Angle(): zero vector",
			u:    pixel.ZV,
			want: 0,
		},
		{
			name: "Vec.Angle(): negative vector",
			u:    pixel.V(-5, -0),
			want: math.Pi,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Angle(); got != tt.want {
				t.Errorf("Vec.Angle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Unit(t *testing.T) {
	tests := []struct {
		name string
		u    pixel.Vec
		want pixel.Vec
	}{
		{
			name: "Vec.Unit(): positive vector",
			u:    pixel.V(0, 30),
			want: pixel.V(0, 1),
		},
		{
			name: "Vec.Unit(): zero vector",
			u:    pixel.ZV,
			want: pixel.V(1, 0),
		},
		{
			name: "Vec.Unit(): negative vector",
			u:    pixel.V(-5, 0),
			want: pixel.V(-1, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Unit(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Unit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Rotated(t *testing.T) {
	type args struct {
		angle float64
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Rotated(): partial rotation",
			u:    pixel.V(0, 1),
			args: args{angle: math.Pi / 2},
			want: pixel.V(-1, 0),
		},
		{
			name: "Vec.Rotated(): full rotation",
			u:    pixel.V(0, 1),
			args: args{angle: 2 * math.Pi},
			want: pixel.V(0, 1),
		},
		{
			name: "Vec.Rotated(): zero rotation",
			u:    pixel.V(0, 1),
			args: args{angle: 0},
			want: pixel.V(0, 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Rotated(tt.args.angle); !closeEnough(got, tt.want) {
				t.Errorf("Vec.Rotated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Normal(t *testing.T) {
	tests := []struct {
		name string
		u    pixel.Vec
		want pixel.Vec
	}{
		{
			name: "Vec.Normal(): positive vector",
			u:    pixel.V(0, 30),
			want: pixel.V(-30, 0),
		},
		{
			name: "Vec.Normal(): zero vector",
			u:    pixel.ZV,
			want: pixel.ZV,
		},
		{
			name: "Vec.Normal(): negative vector",
			u:    pixel.V(-5, 0),
			want: pixel.V(0, -5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Normal(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Normal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Dot(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want float64
	}{
		{
			name: "Vec.Dot(): positive vector",
			u:    pixel.V(0, 30),
			args: args{v: pixel.V(10, 10)},
			want: 300,
		},
		{
			name: "Vec.Dot(): zero vector",
			u:    pixel.ZV,
			args: args{v: pixel.V(10, 10)},
			want: 0,
		},
		{
			name: "Vec.Dot(): negative vector",
			u:    pixel.V(-5, 1),
			args: args{v: pixel.V(10, 10)},
			want: -40,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Dot(tt.args.v); got != tt.want {
				t.Errorf("Vec.Dot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Cross(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want float64
	}{
		{
			name: "Vec.Cross(): positive vector",
			u:    pixel.V(0, 30),
			args: args{v: pixel.V(10, 10)},
			want: -300,
		},
		{
			name: "Vec.Cross(): zero vector",
			u:    pixel.ZV,
			args: args{v: pixel.V(10, 10)},
			want: 0,
		},
		{
			name: "Vec.Cross(): negative vector",
			u:    pixel.V(-5, 1),
			args: args{v: pixel.V(10, 10)},
			want: -60,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Cross(tt.args.v); got != tt.want {
				t.Errorf("Vec.Cross() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Project(t *testing.T) {
	type args struct {
		v pixel.Vec
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Project(): positive vector",
			u:    pixel.V(0, 30),
			args: args{v: pixel.V(10, 10)},
			want: pixel.V(15, 15),
		},
		{
			name: "Vec.Project(): zero vector",
			u:    pixel.ZV,
			args: args{v: pixel.V(10, 10)},
			want: pixel.ZV,
		},
		{
			name: "Vec.Project(): negative vector",
			u:    pixel.V(-30, 0),
			args: args{v: pixel.V(10, 10)},
			want: pixel.V(-15, -15),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Project(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Project() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVec_Map(t *testing.T) {
	type args struct {
		f func(float64) float64
	}
	tests := []struct {
		name string
		u    pixel.Vec
		args args
		want pixel.Vec
	}{
		{
			name: "Vec.Map(): positive vector",
			u:    pixel.V(0, 25),
			args: args{f: math.Sqrt},
			want: pixel.V(0, 5),
		},
		{
			name: "Vec.Map(): zero vector",
			u:    pixel.ZV,
			args: args{f: math.Sqrt},
			want: pixel.ZV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.u.Map(tt.args.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Vec.Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLerp(t *testing.T) {
	type args struct {
		a pixel.Vec
		b pixel.Vec
		t float64
	}
	tests := []struct {
		name string
		args args
		want pixel.Vec
	}{
		{
			name: "Lerp(): t = 0",
			args: args{a: pixel.V(10, 10), b: pixel.ZV, t: 0},
			want: pixel.V(10, 10),
		},
		{
			name: "Lerp(): t = 1/4",
			args: args{a: pixel.V(10, 10), b: pixel.ZV, t: .25},
			want: pixel.V(7.5, 7.5),
		},
		{
			name: "Lerp(): t = 1/2",
			args: args{a: pixel.V(10, 10), b: pixel.ZV, t: .5},
			want: pixel.V(5, 5),
		},
		{
			name: "Lerp(): t = 1",
			args: args{a: pixel.V(10, 10), b: pixel.ZV, t: 1},
			want: pixel.ZV,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pixel.Lerp(tt.args.a, tt.args.b, tt.args.t); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lerp() = %v, want %v", got, tt.want)
			}
		})
	}
}
