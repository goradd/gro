package node

import (
	"testing"

	"github.com/goradd/orm/pkg/query"
	"github.com/stretchr/testify/assert"
)

func TestSerializeForwardRestrictTable(t *testing.T) {
	var n query.NodeI = ForwardRestrict()

	assert.Equal(t, "forward_restrict", n.TableName_())
	assert.Equal(t, query.TableNodeType, n.NodeType_())
	assert.Equal(t, "goradd_unit", n.DatabaseKey_())

	n2 := serNode(t, n)

	assert.Equal(t, "forward_restrict", n2.TableName_())
	assert.Equal(t, query.TableNodeType, n2.NodeType_())
	assert.Equal(t, "goradd_unit", n2.DatabaseKey_())

	nodes := forwardRestrictTable{}.ColumnNodes_()
	for _, cn := range nodes {
		cn2 := serNode(t, cn)
		assert.Equal(t, "forward_restrict", cn2.TableName_())
		assert.Implements(t, (*query.NodeLinker)(nil), cn2)
		assert.Equal(t, query.TableNodeType, cn2.(query.NodeLinker).Parent().NodeType_())
	}
}
