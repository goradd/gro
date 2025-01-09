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

type nodeContainer interface {
	containedNodes() (nodes []NodeI)
}

type Aliaser interface {
	// SetAlias sets a unique name for the node as used in a database query.
	SetAlias(string)
	// GetAlias returns the alias that was used in a database query.
	GetAlias() string
}

// Nodes that can have an alias can mix this in
type nodeAlias struct {
	alias string
}

// SetAlias sets an alias which is an alternate name to use for the node in the result of a query.
// Aliases will generally be assigned during the query build process. You only need to assign a manual
// alias if
func (n *nodeAlias) SetAlias(a string) {
	n.alias = a

}

// GetAlias returns the alias name for the node.
func (n *nodeAlias) GetAlias() string {
	return n.alias
}

// NodeI is the interface that all nodes must satisfy. A node is a representation of an object or a relationship
// between objects in a database that we use to create a query. It lets us abstract the structure of a database
// to be able to query any kind of database. Obviously, this doesn't work for all possible database structures, but
// it generally works well enough to solve most of the situations that you will come across.
type NodeI interface {
	// NodeType_ returns the type of the node
	NodeType_() NodeType
	// TableName_ returns the query name of the table the node is associated with. Not all nodes support this.
	TableName_() string
	// DatabaseKey_ is the database key of the database the node is associated with.
	DatabaseKey_() string
	//Equals(NodeI) bool
	//equals(NodeI) bool
	//log(level int)
}

/**

Public Accessors

The following functions are designed primarily to be used by the db package to help it unpack queries. They are not
given an accessor at the beginning so that they do not show up as a function in editors that provide code hinting when
trying to put together a node chain during the code creation process. Essentially they are trying to create exported
functions for the db package without broadcasting them to the world.

*/

// NodeTableName is used internally by the framework to return the table associated with a node.
func NodeTableName(n NodeI) string {
	return n.TableName_()
}

func NodeDbKey(n NodeI) string {
	return n.DatabaseKey_()
}

// ContainedNodes is used internally by the framework to return the contained nodes.
func ContainedNodes(n NodeI) (nodes []NodeI) {
	if nc, ok := n.(nodeContainer); ok {
		return nc.containedNodes()
	} else {
		return nil
	}
}

// NodePrimaryKey returns the primary key of a node, if it has a primary key. Otherwise, returns nil.
func NodePrimaryKey(n NodeI) NodeI {
	if tn, ok := n.(PrimaryKeyer); ok {
		return tn.PrimaryKeyNode()
	}
	return nil
}

func NodeIsEqual(n NodeI, n2 NodeI) bool {
	return n == n2
}
