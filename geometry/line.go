package geometry

func NewLine(geometry Geometry, material Material) *Object {
	return NewObject("Line", geometry, material)
}
