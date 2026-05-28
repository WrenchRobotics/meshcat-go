package geometry

import "testing"

func TestImageTextureAndObjectIncludeTexturesAndImages(t *testing.T) {
	img := NewPngImage([]byte{0x01, 0x02, 0x03})
	tex := NewImageTexture(img)

	mat := NewMeshPhongMaterial()
	mat.Map = tex
	obj := NewMesh(NewSphere(0.2), mat)
	lowered := obj.Lower()

	textures, ok := lowered["textures"].([]map[string]any)
	if !ok || len(textures) != 1 {
		t.Fatalf("expected one lowered texture, got %T %v", lowered["textures"], lowered["textures"])
	}
	images, ok := lowered["images"].([]map[string]any)
	if !ok || len(images) != 1 {
		t.Fatalf("expected one lowered image, got %T %v", lowered["images"], lowered["images"])
	}

	materials, ok := lowered["materials"].([]map[string]any)
	if !ok || len(materials) != 1 {
		t.Fatalf("expected one lowered material, got %T", lowered["materials"])
	}
	if materials[0]["map"] != textures[0]["uuid"] {
		t.Fatalf("expected material map UUID to reference texture UUID")
	}
	if textures[0]["image"] != images[0]["uuid"] {
		t.Fatalf("expected texture image UUID to reference image UUID")
	}
}
