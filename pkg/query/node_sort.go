package query

// Sorter is the interface a node must satisfy to be able to be used in an OrderBy statement.
type Sorter interface {
	Node
	// Ascending specifies that the node should be sorted in ascending order.
	Ascending() Sorter
	// Descending specifies that the node should be sorted in descending order.
	Descending() Sorter
	// IsDescending returns true if the node is sorted in descending order.
	IsDescending() bool
}

/*
// Sample Ascending function for implementers
func (n *Me) Ascending() QueryNode {
	n.sortDescending = false
	return n
}

// Sample Descending function for implementers
func (n *Me) Descending() {
	n.sortDescending = true
}

// IsDescending returns true if the node is sorted in descending order.
func (n *Me) IsDescending() bool {
	return n.sortDescending
}

*/

// NodeIsDescending is used by the ORM to get the sort state of the node.
// Returns false if the node is not a sorter or is sorted ascending.
func NodeIsDescending(n Node) bool {
	if cn, ok := n.(Sorter); ok {
		return cn.IsDescending()
	}
	return false
}
