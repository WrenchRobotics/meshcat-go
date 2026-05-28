package geometry

import "fmt"

type TriangularMeshGeometry struct {
	GeometryBase
	Vertices [][]float32
	Faces    [][]uint32
	Color    [][]float32
}

func NewTriangularMeshGeometry(vertices [][]float32, faces [][]uint32, color [][]float32) (*TriangularMeshGeometry, error) {
	if len(vertices) == 0 {
		return nil, fmt.Errorf("vertices must not be empty")
	}
	for _, row := range vertices {
		if len(row) != 3 {
			return nil, fmt.Errorf("vertices must be Nx3")
		}
	}
	if len(faces) == 0 {
		return nil, fmt.Errorf("faces must not be empty")
	}
	for _, row := range faces {
		if len(row) != 3 {
			return nil, fmt.Errorf("faces must be Mx3")
		}
	}
	if color != nil {
		if len(color) != len(vertices) {
			return nil, fmt.Errorf("color must match vertices shape")
		}
		for i := range color {
			if len(color[i]) != len(vertices[i]) {
				return nil, fmt.Errorf("color must match vertices shape")
			}
		}
	}

	return &TriangularMeshGeometry{
		GeometryBase: NewGeometryBase(),
		Vertices:     vertices,
		Faces:        faces,
		Color:        color,
	}, nil
}

func (g *TriangularMeshGeometry) Lower(_ *LowerContext) map[string]any {
	attrs := map[string]any{}
	pos, err := packFloat32Array2D(transposeFloat32(g.Vertices))
	if err != nil {
		panic(err)
	}
	attrs["position"] = pos

	if g.Color != nil {
		c, err := packFloat32Array2D(transposeFloat32(g.Color))
		if err != nil {
			panic(err)
		}
		attrs["color"] = c
	}

	index, err := packUint32Array2D(transposeUint32(g.Faces))
	if err != nil {
		panic(err)
	}

	return map[string]any{
		"uuid": g.UUIDString(),
		"type": "BufferGeometry",
		"data": map[string]any{
			"attributes": attrs,
			"index":      index,
		},
	}
}
