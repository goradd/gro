package query

import (
	"bytes"
	"encoding/gob"
)

type ColumnNodeI interface {
	Node
	Sorter
	linker
}

// ColumnNode represents a table or field in a database structure, and is the leaf of a node tree or chain.
type ColumnNode struct {
	// The query name of the column in the database.
	QueryName string
	// The name of the column's data as used in source code.
	Identifier string
	// The receiver type for the column
	ReceiverType ReceiverType
	// True if this is the primary key of its parent table
	IsPrimaryKey   bool
	sortDescending bool
	nodeLink
}

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

func NodeIsPK(n Node) bool {
	if cn, ok := n.(*ColumnNode); !ok {
		return false
	} else {
		return cn.IsPrimaryKey
	}
}

func (n *ColumnNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.QueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Identifier); err != nil {
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
	if err = dec.Decode(&n.Identifier); err != nil {
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

func (n *ColumnNode) id() string {
	return n.Identifier
}

type ider interface {
	id() string
}

// NodeIdentifier returns the Go identifier related to the node.
func NodeIdentifier(n Node) string {
	if id, ok := n.(ider); ok {
		return id.id()
	}
	return ""
}
