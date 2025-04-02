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
		n := Person().Addresses()
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

		assert.True(t, query.NodesMatch(Person().Addresses().ID(), n2.(AddressNode).ID()))
		assert.True(t, query.NodesMatch(Person().Addresses().PersonID(), n2.(AddressNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().Addresses().Street(), n2.(AddressNode).Street()))
		assert.True(t, query.NodesMatch(Person().Addresses().City(), n2.(AddressNode).City()))
	}

	{
		n := Person().EmployeeInfo()
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

		assert.True(t, query.NodesMatch(Person().EmployeeInfo().ID(), n2.(EmployeeInfoNode).ID()))
		assert.True(t, query.NodesMatch(Person().EmployeeInfo().PersonID(), n2.(EmployeeInfoNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().EmployeeInfo().EmployeeNumber(), n2.(EmployeeInfoNode).EmployeeNumber()))
	}

	{
		n := Person().Login()
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

		assert.True(t, query.NodesMatch(Person().Login().ID(), n2.(LoginNode).ID()))
		assert.True(t, query.NodesMatch(Person().Login().PersonID(), n2.(LoginNode).PersonID()))
		assert.True(t, query.NodesMatch(Person().Login().Username(), n2.(LoginNode).Username()))
		assert.True(t, query.NodesMatch(Person().Login().Password(), n2.(LoginNode).Password()))
		assert.True(t, query.NodesMatch(Person().Login().IsEnabled(), n2.(LoginNode).IsEnabled()))
	}

	{
		n := Person().ManagerProjects()
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

		assert.True(t, query.NodesMatch(Person().ManagerProjects().ID(), n2.(ProjectNode).ID()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Status(), n2.(ProjectNode).Status()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().ManagerID(), n2.(ProjectNode).ManagerID()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Person().ManagerProjects().ParentProjectID(), n2.(ProjectNode).ParentProjectID()))
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
	}

}
