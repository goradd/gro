package jointree

import "github.com/goradd/orm/pkg/query"

type reverseNode struct {
	node  query.Node
	child *reverseNode
}

// reverse returns a reverseNode chain pointing to the root of the given node.
func reverse(node query.Node) *reverseNode {
	r := reverseNode{node: node}
	for node.(query.Linker).Parent() != nil {
		np := node.(query.Linker).Parent()
		r2 := reverseNode{np, &r}
		node = np
		r = r2
	}
	return &r
}

func nodeMatch(node1, node2 query.Node) bool {
	if node1.NodeType_() != node2.NodeType_() {
		return false
	}

	if node1.TableName_() != node2.TableName_() {
		return false
	}
	if node1.DatabaseKey_() != node2.DatabaseKey_() {
		return false
	}

	if a, ok := node1.(query.Aliaser); ok {
		if b, ok2 := node2.(query.Aliaser); ok2 {
			return a.Alias() == b.Alias()
		} else {
			return false
		}
	}

	switch node1.NodeType_() {
	case query.ColumnNodeType:
		c1 := node1.(*query.ColumnNode)
		c2 := node2.(*query.ColumnNode)
		return c1.QueryName == c2.QueryName
	case query.TableNodeType:
		return true // already know table names are equal
	case query.ReferenceNodeType:
		return node1.(query.ReferenceNodeI).ColumnName() == node2.(query.ReferenceNodeI).ColumnName()
	case query.ReverseNodeType:
		return node1.(query.ReverseNodeI).ColumnName() == node2.(query.ReverseNodeI).ColumnName()
	case query.ManyManyNodeType:
		return node1.(query.ManyManyNodeI).AssnTableName() == node2.(query.ManyManyNodeI).AssnTableName() &&
			node1.(query.ManyManyNodeI).ColumnName() == node2.(query.ManyManyNodeI).ColumnName()
	default:
		return false
	}
}
