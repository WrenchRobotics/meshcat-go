package geometry

import (
	"bytes"
	"testing"
)

func TestDaeMeshGeometryFromReader(t *testing.T) {
	dae, err := NewDaeMeshGeometryFromReader(bytes.NewBufferString("<COLLADA />"))
	if err != nil {
		t.Fatalf("NewDaeMeshGeometryFromReader returned error: %v", err)
	}
	if dae.MeshFormat != "dae" || dae.Contents != "<COLLADA />" {
		t.Fatalf("unexpected dae geometry: format=%v data=%v", dae.MeshFormat, dae.Contents)
	}
}
