package query

// Aliaser is an interface for nodes that can be given an alias.
type Aliaser interface {
	Node
	// SetAlias sets a unique name for the node as used in a database query.
	SetAlias(string)
	// GetAlias returns the alias that was used in a database query.
	Alias() string
}

// Nodes that can have an alias can mix this in
type nodeAlias struct {
	alias string
}

// SetAlias sets a name to use for the node in the result of a query.
func (n *nodeAlias) SetAlias(a string) {
	n.alias = a
}

// Alias returns the alias name for the node.
func (n *nodeAlias) Alias() string {
	return n.alias
}

// NodeAlias returns the alias used by the node, or an empty string if no alias
// is set, or the node is not an Aliaser.
func NodeAlias(n Node) string {
	if cn, ok := n.(Aliaser); ok {
		return cn.Alias()
	}
	return ""
}
