package geometry

func NewLineLoop(geometry Geometry, material Material) *Object {
	return NewObject("LineLoop", geometry, material)
}
