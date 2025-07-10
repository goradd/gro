package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableAltLeafUnTable(t *testing.T) {
	var n query.Node = AltLeafUn()

	assert.Equal(t, "alt_leaf_un", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "alt_leaf_un", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := altLeafUnTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "alt_leaf_un", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesAltLeafUnTable(t *testing.T) {
	{
		n := AltLeafUn().AltRootUn()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "alt_leaf_un", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(AltLeafUn().AltRootUn().ID(), n2.(AltRootUnNode).ID()))
		assert.True(t, query.NodesMatch(AltLeafUn().AltRootUn().Name(), n2.(AltRootUnNode).Name()))
		assert.True(t, query.NodesMatch(AltLeafUn().AltRootUn().AltRootUnAltLeafUns(), n2.(AltRootUnNode).AltRootUnAltLeafUns()))

	}

}

func TestSerializeReverseReferencesAltLeafUnTable(t *testing.T) {
}

func TestSerializeAssociationsAltLeafUnTable(t *testing.T) {
}
