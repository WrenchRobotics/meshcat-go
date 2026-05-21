package geometry

type Texture interface {
	referenceElement
}

type TextureBase struct {
	SceneElement
}

func NewTextureBase() TextureBase {
	return TextureBase{SceneElement: NewSceneElement("")}
}

func (t *TextureBase) addToContext(ctx *LowerContext, payload map[string]any) {
	ctx.Textures = append(ctx.Textures, payload)
}
