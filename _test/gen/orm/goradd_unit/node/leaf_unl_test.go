package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafUnlTable(t *testing.T) {
	var n query.Node = LeafUnl()

	assert.Equal(t, "leaf_unl", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_unl", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafUnlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_unl", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafUnlTable(t *testing.T) {
	{
		n := LeafUnl().RootUnl()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_unl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafUnl().RootUnl().ID(), n2.(RootUnlNode).ID()))
		assert.True(t, query.NodesMatch(LeafUnl().RootUnl().Name(), n2.(RootUnlNode).Name()))
		assert.True(t, query.NodesMatch(LeafUnl().RootUnl().GroLock(), n2.(RootUnlNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafUnl().RootUnl().LeafUnl(), n2.(RootUnlNode).LeafUnl()))

	}

}

func TestSerializeReverseReferencesLeafUnlTable(t *testing.T) {
}

func TestSerializeAssociationsLeafUnlTable(t *testing.T) {
}
