package geometry

import "testing"

func TestTextTextureLowerDefaults(t *testing.T) {
	tex := NewTextTexture("hi", 0, "")
	lowered := tex.Lower(NewLowerContext())

	if lowered["type"] != "_text" {
		t.Fatalf("unexpected type: %v", lowered["type"])
	}
	if lowered["font_size"] != 100 {
		t.Fatalf("unexpected default font_size: %v", lowered["font_size"])
	}
	if lowered["font_face"] != "sans-serif" {
		t.Fatalf("unexpected default font_face: %v", lowered["font_face"])
	}
}
