package query

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestColumnNodeInterfaces(t *testing.T) {
	n := &ColumnNode{QueryName: "dbName", Identifier: "goName", ReceiverType: ColTypeString, IsPrimaryKey: true}

	assert.Implements(t, (*ColumnNodeI)(nil), n)
}
