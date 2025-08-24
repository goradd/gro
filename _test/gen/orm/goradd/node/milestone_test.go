package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
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
	{
		n := Milestone().Project()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "milestone", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Milestone().Project().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Milestone().Project().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Milestone().Project().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Milestone().Project().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Milestone().Project().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Milestone().Project().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Milestone().Project().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Milestone().Project().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Milestone().Project().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Milestone().Project().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Milestone().Project().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Milestone().Project().ParentID(), n2.(ProjectNode).ParentID()))
		assert.True(t, query.NodesMatch(Milestone().Project().Parent(), n2.(ProjectNode).Parent()))
		assert.True(t, query.NodesMatch(Milestone().Project().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Milestone().Project().Milestones(), n2.(ProjectNode).Milestones()))
		assert.True(t, query.NodesMatch(Milestone().Project().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

}

func TestSerializeReverseReferencesMilestoneTable(t *testing.T) {
}

func TestSerializeAssociationsMilestoneTable(t *testing.T) {
}
