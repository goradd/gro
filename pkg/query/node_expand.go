package query

type Expander interface {
	Expand()
	IsExpanded() bool
}

type nodeExpand struct {
	isExpanded bool
}

func (n *nodeExpand) Expand() {
	n.isExpanded = true
}

func (n *nodeExpand) IsExpanded() bool {
	return n.isExpanded
}
