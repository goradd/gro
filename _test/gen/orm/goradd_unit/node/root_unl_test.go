package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootUnlTable(t *testing.T) {
	var n query.Node = RootUnl()

	assert.Equal(t, "root_unl", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_unl", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootUnlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_unl", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootUnlTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootUnlTable(t *testing.T) {
	{
		n := RootUnl().LeafUnl()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_unl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootUnl().LeafUnl().ID(), n2.(LeafUnlNode).ID()))
		assert.True(t, query.NodesMatch(RootUnl().LeafUnl().Name(), n2.(LeafUnlNode).Name()))
		assert.True(t, query.NodesMatch(RootUnl().LeafUnl().GroLock(), n2.(LeafUnlNode).GroLock()))
		assert.True(t, query.NodesMatch(RootUnl().LeafUnl().RootUnlID(), n2.(LeafUnlNode).RootUnlID()))
		assert.True(t, query.NodesMatch(RootUnl().LeafUnl().RootUnl(), n2.(LeafUnlNode).RootUnl()))

	}

}

func TestSerializeAssociationsRootUnlTable(t *testing.T) {
}
