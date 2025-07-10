package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableAltRootUnTable(t *testing.T) {
	var n query.Node = AltRootUn()

	assert.Equal(t, "alt_root_un", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "alt_root_un", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := altRootUnTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "alt_root_un", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesAltRootUnTable(t *testing.T) {
}

func TestSerializeReverseReferencesAltRootUnTable(t *testing.T) {
	{
		n := AltRootUn().AltRootUnAltLeafUns()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "alt_root_un", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(AltRootUn().AltRootUnAltLeafUns().ID(), n2.(AltLeafUnNode).ID()))
		assert.True(t, query.NodesMatch(AltRootUn().AltRootUnAltLeafUns().Name(), n2.(AltLeafUnNode).Name()))
		assert.True(t, query.NodesMatch(AltRootUn().AltRootUnAltLeafUns().AltRootUnID(), n2.(AltLeafUnNode).AltRootUnID()))
		assert.True(t, query.NodesMatch(AltRootUn().AltRootUnAltLeafUns().AltRootUn(), n2.(AltLeafUnNode).AltRootUn()))

	}

}

func TestSerializeAssociationsAltRootUnTable(t *testing.T) {
}
