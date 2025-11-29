package query

import (
	"bytes"
	"context"
	"encoding/gob"
	"testing"

	"github.com/goradd/gro/ci/tests/gen/goradd/node"
	"github.com/goradd/gro/ci/tests/gen/goradd_unit"
	unit_node "github.com/goradd/gro/ci/tests/gen/goradd_unit/node"
	"github.com/goradd/gro/query"
	"github.com/goradd/gro/query/op"
	"github.com/stretchr/testify/assert"
)

func TestNodeSerialize(t *testing.T) {
	var n query.Node = node.Person().FirstName()

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&n)
	assert.NoError(t, err)

	var n2 query.Node
	dec := gob.NewDecoder(&buf)
	err = dec.Decode(&n2)
	assert.NoError(t, err)
}

func TestNodeRejectTableNodes(t *testing.T) {
	ctx := context.Background()

	// Make sure we panic when a table node is being used as a primary key for a composite key table
	assert.Panics(t, func() {
		_, _ = goradd_unit.QueryTwoKeys(ctx).
			Where(op.Equal(unit_node.TwoKey(), 2)).
			Load()
	})
}
