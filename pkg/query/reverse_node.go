package query

import (
	"bytes"
	"encoding/gob"
)

// ReverseNodeI is the interface to objects that have embedded ReverseNode objects.
type ReverseNodeI interface {
	ColumnName() string
	IsArray() bool
	TableNodeI
	linker
}

// ReverseNode is a mixin for a reverse reference representing a one-to-many relationship or one-to-one
// relationship, depending on whether the foreign key is unique. The other side of the relationship will have
// a matching forward ReferenceNode.
type ReverseNode struct {
	// The query name of the column that is the foreign key pointing to the parent's primary key.
	ColumnQueryName string
	// The identifier that will be used to identify this object in source code.
	// Equals the key for the Get() function on an object. Should be plural.
	Identifier string
	// The type of item acting as a pointer. This should be the same on both sides of the reference.
	ReceiverType ReceiverType
	// IsUnique is true if there is a unique relationship between this node and its parent,
	// which would create a one-to-one relationship rather than a one-to-many relationship.
	IsUnique bool
	nodeLink
}

func (n *ReverseNode) ColumnName() string {
	return n.ColumnQueryName
}

// IsArray returns true if this node creates a one-to-many relationship with its parent.
// Otherwise, it is a one-to-one relationship.
func (n *ReverseNode) IsArray() bool {
	return !n.IsUnique
}

func (n *ReverseNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(n.ColumnQueryName); err != nil {
		panic(err)
	}
	if err = e.Encode(n.Identifier); err != nil {
		panic(err)
	}
	if err = e.Encode(n.ReceiverType); err != nil {
		panic(err)
	}
	if err = e.Encode(n.IsUnique); err != nil {
		panic(err)
	}
	if err = e.Encode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *ReverseNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.ColumnQueryName); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.Identifier); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.ReceiverType); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.IsUnique); err != nil {
		panic(err)
	}
	if err = dec.Decode(&n.nodeLink.parentNode); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&ReverseNode{})
}

func (n *ReverseNode) id() string {
	return n.Identifier
}
