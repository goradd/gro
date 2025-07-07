package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesMilestoneTable(t *testing.T) {
}

func TestSerializeReverseReferencesMilestoneTable(t *testing.T) {
}

func TestSerializeAssociationsMilestoneTable(t *testing.T) {
}
