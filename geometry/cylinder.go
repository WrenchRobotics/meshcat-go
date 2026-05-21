package geometry

type Cylinder struct {
	GeometryBase
	RadiusTop      float64
	RadiusBottom   float64
	Height         float64
	RadialSegments int
}

func NewCylinder(height float64, radius float64, radiusTop *float64, radiusBottom *float64) *Cylinder {
	rt := radius
	rb := radius
	if radiusTop != nil && radiusBottom != nil {
		rt = *radiusTop
		rb = *radiusBottom
	}

	return &Cylinder{
		GeometryBase:   NewGeometryBase(),
		RadiusTop:      rt,
		RadiusBottom:   rb,
		Height:         height,
		RadialSegments: 50,
	}
}

func (c *Cylinder) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":           c.UUIDString(),
		"type":           "CylinderGeometry",
		"radiusTop":      c.RadiusTop,
		"radiusBottom":   c.RadiusBottom,
		"height":         c.Height,
		"radialSegments": c.RadialSegments,
	}
}
