package geometry

func NewLineSegments(geometry Geometry, material Material) *Object {
	return NewObject("LineSegments", geometry, material)
}
