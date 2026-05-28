package geometry

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestStlMeshGeometryFromReader(t *testing.T) {
	want := []byte{0x01, 0x02, 0x03, 0x04}

	stlReader, err := NewStlMeshGeometryFromReader(bytes.NewReader(want))
	if err != nil {
		t.Fatalf("NewStlMeshGeometryFromReader returned error: %v", err)
	}
	got, ok := stlReader.Contents.([]byte)
	if !ok {
		t.Fatalf("expected []byte contents from reader, got %T", stlReader.Contents)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected stl reader bytes: got %v want %v", got, want)
	}
}

func TestStlMeshGeometryFromFile(t *testing.T) {
	want := []byte{0x01, 0x02, 0x03, 0x04}

	tmp := filepath.Join(t.TempDir(), "mesh.stl")
	if err := os.WriteFile(tmp, want, 0o644); err != nil {
		t.Fatalf("failed writing temp stl: %v", err)
	}
	stlFile, err := NewStlMeshGeometryFromFile(tmp)
	if err != nil {
		t.Fatalf("NewStlMeshGeometryFromFile returned error: %v", err)
	}
	got, ok := stlFile.Contents.([]byte)
	if !ok {
		t.Fatalf("expected []byte contents from file, got %T", stlFile.Contents)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected stl file bytes: got %v want %v", got, want)
	}
}
