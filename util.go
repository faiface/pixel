package pixel

import "github.com/go-gl/mathgl/mgl32"

func clamp(x, low, high float64) float64 {
	if x < low {
		return low
	}
	if x > high {
		return high
	}
	return x
}

func transformToMat(t ...Transform) mgl32.Mat3 {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat())
	}
	return mat
}
