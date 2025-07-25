package query

import (
	"bytes"
	"encoding/gob"
)

type ReferenceNodeI interface {
	ColumnNames() (string, string)
	equal(n Node) bool
	TableNodeI
	linker
}

// A ReferenceNode is a mixin for a forward-pointing foreign key relationship.
type ReferenceNode struct {
	// The query name of the column that is the foreign key
	ForeignKey string
	// The name of the matching primary key column in the referenced table
	PrimaryKey string
	// The field that can be used in Get() calls to get the corresponding value from the table.
	Field string
	nodeLink
}

// ColumnNames returns the foreign key column name in this table, and the name of the primary
// key column that it mirrors in the referenced table.
func (n *ReferenceNode) ColumnNames() (string, string) {
	return n.ForeignKey, n.PrimaryKey
}

func (n *ReferenceNode) equal(n2 Node) bool {
	if r, ok := n2.(ReferenceNodeI); ok {
		c1, c2 := r.ColumnNames()
		return c1 == n.ForeignKey && c2 == n.PrimaryKey
	}
	return false
}

// GobEncode encodes the reference in a binary form.
func (n *ReferenceNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.ForeignKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.PrimaryKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Field); err != nil {
		panic(err)
	}
	if err = e.Encode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *ReferenceNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.ForeignKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.PrimaryKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.Field); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&ReferenceNode{})
}

func (n *ReferenceNode) id() string {
	return n.Field
}
