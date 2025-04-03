package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableProjectTable(t *testing.T) {
	var n query.Node = Project()

	assert.Equal(t, "project", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "project", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := projectTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "project", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesProjectTable(t *testing.T) {

	{
		n := Project().Manager()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().Manager().ID(), n2.(PersonNode).ID()))
		assert.True(t, query.NodesMatch(Project().Manager().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(Project().Manager().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(Project().Manager().Types(), n2.(PersonNode).Types()))
		assert.True(t, query.NodesMatch(Project().Manager().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Project().Manager().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Project().Manager().Addresses(), n2.(PersonNode).Addresses()))
		assert.True(t, query.NodesMatch(Project().Manager().EmployeeInfo(), n2.(PersonNode).EmployeeInfo()))
		assert.True(t, query.NodesMatch(Project().Manager().Login(), n2.(PersonNode).Login()))
		assert.True(t, query.NodesMatch(Project().Manager().ManagerProjects(), n2.(PersonNode).ManagerProjects()))
		assert.True(t, query.NodesMatch(Project().Manager().Projects(), n2.(PersonNode).Projects()))

	}

	{
		n := Project().ParentProject()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().ParentProject().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Project().ParentProject().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().ParentProject().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().ParentProject().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().ParentProject().ParentProjectID(), n2.(ProjectNode).ParentProjectID()))
		assert.True(t, query.NodesMatch(Project().ParentProject().ParentProject(), n2.(ProjectNode).ParentProject()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Milestones(), n2.(ProjectNode).Milestones()))
		assert.True(t, query.NodesMatch(Project().ParentProject().ParentProjectProjects(), n2.(ProjectNode).ParentProjectProjects()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Project().ParentProject().Parents(), n2.(ProjectNode).Parents()))
		assert.True(t, query.NodesMatch(Project().ParentProject().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

}

func TestSerializeReverseReferencesProjectTable(t *testing.T) {

	{
		n := Project().Milestones()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().Milestones().ID(), n2.(MilestoneNode).ID()))
		assert.True(t, query.NodesMatch(Project().Milestones().ProjectID(), n2.(MilestoneNode).ProjectID()))
		assert.True(t, query.NodesMatch(Project().Milestones().Project(), n2.(MilestoneNode).Project()))
		assert.True(t, query.NodesMatch(Project().Milestones().Name(), n2.(MilestoneNode).Name()))

	}

	{
		n := Project().ParentProjectProjects()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().ParentProjectID(), n2.(ProjectNode).ParentProjectID()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().ParentProject(), n2.(ProjectNode).ParentProject()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Milestones(), n2.(ProjectNode).Milestones()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().ParentProjectProjects(), n2.(ProjectNode).ParentProjectProjects()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().Parents(), n2.(ProjectNode).Parents()))
		assert.True(t, query.NodesMatch(Project().ParentProjectProjects().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

}

func TestSerializeAssociationsProjectTable(t *testing.T) {

	{
		n := Project().Children()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().Children().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Project().Children().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().Children().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Project().Children().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Project().Children().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Project().Children().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().Children().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().Children().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().Children().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().Children().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().Children().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().Children().ParentProjectID(), n2.(ProjectNode).ParentProjectID()))
		assert.True(t, query.NodesMatch(Project().Children().ParentProject(), n2.(ProjectNode).ParentProject()))
		assert.True(t, query.NodesMatch(Project().Children().Milestones(), n2.(ProjectNode).Milestones()))
		assert.True(t, query.NodesMatch(Project().Children().ParentProjectProjects(), n2.(ProjectNode).ParentProjectProjects()))
		assert.True(t, query.NodesMatch(Project().Children().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Project().Children().Parents(), n2.(ProjectNode).Parents()))
		assert.True(t, query.NodesMatch(Project().Children().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

	{
		n := Project().Parents()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().Parents().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Project().Parents().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().Parents().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Project().Parents().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Project().Parents().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Project().Parents().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().Parents().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().Parents().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().Parents().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().Parents().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().Parents().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().Parents().ParentProjectID(), n2.(ProjectNode).ParentProjectID()))
		assert.True(t, query.NodesMatch(Project().Parents().ParentProject(), n2.(ProjectNode).ParentProject()))
		assert.True(t, query.NodesMatch(Project().Parents().Milestones(), n2.(ProjectNode).Milestones()))
		assert.True(t, query.NodesMatch(Project().Parents().ParentProjectProjects(), n2.(ProjectNode).ParentProjectProjects()))
		assert.True(t, query.NodesMatch(Project().Parents().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Project().Parents().Parents(), n2.(ProjectNode).Parents()))
		assert.True(t, query.NodesMatch(Project().Parents().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

	{
		n := Project().TeamMembers()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "project", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(Project().TeamMembers().ID(), n2.(PersonNode).ID()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Types(), n2.(PersonNode).Types()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Addresses(), n2.(PersonNode).Addresses()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().EmployeeInfo(), n2.(PersonNode).EmployeeInfo()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Login(), n2.(PersonNode).Login()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().ManagerProjects(), n2.(PersonNode).ManagerProjects()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Projects(), n2.(PersonNode).Projects()))

	}

}
