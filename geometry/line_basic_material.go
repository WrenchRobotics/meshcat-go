package geometry

type LineBasicMaterial struct {
	*GenericMaterial
}

func NewLineBasicMaterial() *LineBasicMaterial {
	return &LineBasicMaterial{GenericMaterial: NewGenericMaterial("LineBasicMaterial")}
}
