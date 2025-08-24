package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableAddressTable(t *testing.T) {
	var n query.Node = Address()

	assert.Equal(t, "address", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "address", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := addressTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "address", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesAddressTable(t *testing.T) {
	{
		n := Address().Person()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "address", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Address().Person().ID(), n2.(PersonNode).ID()))
		assert.True(t, query.NodesMatch(Address().Person().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(Address().Person().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(Address().Person().PersonType(), n2.(PersonNode).PersonType()))
		assert.True(t, query.NodesMatch(Address().Person().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Address().Person().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Address().Person().ManagerProjects(), n2.(PersonNode).ManagerProjects()))
		assert.True(t, query.NodesMatch(Address().Person().Addresses(), n2.(PersonNode).Addresses()))
		assert.True(t, query.NodesMatch(Address().Person().EmployeeInfo(), n2.(PersonNode).EmployeeInfo()))
		assert.True(t, query.NodesMatch(Address().Person().Login(), n2.(PersonNode).Login()))
		assert.True(t, query.NodesMatch(Address().Person().Projects(), n2.(PersonNode).Projects()))

	}

}

func TestSerializeReverseReferencesAddressTable(t *testing.T) {
}

func TestSerializeAssociationsAddressTable(t *testing.T) {
}
