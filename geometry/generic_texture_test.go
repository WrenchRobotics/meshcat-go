package geometry

import "testing"

func TestGenericTextureLowersImagePropertyReference(t *testing.T) {
	img := NewPngImage([]byte{0x0a})
	gt := NewGenericTexture(map[string]any{
		"image": img,
		"foo":   "bar",
	})

	ctx := NewLowerContext()
	lowered := gt.Lower(ctx)

	if lowered["foo"] != "bar" {
		t.Fatalf("expected passthrough property foo=bar, got %v", lowered["foo"])
	}
	if _, ok := lowered["image"].(string); !ok {
		t.Fatalf("expected lowered image to be UUID string, got %T", lowered["image"])
	}
	if len(ctx.Images) != 1 {
		t.Fatalf("expected one image in context, got %d", len(ctx.Images))
	}
}
