package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootUTable(t *testing.T) {
	var n query.Node = RootU()

	assert.Equal(t, "root_u", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_u", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootUTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_u", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootUTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootUTable(t *testing.T) {
	{
		n := RootU().LeafU()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_u", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootU().LeafU().ID(), n2.(LeafUNode).ID()))
		assert.True(t, query.NodesMatch(RootU().LeafU().Name(), n2.(LeafUNode).Name()))
		assert.True(t, query.NodesMatch(RootU().LeafU().RootUID(), n2.(LeafUNode).RootUID()))
		assert.True(t, query.NodesMatch(RootU().LeafU().RootU(), n2.(LeafUNode).RootU()))

	}

}

func TestSerializeAssociationsRootUTable(t *testing.T) {
}
