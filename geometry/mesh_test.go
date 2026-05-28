package geometry

import "testing"

func TestMeshLowerContainsGeometryAndMaterial(t *testing.T) {
	geom := NewSphere(0.4)
	mat := NewMeshPhongMaterial()
	mesh := NewMesh(geom, mat)

	lowered := mesh.Lower()

	geoms, ok := lowered["geometries"].([]map[string]any)
	if !ok || len(geoms) != 1 {
		t.Fatalf("expected one lowered geometry, got %T %v", lowered["geometries"], lowered["geometries"])
	}
	mats, ok := lowered["materials"].([]map[string]any)
	if !ok || len(mats) != 1 {
		t.Fatalf("expected one lowered material, got %T %v", lowered["materials"], lowered["materials"])
	}

	obj, ok := lowered["object"].(map[string]any)
	if !ok {
		t.Fatalf("expected object payload map, got %T", lowered["object"])
	}
	if obj["type"] != "Mesh" {
		t.Fatalf("unexpected object type: %v", obj["type"])
	}
	matrix, ok := obj["matrix"].([]float64)
	if !ok || len(matrix) != 16 {
		t.Fatalf("expected 4x4 flattened intrinsic matrix, got %T len=%d", obj["matrix"], len(matrix))
	}
}
