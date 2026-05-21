package geometry

type Box struct {
	GeometryBase
	Lengths [3]float64
}

func NewBox(lengths [3]float64) *Box {
	return &Box{
		GeometryBase: NewGeometryBase(),
		Lengths:      lengths,
	}
}

func (b *Box) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":  b.UUIDString(),
		"type":  "BoxGeometry",
		"width": b.Lengths[0],
		"height": b.Lengths[1],
		"depth": b.Lengths[2],
	}
}
