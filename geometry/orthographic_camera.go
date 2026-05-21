package geometry

type OrthographicCamera struct {
	SceneElement
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
	Near   float64
	Far    float64
	Zoom   float64
}

func NewOrthographicCamera(left, right, top, bottom, near, far float64, zoom float64) *OrthographicCamera {
	return &OrthographicCamera{
		SceneElement: NewSceneElement(""),
		Left:         left,
		Right:        right,
		Top:          top,
		Bottom:       bottom,
		Near:         near,
		Far:          far,
		Zoom:         zoom,
	}
}

func (c *OrthographicCamera) Lower() map[string]any {
	return map[string]any{
		"object": map[string]any{
			"uuid":   c.UUIDString(),
			"type":   "OrthographicCamera",
			"left":   c.Left,
			"right":  c.Right,
			"top":    c.Top,
			"bottom": c.Bottom,
			"near":   c.Near,
			"far":    c.Far,
			"zoom":   c.Zoom,
		},
	}
}
