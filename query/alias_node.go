package query

import (
	"bytes"
	"encoding/gob"
)

type AliasNodeI interface {
	Node
	Alias() string
}

// An AliasNode is a reference to a prior aliased operation later in a query. An alias is a name given
// to a calculated value.
type AliasNode struct {
	alias string
}

// Alias returns an alias node for the given labeled alias that was
// previously defined in a query.
func Alias(alias string) AliasNodeI {
	return &AliasNode{alias: alias}
}

func (n *AliasNode) NodeType_() NodeType {
	return AliasNodeType
}

func (n *AliasNode) TableName_() string {
	return ""
}

func (n *AliasNode) DatabaseKey_() string {
	return ""
}

func (n *AliasNode) Alias() string {
	return n.alias
}

func (n *AliasNode) GobEncode() (data []byte, err error) {
	var buf bytes.Buffer
	e := gob.NewEncoder(&buf)

	if err = e.Encode(&n.alias); err != nil {
		panic(err)
	}
	data = buf.Bytes()
	return
}

func (n *AliasNode) GobDecode(data []byte) (err error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	if err = dec.Decode(&n.alias); err != nil {
		panic(err)
	}
	return
}

func init() {
	gob.Register(&AliasNode{})
}

// id is a unique identifier within the parent namespace and satisfies the queryKeyer interface
func (n *AliasNode) queryKey() string {
	return n.alias
}
