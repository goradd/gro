package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootNlTable(t *testing.T) {
	var n query.Node = RootNl()

	assert.Equal(t, "root_nl", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_nl", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootNlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_nl", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootNlTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootNlTable(t *testing.T) {
	{
		n := RootNl().LeafNls()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_nl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootNl().LeafNls().ID(), n2.(LeafNlNode).ID()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().Name(), n2.(LeafNlNode).Name()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().GroLock(), n2.(LeafNlNode).GroLock()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().RootNlID(), n2.(LeafNlNode).RootNlID()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().RootNl(), n2.(LeafNlNode).RootNl()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().Leaf2s(), n2.(LeafNlNode).Leaf2s()))
		assert.True(t, query.NodesMatch(RootNl().LeafNls().Leaf1s(), n2.(LeafNlNode).Leaf1s()))

	}

}

func TestSerializeAssociationsRootNlTable(t *testing.T) {
}
