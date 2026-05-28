package geometry

import (
	"testing"
)

func TestSceneElementAutoUUID(t *testing.T) {
	a := NewSceneElement("")
	b := NewSceneElement("")

	if a.UUID == "" || b.UUID == "" {
		t.Fatal("expected auto-generated UUIDs to be non-empty")
	}
	if a.UUID == b.UUID {
		t.Fatal("expected unique UUIDs for separate scene elements")
	}
}
