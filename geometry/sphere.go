package geometry

type Sphere struct {
	GeometryBase
	Radius float64
}

func NewSphere(radius float64) *Sphere {
	return &Sphere{
		GeometryBase: NewGeometryBase(),
		Radius:       radius,
	}
}

func (s *Sphere) Lower(_ *LowerContext) map[string]any {
	return map[string]any{
		"uuid":           s.UUIDString(),
		"type":           "SphereGeometry",
		"radius":         s.Radius,
		"widthSegments":  20,
		"heightSegments": 20,
	}
}
