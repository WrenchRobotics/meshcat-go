package geometry

type GenericTexture struct {
	TextureBase
	Properties map[string]any
}

func NewGenericTexture(properties map[string]any) *GenericTexture {
	if properties == nil {
		properties = map[string]any{}
	}

	return &GenericTexture{
		TextureBase: NewTextureBase(),
		Properties:  properties,
	}
}

func (t *GenericTexture) Lower(ctx *LowerContext) map[string]any {
	data := map[string]any{"uuid": t.UUIDString()}
	for k, v := range t.Properties {
		data[k] = v
	}

	if imageAny, ok := data["image"]; ok {
		if image, ok := imageAny.(Image); ok {
			data["image"] = lowerInObject(image, ctx)
		}
	}

	return data
}
