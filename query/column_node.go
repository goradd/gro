package query

import (
	"bytes"
	"encoding/gob"

	"github.com/goradd/gro/schema"
)

type ColumnNodeI interface {
	Node
	Sorter
	linker
}

// ColumnNode represents a table or field in a database structure, and is the leaf of a node tree or chain.
//
// You would not normally create a column node directly. Use the code generated node functions to create column nodes.
type ColumnNode struct {
	// The query name of the column in the database.
	QueryName string
	// The identifier of the data in the corresponding object.
	// Pass this to Get() to retrieve a value.
	Field string
	// The receiver type for the column
	ReceiverType ReceiverType
	// The schema type for the node
	SchemaType schema.ColumnType
	// The schema subtype for the node
	SchemaSubType schema.ColumnSubType
	// True if this is the single primary key of its parent table
	IsPrimaryKey   bool
	sortDescending bool
	nodeLink
}

// NewColumnNode is used by the code generated framework to create a new column node that refers to a
// specific column in a table. You would not normally call this function directly.
func NewColumnNode(
	queryName string,
	field string,
	receiverType ReceiverType,
	schemaType schema.ColumnType,
	schemaSubType schema.ColumnSubType,
	isPrimaryKey bool,
	parent Node,
) *ColumnNode {
	n := &ColumnNode{
		QueryName:     queryName,
		Field:         field,
		ReceiverType:  receiverType,
		SchemaType:    schemaType,
		SchemaSubType: schemaSubType,
		IsPrimaryKey:  isPrimaryKey,
	}
	n.parentNode = parent
	return n
}

// NodeType_ is used by the framework to return the type of node this is.
func (n *ColumnNode) NodeType_() NodeType {
	return ColumnNodeType
}

func (n *ColumnNode) TableName_() string {
	return n.parent_().TableName_()
}

func (n *ColumnNode) DatabaseKey_() string {
	return n.parent_().DatabaseKey_()
}

// Ascending sets the column to sort ascending when used in an OrderBy statement.
func (n *ColumnNode) Ascending() Sorter {
	n.sortDescending = false
	return n
}

// Descending sets the column to sort descending when used in an OrderBy statement.
func (n *ColumnNode) Descending() Sorter {
	n.sortDescending = true
	return n
}

// IsDescending returns true if the node is sorted in descending order.
func (n *ColumnNode) IsDescending() bool {
	return n.sortDescending
}

func (n *ColumnNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.QueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Field); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ReceiverType); err != nil {
		panic(err)
	}
	if err = e.Encode(n.IsPrimaryKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.sortDescending); err != nil {
		panic(err)
	}
	if err = e.Encode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *ColumnNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.QueryName); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.Field); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.ReceiverType); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.IsPrimaryKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.sortDescending); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&ColumnNode{})
}

func (n *ColumnNode) queryKey() string {
	return n.Field
}

// ColumnNodeQueryName returns the name used in the database of the column that corresponds to the node.
func ColumnNodeQueryName(n Node) string {
	return n.(*ColumnNode).QueryName
}
