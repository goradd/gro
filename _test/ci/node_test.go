package ci

import (
	"bytes"
	"encoding/gob"
	"github.com/goradd/orm/_test/gen/orm/goradd"
	"github.com/goradd/orm/_test/gen/orm/goradd/node"
	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNodeEquality(t *testing.T) {

	n := node.Person()
	if !query.NodeIsEqual(n, n) {
		t.Error("Table node not equal to self")
	}

	n = node.Project().Manager()
	if !query.NodeIsEqual(n, n) {
		t.Error("Reference node not equal to self")
	}

	n2 := node.Person().ManagerProjects()
	if !query.NodeIsEqual(n2, n2) {
		t.Error("Reverse Reference node not equal to self")
	}

	n3 := node.Person().Projects()
	if !query.NodeIsEqual(n3, n3) {
		t.Error("Many-Many node not equal to self")
	}

	n4 := query.NewValueNode(goradd.PersonTypeContractor)
	if !query.NodeIsEqual(n4, n4) {
		t.Error("ReceiverType node not equal to self")
	}

}

/*
func BenchmarkNodeType1(b *testing.B) {
	n := node.Project().Manager()

	for i := 0; i < b.N; i++ {
		t := query.NodeGetType(n)
		if t == query.ReferenceNodeType {
			_ = n
		}
	}
}

func BenchmarkNodeType2(b *testing.B) {
	n := node.Project().Manager().(query.ReferenceNodeI)

	for i := 0; i < b.N; i++ {
		if r, ok := n.EmbeddedNode_().(*query.ReferenceNode); ok {
			_ = r
		}
	}
}*/

func TestNodeSerialize(t *testing.T) {
	var n query.Node = node.Person().FirstName()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&n)
	assert.NoError(t, err)

	var n2 query.Node
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&n2)
	assert.NoError(t, err)

	assert.True(t, query.NodeIsEqual(n2, n))
}
