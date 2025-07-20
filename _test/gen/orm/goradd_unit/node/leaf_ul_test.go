package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafUlTable(t *testing.T) {
	var n query.Node = LeafUl()

	assert.Equal(t, "leaf_ul", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_ul", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafUlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_ul", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafUlTable(t *testing.T) {
	{
		n := LeafUl().RootUl()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_ul", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafUl().RootUl().ID(), n2.(RootUlNode).ID()))
		assert.True(t, query.NodesMatch(LeafUl().RootUl().Name(), n2.(RootUlNode).Name()))
		assert.True(t, query.NodesMatch(LeafUl().RootUl().GroLock(), n2.(RootUlNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafUl().RootUl().LeafUl(), n2.(RootUlNode).LeafUl()))

	}

}

func TestSerializeReverseReferencesLeafUlTable(t *testing.T) {
}

func TestSerializeAssociationsLeafUlTable(t *testing.T) {
}
