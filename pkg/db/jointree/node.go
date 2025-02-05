package jointree

import "github.com/goradd/orm/pkg/query"

type reverseNode struct {
	node  query.Node
	child *reverseNode
}

// reverse returns a reverseNode chain pointing to the root of the given node.
func reverse(node query.Node) *reverseNode {
	r := &reverseNode{node: node}
	for query.NodeParent(node) != nil {
		np := query.NodeParent(node)

		r = &reverseNode{np, r}
		node = np
	}
	return r
}

func nodeMatch(node1, node2 query.Node) bool {
	return query.NodesMatch(node1, node2)
}
