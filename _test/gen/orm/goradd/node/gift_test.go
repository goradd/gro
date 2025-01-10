package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSerializeTableGiftTable(t *testing.T) {
	var n query.NodeI = Gift()

	assert.Equal(t, "gift", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "gift", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd", n2.DatabaseKey_())

	nodes := giftTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "gift", cn2.TableName_())
		require.Implements(t, (*query.NodeLinker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.NodeLinker).Parent().NodeType_())
	}
}

func TestSerializeReferencesGiftTable(t *testing.T) {
}

func TestSerializeReverseReferencesGiftTable(t *testing.T) {
}

func TestSerializeAssociationsGiftTable(t *testing.T) {
}
