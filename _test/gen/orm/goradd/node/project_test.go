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
}

func TestSerializeReverseReferencesProjectTable(t *testing.T) {
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

		assert.True(t, query.NodesMatch(Project().Children().Id(), n2.(ProjectNode).Id()))
		assert.True(t, query.NodesMatch(Project().Children().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().Children().StatusEnum(), n2.(ProjectNode).StatusEnum()))
		assert.True(t, query.NodesMatch(Project().Children().ManagerId(), n2.(ProjectNode).ManagerId()))
		assert.True(t, query.NodesMatch(Project().Children().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().Children().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().Children().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().Children().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().Children().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().Children().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().Children().ParentProjectId(), n2.(ProjectNode).ParentProjectId()))
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

		assert.True(t, query.NodesMatch(Project().Parents().Id(), n2.(ProjectNode).Id()))
		assert.True(t, query.NodesMatch(Project().Parents().Num(), n2.(ProjectNode).Num()))
		assert.True(t, query.NodesMatch(Project().Parents().StatusEnum(), n2.(ProjectNode).StatusEnum()))
		assert.True(t, query.NodesMatch(Project().Parents().ManagerId(), n2.(ProjectNode).ManagerId()))
		assert.True(t, query.NodesMatch(Project().Parents().Name(), n2.(ProjectNode).Name()))
		assert.True(t, query.NodesMatch(Project().Parents().Description(), n2.(ProjectNode).Description()))
		assert.True(t, query.NodesMatch(Project().Parents().StartDate(), n2.(ProjectNode).StartDate()))
		assert.True(t, query.NodesMatch(Project().Parents().EndDate(), n2.(ProjectNode).EndDate()))
		assert.True(t, query.NodesMatch(Project().Parents().Budget(), n2.(ProjectNode).Budget()))
		assert.True(t, query.NodesMatch(Project().Parents().Spent(), n2.(ProjectNode).Spent()))
		assert.True(t, query.NodesMatch(Project().Parents().ParentProjectId(), n2.(ProjectNode).ParentProjectId()))
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

		assert.True(t, query.NodesMatch(Project().TeamMembers().Id(), n2.(PersonNode).Id()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().FirstName(), n2.(PersonNode).FirstName()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().LastName(), n2.(PersonNode).LastName()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().PersonTypeEnum(), n2.(PersonNode).PersonTypeEnum()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Created(), n2.(PersonNode).Created()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Modified(), n2.(PersonNode).Modified()))
		assert.True(t, query.NodesMatch(Project().TeamMembers().Projects(), n2.(PersonNode).Projects()))

	}

}
