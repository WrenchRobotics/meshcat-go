package geometry

import "encoding/base64"

type PngImage struct {
	ImageBase
	Data []byte
}

func NewPngImage(data []byte) *PngImage {
	return &PngImage{
		ImageBase: NewImageBase(),
		Data:      data,
	}
}

func (p *PngImage) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid": p.UUIDString(),
		"url":  "data:image/png;base64," + base64.StdEncoding.EncodeToString(p.Data),
	}
}
