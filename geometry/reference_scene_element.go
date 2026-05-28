package geometry

type referenceElement interface {
	Lower(*LowerContext) map[string]any
	addToContext(*LowerContext, map[string]any)
	UUIDString() string
}

func lowerInObject(el referenceElement, ctx *LowerContext) string {
	payload := el.Lower(ctx)
	el.addToContext(ctx, payload)
	return el.UUIDString()
}
