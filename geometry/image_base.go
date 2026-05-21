package geometry

type Image interface {
	referenceElement
}

type ImageBase struct {
	SceneElement
}

func NewImageBase() ImageBase {
	return ImageBase{SceneElement: NewSceneElement("")}
}

func (i *ImageBase) addToContext(ctx *LowerContext, payload map[string]any) {
	ctx.Images = append(ctx.Images, payload)
}
