package geometry

type PerspectiveCamera struct {
	SceneElement
	Fov        float64
	Aspect     float64
	Near       float64
	Far        float64
	Zoom       float64
	FilmGauge  float64
	FilmOffset float64
	Focus      float64
}

func NewPerspectiveCamera() *PerspectiveCamera {
	return &PerspectiveCamera{
		SceneElement: NewSceneElement(""),
		Fov:          50,
		Aspect:       1,
		Near:         0.1,
		Far:          2000,
		Zoom:         1,
		FilmGauge:    35,
		FilmOffset:   0,
		Focus:        10,
	}
}

func (c *PerspectiveCamera) Lower() map[string]any {
	return map[string]any{
		"object": map[string]any{
			"uuid":       c.UUIDString(),
			"type":       "PerspectiveCamera",
			"aspect":     c.Aspect,
			"far":        c.Far,
			"filmGauge":  c.FilmGauge,
			"filmOffset": c.FilmOffset,
			"focus":      c.Focus,
			"fov":        c.Fov,
			"near":       c.Near,
			"zoom":       c.Zoom,
		},
	}
}
