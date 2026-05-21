package geometry

import "testing"

func TestTriangularMeshGeometryValidation(t *testing.T) {
	if _, err := NewTriangularMeshGeometry([][]float32{{0, 1}}, [][]uint32{{0, 1, 2}}, nil); err == nil {
		t.Fatal("expected vertices shape validation error")
	}
	if _, err := NewTriangularMeshGeometry([][]float32{{0, 1, 2}}, [][]uint32{{0, 1}}, nil); err == nil {
		t.Fatal("expected faces shape validation error")
	}
	if _, err := NewTriangularMeshGeometry(
		[][]float32{{0, 1, 2}, {3, 4, 5}},
		[][]uint32{{0, 1, 1}},
		[][]float32{{1, 0, 0}},
	); err == nil {
		t.Fatal("expected color shape validation error")
	}
}

func TestTriangularMeshGeometryLowerIncludesAttributesAndIndex(t *testing.T) {
	g, err := NewTriangularMeshGeometry(
		[][]float32{{0, 0, 0}, {1, 0, 0}, {0, 1, 0}},
		[][]uint32{{0, 1, 2}},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}

	lowered := g.Lower(NewLowerContext())
	if lowered["type"] != "BufferGeometry" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
	data, ok := lowered["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data map, got %T", lowered["data"])
	}
	attrs, ok := data["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("expected attributes map, got %T", data["attributes"])
	}
	if _, ok := attrs["position"].(map[string]any); !ok {
		t.Fatalf("expected packed position attribute, got %T", attrs["position"])
	}
	if _, ok := data["index"].(map[string]any); !ok {
		t.Fatalf("expected packed index attribute, got %T", data["index"])
	}
}
