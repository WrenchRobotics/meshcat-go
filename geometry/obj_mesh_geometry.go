package geometry

import (
	"io"
	"os"
)

type ObjMeshGeometry struct {
	*MeshGeometry
}

func NewObjMeshGeometry(contents string) *ObjMeshGeometry {
	return &ObjMeshGeometry{MeshGeometry: NewMeshGeometry(contents, "obj")}
}

func NewObjMeshGeometryFromFile(path string) (*ObjMeshGeometry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewObjMeshGeometry(string(b)), nil
}

func NewObjMeshGeometryFromReader(r io.Reader) (*ObjMeshGeometry, error) {
	b, err := readAll(r)
	if err != nil {
		return nil, err
	}
	return NewObjMeshGeometry(string(b)), nil
}
