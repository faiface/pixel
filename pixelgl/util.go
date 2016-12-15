package pixelgl

import "github.com/go-gl/gl/v3.3-core/gl"

type binder struct {
	restoreLoc uint32
	bindFunc   func(uint32)

	obj uint32

	prev []uint32
}

func (b *binder) bind() *binder {
	var prev int32
	gl.GetIntegerv(b.restoreLoc, &prev)
	b.prev = append(b.prev, uint32(prev))

	if b.prev[len(b.prev)-1] != b.obj {
		b.bindFunc(b.obj)
	}
	return b
}

func (b *binder) restore() *binder {
	if b.prev[len(b.prev)-1] != b.obj {
		b.bindFunc(b.prev[len(b.prev)-1])
	}
	b.prev = b.prev[:len(b.prev)-1]
	return b
}
