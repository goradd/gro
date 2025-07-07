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

		assert.True(t, query.NodesMatch(Person().Projects().Id(), n2.(ProjectNode).Id()))
		assert.True(t, query.NodesMatch(Person().Projects().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Person().Projects().StatusEnum(), n2.(ProjectNode).StatusEnum()))
		assert.True(t, query.NodesMatch(Person().Projects().ManagerId(), n2.(ProjectNode).ManagerId()))
		assert.True(t, query.NodesMatch(Person().Projects().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Person().Projects().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Person().Projects().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Person().Projects().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Person().Projects().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Person().Projects().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Person().Projects().ParentProjectId(), n2.(ProjectNode).ParentProjectId()))
		assert.True(t, query.NodesMatch(Person().Projects().Children(), n2.(ProjectNode).Children()))
		assert.True(t, query.NodesMatch(Person().Projects().Parents(), n2.(ProjectNode).Parents()))
		assert.True(t, query.NodesMatch(Person().Projects().TeamMembers(), n2.(ProjectNode).TeamMembers()))

	}

}
