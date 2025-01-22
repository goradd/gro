package query

type PrimaryKeyer interface {
	PrimaryKeyNode() *ColumnNode
}

// TableNodeI is the interface that all table-like nodes must adhere to
type TableNodeI interface {
	Node
	PrimaryKeyer
	ColumnNodes_() []Node
}

// NodeIsJoinable returns true if n is a node that can be used in a Builder.Join.
func NodeIsJoinable(n Node) bool {
	_, ok := n.(TableNodeI)
	return ok
}

// NodePrimaryKey returns the primary key of a node, if it has a primary key. Otherwise, returns nil.
func NodePrimaryKey(n Node) Node {
	if tn, ok := n.(PrimaryKeyer); ok {
		return tn.PrimaryKeyNode()
	}
	return nil
}
