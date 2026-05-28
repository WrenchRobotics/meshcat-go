package geometry

type MeshBasicMaterial struct {
	*GenericMaterial
}

func NewMeshBasicMaterial() *MeshBasicMaterial {
	return &MeshBasicMaterial{GenericMaterial: NewGenericMaterial("MeshBasicMaterial")}
}
