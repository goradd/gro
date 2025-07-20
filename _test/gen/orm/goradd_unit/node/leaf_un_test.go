package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafUnTable(t *testing.T) {
	var n query.Node = LeafUn()

	assert.Equal(t, "leaf_un", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_un", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafUnTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_un", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafUnTable(t *testing.T) {
	{
		n := LeafUn().RootUn()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_un", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafUn().RootUn().ID(), n2.(RootUnNode).ID()))
		assert.True(t, query.NodesMatch(LeafUn().RootUn().Name(), n2.(RootUnNode).Name()))
		assert.True(t, query.NodesMatch(LeafUn().RootUn().LeafUn(), n2.(RootUnNode).LeafUn()))

	}

}

func TestSerializeReverseReferencesLeafUnTable(t *testing.T) {
}

func TestSerializeAssociationsLeafUnTable(t *testing.T) {
}
