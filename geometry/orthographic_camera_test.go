package geometry

import "testing"

func TestOrthographicCameraLower(t *testing.T) {
	cam := NewOrthographicCamera(-1, 1, 2, -2, 0.1, 10, 1.5)
	lowered := cam.Lower()

	obj, ok := lowered["object"].(map[string]any)
	if !ok {
		t.Fatalf("expected camera object payload map, got %T", lowered["object"])
	}

	if obj["type"] != "OrthographicCamera" {
		t.Fatalf("unexpected camera type: %v", obj["type"])
	}
	if obj["left"] != -1.0 || obj["right"] != 1.0 || obj["top"] != 2.0 || obj["bottom"] != -2.0 {
		t.Fatalf("unexpected orthographic extents: %v", obj)
	}
	if obj["near"] != 0.1 || obj["far"] != 10.0 || obj["zoom"] != 1.5 {
		t.Fatalf("unexpected near/far/zoom: %v", obj)
	}
}
