package geometry
package geometry

import "testing"

func TestPerspectiveCameraDefaultsAndLower(t *testing.T) {
	cam := NewPerspectiveCamera()
	lowered := cam.Lower()

	obj, ok := lowered["object"].(map[string]any)
	if !ok {
		t.Fatalf("expected camera object payload map, got %T", lowered["object"])
	}

	if obj["type"] != "PerspectiveCamera" {
		t.Fatalf("unexpected camera type: %v", obj["type"])
	}

	if obj["fov"] != 50.0 || obj["aspect"] != 1.0 || obj["near"] != 0.1 || obj["far"] != 2000.0 {
		t.Fatalf("unexpected perspective defaults: %v", obj)
	}
	if obj["zoom"] != 1.0 || obj["filmGauge"] != 35.0 || obj["filmOffset"] != 0.0 || obj["focus"] != 10.0 {
		t.Fatalf("unexpected perspective camera settings: %v", obj)
	}
}
