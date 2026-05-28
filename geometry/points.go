package geometry

func NewPoints(geometry Geometry, material Material) *Object {
	return NewObject("Points", geometry, material)
}
