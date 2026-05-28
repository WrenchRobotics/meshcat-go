package geometry

import "testing"

func TestMeshLambertMaterialType(t *testing.T) {
	lowered := NewMeshLambertMaterial().Lower(NewLowerContext())
	if lowered["type"] != "MeshLambertMaterial" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
}
