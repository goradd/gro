package query

// The linker interface provides an interface to allow nodes to be linked in a parent chain
type linker interface {
	// SetParent sets the parent node of this node.
	setParent_(Node)
	// Parent returns the parent node of this node.
	parent_() Node
}

// The nodeLink is designed to be a mixin for the basic node structure. It encapsulates the joining of nodes.
type nodeLink struct {
	// Parent of the join
	parentNode Node
}

func (n *nodeLink) setParent_(pn Node) {
	n.parentNode = pn
}

func (n *nodeLink) parent_() Node {
	return n.parentNode
}

// NodeParent returns the parent of the node, or nil if the node has no parent.
func NodeParent(n Node) Node {
	if cn, ok := n.(linker); ok {
		return cn.parent_()
	}
	return nil
}

// NodeSetParent returns the parent of the node, or nil if the node has no parent.
func NodeSetParent(n Node, parent Node) {
	if cn, ok := n.(linker); ok {
		cn.setParent_(parent)
	}
}

// RootNode returns the end of the node chain, or the top most parent in the chain.
// Returns nil if this type of node does not have a root node.
func RootNode(n Node) Node {
	if linker, ok := n.(linker); !ok {
		if _, ok := n.(TableNodeI); ok {
			return n // found the top table
		}
		return nil // a node that does not connect to a root, like an operation node
	} else {
		return RootNode(linker.parent_())
	}
}
