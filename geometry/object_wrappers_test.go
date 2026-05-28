package geometry

import "testing"

func TestObjectWrapperTypes(t *testing.T) {
	geom := NewBox([3]float64{1, 1, 1})
	mat := NewLineBasicMaterial()

	cases := []struct {
		name string
		obj  *Object
		want string
	}{
		{name: "Points", obj: NewPoints(geom, mat), want: "Points"},
		{name: "Line", obj: NewLine(geom, mat), want: "Line"},
		{name: "LineSegments", obj: NewLineSegments(geom, mat), want: "LineSegments"},
		{name: "LineLoop", obj: NewLineLoop(geom, mat), want: "LineLoop"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lowered := tc.obj.Lower()
			objData, ok := lowered["object"].(map[string]any)
			if !ok {
				t.Fatalf("expected object payload map, got %T", lowered["object"])
			}
			if objData["type"] != tc.want {
				t.Fatalf("unexpected wrapper type: got %v want %v", objData["type"], tc.want)
			}
		})
	}
}
