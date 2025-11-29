package node

import (
	"testing"

	"github.com/goradd/gro/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootUlTable(t *testing.T) {
	var n query.Node = RootUl()

	assert.Equal(t, "root_ul", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_ul", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootUlTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_ul", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootUlTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootUlTable(t *testing.T) {
	{
		n := RootUl().LeafUl()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_ul", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootUl().LeafUl().ID(), n2.(LeafUlNode).ID()))
		assert.True(t, query.NodesMatch(RootUl().LeafUl().Name(), n2.(LeafUlNode).Name()))
		assert.True(t, query.NodesMatch(RootUl().LeafUl().GroLock(), n2.(LeafUlNode).GroLock()))
		assert.True(t, query.NodesMatch(RootUl().LeafUl().RootUlID(), n2.(LeafUlNode).RootUlID()))
		assert.True(t, query.NodesMatch(RootUl().LeafUl().RootUl(), n2.(LeafUlNode).RootUl()))

	}

}

func TestSerializeAssociationsRootUlTable(t *testing.T) {
}
