package query

type PrimaryKeyer interface {
	PrimaryKey() *ColumnNode
}

// TableNodeI is the interface that all table-like nodes must adhere to
type TableNodeI interface {
	Node
	PrimaryKeyer
	ColumnNodes_() []Node
}

// NodeIsTable returns true if n is a table-like node.
// This includes top level table nodes, forward and reverse references and many-many references.
func NodeIsTable(n Node) bool {
	_, ok := n.(TableNodeI)
	return ok
}

// NodePrimaryKey returns the primary key of a node, if it has a primary key. Otherwise, returns nil.
func NodePrimaryKey(n Node) Node {
	if tn, ok := n.(PrimaryKeyer); ok {
		return tn.PrimaryKey()
	}
	return nil
}
