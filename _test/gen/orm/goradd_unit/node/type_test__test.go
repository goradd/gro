package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeTableTypeTestTable(t *testing.T) {
	var n query.NodeI = TypeTest()

	assert.Equal(t, "type_test", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "type_test", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := typeTestTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "type_test", cn2.TableName_())
		require.Implements(t, (*query.NodeLinker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.NodeLinker).Parent().NodeType_())
	}
}

func TestSerializeReferencesTypeTestTable(t *testing.T) {
}

func TestSerializeReverseReferencesTypeTestTable(t *testing.T) {
}

func TestSerializeAssociationsTypeTestTable(t *testing.T) {
}
