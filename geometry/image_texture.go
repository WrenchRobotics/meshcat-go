package geometry

type ImageTexture struct {
	TextureBase
	Image      Image
	Wrap       [2]int
	Repeat     [2]int
	Properties map[string]any
}

func NewImageTexture(image Image) *ImageTexture {
	return &ImageTexture{
		TextureBase: NewTextureBase(),
		Image:       image,
		Wrap:        [2]int{1001, 1001},
		Repeat:      [2]int{1, 1},
		Properties:  map[string]any{},
	}
}

func (t *ImageTexture) Lower(ctx *LowerContext) map[string]any {
	data := map[string]any{
		"uuid":   t.UUIDString(),
		"wrap":   []int{t.Wrap[0], t.Wrap[1]},
		"repeat": []int{t.Repeat[0], t.Repeat[1]},
	}
	if t.Image != nil {
		data["image"] = lowerInObject(t.Image, ctx)
	}

	for k, v := range t.Properties {
		data[k] = v
	}

	return data
}
