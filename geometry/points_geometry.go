package geometry

import "fmt"

type PointsGeometry struct {
	GeometryBase
	Position [][]float32
	Color    [][]float32
}

func NewPointsGeometry(position [][]float32, color [][]float32) (*PointsGeometry, error) {
	if len(position) == 0 {
		return nil, fmt.Errorf("position must not be empty")
	}
	_, err := itemSize2D(position)
	if err != nil {
		return nil, err
	}
	if color != nil {
		if len(color) != len(position) {
			return nil, fmt.Errorf("color must match position shape")
		}
		for i := range color {
			if len(color[i]) != len(position[i]) {
				return nil, fmt.Errorf("color must match position shape")
			}
		}
	}

	return &PointsGeometry{
		GeometryBase: NewGeometryBase(),
		Position:     position,
		Color:        color,
	}, nil
}

func (g *PointsGeometry) Lower(_ *LowerContext) map[string]any {
	attrs := map[string]any{}
	pos, err := packFloat32Array2D(g.Position)
	if err != nil {
		panic(err)
	}
	attrs["position"] = pos
	if g.Color != nil {
		c, err := packFloat32Array2D(g.Color)
		if err != nil {
			panic(err)
		}
		attrs["color"] = c
	}

	return map[string]any{
		"uuid": g.UUIDString(),
		"type": "BufferGeometry",
		"data": map[string]any{
			"attributes": attrs,
		},
	}
}
