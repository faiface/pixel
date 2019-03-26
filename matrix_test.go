package pixel_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/faiface/pixel"
)

func BenchmarkMatrix(b *testing.B) {
	b.Run("Moved", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.Moved(pixel.V(4.217, -132.99))
		}
	})
	b.Run("ScaledXY", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.ScaledXY(pixel.V(-5.1, 9.3), pixel.V(2.1, 0.98))
		}
	})
	b.Run("Rotated", func(b *testing.B) {
		m := pixel.IM
		for i := 0; i < b.N; i++ {
			m = m.Rotated(pixel.V(-5.1, 9.3), 1.4)
		}
	})
	b.Run("Chained", func(b *testing.B) {
		var m1, m2 pixel.Matrix
		for i := range m1 {
			m1[i] = rand.Float64()
			m2[i] = rand.Float64()
		}
		for i := 0; i < b.N; i++ {
			m1 = m1.Chained(m2)
		}
	})
	b.Run("Project", func(b *testing.B) {
		var m pixel.Matrix
		for i := range m {
			m[i] = rand.Float64()
		}
		u := pixel.V(1, 1)
		for i := 0; i < b.N; i++ {
			u = m.Project(u)
		}
	})
	b.Run("Unproject", func(b *testing.B) {
	again:
		var m pixel.Matrix
		for i := range m {
			m[i] = rand.Float64()
		}
		if (m[0]*m[3])-(m[1]*m[2]) == 0 { // zero determinant, not invertible
			goto again
		}
		u := pixel.V(1, 1)
		for i := 0; i < b.N; i++ {
			u = m.Unproject(u)
		}
	})
}

func TestMatrix_String(t *testing.T) {
	tests := []struct {
		name string
		m    pixel.Matrix
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("Matrix.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Moved(t *testing.T) {
	type args struct {
		delta pixel.Vec
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Matrix
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Moved(tt.args.delta); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Moved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_ScaledXY(t *testing.T) {
	type args struct {
		around pixel.Vec
		scale  pixel.Vec
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Matrix
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.ScaledXY(tt.args.around, tt.args.scale); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.ScaledXY() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Scaled(t *testing.T) {
	type args struct {
		around pixel.Vec
		scale  float64
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Matrix
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Scaled(tt.args.around, tt.args.scale); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Scaled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Rotated(t *testing.T) {
	type args struct {
		around pixel.Vec
		angle  float64
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Matrix
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Rotated(tt.args.around, tt.args.angle); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Rotated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Chained(t *testing.T) {
	type args struct {
		next pixel.Matrix
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Matrix
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Chained(tt.args.next); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Chained() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Project(t *testing.T) {
	type args struct {
		u pixel.Vec
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Vec
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Project(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Project() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatrix_Unproject(t *testing.T) {
	type args struct {
		u pixel.Vec
	}
	tests := []struct {
		name string
		m    pixel.Matrix
		args args
		want pixel.Vec
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Unproject(tt.args.u); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Matrix.Unproject() = %v, want %v", got, tt.want)
			}
		})
	}
}
