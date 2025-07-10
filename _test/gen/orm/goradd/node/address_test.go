package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
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
		assert.True(t, query.NodesMatch(Address().Person().PersonTypeEnum(), n2.(PersonNode).PersonTypeEnum()))
		assert.True(t, query.NodesMatch(Address().Person().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Address().Person().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Address().Person().ManagerProject(), n2.(PersonNode).ManagerProject()))
		assert.True(t, query.NodesMatch(Address().Person().PersonAddress(), n2.(PersonNode).PersonAddress()))
		assert.True(t, query.NodesMatch(Address().Person().PersonEmployeeInfo(), n2.(PersonNode).PersonEmployeeInfo()))
		assert.True(t, query.NodesMatch(Address().Person().PersonLogin(), n2.(PersonNode).PersonLogin()))
		assert.True(t, query.NodesMatch(Address().Person().Projects(), n2.(PersonNode).Projects()))

	}

}

func TestSerializeReverseReferencesAddressTable(t *testing.T) {
}

func TestSerializeAssociationsAddressTable(t *testing.T) {
}
