package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafNTable(t *testing.T) {
	var n query.Node = LeafN()

	assert.Equal(t, "leaf_n", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_n", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafNTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_n", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafNTable(t *testing.T) {
	{
		n := LeafN().RootN()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_n", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafN().RootN().ID(), n2.(RootNNode).ID()))
		assert.True(t, query.NodesMatch(LeafN().RootN().Name(), n2.(RootNNode).Name()))
		assert.True(t, query.NodesMatch(LeafN().RootN().LeafNs(), n2.(RootNNode).LeafNs()))

	}

}

func TestSerializeReverseReferencesLeafNTable(t *testing.T) {
}

func TestSerializeAssociationsLeafNTable(t *testing.T) {
}
