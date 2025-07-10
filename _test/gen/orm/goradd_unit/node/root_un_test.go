package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootUnTable(t *testing.T) {
	var n query.Node = RootUn()

	assert.Equal(t, "root_un", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_un", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootUnTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_un", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootUnTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootUnTable(t *testing.T) {
	{
		n := RootUn().RootUnLeafUns()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_un", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootUn().RootUnLeafUns().ID(), n2.(LeafUnNode).ID()))
		assert.True(t, query.NodesMatch(RootUn().RootUnLeafUns().Name(), n2.(LeafUnNode).Name()))
		assert.True(t, query.NodesMatch(RootUn().RootUnLeafUns().RootUnID(), n2.(LeafUnNode).RootUnID()))
		assert.True(t, query.NodesMatch(RootUn().RootUnLeafUns().RootUn(), n2.(LeafUnNode).RootUn()))

	}

}

func TestSerializeAssociationsRootUnTable(t *testing.T) {
}
