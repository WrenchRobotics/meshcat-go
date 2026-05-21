package geometry
package geometry

import (
	"bytes"
	"testing"
)

func TestObjMeshGeometryFromReader(t *testing.T) {
	obj, err := NewObjMeshGeometryFromReader(bytes.NewBufferString("o cube"))
	if err != nil {
		t.Fatalf("NewObjMeshGeometryFromReader returned error: %v", err)
	}
	if obj.MeshFormat != "obj" || obj.Contents != "o cube" {
		t.Fatalf("unexpected obj geometry: format=%v data=%v", obj.MeshFormat, obj.Contents)
	}
}
