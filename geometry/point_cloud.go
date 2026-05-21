package geometry

type PointsMaterialOption func(*PointsMaterial)

func WithPointsSize(size float64) PointsMaterialOption {
	return func(m *PointsMaterial) {
		m.Size = size
	}
}

func WithPointsColor(color int) PointsMaterialOption {
	return func(m *PointsMaterial) {
		m.Color = color
	}
}

func NewPointCloud(position [][]float32, color [][]float32, opts ...PointsMaterialOption) (*Object, error) {
	geom, err := NewPointsGeometry(position, color)
	if err != nil {
		return nil, err
	}
	mat := NewPointsMaterial()
	for _, opt := range opts {
		opt(mat)
	}
	return NewPoints(geom, mat), nil
}
