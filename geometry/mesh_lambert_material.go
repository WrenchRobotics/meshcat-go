package geometry

type MeshLambertMaterial struct {
	*GenericMaterial
}

func NewMeshLambertMaterial() *MeshLambertMaterial {
	return &MeshLambertMaterial{GenericMaterial: NewGenericMaterial("MeshLambertMaterial")}
}
