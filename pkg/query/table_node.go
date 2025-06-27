package query

type PrimaryKeyer interface {
	PrimaryKeys() []*ColumnNode
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

// NodePrimaryKeys returns the primary key nodes of a table type node.
func NodePrimaryKeys(n Node) (nodes []Node) {
	if tn, ok := n.(PrimaryKeyer); ok {
		for _, n2 := range tn.PrimaryKeys() {
			nodes = append(nodes, n2)
		}
	}
	return
}
