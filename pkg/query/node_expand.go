package query

// Expander is an interface used by reverse and many-many references to turn an array node into multiple
// rows in the result set, each row having a single item from the original array.
type Expander interface {
	// Expand will expand the array node into individual rows in the result set.
	Expand()
	// IsExpanded returns true if the node is expanded.
	IsExpanded() bool
}

// nodeExpand is a mixin for nodes that can be expanded.
type nodeExpand struct {
	isExpanded bool
}

// Expand will expand the array node into individual rows in the result set.
func (n *nodeExpand) Expand() {
	n.isExpanded = true
}

// IsExpanded returns true if the node is expanded.
func (n *nodeExpand) IsExpanded() bool {
	return n.isExpanded
}

// NodeIsExpanded is used by the ORM to see if a node is expanded.
func NodeIsExpanded(n Node) bool {
	if cn, ok := n.(Expander); ok {
		return cn.IsExpanded()
	}
	return false
}
