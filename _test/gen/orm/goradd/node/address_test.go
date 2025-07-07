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
}

func TestSerializeReverseReferencesAddressTable(t *testing.T) {
}

func TestSerializeAssociationsAddressTable(t *testing.T) {
}
