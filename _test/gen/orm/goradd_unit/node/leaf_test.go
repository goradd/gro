package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafTable(t *testing.T) {
	var n query.Node = Leaf()

	assert.Equal(t, "leaf", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafTable(t *testing.T) {
	{
		n := Leaf().Root()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Leaf().Root().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Leaf().Root().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Leaf().Root().Leafs(), n2.(RootNode).Leafs()))

	}

}

func TestSerializeReverseReferencesLeafTable(t *testing.T) {
}

func TestSerializeAssociationsLeafTable(t *testing.T) {
}
