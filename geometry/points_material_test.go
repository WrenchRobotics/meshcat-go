package geometry

import "testing"

func TestPointsMaterialDefaultsAndLower(t *testing.T) {
	pm := NewPointsMaterial()
	lowered := pm.Lower(NewLowerContext())

	if lowered["type"] != "PointsMaterial" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
	if lowered["color"] != 0xffffff {
		t.Fatalf("unexpected default color: %v", lowered["color"])
	}
	if lowered["size"] != 0.001 {
		t.Fatalf("unexpected default size: %v", lowered["size"])
	}
	if lowered["vertexColors"] != 2 {
		t.Fatalf("unexpected vertexColors enum: %v", lowered["vertexColors"])
	}
}
