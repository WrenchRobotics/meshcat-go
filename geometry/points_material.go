package geometry

type PointsMaterial struct {
	MaterialBase
	Size  float64
	Color int
}

func NewPointsMaterial() *PointsMaterial {
	return &PointsMaterial{
		MaterialBase: NewMaterialBase(),
		Size:         0.001,
		Color:        0xffffff,
	}
}

func (m *PointsMaterial) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":         m.UUIDString(),
		"type":         "PointsMaterial",
		"color":        m.Color,
		"size":         m.Size,
		"vertexColors": 2,
	}
}
