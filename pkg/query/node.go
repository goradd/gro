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
	EnumNodeType
	ManyEnumNodeType
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
	case EnumNodeType:
		return "EnumNodeType"
	case ManyEnumNodeType:
		return "ManyEnumNodeType"
	default:
		return "Unknown"
	}
}

type container interface {
	containedNodes() (nodes []Node)
}

// ider returns a unique value within the parent node's namespace
type ider interface {
	id() string
}

// NodeId returns a unique value within the parent namespace, if the node supports it.
// Top level table nodes do not support this, but that should be fine since there is only one in a node chain.
func NodeId(n Node) string {
	if i, ok := n.(ider); ok {
		return i.id()
	}
	return ""
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
