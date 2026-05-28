package geometry

import "testing"

func TestNewSceneTextIncludesMappedTexture(t *testing.T) {
	obj := NewSceneText("hello", 0, 0, 0, "")
	lowered := obj.Lower()

	o := lowered["object"].(map[string]any)
	if o["type"] != "Mesh" {
		t.Fatalf("unexpected object type: %v", o["type"])
	}
	textures, ok := lowered["textures"].([]map[string]any)
	if !ok || len(textures) != 1 {
		t.Fatalf("expected one texture in lowered scene text payload")
	}
	mats := lowered["materials"].([]map[string]any)
	if mats[0]["map"] != textures[0]["uuid"] {
		t.Fatalf("material map should reference texture UUID")
	}
	if mats[0]["transparent"] != true {
		t.Fatalf("scene text material should force transparent=true")
	}
	if mats[0]["needsUpdate"] != true {
		t.Fatalf("scene text material should set needsUpdate=true")
	}
}
