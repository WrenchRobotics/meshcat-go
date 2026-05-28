package geometry

import "testing"

func TestMeshBasicMaterialType(t *testing.T) {
	lowered := NewMeshBasicMaterial().Lower(NewLowerContext())
	if lowered["type"] != "MeshBasicMaterial" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
}
