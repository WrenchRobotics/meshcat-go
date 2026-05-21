package geometry

import "testing"

func TestMeshToonMaterialType(t *testing.T) {
	lowered := NewMeshToonMaterial().Lower(NewLowerContext())
	if lowered["type"] != "MeshToonMaterial" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
}
