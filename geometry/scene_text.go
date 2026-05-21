package geometry

func NewSceneText(text string, width float64, height float64, fontSize int, fontFace string) *Object {
	if width == 0 {
		width = 10
	}
	if height == 0 {
		height = 10
	}

	plane := NewPlane(width, height, 1, 1)
	mat := NewMeshPhongMaterial()
	mat.Map = NewTextTexture(text, fontSize, fontFace)
	transparent := true
	mat.Transparent = &transparent
	mat.Properties["needsUpdate"] = true

	return NewMesh(plane, mat)
}
