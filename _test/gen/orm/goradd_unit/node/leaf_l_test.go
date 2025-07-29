package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafLTable(t *testing.T) {
	var n query.Node = LeafL()

	assert.Equal(t, "leaf_l", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_l", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafLTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_l", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafLTable(t *testing.T) {
	{
		n := LeafL().RootL()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_l", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafL().RootL().ID(), n2.(RootLNode).ID()))
		assert.True(t, query.NodesMatch(LeafL().RootL().Name(), n2.(RootLNode).Name()))
		assert.True(t, query.NodesMatch(LeafL().RootL().GroLock(), n2.(RootLNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafL().RootL().LeafLs(), n2.(RootLNode).LeafLs()))

	}

}

func TestSerializeReverseReferencesLeafLTable(t *testing.T) {
}

func TestSerializeAssociationsLeafLTable(t *testing.T) {
}
