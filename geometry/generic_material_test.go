package geometry

import "testing"

func TestGenericMaterialTransparencyDefaultBehavior(t *testing.T) {
	mat := NewMeshPhongMaterial()
	mat.Opacity = 0.5
	lowered := mat.Lower(NewLowerContext())
	if lowered["transparent"] != true {
		t.Fatalf("expected transparent=true when opacity != 1.0, got %v", lowered["transparent"])
	}

	forced := false
	mat.Transparent = &forced
	lowered2 := mat.Lower(NewLowerContext())
	if lowered2["transparent"] != false {
		t.Fatalf("expected explicit transparent override to win, got %v", lowered2["transparent"])
	}
}

func TestGenericMaterialPropertiesPassThrough(t *testing.T) {
	m := NewMeshBasicMaterial()
	m.Properties["needsUpdate"] = true
	m.Properties["custom"] = "x"

	lowered := m.Lower(NewLowerContext())
	if lowered["needsUpdate"] != true {
		t.Fatalf("expected needsUpdate passthrough, got %v", lowered["needsUpdate"])
	}
	if lowered["custom"] != "x" {
		t.Fatalf("expected custom passthrough, got %v", lowered["custom"])
	}
}
