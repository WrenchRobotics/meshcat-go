package geometry

import (
	"testing"
)

func TestBoxLowerShape(t *testing.T) {
	box := NewBox([3]float64{1, 2, 3})
	m := box.Lower(NewLowerContext())

	if m["type"] != "BoxGeometry" {
		t.Fatalf("unexpected type: %v", m["type"])
	}
	if m["width"] != 1.0 || m["height"] != 2.0 || m["depth"] != 3.0 {
		t.Fatalf("unexpected dimensions: %v", m)
	}
}
