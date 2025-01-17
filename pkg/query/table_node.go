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

// NodePrimaryKey returns the primary key of a node, if it has a primary key. Otherwise, returns nil.
func NodePrimaryKey(n Node) Node {
	if tn, ok := n.(PrimaryKeyer); ok {
		return tn.PrimaryKeyNode()
	}
	return nil
}
