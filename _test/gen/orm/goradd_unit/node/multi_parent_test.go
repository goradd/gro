package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableMultiParentTable(t *testing.T) {
	var n query.Node = MultiParent()

	assert.Equal(t, "multi_parent", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "multi_parent", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := multiParentTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "multi_parent", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesMultiParentTable(t *testing.T) {
	{
		n := MultiParent().Parent1()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "multi_parent", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(MultiParent().Parent1().ID(), n2.(MultiParentNode).ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Name(), n2.(MultiParentNode).Name()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent1ID(), n2.(MultiParentNode).Parent1ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent1(), n2.(MultiParentNode).Parent1()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent2ID(), n2.(MultiParentNode).Parent2ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent2(), n2.(MultiParentNode).Parent2()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent1MultiParents(), n2.(MultiParentNode).Parent1MultiParents()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1().Parent2MultiParents(), n2.(MultiParentNode).Parent2MultiParents()))

	}

	{
		n := MultiParent().Parent2()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "multi_parent", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(MultiParent().Parent2().ID(), n2.(MultiParentNode).ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Name(), n2.(MultiParentNode).Name()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent1ID(), n2.(MultiParentNode).Parent1ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent1(), n2.(MultiParentNode).Parent1()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent2ID(), n2.(MultiParentNode).Parent2ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent2(), n2.(MultiParentNode).Parent2()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent1MultiParents(), n2.(MultiParentNode).Parent1MultiParents()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2().Parent2MultiParents(), n2.(MultiParentNode).Parent2MultiParents()))

	}

}

func TestSerializeReverseReferencesMultiParentTable(t *testing.T) {
	{
		n := MultiParent().Parent1MultiParents()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "multi_parent", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().ID(), n2.(MultiParentNode).ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Name(), n2.(MultiParentNode).Name()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent1ID(), n2.(MultiParentNode).Parent1ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent1(), n2.(MultiParentNode).Parent1()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent2ID(), n2.(MultiParentNode).Parent2ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent2(), n2.(MultiParentNode).Parent2()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent1MultiParents(), n2.(MultiParentNode).Parent1MultiParents()))
		assert.True(t, query.NodesMatch(MultiParent().Parent1MultiParents().Parent2MultiParents(), n2.(MultiParentNode).Parent2MultiParents()))

	}

	{
		n := MultiParent().Parent2MultiParents()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "multi_parent", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().ID(), n2.(MultiParentNode).ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Name(), n2.(MultiParentNode).Name()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent1ID(), n2.(MultiParentNode).Parent1ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent1(), n2.(MultiParentNode).Parent1()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent2ID(), n2.(MultiParentNode).Parent2ID()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent2(), n2.(MultiParentNode).Parent2()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent1MultiParents(), n2.(MultiParentNode).Parent1MultiParents()))
		assert.True(t, query.NodesMatch(MultiParent().Parent2MultiParents().Parent2MultiParents(), n2.(MultiParentNode).Parent2MultiParents()))

	}

}

func TestSerializeAssociationsMultiParentTable(t *testing.T) {
}
