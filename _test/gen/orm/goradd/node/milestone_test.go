package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeTableMilestoneTable(t *testing.T) {
	var n query.Node = Milestone()

	assert.Equal(t, "milestone", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "milestone", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := milestoneTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "milestone", cn2.TableName_())
		require.Implements(t, (*query.Linker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.Linker).Parent().NodeType_())
	}
}

func TestSerializeReferencesMilestoneTable(t *testing.T) {

	{
		n := Milestone().Project()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "milestone", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReferenceNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

}

func TestSerializeReverseReferencesMilestoneTable(t *testing.T) {
}

func TestSerializeAssociationsMilestoneTable(t *testing.T) {
}
