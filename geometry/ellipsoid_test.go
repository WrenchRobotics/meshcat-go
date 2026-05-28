package geometry

import (
	"testing"
)

func TestEllipsoidIntrinsicTransform(t *testing.T) {
	e := NewEllipsoid([3]float64{2, 3, 4})
	m := e.IntrinsicTransform()

	if m.At(0, 0) != 2 || m.At(1, 1) != 3 || m.At(2, 2) != 4 || m.At(3, 3) != 1 {
		t.Fatalf("unexpected ellipsoid intrinsic matrix")
	}
}
