package query

import (
	"bytes"
	"encoding/gob"
)

type ColumnNodeI interface {
	NodeI
	NodeSorter
	NodeLinker
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
	IsPrimaryKey bool
	nodeSort
	nodeLink
}

func (n *ColumnNode) NodeType_() NodeType {
	return ColumnNodeType
}

func (n *ColumnNode) TableName_() string {
	return n.Parent().TableName_()
}

func (n *ColumnNode) DatabaseKey_() string {
	return n.Parent().DatabaseKey_()
}

func NodeIsPK(n NodeI) bool {
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
	if err = e.Encode(n.nodeSort.sortDescending); err != nil {
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
	if err = dec.Decode(&n.nodeSort.sortDescending); err != nil {
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
