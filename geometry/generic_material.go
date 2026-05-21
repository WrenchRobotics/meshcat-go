package geometry

type GenericMaterial struct {
	MaterialBase
	Type               string
	Color              int
	Reflectivity       float64
	Map                Texture
	Side               int
	Transparent        *bool
	Opacity            float64
	LineWidth          float64
	Wireframe          bool
	WireframeLineWidth float64
	VertexColors       bool
	Properties         map[string]any
}

func NewGenericMaterial(materialType string) *GenericMaterial {
	return &GenericMaterial{
		MaterialBase:       NewMaterialBase(),
		Type:               materialType,
		Color:              0xffffff,
		Reflectivity:       0.5,
		Side:               2,
		Opacity:            1.0,
		LineWidth:          1.0,
		Wireframe:          false,
		WireframeLineWidth: 1.0,
		VertexColors:       false,
		Properties:         map[string]any{},
	}
}

func (m *GenericMaterial) Lower(ctx *LowerContext) map[string]any {
	transparent := m.Opacity != 1.0
	if m.Transparent != nil {
		transparent = *m.Transparent
	}

	data := map[string]any{
		"uuid":               m.UUIDString(),
		"type":               m.Type,
		"color":              m.Color,
		"reflectivity":       m.Reflectivity,
		"side":               m.Side,
		"transparent":        transparent,
		"opacity":            m.Opacity,
		"linewidth":          m.LineWidth,
		"wireframe":          m.Wireframe,
		"wireframeLinewidth": m.WireframeLineWidth,
		"vertexColors":       map[bool]int{true: 2, false: 0}[m.VertexColors],
	}

	for k, v := range m.Properties {
		data[k] = v
	}

	if m.Map != nil {
		data["map"] = lowerInObject(m.Map, ctx)
	}

	return data
}
