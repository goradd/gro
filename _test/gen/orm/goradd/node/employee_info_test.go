package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableEmployeeInfoTable(t *testing.T) {
	var n query.Node = EmployeeInfo()

	assert.Equal(t, "employee_info", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "employee_info", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := employeeInfoTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "employee_info", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesEmployeeInfoTable(t *testing.T) {

	{
		n := EmployeeInfo().Person()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "employee_info", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(EmployeeInfo().Person().ID(), n2.(PersonNode).ID()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Types(), n2.(PersonNode).Types()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Addresses(), n2.(PersonNode).Addresses()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().EmployeeInfo(), n2.(PersonNode).EmployeeInfo()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Login(), n2.(PersonNode).Login()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().ManagerProjects(), n2.(PersonNode).ManagerProjects()))
		assert.True(t, query.NodesMatch(EmployeeInfo().Person().Projects(), n2.(PersonNode).Projects()))

	}

}

func TestSerializeReverseReferencesEmployeeInfoTable(t *testing.T) {
}

func TestSerializeAssociationsEmployeeInfoTable(t *testing.T) {
}
