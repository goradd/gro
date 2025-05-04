package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableRootNTable(t *testing.T) {
	var n query.Node = RootN()

	assert.Equal(t, "root_n", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "root_n", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := rootNTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "root_n", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesRootNTable(t *testing.T) {
}

func TestSerializeReverseReferencesRootNTable(t *testing.T) {

	{
		n := RootN().LeafNs()
		n2 := serNode(t, n)
		parentNode := query.NodeParent(n2)
		assert.Equal(t, query.TableNodeType, parentNode.NodeType_())
		assert.Equal(t, "root_n", parentNode.TableName_())

		nodes := n.(query.TableNodeI).ColumnNodes_()
		for _, cn := range nodes {
			cn2 := serNode(t, cn)
			assert.Equal(t, n.TableName_(), cn2.TableName_())
			assert.Equal(t, query.ReverseNodeType, query.NodeParent(cn2).NodeType_())
		}

		assert.True(t, query.NodesMatch(RootN().LeafNs().ID(), n2.(LeafNNode).ID()))
		assert.True(t, query.NodesMatch(RootN().LeafNs().Name(), n2.(LeafNNode).Name()))
		assert.True(t, query.NodesMatch(RootN().LeafNs().RootNID(), n2.(LeafNNode).RootNID()))
		assert.True(t, query.NodesMatch(RootN().LeafNs().RootN(), n2.(LeafNNode).RootN()))

	}

}

func TestSerializeAssociationsRootNTable(t *testing.T) {
}
