package geometry

type Plane struct {
	GeometryBase
	Width          float64
	Height         float64
	WidthSegments  int
	HeightSegments int
}

func NewPlane(width float64, height float64, widthSegments int, heightSegments int) *Plane {
	return &Plane{
		GeometryBase:    NewGeometryBase(),
		Width:           width,
		Height:          height,
		WidthSegments:   widthSegments,
		HeightSegments:  heightSegments,
	}
}

func (p *Plane) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":           p.UUIDString(),
		"type":           "PlaneGeometry",
		"width":          p.Width,
		"height":         p.Height,
		"widthSegments":  p.WidthSegments,
		"heightSegments": p.HeightSegments,
	}
}
