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

func lerp(x float64, a, b Vec) Vec {
	return a.Scaled(1-x) + b.Scaled(x)
}

func lerp2d(x, a, b Vec) Vec {
	return V(
		lerp(x.X(), a, b).X(),
		lerp(x.Y(), a, b).Y(),
	)
}

func transformToMat(t ...Transform) mgl32.Mat3 {
	mat := mgl32.Ident3()
	for i := range t {
		mat = mat.Mul3(t[i].Mat())
	}
	return mat
}

func pictureBounds(p Picture, v Vec) Vec {
	w, h := p.Bounds().Size.XY()
	a := p.Bounds().Pos
	b := p.Bounds().Pos + p.Bounds().Size
	u := lerp2d(v, a, b)
	return V(u.X()/w, u.Y()/h)
}
