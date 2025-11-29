package query

import (
	"bytes"
	"encoding/gob"
)

type ManyManyNodeI interface {
	AssnTableName() string
	RefColumnNames() (string, string)
	ParentColumnNames() (string, string)
	equal(n Node) bool
	TableNodeI
	linker
}

// A ManyManyNode is a mixin for an association node that links one table to another table with a many-to-many relationship.
type ManyManyNode struct {
	// The association table
	AssnTableQueryName string
	// The column in the association table pointing toward the parent node.
	ParentForeignKey string
	// The primary key column in the parent table that matches the ParentForeignKey.
	ParentPrimaryKey string

	// Identifier used to refer to the collection of objects. This is a plural name.
	Field string
	// Column in the association table pointing forwards to the embedding node
	RefForeignKey string
	// Primary key column in the child table that matches the RefForeignKey
	RefPrimaryKey string

	nodeLink
}

func (n *ManyManyNode) AssnTableName() string {
	return n.AssnTableQueryName
}

func (n *ManyManyNode) RefColumnNames() (string, string) {
	return n.RefForeignKey, n.RefPrimaryKey
}

func (n *ManyManyNode) ParentColumnNames() (string, string) {
	return n.ParentForeignKey, n.ParentPrimaryKey
}

func (n *ManyManyNode) equal(n2 Node) bool {
	if r, ok := n2.(ManyManyNodeI); ok {
		pfk, ppk := r.ParentColumnNames()
		rfk, rpk := r.RefColumnNames()
		return pfk == n.ParentForeignKey && ppk == n.ParentPrimaryKey &&
			rpk == n.RefPrimaryKey && rfk == n.RefForeignKey
	}
	return false
}

func (n *ManyManyNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.AssnTableQueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ParentForeignKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ParentPrimaryKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Field); err != nil {
		panic(err)
	}
	if err = e.Encode(n.RefForeignKey); err != nil {
		panic(err)
	}
	if err = e.Encode(n.RefPrimaryKey); err != nil {
		panic(err)
	}
	if err = e.Encode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *ManyManyNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.AssnTableQueryName); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.ParentForeignKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.ParentPrimaryKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.Field); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.RefForeignKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.RefPrimaryKey); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&ManyManyNode{})
}

func (n *ManyManyNode) queryKey() string {
	return n.Field
}
