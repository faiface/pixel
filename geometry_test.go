package pixel

import (
	"testing"
)

func TestRectArea(t *testing.T) {
	r1 := Rect{Vec{0, 0}, Vec{100, 100}}
	r2 := Rect{Vec{200, 200}, Vec{300, 300}}
	area := r1.Intersect(r2).Area()
	if area != 0 {
		t.Fail()
		t.Logf("Expected area of 0, got: %v", area)
	}
	r1 = Rect{Vec{0, 0}, Vec{100, 100}}
	r2 = Rect{Vec{50, 50}, Vec{500, 500}}
	area = r1.Intersect(r2).Area()
	if area != 2500 {
		t.Fail()
		t.Logf("Expected area of 2500, got: %v", area)
	}
}
