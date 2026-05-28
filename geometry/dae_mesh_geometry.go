package geometry

import (
	"io"
	"os"
)

type DaeMeshGeometry struct {
	*MeshGeometry
}

func NewDaeMeshGeometry(contents string) *DaeMeshGeometry {
	return &DaeMeshGeometry{MeshGeometry: NewMeshGeometry(contents, "dae")}
}

func NewDaeMeshGeometryFromFile(path string) (*DaeMeshGeometry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewDaeMeshGeometry(string(b)), nil
}

func NewDaeMeshGeometryFromReader(r io.Reader) (*DaeMeshGeometry, error) {
	b, err := readAll(r)
	if err != nil {
		return nil, err
	}
	return NewDaeMeshGeometry(string(b)), nil
}
