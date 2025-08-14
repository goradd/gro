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

// Node is a representation of an object or a relationship between objects in a database.
//
// The ORM will create Node types during code generation. The develpoper chains these Nodes to
// indicate a specific table or column in a query. A developer typically would not create a Node directly.
//
// For example, a developer might use node.Project().TeamMembers().FirstName() to get the node pointing
// to the first name of a team member in a project.
type Node interface {
	// NodeType_ returns the type of the node
	NodeType_() NodeType
	// TableName_ returns the query name of the table the node is associated with.
	// Nodes that are not table specific will return an empty string.
	TableName_() string
	// DatabaseKey_ is the key of the database the node is associated with.
	DatabaseKey_() string
}

/**

Public Accessors

The following functions are designed primarily to be used by the db package to help it unpack queries. They are not
given an accessor at the beginning so that they do not show up as a function in editors that provide code hinting when
trying to put together a node chain during the code creation process. Essentially they are trying to create exported
functions for the db package without broadcasting them to the world.

*/

// ContainedNodes is used internally by the ORM to return the contained nodes.
func ContainedNodes(n Node) (nodes []Node) {
	if nc, ok := n.(container); ok {
		return nc.containedNodes()
	} else {
		return nil
	}
}

// NodesMatch is used internally by the ORM to indicate whether two nodes point to the same database item.
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
		return ColumnNodeQueryName(node1) == ColumnNodeQueryName(node2)
	case TableNodeType:
		return true // already know table names are equal
	case ReferenceNodeType:
		return node1.(ReferenceNodeI).equal(node2.(ReferenceNodeI)) &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	case ReverseNodeType:
		return node1.(ReverseNodeI).equal(node2.(ReverseNodeI)) &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	case ManyManyNodeType:
		return node1.(ManyManyNodeI).equal(node2.(ManyManyNodeI)) &&
			NodesMatch(NodeParent(node1), NodeParent(node2))
	default:
		return false
	}
}

// NodeIsArray returns true if the node is an array connection, like a ManyMany relationship or one-to-many Reverse node.
func NodeIsArray(n Node) bool {
	t := n.NodeType_()
	if t == ManyManyNodeType {
		return true
	} else if t == ReverseNodeType {
		return n.(ReverseNodeI).IsArray()
	}
	return false
}

type queryKeyer interface {
	queryKey() string
}

// NodeQueryKey is used by the ORM as the key used in query result sets to refer to the data corresponding to the node.
func NodeQueryKey(n Node) string {
	if id, ok := n.(queryKeyer); ok {
		return id.queryKey()
	}
	return ""
}
