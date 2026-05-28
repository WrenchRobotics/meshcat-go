package geometry

import (
	"io"
	"os"
)

type StlMeshGeometry struct {
	*MeshGeometry
}

func NewStlMeshGeometry(contents []byte) *StlMeshGeometry {
	return &StlMeshGeometry{MeshGeometry: NewMeshGeometry(contents, "stl")}
}

func NewStlMeshGeometryFromFile(path string) (*StlMeshGeometry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewStlMeshGeometry(b), nil
}

func NewStlMeshGeometryFromReader(r io.Reader) (*StlMeshGeometry, error) {
	b, err := readAll(r)
	if err != nil {
		return nil, err
	}
	return NewStlMeshGeometry(b), nil
}
