package query

type PrimaryKeyer interface {
	PrimaryKeys() []*ColumnNode
}

// TableNodeI is the interface that all table-like nodes must satisfy
type TableNodeI interface {
	Node
	PrimaryKeyer
	ColumnNodes_() []Node
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
