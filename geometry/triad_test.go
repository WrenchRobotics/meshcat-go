package geometry

import "testing"

func TestNewTriadCreatesLineSegmentsWithVertexColors(t *testing.T) {
	triad := NewTriad(2)
	lowered := triad.Lower()

	o := lowered["object"].(map[string]any)
	if o["type"] != "LineSegments" {
		t.Fatalf("unexpected triad object type: %v", o["type"])
	}
	mats := lowered["materials"].([]map[string]any)
	if mats[0]["type"] != "LineBasicMaterial" {
		t.Fatalf("unexpected triad material type: %v", mats[0]["type"])
	}
	if mats[0]["vertexColors"] != 2 {
		t.Fatalf("triad material should have vertexColors enum=2")
	}

	geoms := lowered["geometries"].([]map[string]any)
	data := geoms[0]["data"].(map[string]any)
	attrs := data["attributes"].(map[string]any)
	pos := attrs["position"].(map[string]any)
	if pos["itemSize"] != 3 {
		t.Fatalf("unexpected triad position itemSize: %v", pos["itemSize"])
	}
}
