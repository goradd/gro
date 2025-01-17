package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeTableForwardCascadeTable(t *testing.T) {
	var n query.Node = ForwardCascade()

	assert.Equal(t, "forward_cascade", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "forward_cascade", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := forwardCascadeTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "forward_cascade", cn2.TableName_())
		require.Implements(t, (*query.Linker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.Linker).Parent().NodeType_())
	}
}

func TestSerializeReferencesForwardCascadeTable(t *testing.T) {

	{
		n := ForwardCascade().Reverse()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "forward_cascade", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReferenceNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

}

func TestSerializeReverseReferencesForwardCascadeTable(t *testing.T) {
}

func TestSerializeAssociationsForwardCascadeTable(t *testing.T) {
}
