package meshcat

import (
	"testing"
)

type mockClient struct {
	lastPath      []string
	lastObject    interface{}
	lastTransform [4][4]float64
	lastProperty  string
	lastValue     interface{}
	lastDelete    []string
	openCalled    bool
}

func (m *mockClient) SetObject(path []string, obj interface{}) error {
	m.lastPath = path
	m.lastObject = obj
	return nil
}
func (m *mockClient) SetTransform(path []string, transform [4][4]float64) error {
	m.lastPath = path
	m.lastTransform = transform
	return nil
}
func (m *mockClient) SetProperty(path []string, property string, value interface{}) error {
	m.lastPath = path
	m.lastProperty = property
	m.lastValue = value
	return nil
}
func (m *mockClient) Delete(path []string) error {
	m.lastDelete = path
	return nil
}
func (m *mockClient) Open() error {
	m.openCalled = true
	return nil
}

func TestVisualizerPathNavigation(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc)
	v2 := v.At("foo").At("bar")
	if len(v2.path) != 2 || v2.path[0] != "foo" || v2.path[1] != "bar" {
		t.Errorf("unexpected path: %v", v2.path)
	}
}

func TestSetObject(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc).At("obj")
	obj := struct{ Name string }{"cube"}
	if err := v.SetObject(obj); err != nil {
		t.Fatal(err)
	}
	if mc.lastPath[0] != "obj" || mc.lastObject != obj {
		t.Errorf("SetObject failed: %v, %v", mc.lastPath, mc.lastObject)
	}
}

func TestSetTransform(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc).At("foo")
	var tf [4][4]float64
	tf[0][0] = 1
	if err := v.SetTransform(tf); err != nil {
		t.Fatal(err)
	}
	if mc.lastPath[0] != "foo" || mc.lastTransform != tf {
		t.Errorf("SetTransform failed: %v, %v", mc.lastPath, mc.lastTransform)
	}
}

func TestDelete(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc).At("bar")
	if err := v.Delete(); err != nil {
		t.Fatal(err)
	}
	if mc.lastDelete[0] != "bar" {
		t.Errorf("Delete failed: %v", mc.lastDelete)
	}
}

func TestSetProperty(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc).At("Background")
	value := []float64{1, 0, 0}
	if err := v.SetProperty("top_color", value); err != nil {
		t.Fatal(err)
	}
	if mc.lastPath[0] != "Background" || mc.lastProperty != "top_color" {
		t.Errorf("SetProperty failed: %v, %q", mc.lastPath, mc.lastProperty)
	}
	got, ok := mc.lastValue.([]float64)
	if !ok || len(got) != 3 || got[0] != 1 || got[1] != 0 || got[2] != 0 {
		t.Errorf("unexpected property value: %#v", mc.lastValue)
	}
}

func TestOpen(t *testing.T) {
	mc := &mockClient{}
	v := NewVisualizer(mc)
	if err := v.Open(); err != nil {
		t.Fatal(err)
	}
	if !mc.openCalled {
		t.Errorf("Open not called")
	}
}
