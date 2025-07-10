package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootLTable(t *testing.T) {
	var n query.Node = RootL()

	assert.Equal(t, "root_l", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_l", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootLTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_l", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootLTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootLTable(t *testing.T) {
	{
		n := RootL().RootLLeafLs()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_l", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootL().RootLLeafLs().ID(), n2.(LeafLNode).ID()))
		assert.True(t, query.NodesMatch(RootL().RootLLeafLs().Name(), n2.(LeafLNode).Name()))
		assert.True(t, query.NodesMatch(RootL().RootLLeafLs().GroLock(), n2.(LeafLNode).GroLock()))
		assert.True(t, query.NodesMatch(RootL().RootLLeafLs().RootLID(), n2.(LeafLNode).RootLID()))
		assert.True(t, query.NodesMatch(RootL().RootLLeafLs().RootL(), n2.(LeafLNode).RootL()))

	}

}

func TestSerializeAssociationsRootLTable(t *testing.T) {
}
