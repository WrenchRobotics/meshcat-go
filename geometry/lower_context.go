package geometry

// LowerContext accumulates referenced assets while lowering an Object payload.
type LowerContext struct {
	Geometries []map[string]any
	Materials  []map[string]any
	Textures   []map[string]any
	Images     []map[string]any
}

func NewLowerContext() *LowerContext {
	return &LowerContext{}
}
