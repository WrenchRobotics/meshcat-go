package geometry

type MeshToonMaterial struct {
	*GenericMaterial
}

func NewMeshToonMaterial() *MeshToonMaterial {
	return &MeshToonMaterial{GenericMaterial: NewGenericMaterial("MeshToonMaterial")}
}
