package geometry

func NewMesh(geometry Geometry, material Material) *Object {
	return NewObject("Mesh", geometry, material)
}
