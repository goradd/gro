package query

// NodeSorter is the interface a node must satisfy to be able to be used in an OrderBy statement.
type NodeSorter interface {
	Ascending() NodeI
	Descending() NodeI
	IsDescending() bool
}

type nodeSort struct {
	// Used by OrderBy clauses
	sortDescending bool
}

func (n *nodeSort) Ascending() {
	n.sortDescending = false
}

func (n *nodeSort) Descending() {
	n.sortDescending = true
}

func (n *nodeSort) IsDescending() bool {
	return n.sortDescending
}
