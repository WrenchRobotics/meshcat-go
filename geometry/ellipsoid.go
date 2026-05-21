package geometry

import "gonum.org/v1/gonum/mat"

type Ellipsoid struct {
	*Sphere
	Radii [3]float64
}

func NewEllipsoid(radii [3]float64) *Ellipsoid {
	return &Ellipsoid{
		Sphere: NewSphere(1.0),
		Radii:  radii,
	}
}

func (e *Ellipsoid) IntrinsicTransform() *mat.Dense {
	return mat.NewDense(4, 4, []float64{
		e.Radii[0], 0, 0, 0,
		0, e.Radii[1], 0, 0,
		0, 0, e.Radii[2], 0,
		0, 0, 0, 1,
	})
}
