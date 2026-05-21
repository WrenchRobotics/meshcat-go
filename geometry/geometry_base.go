package geometry

import "gonum.org/v1/gonum/mat"

type Geometry interface {
	referenceElement
	IntrinsicTransform() *mat.Dense
}

type GeometryBase struct {
	SceneElement
}

func NewGeometryBase() GeometryBase {
	return GeometryBase{SceneElement: NewSceneElement("")}
}

func (g *GeometryBase) IntrinsicTransform() *mat.Dense {
	return mat.NewDense(4, 4, []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	})
}

func (g *GeometryBase) addToContext(ctx *LowerContext, payload map[string]any) {
	ctx.Geometries = append(ctx.Geometries, payload)
}
