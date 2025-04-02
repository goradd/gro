package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
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

	{
		n := Root().OptionalLeaf()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().OptionalLeaf().ID(), n2.(LeafNode).ID()))
		assert.True(t, query.NodesMatch(Root().OptionalLeaf().Name(), n2.(LeafNode).Name()))
		assert.True(t, query.NodesMatch(Root().OptionalLeaf().OptionalLeafRoots(), n2.(LeafNode).OptionalLeafRoots()))
		assert.True(t, query.NodesMatch(Root().OptionalLeaf().RequiredLeafRoots(), n2.(LeafNode).RequiredLeafRoots()))
		assert.True(t, query.NodesMatch(Root().OptionalLeaf().OptionalLeafUniqueRoot(), n2.(LeafNode).OptionalLeafUniqueRoot()))
		assert.True(t, query.NodesMatch(Root().OptionalLeaf().RequiredLeafUniqueRoot(), n2.(LeafNode).RequiredLeafUniqueRoot()))

	}

	{
		n := Root().RequiredLeaf()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().RequiredLeaf().ID(), n2.(LeafNode).ID()))
		assert.True(t, query.NodesMatch(Root().RequiredLeaf().Name(), n2.(LeafNode).Name()))
		assert.True(t, query.NodesMatch(Root().RequiredLeaf().OptionalLeafRoots(), n2.(LeafNode).OptionalLeafRoots()))
		assert.True(t, query.NodesMatch(Root().RequiredLeaf().RequiredLeafRoots(), n2.(LeafNode).RequiredLeafRoots()))
		assert.True(t, query.NodesMatch(Root().RequiredLeaf().OptionalLeafUniqueRoot(), n2.(LeafNode).OptionalLeafUniqueRoot()))
		assert.True(t, query.NodesMatch(Root().RequiredLeaf().RequiredLeafUniqueRoot(), n2.(LeafNode).RequiredLeafUniqueRoot()))

	}

	{
		n := Root().OptionalLeafUnique()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().ID(), n2.(LeafNode).ID()))
		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().Name(), n2.(LeafNode).Name()))
		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().OptionalLeafRoots(), n2.(LeafNode).OptionalLeafRoots()))
		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().RequiredLeafRoots(), n2.(LeafNode).RequiredLeafRoots()))
		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().OptionalLeafUniqueRoot(), n2.(LeafNode).OptionalLeafUniqueRoot()))
		assert.True(t, query.NodesMatch(Root().OptionalLeafUnique().RequiredLeafUniqueRoot(), n2.(LeafNode).RequiredLeafUniqueRoot()))

	}

	{
		n := Root().RequiredLeafUnique()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().ID(), n2.(LeafNode).ID()))
		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().Name(), n2.(LeafNode).Name()))
		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().OptionalLeafRoots(), n2.(LeafNode).OptionalLeafRoots()))
		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().RequiredLeafRoots(), n2.(LeafNode).RequiredLeafRoots()))
		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().OptionalLeafUniqueRoot(), n2.(LeafNode).OptionalLeafUniqueRoot()))
		assert.True(t, query.NodesMatch(Root().RequiredLeafUnique().RequiredLeafUniqueRoot(), n2.(LeafNode).RequiredLeafUniqueRoot()))

	}

	{
		n := Root().Parent()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Root().Parent().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Root().Parent().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Root().Parent().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Root().Parent().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Root().Parent().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Root().Parent().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Root().Parent().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Root().Parent().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Root().Parent().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Root().Parent().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Root().Parent().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Root().Parent().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Root().Parent().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

}

func TestSerializeReverseReferencesRootTable(t *testing.T) {

	{
		n := Root().ParentRoots()
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

		assert.True(t, query.NodesMatch(Root().ParentRoots().ID(), n2.(RootNode).ID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().Name(), n2.(RootNode).Name()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().OptionalLeafID(), n2.(RootNode).OptionalLeafID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().OptionalLeaf(), n2.(RootNode).OptionalLeaf()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().RequiredLeafID(), n2.(RootNode).RequiredLeafID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().RequiredLeaf(), n2.(RootNode).RequiredLeaf()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().OptionalLeafUniqueID(), n2.(RootNode).OptionalLeafUniqueID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().OptionalLeafUnique(), n2.(RootNode).OptionalLeafUnique()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().RequiredLeafUniqueID(), n2.(RootNode).RequiredLeafUniqueID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().RequiredLeafUnique(), n2.(RootNode).RequiredLeafUnique()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().ParentID(), n2.(RootNode).ParentID()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().Parent(), n2.(RootNode).Parent()))
		assert.True(t, query.NodesMatch(Root().ParentRoots().ParentRoots(), n2.(RootNode).ParentRoots()))

	}

}

func TestSerializeAssociationsRootTable(t *testing.T) {
}
