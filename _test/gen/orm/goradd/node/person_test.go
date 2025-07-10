package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTablePersonTable(t *testing.T) {
	var n query.Node = Person()

	assert.Equal(t, "person", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "person", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := personTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "person", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesPersonTable(t *testing.T) {
}

func TestSerializeReverseReferencesPersonTable(t *testing.T) {

	{
		n := Person().ManagerProject()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Person().ManagerProject().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().StatusEnum(), n2.(ProjectNode).StatusEnum()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().ParentID(), n2.(ProjectNode).ParentID()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Parent(), n2.(ProjectNode).Parent()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().Child(), n2.(ProjectNode).Child()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().ProjectMilestone(), n2.(ProjectNode).ProjectMilestone()))
		assert.True(t, query.NodesMatch(Person().ManagerProject().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

	{
		n := Person().PersonAddress()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Person().PersonAddress().ID(), n2.(AddressNode).ID()))
		assert.True(t, query.NodesMatch(Person().PersonAddress().Street(), n2.(AddressNode).Street()))
		assert.True(t, query.NodesMatch(Person().PersonAddress().City(), n2.(AddressNode).City()))
		assert.True(t, query.NodesMatch(Person().PersonAddress().PersonID(), n2.(AddressNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().PersonAddress().Person(), n2.(AddressNode).Person()))

	}

	{
		n := Person().PersonEmployeeInfo()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Person().PersonEmployeeInfo().ID(), n2.(EmployeeInfoNode).ID()))
		assert.True(t, query.NodesMatch(Person().PersonEmployeeInfo().EmployeeNumber(), n2.(EmployeeInfoNode).EmployeeNumber()))
		assert.True(t, query.NodesMatch(Person().PersonEmployeeInfo().PersonID(), n2.(EmployeeInfoNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().PersonEmployeeInfo().Person(), n2.(EmployeeInfoNode).Person()))

	}

	{
		n := Person().PersonLogin()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Person().PersonLogin().ID(), n2.(LoginNode).ID()))
		assert.True(t, query.NodesMatch(Person().PersonLogin().Username(), n2.(LoginNode).Username()))
		assert.True(t, query.NodesMatch(Person().PersonLogin().Password(), n2.(LoginNode).Password()))
		assert.True(t, query.NodesMatch(Person().PersonLogin().IsEnabled(), n2.(LoginNode).IsEnabled()))
		assert.True(t, query.NodesMatch(Person().PersonLogin().PersonID(), n2.(LoginNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().PersonLogin().Person(), n2.(LoginNode).Person()))

	}

}

func TestSerializeAssociationsPersonTable(t *testing.T) {

	{
		n := Person().Projects()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = query.NodeParent(cn2)
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}

		assert.True(t, query.NodesMatch(Person().Projects().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Person().Projects().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Person().Projects().StatusEnum(), n2.(ProjectNode).StatusEnum()))
		assert.True(t, query.NodesMatch(Person().Projects().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Person().Projects().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Person().Projects().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Person().Projects().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Person().Projects().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Person().Projects().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Person().Projects().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Person().Projects().Manager(), n2.(ProjectNode).Manager()))
		assert.True(t, query.NodesMatch(Person().Projects().ParentID(), n2.(ProjectNode).ParentID()))
		assert.True(t, query.NodesMatch(Person().Projects().Parent(), n2.(ProjectNode).Parent()))
		assert.True(t, query.NodesMatch(Person().Projects().Child(), n2.(ProjectNode).Child()))
		assert.True(t, query.NodesMatch(Person().Projects().ProjectMilestone(), n2.(ProjectNode).ProjectMilestone()))
		assert.True(t, query.NodesMatch(Person().Projects().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

}
