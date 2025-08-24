package node

import (
	"testing"

	"github.com/goradd/gro/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeTableTimeoutTestTable(t *testing.T) {
	var n query.Node = TimeoutTest()

	assert.Equal(t, "timeout_test", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "timeout_test", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := timeoutTestTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "timeout_test", cn2.TableName_())
		assert.Equal(t, query.TableNodeType, query.NodeParent(cn2).NodeType_())
	}
}

func TestSerializeReferencesTimeoutTestTable(t *testing.T) {
}

func TestSerializeReverseReferencesTimeoutTestTable(t *testing.T) {
}

func TestSerializeAssociationsTimeoutTestTable(t *testing.T) {
}
