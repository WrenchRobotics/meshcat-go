package scene

import (
	"testing"
)

func TestNewTreeNodeInitializesDefaults(t *testing.T) {
	node := NewTreeNode()

	if node == nil {
		t.Fatal("expected NewTreeNode to return a non-nil node")
	}
	if node.Children == nil {
		t.Fatal("expected Children map to be initialized")
	}
	if len(node.Children) != 0 {
		t.Fatalf("expected empty Children map, got %d entries", len(node.Children))
	}
	if node.Properties == nil {
		t.Fatal("expected Properties slice to be initialized")
	}
	if len(node.Properties) != 0 {
		t.Fatalf("expected empty Properties slice, got %d entries", len(node.Properties))
	}
	if node.Object != nil {
		t.Fatalf("expected Object to be nil by default, got %v", node.Object)
	}
	if node.Transform != nil {
		t.Fatalf("expected Transform to be nil by default, got %v", node.Transform)
	}
	if node.Animation != nil {
		t.Fatalf("expected Animation to be nil by default, got %v", node.Animation)
	}
}

func TestChildCreatesAndReusesChildNode(t *testing.T) {
	root := NewTreeNode()

	child1 := root.Child("camera")
	if child1 == nil {
		t.Fatal("expected Child to return a non-nil node")
	}
	if got := len(root.Children); got != 1 {
		t.Fatalf("expected one child after first lookup, got %d", got)
	}

	child2 := root.Child("camera")
	if child2 != child1 {
		t.Fatal("expected Child to return the existing node for the same key")
	}
	if got := len(root.Children); got != 1 {
		t.Fatalf("expected child count to remain one, got %d", got)
	}
}

func TestChildInitializesNilChildrenMap(t *testing.T) {
	root := &TreeNode{}

	child := root.Child("lights")
	if child == nil {
		t.Fatal("expected Child to create a node when map is nil")
	}
	if root.Children == nil {
		t.Fatal("expected Child to initialize nil Children map")
	}
	if got := len(root.Children); got != 1 {
		t.Fatalf("expected one child after creation, got %d", got)
	}
}

func TestGetPathCreatesPathAndReturnsTerminalNode(t *testing.T) {
	root := NewTreeNode()
	path := []string{"world", "robot", "arm"}

	terminal := root.GetPath(path)
	if terminal == nil {
		t.Fatal("expected GetPath to return a non-nil terminal node")
	}

	second := root.GetPath(path)
	if second != terminal {
		t.Fatal("expected GetPath to return the same terminal node for an existing path")
	}

	world, ok := root.Children["world"]
	if !ok || world == nil {
		t.Fatal("expected GetPath to create first segment 'world'")
	}
	robot, ok := world.Children["robot"]
	if !ok || robot == nil {
		t.Fatal("expected GetPath to create second segment 'robot'")
	}
	arm, ok := robot.Children["arm"]
	if !ok || arm == nil {
		t.Fatal("expected GetPath to create terminal segment 'arm'")
	}
	if arm != terminal {
		t.Fatal("expected terminal node to match created path node")
	}
}

func TestGetPathWithEmptyPathReturnsRoot(t *testing.T) {
	root := NewTreeNode()

	got := root.GetPath(nil)
	if got != root {
		t.Fatal("expected GetPath with empty path to return root")
	}

	got = root.GetPath([]string{})
	if got != root {
		t.Fatal("expected GetPath with empty slice to return root")
	}
}

func TestFindPathReturnsNodeWhenPathExists(t *testing.T) {
	root := NewTreeNode()
	expected := root.GetPath([]string{"scene", "table", "leg"})

	got, ok := root.FindPath([]string{"scene", "table", "leg"})
	if !ok {
		t.Fatal("expected FindPath to report existing path")
	}
	if got != expected {
		t.Fatal("expected FindPath to return terminal node for existing path")
	}
}

func TestFindPathReturnsNilAndDoesNotCreateWhenMissing(t *testing.T) {
	root := NewTreeNode()
	root.GetPath([]string{"existing"})

	before := len(root.Children)
	got, ok := root.FindPath([]string{"missing", "branch"})
	if ok {
		t.Fatal("expected FindPath to report missing path")
	}
	if got != nil {
		t.Fatalf("expected nil node for missing path, got %v", got)
	}
	if after := len(root.Children); after != before {
		t.Fatalf("expected FindPath not to create nodes; child count changed from %d to %d", before, after)
	}
	if _, exists := root.Children["missing"]; exists {
		t.Fatal("expected missing path segment not to be created")
	}
}

func TestFindPathWithEmptyPathReturnsRoot(t *testing.T) {
	root := NewTreeNode()

	got, ok := root.FindPath(nil)
	if !ok {
		t.Fatal("expected FindPath with empty path to succeed")
	}
	if got != root {
		t.Fatal("expected FindPath with empty path to return root")
	}

	got, ok = root.FindPath([]string{})
	if !ok {
		t.Fatal("expected FindPath with empty slice to succeed")
	}
	if got != root {
		t.Fatal("expected FindPath with empty slice to return root")
	}
}
