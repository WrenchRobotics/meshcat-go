package scene

type TreeNode struct {
	Children   map[string]*TreeNode
	Object     any
	Transform  any
	Properties []any
	Animation  any
}

func NewTreeNode() *TreeNode {
	return &TreeNode{
		Children:   make(map[string]*TreeNode),
		Properties: make([]any, 0),
	}
}

// Child returns the named child node, creating it if missing (Python defaultdict behavior).
func (n *TreeNode) Child(key string) *TreeNode {
	if n.Children == nil {
		n.Children = make(map[string]*TreeNode)
	}
	if child, ok := n.Children[key]; ok {
		return child
	}
	child := NewTreeNode()
	n.Children[key] = child
	return child
}

// GetPath walks/creates the path and returns the terminal node.
func (n *TreeNode) GetPath(path []string) *TreeNode {
	cur := n
	for _, k := range path {
		cur = cur.Child(k)
	}
	return cur
}

// FindPath walks the path WITHOUT creating nodes (safe "find").
// Returns (node, true) if found, else (nil, false).
func (n *TreeNode) FindPath(path []string) (*TreeNode, bool) {
	cur := n
	for _, k := range path {
		next, ok := cur.Children[k]
		if !ok {
			return nil, false
		}
		cur = next
	}
	return cur, true
}
