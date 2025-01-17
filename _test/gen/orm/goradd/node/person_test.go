package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		require.Implements(t, (*query.Linker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.Linker).Parent().NodeType_())
	}
}

func TestSerializeReferencesPersonTable(t *testing.T) {
}

func TestSerializeReverseReferencesPersonTable(t *testing.T) {

	{
		n := Person().Addresses()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReverseNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

	{
		n := Person().EmployeeInfo()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReverseNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

	{
		n := Person().Login()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReverseNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

	{
		n := Person().ManagerProjects()
		n2 := serNode(t, n)
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			require.Implements(t, (*query.Linker)(nil), cn2)
			assert.Equal(t, query.ReverseNodeType, cn2.(query.Linker).Parent().NodeType_())
		}
	}

}

func TestSerializeAssociationsPersonTable(t *testing.T) {

	{
		n := Person().PersonTypes()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyEnumNodeType, n2.NodeType_())
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = cn2.(query.Linker).Parent()
			assert.Equal(t, query.ManyEnumNodeType, parentNode.NodeType_())
		}
	}

	{
		n := Person().Projects()
		n2 := serNode(t, n)
		assert.Equal(t, query.ManyManyNodeType, n2.NodeType_())
		parentNode := n2.(query.Linker).Parent()
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "person", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			//        assert.Equal(t, query.ColumnNodeType, cn2.NodeType_())
			parentNode = cn2.(query.Linker).Parent()
			assert.Equal(t, query.ManyManyNodeType, parentNode.NodeType_())
		}
	}

}
