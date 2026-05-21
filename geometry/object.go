package geometry

import "gonum.org/v1/gonum/mat"

type Object struct {
	SceneElement
	Type     string
	Geometry Geometry
	Material Material
}

func NewObject(objectType string, geometry Geometry, material Material) *Object {
	return &Object{
		SceneElement: NewSceneElement(""),
		Type:         objectType,
		Geometry:     geometry,
		Material:     material,
	}
}

func (o *Object) Lower() map[string]any {
	ctx := NewLowerContext()
	geomUUID := lowerInObject(o.Geometry, ctx)
	matUUID := lowerInObject(o.Material, ctx)

	intrinsic := o.Geometry.IntrinsicTransform()

	data := map[string]any{
		"metadata": map[string]any{
			"version": 4.5,
			"type":    "Object",
		},
		"geometries": ctx.Geometries,
		"materials":  ctx.Materials,
		"object": map[string]any{
			"uuid":     o.UUIDString(),
			"type":     o.Type,
			"geometry": geomUUID,
			"material": matUUID,
			"matrix":   flattenDense(intrinsic),
		},
	}

	if len(ctx.Textures) > 0 {
		data["textures"] = ctx.Textures
	}
	if len(ctx.Images) > 0 {
		data["images"] = ctx.Images
	}

	return data
}

func flattenDense(m *mat.Dense) []float64 {
	r, c := m.Dims()
	out := make([]float64, 0, r*c)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			out = append(out, m.At(i, j))
		}
	}
	return out
}
