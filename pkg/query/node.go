package query

// NodeType indicates the type of node, which saves us from having to use reflection to determine this
type NodeType int

const (
	UnknownNodeType NodeType = iota
	TableNodeType
	ColumnNodeType
	ReferenceNodeType // forward reference from a foreign key
	ManyManyNodeType
	ReverseNodeType
	ValueNodeType
	OperationNodeType
	AliasNodeType
	SubqueryNodeType
)

// String satisfies the fmt.Stringer interface for NodeType.
func (nt NodeType) String() string {
	switch nt {
	case UnknownNodeType:
		return "UnknownNodeType"
	case TableNodeType:
		return "TableNodeType"
	case ColumnNodeType:
		return "ColumnNodeType"
	case ReferenceNodeType:
		return "ReferenceNodeType"
	case ManyManyNodeType:
		return "ManyManyNodeType"
	case ReverseNodeType:
		return "ReverseNodeType"
	case ValueNodeType:
		return "ValueNodeType"
	case OperationNodeType:
		return "OperationNodeType"
	case AliasNodeType:
		return "AliasNodeType"
	case SubqueryNodeType:
		return "SubqueryNodeType"
	default:
		return "Unknown"
	}
}

type container interface {
	containedNodes() (nodes []Node)
}

// Node is the interface that all nodes must satisfy. A node is a representation of an object or a relationship
// between objects in a database that we use to create a query. It lets us abstract the structure of a database
// to be able to query any kind of database. Obviously, this doesn't work for all possible database structures, but
// it generally works well enough to solve most of the situations that you will come across.
type Node interface {
	// NodeType_ returns the type of the node
	NodeType_() NodeType
	// TableName_ returns the query name of the table the node is associated with. Not all nodes support this.
	TableName_() string
	// DatabaseKey_ is the database key of the database the node is associated with.
	DatabaseKey_() string
}

/**

Public Accessors

The following functions are designed primarily to be used by the db package to help it unpack queries. They are not
given an accessor at the beginning so that they do not show up as a function in editors that provide code hinting when
trying to put together a node chain during the code creation process. Essentially they are trying to create exported
functions for the db package without broadcasting them to the world.

*/

// ContainedNodes is used internally by the framework to return the contained nodes.
func ContainedNodes(n Node) (nodes []Node) {
	if nc, ok := n.(container); ok {
		return nc.containedNodes()
	} else {
		return nil
	}
}

func NodesMatch(node1, node2 Node) bool {
	if node1.NodeType_() != node2.NodeType_() {
		return false
	}

	if node1.TableName_() != node2.TableName_() {
		return false
	}
	if node1.DatabaseKey_() != node2.DatabaseKey_() {
		return false
	}

	switch node1.NodeType_() {
	case AliasNodeType:
		return node1.(AliasNodeI).Alias() == node2.(AliasNodeI).Alias()
	case ColumnNodeType:
		c1 := node1.(*ColumnNode)
		c2 := node2.(*ColumnNode)
		return c1.QueryName == c2.QueryName
	case TableNodeType:
		return true // already know table names are equal
	case ReferenceNodeType:
		return node1.(ReferenceNodeI).ColumnName() == node2.(ReferenceNodeI).ColumnName() &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	case ReverseNodeType:
		return node1.(ReverseNodeI).ColumnName() == node2.(ReverseNodeI).ColumnName() &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	case ManyManyNodeType:
		return node1.(ManyManyNodeI).AssnTableName() == node2.(ManyManyNodeI).AssnTableName() &&
			node1.(ManyManyNodeI).RefColumnName() == node2.(ManyManyNodeI).RefColumnName() &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	default:
		return false
	}
}
