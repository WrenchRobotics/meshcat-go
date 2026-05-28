package geometry

import "io"

type MeshGeometry struct {
	GeometryBase
	Contents   any
	MeshFormat string
}

func NewMeshGeometry(contents any, meshFormat string) *MeshGeometry {
	return &MeshGeometry{
		GeometryBase: NewGeometryBase(),
		Contents:     contents,
		MeshFormat:   meshFormat,
	}
}

func (g *MeshGeometry) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"type":   "_meshfile_geometry",
		"uuid":   g.UUIDString(),
		"format": g.MeshFormat,
		"data":   g.Contents,
	}
}

func readAll(r io.Reader) ([]byte, error) {
	if r == nil {
		return nil, nil
	}
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 2048)
	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return buf, nil
}
