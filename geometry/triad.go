package geometry

func NewTriad(scale float32) *Object {
	if scale == 0 {
		scale = 1
	}

	position := [][]float32{
		{0, scale, 0, 0, 0, 0},
		{0, 0, 0, scale, 0, 0},
		{0, 0, 0, 0, 0, scale},
	}
	color := [][]float32{
		{1, 1, 0, 0.6, 0, 0},
		{0, 0.6, 1, 1, 0, 0.6},
		{0, 0, 0, 0, 1, 1},
	}

	geom, err := NewPointsGeometry(position, color)
	if err != nil {
		panic(err)
	}
	mat := NewLineBasicMaterial()
	mat.VertexColors = true
	return NewLineSegments(geom, mat)
}
