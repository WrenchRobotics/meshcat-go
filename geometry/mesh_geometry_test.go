package geometry

import "testing"

func TestMeshGeometryLowerSchema(t *testing.T) {
	g := NewMeshGeometry("v 0 0 0", "obj")
	lowered := g.Lower(NewLowerContext())

	if lowered["type"] != "_meshfile_geometry" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
	if lowered["format"] != "obj" {
		t.Fatalf("unexpected format: %v", lowered["format"])
	}
	if lowered["data"] != "v 0 0 0" {
		t.Fatalf("unexpected data payload: %v", lowered["data"])
	}
}
