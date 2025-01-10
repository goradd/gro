package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeTableEmployeeInfoTable(t *testing.T) {
	var n query.NodeI = EmployeeInfo()

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
		require.Implements(t, (*query.NodeLinker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.NodeLinker).Parent().NodeType_())
	}
}

func TestSerializeReferencesEmployeeInfoTable(t *testing.T) {

	{
		n := EmployeeInfo().Person()
		n2 := serNode(t, n)
		parentNode := n2.(query.NodeLinker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "employee_info", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.NodeLinker)(nil), cn2)
			assert.Equal(t, query.ReferenceNodeType, cn2.(query.NodeLinker).Parent().NodeType_())
		}
	}

}

func TestSerializeReverseReferencesEmployeeInfoTable(t *testing.T) {
}

func TestSerializeAssociationsEmployeeInfoTable(t *testing.T) {
}
