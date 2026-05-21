package geometry

import (
	"testing"
)

func TestCylinderDefaultsAndOverrides(t *testing.T) {
	c1 := NewCylinder(1.5, 0.2, nil, nil)
	m1 := c1.Lower(NewLowerContext())
	if m1["radiusTop"] != 0.2 || m1["radiusBottom"] != 0.2 {
		t.Fatalf("expected default top/bottom radius from radius arg, got %v", m1)
	}

	rt := 0.3
	rb := 0.1
	c2 := NewCylinder(1.5, 0.2, &rt, &rb)
	m2 := c2.Lower(NewLowerContext())
	if m2["radiusTop"] != 0.3 || m2["radiusBottom"] != 0.1 {
		t.Fatalf("expected explicit top/bottom radius overrides, got %v", m2)
	}
}
