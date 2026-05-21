package geometry

import "testing"

func TestNewPointCloudBuildsPointsObject(t *testing.T) {
	pc, err := NewPointCloud(
		[][]float32{{0, 1}, {2, 3}, {4, 5}},
		[][]float32{{1, 0}, {0, 1}, {0.5, 0.5}},
		WithPointsSize(0.01),
	)
	if err != nil {
		t.Fatalf("NewPointCloud returned error: %v", err)
	}
	lowered := pc.Lower()
	obj := lowered["object"].(map[string]any)
	if obj["type"] != "Points" {
		t.Fatalf("unexpected object type: %v", obj["type"])
	}
	mats := lowered["materials"].([]map[string]any)
	if mats[0]["type"] != "PointsMaterial" {
		t.Fatalf("unexpected material type: %v", mats[0]["type"])
	}
	if mats[0]["size"] != 0.01 {
		t.Fatalf("unexpected points size: %v", mats[0]["size"])
	}
}
