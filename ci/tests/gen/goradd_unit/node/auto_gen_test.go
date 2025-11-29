package node

import (
	"testing"

	"github.com/goradd/gro/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableAutoGenTable(t *testing.T) {
	var n query.Node = AutoGen()

	assert.Equal(t, "auto_gen", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "auto_gen", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := autoGenTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "auto_gen", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesAutoGenTable(t *testing.T) {
}

func TestSerializeReverseReferencesAutoGenTable(t *testing.T) {
}

func TestSerializeAssociationsAutoGenTable(t *testing.T) {
}
