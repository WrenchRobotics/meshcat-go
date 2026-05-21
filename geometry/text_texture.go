package geometry

type TextTexture struct {
	TextureBase
	Text     string
	FontSize int
	FontFace string
}

func NewTextTexture(text string, fontSize int, fontFace string) *TextTexture {
	if fontSize == 0 {
		fontSize = 100
	}
	if fontFace == "" {
		fontFace = "sans-serif"
	}

	return &TextTexture{
		TextureBase: NewTextureBase(),
		Text:        text,
		FontSize:    fontSize,
		FontFace:    fontFace,
	}
}

func (t *TextTexture) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":      t.UUIDString(),
		"type":      "_text",
		"text":      t.Text,
		"font_size": t.FontSize,
		"font_face": t.FontFace,
	}
}
