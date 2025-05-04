package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafNlTable(t *testing.T) {
	var n query.Node = LeafNl()

	assert.Equal(t, "leaf_nl", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf_nl", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafNlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf_nl", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafNlTable(t *testing.T) {

	{
		n := LeafNl().RootNl()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_nl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafNl().RootNl().ID(), n2.(RootNlNode).ID()))
		assert.True(t, query.NodesMatch(LeafNl().RootNl().Name(), n2.(RootNlNode).Name()))
		assert.True(t, query.NodesMatch(LeafNl().RootNl().GroLock(), n2.(RootNlNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafNl().RootNl().LeafNls(), n2.(RootNlNode).LeafNls()))

	}

}

func TestSerializeReverseReferencesLeafNlTable(t *testing.T) {
}

func TestSerializeAssociationsLeafNlTable(t *testing.T) {

	{
		n := LeafNl().Leaf2s()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_nl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().ID(), n2.(LeafNlNode).ID()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().Name(), n2.(LeafNlNode).Name()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().RootNlID(), n2.(LeafNlNode).RootNlID()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().RootNl(), n2.(LeafNlNode).RootNl()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().GroLock(), n2.(LeafNlNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().Leaf2s(), n2.(LeafNlNode).Leaf2s()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf2s().Leaf1s(), n2.(LeafNlNode).Leaf1s()))

	}

	{
		n := LeafNl().Leaf1s()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf_nl", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().ID(), n2.(LeafNlNode).ID()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().Name(), n2.(LeafNlNode).Name()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().RootNlID(), n2.(LeafNlNode).RootNlID()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().RootNl(), n2.(LeafNlNode).RootNl()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().GroLock(), n2.(LeafNlNode).GroLock()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().Leaf2s(), n2.(LeafNlNode).Leaf2s()))
		assert.True(t, query.NodesMatch(LeafNl().Leaf1s().Leaf1s(), n2.(LeafNlNode).Leaf1s()))

	}

}
