package query

import (
	"bytes"
	"encoding/gob"
)

type AliasNodeI interface {
	Node
	Aliaser
}

// An AliasNode is a reference to a prior aliased operation later in a query. An alias is a name given
// to a computed value.
type AliasNode struct {
	nodeAlias
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

// Alias returns an AliasNode type, which allows you to refer to a prior created named alias operation.
// TODO: Add this to all node structures instead
func Alias(name string) *AliasNode {
	return &AliasNode{
		nodeAlias{
			alias: name,
		},
	}
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

// id is a unique identifier within the parent namespace and satisfies the ider interface
func (n *AliasNode) id() string {
	return n.alias
}
