package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafUTable(t *testing.T) {
	var n query.Node = LeafU()

	assert.Equal(t, "leaf_u", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_u", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafUTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_u", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafUTable(t *testing.T) {

	{
		n := LeafU().RootU()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_u", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafU().RootU().ID(), n2.(RootUNode).ID()))
		assert.True(t, query.NodesMatch(LeafU().RootU().Name(), n2.(RootUNode).Name()))
		assert.True(t, query.NodesMatch(LeafU().RootU().LeafU(), n2.(RootUNode).LeafU()))

	}

}

func TestSerializeReverseReferencesLeafUTable(t *testing.T) {
}

func TestSerializeAssociationsLeafUTable(t *testing.T) {
}
