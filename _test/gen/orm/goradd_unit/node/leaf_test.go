package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLeafTable(t *testing.T) {
	var n query.Node = Leaf()

	assert.Equal(t, "leaf", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "leaf", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := leafTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "leaf", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLeafTable(t *testing.T) {
}

func TestSerializeReverseReferencesLeafTable(t *testing.T) {

	{
		n := Leaf().OptionalLeafRoots()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafRoots().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

	{
		n := Leaf().RequiredLeafRoots()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafRoots().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

	{
		n := Leaf().OptionalLeafUniqueRoot()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Leaf().OptionalLeafUniqueRoot().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

	{
		n := Leaf().RequiredLeafUniqueRoot()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "leaf", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Leaf().RequiredLeafUniqueRoot().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

}

func TestSerializeAssociationsLeafTable(t *testing.T) {
}
