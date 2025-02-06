package query

import (
	"bytes"
	"encoding/gob"
)

type ManyManyNodeI interface {
	AssnTableName() string
	RefColumnName() string
	ParentColumnName() string
	TableNodeI
	linker
}

// A ManyManyNode is a mixin for an association node that links one table to another table with a many-to-many relationship.
type ManyManyNode struct {
	// The association table
	AssnTableQueryName string
	// The column in the association table pointing toward the parent node.
	ParentColumnQueryName string
	// The parent column's type.
	ParentColumnReceiverType ReceiverType

	// Identifier used to refer to the collection of objects. This is a plural name.
	Identifier string
	// Column in the association table pointing forwards to the embedding node
	RefColumnQueryName string
	// The ref column's type
	RefColumnReceiverType ReceiverType

	nodeLink
}

func (n *ManyManyNode) AssnTableName() string {
	return n.AssnTableQueryName
}

func (n *ManyManyNode) RefColumnName() string {
	return n.RefColumnQueryName
}

func (n *ManyManyNode) ParentColumnName() string {
	return n.ParentColumnQueryName
}

func (n *ManyManyNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.AssnTableQueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ParentColumnQueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ParentColumnReceiverType); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Identifier); err != nil {
		panic(err)
	}
	if err = e.Encode(n.RefColumnQueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.RefColumnReceiverType); err != nil {
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
	if err = dec.Decode(&n.ParentColumnQueryName); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.ParentColumnReceiverType); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.Identifier); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.RefColumnQueryName); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.RefColumnReceiverType); err != nil {
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

func (n *ManyManyNode) id() string {
	return n.Identifier
}
