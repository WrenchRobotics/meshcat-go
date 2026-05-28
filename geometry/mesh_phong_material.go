package geometry

type MeshPhongMaterial struct {
	*GenericMaterial
}

func NewMeshPhongMaterial() *MeshPhongMaterial {
	return &MeshPhongMaterial{GenericMaterial: NewGenericMaterial("MeshPhongMaterial")}
}
