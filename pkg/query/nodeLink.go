package query

type Parenter interface {
	Parent() NodeI
}

// The nodeLinker interface provides an interface to allow nodes to be linked in a parent chain
type nodeLinker interface {
	SetParent(NodeI)
	Parent() NodeI
}

// The nodeLink is designed to be a mixin for the basic node structure. It encapsulates the joining of nodes.
type nodeLink struct {
	// Parent of the join
	parentNode NodeI
}

func (n *nodeLink) SetParent(pn NodeI) {
	n.parentNode = pn
}

func (n *nodeLink) Parent() NodeI {
	return n.parentNode
}
