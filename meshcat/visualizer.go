package meshcat

import (
	"fmt"
)

// Visualizer provides a high-level interface for manipulating a Meshcat scene.
type Visualizer struct {
	conn MeshcatClient // Abstract connection to Meshcat server
	path []string      // Path in the scene tree
}

// MeshcatClient abstracts the communication with the Meshcat server.
type MeshcatClient interface {
	SetObject(path []string, obj interface{}) error
	SetTransform(path []string, transform [4][4]float64) error
	SetProperty(path []string, property string, value interface{}) error
	Delete(path []string) error
	Open() error
}

// NewVisualizer creates a new Visualizer at the root path.
func NewVisualizer(conn MeshcatClient) *Visualizer {
	return &Visualizer{conn: conn, path: []string{}}
}

// At returns a new Visualizer at the given child path.
func (v *Visualizer) At(child string) *Visualizer {
	newPath := append([]string{}, v.path...)
	newPath = append(newPath, child)
	return &Visualizer{conn: v.conn, path: newPath}
}

// AtPath returns a new Visualizer at the given subpath.
func (v *Visualizer) AtPath(path ...string) *Visualizer {
	newPath := append([]string{}, v.path...)
	newPath = append(newPath, path...)
	return &Visualizer{conn: v.conn, path: newPath}
}

// SetObject sets a geometry object at the current path.
func (v *Visualizer) SetObject(obj interface{}) error {
	return v.conn.SetObject(v.path, obj)
}

// SetTransform sets a transform at the current path.
func (v *Visualizer) SetTransform(transform [4][4]float64) error {
	return v.conn.SetTransform(v.path, transform)
}

// SetProperty sets a named property at the current path.
func (v *Visualizer) SetProperty(property string, value interface{}) error {
	return v.conn.SetProperty(v.path, property, value)
}

// Delete removes the object or subtree at the current path.
func (v *Visualizer) Delete() error {
	return v.conn.Delete(v.path)
}

// Open opens the Meshcat viewer in a browser.
func (v *Visualizer) Open() error {
	return v.conn.Open()
}

// String returns a string representation of the current path.
func (v *Visualizer) String() string {
	return fmt.Sprintf("/scene/%v", v.path)
}
