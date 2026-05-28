package geometry

type Material interface {
	referenceElement
}

type MaterialBase struct {
	SceneElement
}

func NewMaterialBase() MaterialBase {
	return MaterialBase{SceneElement: NewSceneElement("")}
}

func (m *MaterialBase) addToContext(ctx *LowerContext, payload map[string]any) {
	ctx.Materials = append(ctx.Materials, payload)
}
