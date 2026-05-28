package geometry

import (
	"strings"
	"testing"
)

func TestPngImageLowerDataURL(t *testing.T) {
	img := NewPngImage([]byte{0x89, 0x50, 0x4e, 0x47})
	lowered := img.Lower(NewLowerContext())

	url, ok := lowered["url"].(string)
	if !ok {
		t.Fatalf("expected string url, got %T", lowered["url"])
	}
	if !strings.HasPrefix(url, "data:image/png;base64,") {
		t.Fatalf("unexpected png data url prefix: %q", url)
	}
}
