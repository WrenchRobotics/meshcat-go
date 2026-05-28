package geometry

import "testing"

func TestLineBasicMaterialType(t *testing.T) {
	lowered := NewLineBasicMaterial().Lower(NewLowerContext())
	if lowered["type"] != "LineBasicMaterial" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
}
