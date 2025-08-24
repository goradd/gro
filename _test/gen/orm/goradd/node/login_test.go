package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableLoginTable(t *testing.T) {
	var n query.Node = Login()

	assert.Equal(t, "login", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "login", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := loginTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "login", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesLoginTable(t *testing.T) {
	{
		n := Login().Person()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "login", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReferenceNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(Login().Person().ID(), n2.(PersonNode).ID()))
		assert.True(t, query.NodesMatch(Login().Person().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(Login().Person().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(Login().Person().PersonType(), n2.(PersonNode).PersonType()))
		assert.True(t, query.NodesMatch(Login().Person().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Login().Person().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Login().Person().ManagerProjects(), n2.(PersonNode).ManagerProjects()))
		assert.True(t, query.NodesMatch(Login().Person().Addresses(), n2.(PersonNode).Addresses()))
		assert.True(t, query.NodesMatch(Login().Person().EmployeeInfo(), n2.(PersonNode).EmployeeInfo()))
		assert.True(t, query.NodesMatch(Login().Person().Login(), n2.(PersonNode).Login()))
		assert.True(t, query.NodesMatch(Login().Person().Projects(), n2.(PersonNode).Projects()))

	}

}

func TestSerializeReverseReferencesLoginTable(t *testing.T) {
}

func TestSerializeAssociationsLoginTable(t *testing.T) {
}
