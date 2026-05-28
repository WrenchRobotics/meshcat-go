package geometry

import "testing"

func TestPointsGeometryValidationAndLower(t *testing.T) {
	if _, err := NewPointsGeometry([][]float32{}, nil); err == nil {
		t.Fatal("expected empty position validation error")
	}
	if _, err := NewPointsGeometry(
		[][]float32{{0, 1}, {2, 3}},
		[][]float32{{0, 1}},
	); err == nil {
		t.Fatal("expected color shape validation error")
	}

	g, err := NewPointsGeometry(
		[][]float32{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	lowered := g.Lower(NewLowerContext())
	if lowered["type"] != "BufferGeometry" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
	data := lowered["data"].(map[string]any)
	attrs := data["attributes"].(map[string]any)
	pos := attrs["position"].(map[string]any)
	if pos["itemSize"] != 3 {
		t.Fatalf("unexpected points position itemSize: %v", pos["itemSize"])
	}
	if pos["type"] != "Float32Array" {
		t.Fatalf("unexpected points position type: %v", pos["type"])
	}
}
