package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootTable(t *testing.T) {
	var n query.Node = Root()

	assert.Equal(t, "root", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootTable(t *testing.T) {
	{
		n := Root().Leafs()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().Leafs().ID(), n2.(LeafNode).ID()))
		assert.True(t, query.NodesMatch(Root().Leafs().Name(), n2.(LeafNode).Name()))
		assert.True(t, query.NodesMatch(Root().Leafs().RootID(), n2.(LeafNode).RootID()))
		assert.True(t, query.NodesMatch(Root().Leafs().Root(), n2.(LeafNode).Root()))

	}

}

func TestSerializeAssociationsRootTable(t *testing.T) {
}
